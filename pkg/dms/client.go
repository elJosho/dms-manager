package dms

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/databasemigrationservice"
)

// Client wraps the AWS DMS client with additional functionality
type Client struct {
	svc     *databasemigrationservice.Client
	profile string
	region  string
}

// NewClient creates a new DMS client with the specified profile and region
func NewClient(ctx context.Context, profile, region string) (*Client, error) {
	var opts []func(*config.LoadOptions) error

	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Support custom endpoint for LocalStack
	var clientOpts []func(*databasemigrationservice.Options)
	if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
		clientOpts = append(clientOpts, func(o *databasemigrationservice.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	}

	return &Client{
		svc:     databasemigrationservice.NewFromConfig(cfg, clientOpts...),
		profile: profile,
		region:  cfg.Region,
	}, nil
}

// GetProfile returns the AWS profile being used
func (c *Client) GetProfile() string {
	return c.profile
}

// GetRegion returns the AWS region being used
func (c *Client) GetRegion() string {
	return c.region
}

// GetService returns the underlying AWS DMS client
func (c *Client) GetService() *databasemigrationservice.Client {
	return c.svc
}
