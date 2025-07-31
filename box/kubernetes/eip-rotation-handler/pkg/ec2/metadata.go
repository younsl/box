package ec2

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/younsl/eip-rotation-handler/pkg/configs"
)

const (
	// IMDS endpoints
	defaultIMDSURL      = "http://169.254.169.254"
	metadataBasePath    = "/latest/meta-data"
	tokenEndpoint       = "/latest/api/token"
	defaultTimeoutSec   = 10
	defaultTokenTTL     = "21600" // 6 hours
	tokenBufferDuration = 10 * time.Minute

	// HTTP constants
	httpMethodGET  = "GET"
	httpMethodPUT  = "PUT"
	tokenHeader    = "X-aws-ec2-metadata-token"
	tokenTTLHeader = "X-aws-ec2-metadata-token-ttl-seconds"

	// Metadata endpoints
	EndpointInstanceID = "instance-id"
	EndpointRegion     = "placement/region"
	EndpointPublicIPv4 = "public-ipv4"
)

type IMDSError struct {
	Endpoint   string
	StatusCode int
	Message    string
	Err        error
}

func (e *IMDSError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("IMDS error for %s: %s (caused by: %v)", e.Endpoint, e.Message, e.Err)
	}
	return fmt.Sprintf("IMDS error for %s: %s", e.Endpoint, e.Message)
}

func (e *IMDSError) Unwrap() error {
	return e.Err
}

type MetadataService interface {
	GetInstanceID(ctx context.Context) (string, error)
	GetRegion(ctx context.Context) (string, error)
	GetPublicIPv4(ctx context.Context) (string, error)
}

type MetadataClient struct {
	baseURL     string
	client      *http.Client
	logger      *logrus.Logger
	imdsVersion string
	token       string
	tokenExpiry time.Time
}

func NewMetadataClient(baseURL string, imdsVersion string, logger *logrus.Logger) *MetadataClient {
	if baseURL == "" {
		baseURL = defaultIMDSURL
	}

	return &MetadataClient{
		baseURL:     baseURL,
		imdsVersion: imdsVersion,
		client: &http.Client{
			Timeout: defaultTimeoutSec * time.Second,
		},
		logger: logger,
	}
}

func (m *MetadataClient) GetInstanceID(ctx context.Context) (string, error) {
	return m.getMetadata(ctx, EndpointInstanceID)
}

func (m *MetadataClient) GetRegion(ctx context.Context) (string, error) {
	return m.getMetadata(ctx, EndpointRegion)
}

func (m *MetadataClient) GetPublicIPv4(ctx context.Context) (string, error) {
	return m.getMetadata(ctx, EndpointPublicIPv4)
}

func (m *MetadataClient) getMetadata(ctx context.Context, endpoint string) (string, error) {
	switch m.imdsVersion {
	case configs.IMDSVersionV1:
		return m.fetchMetadata(ctx, endpoint, nil)
	case configs.IMDSVersionV2:
		return m.fetchWithToken(ctx, endpoint)
	case configs.IMDSVersionAuto:
		fallthrough
	default:
		return m.fetchWithFallback(ctx, endpoint)
	}
}

func (m *MetadataClient) fetchWithToken(ctx context.Context, endpoint string) (string, error) {
	token, err := m.getValidToken(ctx)
	if err != nil {
		return "", m.wrapError(endpoint, "failed to get IMDS v2 token", err)
	}

	return m.fetchMetadata(ctx, endpoint, map[string]string{
		tokenHeader: token,
	})
}

func (m *MetadataClient) fetchWithFallback(ctx context.Context, endpoint string) (string, error) {
	// Try IMDS v2 first
	if result, err := m.fetchWithToken(ctx, endpoint); err == nil {
		m.logger.WithField("endpoint", endpoint).Debug("Retrieved metadata using IMDS v2")
		return result, nil
	} else {
		m.logger.WithFields(logrus.Fields{
			"endpoint": endpoint,
			"v2_error": err.Error(),
		}).Warn("IMDS v2 failed, falling back to IMDS v1")
	}

	// Fallback to IMDS v1
	if result, err := m.fetchMetadata(ctx, endpoint, nil); err == nil {
		m.logger.WithField("endpoint", endpoint).Debug("Retrieved metadata using IMDS v1")
		return result, nil
	} else {
		return "", m.wrapError(endpoint, "both IMDS v2 and v1 failed", err)
	}
}

func (m *MetadataClient) fetchMetadata(ctx context.Context, endpoint string, headers map[string]string) (string, error) {
	url := fmt.Sprintf("%s%s/%s", m.baseURL, metadataBasePath, endpoint)

	req, err := http.NewRequestWithContext(ctx, httpMethodGET, url, nil)
	if err != nil {
		return "", m.wrapError(endpoint, "failed to create request", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return "", m.wrapError(endpoint, "HTTP request failed", err)
	}
	defer resp.Body.Close()

	if err := m.checkStatus(resp.StatusCode, endpoint); err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", m.wrapError(endpoint, "failed to read response body", err)
	}

	result := strings.TrimSpace(string(body))
	m.logger.WithFields(logrus.Fields{
		"endpoint": endpoint,
		"value":    result,
	}).Debug("Retrieved metadata")

	return result, nil
}

func (m *MetadataClient) getValidToken(ctx context.Context) (string, error) {
	if m.token != "" && time.Now().Before(m.tokenExpiry) {
		return m.token, nil
	}
	return m.requestToken(ctx)
}

func (m *MetadataClient) requestToken(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s%s", m.baseURL, tokenEndpoint)

	req, err := http.NewRequestWithContext(ctx, httpMethodPUT, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set(tokenTTLHeader, defaultTokenTTL)

	resp, err := m.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response: %w", err)
	}

	m.token = strings.TrimSpace(string(body))
	m.tokenExpiry = time.Now().Add(6*time.Hour - tokenBufferDuration)

	m.logger.Debug("Successfully obtained IMDS v2 token")
	return m.token, nil
}

func (m *MetadataClient) checkStatus(statusCode int, endpoint string) error {
	switch statusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return &IMDSError{
			Endpoint:   endpoint,
			StatusCode: statusCode,
			Message:    "metadata not found",
		}
	case http.StatusUnauthorized:
		return &IMDSError{
			Endpoint:   endpoint,
			StatusCode: statusCode,
			Message:    "unauthorized access - IMDS v2 may be enforced",
		}
	case http.StatusForbidden:
		return &IMDSError{
			Endpoint:   endpoint,
			StatusCode: statusCode,
			Message:    "forbidden access - check IMDS configuration",
		}
	default:
		return &IMDSError{
			Endpoint:   endpoint,
			StatusCode: statusCode,
			Message:    fmt.Sprintf("unexpected status code: %d", statusCode),
		}
	}
}

func (m *MetadataClient) wrapError(endpoint, message string, err error) error {
	return &IMDSError{
		Endpoint: endpoint,
		Message:  message,
		Err:      err,
	}
}
