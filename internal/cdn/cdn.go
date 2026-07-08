package cdn

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type CDNConfig struct {
	Provider      string // "cloudflare-r2", "aws-s3", "local"
	BucketName    string
	Region        string
	AccessKey     string
	SecretKey     string
	CustomDomain  string // e.g., cdn.otexads.com
	PublicURLBase string // e.g., https://cdn.otexads.com
}

type CDNClient interface {
	Upload(ctx context.Context, key string, reader io.Reader, contentType string) (string, error)
	GetURL(key string) string
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

type CloudflareR2Client struct {
	client *s3.Client
	config CDNConfig
}

type AWSS3Client struct {
	client *s3.Client
	config CDNConfig
}

type LocalCDNClient struct {
	basePath string
	config   CDNConfig
}

func NewCDNClient(config CDNConfig) (CDNClient, error) {
	switch strings.ToLower(config.Provider) {
	case "cloudflare-r2", "aws-s3":
		cfg, err := loadAWSConfig(config)
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}

		client := s3.NewFromConfig(cfg)

		if strings.ToLower(config.Provider) == "cloudflare-r2" {
			return &CloudflareR2Client{
				client: client,
				config: config,
			}, nil
		}

		return &AWSS3Client{
			client: client,
			config: config,
		}, nil

	case "local":
		if err := os.MkdirAll(config.BucketName, 0755); err != nil {
			return nil, fmt.Errorf("failed to create local CDN directory: %w", err)
		}
		return &LocalCDNClient{
			basePath: config.BucketName,
			config:   config,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported CDN provider: %s", config.Provider)
	}
}

func loadAWSConfig(config CDNConfig) (aws.Config, error) {
	customResolver := aws.NewCredentialsCache(func(ctx context.Context) (aws.Credentials, error) {
		return aws.Credentials{
			AccessKeyID:     config.AccessKey,
			SecretAccessKey: config.SecretKey,
		}, nil
	})

	// For Cloudflare R2, we need to use a custom endpoint
	var opts []func(*config.LoadOptions) error
	opts = append(opts, config.WithCredentialsProvider(customResolver))

	if strings.Contains(config.Region, "r2.cloudflarestorage.com") {
		opts = append(opts, config.WithRegion("auto"))
	} else {
		opts = append(opts, config.WithRegion(config.Region))
	}

	return config.LoadDefaultConfig(context.Background(), opts...)
}

func (c *CloudflareR2Client) Upload(ctx context.Context, key string, reader io.Reader, contentType string) (string, error) {
	uploader := manager.NewUploader(c.client)

	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.config.BucketName),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
		// Cache for 1 year by default
		CacheControl: aws.String("public, max-age=31536000, immutable"),
		// Enable CORS
		ACL: "public-read",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to R2: %w", err)
	}

	return c.GetURL(key), nil
}

func (c *CloudflareR2Client) GetURL(key string) string {
	if c.config.CustomDomain != "" {
		return fmt.Sprintf("https://%s/%s", c.config.CustomDomain, key)
	}
	if c.config.PublicURLBase != "" {
		return fmt.Sprintf("%s/%s", c.config.PublicURLBase, key)
	}
	// Fallback to R2 public URL
	return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/%s", c.config.BucketName, c.config.Region, key)
}

func (c *CloudflareR2Client) Delete(ctx context.Context, key string) error {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.config.BucketName),
		Key:    aws.String(key),
	})
	return err
}

func (c *CloudflareR2Client) Exists(ctx context.Context, key string) (bool, error) {
	_, err := c.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.config.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (c *AWSS3Client) Upload(ctx context.Context, key string, reader io.Reader, contentType string) (string, error) {
	uploader := manager.NewUploader(c.client)

	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.config.BucketName),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
		CacheControl: aws.String("public, max-age=31536000, immutable"),
		ACL:         "public-read",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return c.GetURL(key), nil
}

func (c *AWSS3Client) GetURL(key string) string {
	if c.config.CustomDomain != "" {
		return fmt.Sprintf("https://%s/%s", c.config.CustomDomain, key)
	}
	if c.config.PublicURLBase != "" {
		return fmt.Sprintf("%s/%s", c.config.PublicURLBase, key)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", c.config.BucketName, c.config.Region, key)
}

func (c *AWSS3Client) Delete(ctx context.Context, key string) error {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.config.BucketName),
		Key:    aws.String(key),
	})
	return err
}

func (c *AWSS3Client) Exists(ctx context.Context, key string) (bool, error) {
	_, err := c.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(c.config.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (c *LocalCDNClient) Upload(ctx context.Context, key string, reader io.Reader, contentType string) (string, error) {
	filePath := fmt.Sprintf("%s/%s", c.basePath, key)
	
	// Create directory if needed
	dir := filePath[:strings.LastIndex(filePath, "/")]
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return c.GetURL(key), nil
}

func (c *LocalCDNClient) GetURL(key string) string {
	if c.config.CustomDomain != "" {
		return fmt.Sprintf("https://%s/%s", c.config.CustomDomain, key)
	}
	if c.config.PublicURLBase != "" {
		return fmt.Sprintf("%s/%s", c.config.PublicURLBase, key)
	}
	return fmt.Sprintf("/cdn/%s", key)
}

func (c *LocalCDNClient) Delete(ctx context.Context, key string) error {
	filePath := fmt.Sprintf("%s/%s", c.basePath, key)
	return os.Remove(filePath)
}

func (c *LocalCDNClient) Exists(ctx context.Context, key string) (bool, error) {
	filePath := fmt.Sprintf("%s/%s", c.basePath, key)
	_, err := os.Stat(filePath)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GenerateCDNKey creates a unique key for storing files in CDN
func GenerateCDNKey(campaignID, creativeID string) string {
	return fmt.Sprintf("creatives/%s/%s", campaignID, creativeID)
}

// GenerateCacheHeaders returns appropriate cache headers for CDN
func GenerateCacheHeaders(ttl time.Duration) map[string]string {
	return map[string]string{
		"Cache-Control": fmt.Sprintf("public, max-age=%d", int(ttl.Seconds())),
		"Expires":       time.Now().Add(ttl).UTC().Format(http.TimeFormat),
	}
}
