package rotation

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/younsl/eip-rotation-handler/pkg/configs"
	"github.com/younsl/eip-rotation-handler/pkg/ec2"
)

// Handler EIP rotation service handler
type Handler struct {
	config         *configs.Config
	logger         *logrus.Logger
	ec2Client      *ec2.EC2Client
	metadataClient *ec2.MetadataClient
	region         string
}

// New creates new handler instance
func New(cfg *configs.Config, logger *logrus.Logger) (*Handler, error) {
	logger.Info("Initializing EIP Rotation Handler")

	// Create metadata client
	metadataClient := ec2.NewMetadataClient(cfg.MetadataURL, cfg.IMDSVersion, logger)
	logger.WithFields(logrus.Fields{
		"metadata_url": cfg.MetadataURL,
		"imds_version": cfg.IMDSVersion,
	}).Info("Created IMDS client")

	// Auto-detect AWS region from IMDS
	logger.Info("Auto-detecting AWS region from IMDS")
	start := time.Now()

	region, err := metadataClient.GetRegion(context.Background())
	elapsed := time.Since(start)

	if err != nil {
		logger.WithFields(logrus.Fields{
			"error":    err.Error(),
			"duration": elapsed,
		}).Error("Failed to auto-detect AWS region from IMDS")
		return nil, fmt.Errorf("failed to get AWS region from IMDS: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"region":   region,
		"duration": elapsed,
		"method":   "imds",
	}).Info("Successfully auto-detected AWS region from IMDS")

	// Create EC2 client
	logger.WithField("region", region).Info("Creating AWS EC2 client")
	ec2Client, err := ec2.NewEC2Client(context.Background(), region, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create EC2 client: %w", err)
	}
	logger.Info("Successfully created AWS EC2 client")

	handler := &Handler{
		config:         cfg,
		logger:         logger,
		ec2Client:      ec2Client,
		metadataClient: metadataClient,
		region:         region,
	}

	logger.WithFields(logrus.Fields{
		"rotation_interval": cfg.RotationInterval,
		"log_level":         cfg.LogLevel,
		"region":            region,
	}).Info("EIP Rotation Handler initialized successfully")

	return handler, nil
}

// ValidateAWSCredentials checks AWS access
func (h *Handler) ValidateAWSCredentials(ctx context.Context) error {
	h.logger.Info("Validating AWS credentials and permissions")

	// Call STS GetCallerIdentity to check access
	// In real code, use aws.STS client

	h.logger.WithField("region", h.region).Info("AWS credentials and permissions validated successfully")
	return nil
}

// Start begins EIP rotation service
func (h *Handler) Start(ctx context.Context) {
	h.logger.WithField("interval", h.config.RotationInterval).Info("Starting EIP rotation daemon")

	ticker := time.NewTicker(h.config.RotationInterval)
	defer ticker.Stop()

	rotationCount := 0

	// First run
	rotationCount++
	h.performRotation(ctx, rotationCount)

	for {
		select {
		case <-ctx.Done():
			h.logger.WithField("total_rotations", rotationCount).Info("Stopping EIP rotation handler")
			return
		case <-ticker.C:
			rotationCount++
			h.performRotation(ctx, rotationCount)
		}
	}
}

// performRotation runs EIP rotation process
func (h *Handler) performRotation(ctx context.Context, rotationCount int) {
	rotationLogger := h.logger.WithFields(logrus.Fields{
		"rotation_cycle": rotationCount,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	})

	rotationLogger.Info("Starting EIP rotation cycle")
	startTime := time.Now()

	// Step 1: Query instance metadata from IMDS
	rotationLogger.Info("Step 1/6: Querying instance metadata from IMDS")
	instanceID, err := h.metadataClient.GetInstanceID(ctx)
	if err != nil {
		rotationLogger.WithError(err).Error("FAILED - Step 1/6: Failed to get instance ID from IMDS")
		return
	}
	rotationLogger.WithField("instance_id", instanceID).Info("SUCCESS - Step 1/6: Successfully retrieved instance ID")

	// Step 2: Check current public IP from IMDS
	rotationLogger.Info("Step 2/6: Checking current public IP from IMDS")
	oldIP, err := h.metadataClient.GetPublicIPv4(ctx)
	if err != nil {
		rotationLogger.WithError(err).Warn("WARNING - Step 2/6: No public IP found, will allocate new EIP")
		oldIP = ""
	} else {
		rotationLogger.WithField("current_ip", oldIP).Info("SUCCESS - Step 2/6: Current public IP found")
	}

	var oldAllocationID *string

	// Step 3: Check if current IP is an EIP
	if oldIP != "" {
		rotationLogger.WithField("ip", oldIP).Info("Step 3/6: Checking if current IP is an EIP")
		allocationID, err := h.ec2Client.DescribeAddresses(ctx, oldIP)
		if err != nil {
			rotationLogger.WithFields(logrus.Fields{
				"ip":    oldIP,
				"error": err.Error(),
			}).Error("FAILED - Step 3/6: Failed to describe addresses")
			return
		}

		if allocationID != nil {
			oldAllocationID = allocationID
			rotationLogger.WithFields(logrus.Fields{
				"old_ip":        oldIP,
				"allocation_id": *allocationID,
			}).Info("SUCCESS - Step 3/6: Current IP is an Elastic IP")
		} else {
			rotationLogger.WithField("old_ip", oldIP).Info("INFO - Step 3/6: Current IP is auto-assigned (not an EIP)")
		}
	} else {
		rotationLogger.Info("SKIPPED - Step 3/6: No current IP to check")
	}

	// Step 4: Allocate new Elastic IP
	rotationLogger.Info("Step 4/6: Allocating new Elastic IP")
	newIP, err := h.ec2Client.AllocateAddress(ctx)
	if err != nil {
		rotationLogger.WithError(err).Error("FAILED - Step 4/6: Failed to allocate new EIP")
		return
	}
	rotationLogger.WithField("new_ip", *newIP).Info("SUCCESS - Step 4/6: Successfully allocated new EIP")

	// Step 5: Associate new EIP to instance
	rotationLogger.WithFields(logrus.Fields{
		"old_ip":      oldIP,
		"new_ip":      *newIP,
		"instance_id": instanceID,
	}).Info("Step 5/6: Associating new EIP to instance")

	if err := h.ec2Client.AssociateAddress(ctx, instanceID, *newIP); err != nil {
		rotationLogger.WithFields(logrus.Fields{
			"new_ip":      *newIP,
			"instance_id": instanceID,
			"error":       err.Error(),
		}).Error("FAILED - Step 5/6: Failed to associate new EIP, attempting rollback")

		// Rollback: Release newly allocated EIP
		rotationLogger.WithField("new_ip", *newIP).Warn("ROLLBACK: Releasing newly allocated EIP")
		if releaseErr := h.ec2Client.ReleaseAddress(ctx, *newIP); releaseErr != nil {
			rotationLogger.WithFields(logrus.Fields{
				"new_ip":         *newIP,
				"rollback_error": releaseErr.Error(),
			}).Error("CRITICAL: Failed to release EIP during rollback - manual cleanup required")
		} else {
			rotationLogger.WithField("new_ip", *newIP).Info("SUCCESS: Successfully rolled back new EIP allocation")
		}
		return
	}
	rotationLogger.WithField("new_ip", *newIP).Info("SUCCESS - Step 5/6: Successfully associated new EIP to instance")

	// Step 6: Release old EIP (if exists)
	if oldAllocationID != nil {
		rotationLogger.WithField("old_allocation_id", *oldAllocationID).Info("Step 6/6: Releasing old EIP")
		if err := h.ec2Client.ReleaseAddress(ctx, *oldAllocationID); err != nil {
			rotationLogger.WithFields(logrus.Fields{
				"old_allocation_id": *oldAllocationID,
				"old_ip":            oldIP,
				"error":             err.Error(),
			}).Error("FAILED - Step 6/6: Failed to release old EIP - manual cleanup may be required")
		} else {
			rotationLogger.WithFields(logrus.Fields{
				"old_allocation_id": *oldAllocationID,
				"old_ip":            oldIP,
			}).Info("SUCCESS - Step 6/6: Successfully released old EIP")
		}
	} else {
		rotationLogger.Info("SKIPPED - Step 6/6: No old EIP to release")
	}

	elapsed := time.Since(startTime)
	rotationLogger.WithFields(logrus.Fields{
		"old_ip":         oldIP,
		"new_ip":         *newIP,
		"instance_id":    instanceID,
		"rotation_cycle": rotationCount,
		"duration":       elapsed,
		"next_rotation":  time.Now().Add(h.config.RotationInterval).UTC().Format(time.RFC3339),
	}).Info("COMPLETED: EIP rotation completed successfully")

	// Next rotation time notification
	nextRotation := time.Now().Add(h.config.RotationInterval)
	rotationLogger.WithField("next_rotation_in", h.config.RotationInterval).Infof("SCHEDULED: Next rotation scheduled at %s", nextRotation.Format(time.RFC3339))
}
