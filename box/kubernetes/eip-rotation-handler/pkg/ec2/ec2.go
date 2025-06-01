package ec2

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/sirupsen/logrus"
)

const (
	requiredEC2Permissions = "ec2:AllocateAddress, ec2:AssociateAddress, ec2:DescribeAddresses, ec2:ReleaseAddress"
)

// EC2Client AWS EC2 service client wrapper
type EC2Client struct {
	client *ec2.Client
	region string
	logger *logrus.Logger
}

// NewEC2Client creates new EC2 client
func NewEC2Client(ctx context.Context, region string, logger *logrus.Logger) (*EC2Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &EC2Client{
		client: ec2.NewFromConfig(cfg),
		region: region,
		logger: logger,
	}, nil
}

// AllocateAddress allocates new EIP
func (c *EC2Client) AllocateAddress(ctx context.Context) (*string, error) {
	input := &ec2.AllocateAddressInput{
		Domain: types.DomainIdentifierVpc,
	}

	result, err := c.client.AllocateAddress(ctx, input)
	if err != nil {
		return nil, c.handleEC2Error("ec2:AllocateAddress", "failed to allocate EIP", err)
	}

	c.logger.WithFields(logrus.Fields{
		"public_ip":     aws.ToString(result.PublicIp),
		"allocation_id": aws.ToString(result.AllocationId),
	}).Info("Successfully allocated new EIP")

	return result.PublicIp, nil
}

// AssociateAddress connects EIP to instance
func (c *EC2Client) AssociateAddress(ctx context.Context, instanceID, publicIP string) error {
	input := &ec2.AssociateAddressInput{
		InstanceId: aws.String(instanceID),
		PublicIp:   aws.String(publicIP),
	}

	_, err := c.client.AssociateAddress(ctx, input)
	if err != nil {
		return c.handleEC2Error("ec2:AssociateAddress",
			fmt.Sprintf("failed to associate EIP %s to instance %s", publicIP, instanceID), err)
	}

	c.logger.WithFields(logrus.Fields{
		"instance_id": instanceID,
		"public_ip":   publicIP,
	}).Info("Successfully associated EIP to instance")

	return nil
}

// DescribeAddresses gets EIP info for given public IP
func (c *EC2Client) DescribeAddresses(ctx context.Context, publicIP string) (*string, error) {
	input := &ec2.DescribeAddressesInput{
		PublicIps: []string{publicIP},
	}

	result, err := c.client.DescribeAddresses(ctx, input)
	if err != nil {
		return nil, c.handleEC2Error("ec2:DescribeAddresses", "failed to describe addresses", err)
	}

	if len(result.Addresses) == 0 {
		return nil, nil // Not an EIP (auto-assigned IP)
	}

	return result.Addresses[0].AllocationId, nil
}

// ReleaseAddress releases EIP
func (c *EC2Client) ReleaseAddress(ctx context.Context, allocationID string) error {
	input := &ec2.ReleaseAddressInput{
		AllocationId: aws.String(allocationID),
	}

	_, err := c.client.ReleaseAddress(ctx, input)
	if err != nil {
		return c.handleEC2Error("ec2:ReleaseAddress",
			fmt.Sprintf("failed to release EIP with allocation ID %s", allocationID), err)
	}

	c.logger.WithFields(logrus.Fields{
		"allocation_id": allocationID,
	}).Info("Successfully released EIP")

	return nil
}

// handleEC2Error handles EC2 errors and logging
func (c *EC2Client) handleEC2Error(action, baseMessage string, err error) error {
	if c.isPermissionError(err) {
		c.logPermissionError(action, err)
		return fmt.Errorf("permission denied: %s: %w", baseMessage, err)
	}
	return fmt.Errorf("%s: %w", baseMessage, err)
}

// isPermissionError checks if error is AWS permission related
func (c *EC2Client) isPermissionError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	permissionKeywords := []string{
		"unauthorizedoperation",
		"accessdenied",
		"forbidden",
		"not authorized",
		"permission",
	}

	for _, keyword := range permissionKeywords {
		if strings.Contains(errStr, keyword) {
			return true
		}
	}
	return false
}

// logPermissionError logs permission errors with details
func (c *EC2Client) logPermissionError(action string, err error) {
	c.logger.WithFields(logrus.Fields{
		"error":                err.Error(),
		"action":               action,
		"region":               c.region,
		"required_permissions": requiredEC2Permissions,
		"solution":             "Add the required EC2 permissions to your EKS worker node IAM role",
	}).Error("PERMISSION_DENIED: IAM role lacks required EC2 permissions")
}

// GetInstanceID gets EC2 instance ID
func (c *EC2Client) GetInstanceID(ctx context.Context, metadataURL string) (string, error) {
	// Instance ID query logic using IMDS is in metadata package
	return "", fmt.Errorf("implement in metadata package")
}
