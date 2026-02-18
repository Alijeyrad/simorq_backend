package s3

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/Alijeyrad/simorq_backend/config"
)

// Client wraps the AWS S3 client configured for ArvanCloud S3-compatible storage.
type Client struct {
	s3     *s3.Client
	presig *s3.PresignClient
	bucket string
	ttl    time.Duration
}

// New creates a new S3 client configured for ArvanCloud.
func New(cfg config.S3Config) (*Client, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("s3: bucket name is required")
	}

	resolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...any) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               cfg.Endpoint,
				SigningRegion:     cfg.Region,
				HostnameImmutable: true,
			}, nil
		},
	)

	awsCfg, err := awscfg.LoadDefaultConfig(context.Background(),
		awscfg.WithRegion(cfg.Region),
		awscfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
		awscfg.WithEndpointResolverWithOptions(resolver),
	)
	if err != nil {
		return nil, fmt.Errorf("s3: load config: %w", err)
	}

	cli := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true // ArvanCloud requires path-style
	})

	ttl := time.Duration(cfg.PresignTTLSec) * time.Second
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}

	return &Client{
		s3:     cli,
		presig: s3.NewPresignClient(cli),
		bucket: cfg.Bucket,
		ttl:    ttl,
	}, nil
}

// Upload puts an object into S3. The key should follow the convention
// {entity}/{clinic_id}/{uuid}.{ext}.
func (c *Client) Upload(ctx context.Context, key, contentType string, body io.Reader, size int64) error {
	_, err := c.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(c.bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(contentType),
		ACL:           types.ObjectCannedACLPrivate,
	})
	if err != nil {
		return fmt.Errorf("s3 upload %q: %w", key, err)
	}
	return nil
}

// PresignDownload generates a presigned GET URL valid for the configured TTL.
func (c *Client) PresignDownload(ctx context.Context, key string) (string, error) {
	req, err := c.presig.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(c.ttl))
	if err != nil {
		return "", fmt.Errorf("s3 presign %q: %w", key, err)
	}
	return req.URL, nil
}

// Delete removes an object from S3.
func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("s3 delete %q: %w", key, err)
	}
	return nil
}
