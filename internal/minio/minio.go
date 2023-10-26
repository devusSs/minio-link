package minio

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	miniolib "github.com/minio/minio-go/v7"
	credentials "github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/devusSs/minio-link/internal/config/environment"
	"github.com/devusSs/minio-link/pkg/log"
)

// Wrapper around minio library
type MinioClient struct {
	logger        *log.Logger
	client        *miniolib.Client
	bucketName    string
	bucketRegion  string
	objectLocking bool
	expiry        time.Duration
}

// UploadFile uploads a file to minio and returns the share link
func (c *MinioClient) UploadFile(
	ctx context.Context,
	filePath string,
	public bool,
) (string, error) {
	c.logger.Debug(fmt.Sprintf("trying to upload file %s", filepath.Base(filePath)))
	if err := c.createBucket(ctx, public); err != nil {
		c.logger.Error(err.Error())
		return "", err
	}
	contentType, err := findContentType(filePath)
	if err != nil {
		c.logger.Error(fmt.Sprintf("failed to get mime type: %s", err.Error()))
		return "", fmt.Errorf("failed to get mime type: %w", err)
	}
	c.logger.Debug(fmt.Sprintf("set content type: %s", contentType))
	fileName := randomiseFileName(filePath)
	info, err := c.client.FPutObject(
		ctx,
		c.bucketName,
		fileName,
		filePath,
		miniolib.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		c.logger.Error(fmt.Sprintf("failed to upload file: %s", err.Error()))
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	c.logger.Debug(
		fmt.Sprintf(
			"upload done, bucket: %s, name: %s, tag: %s, size: %d",
			info.Bucket,
			info.Key,
			info.ETag,
			info.Size,
		),
	)
	if public {
		baseURL := c.client.EndpointURL().String()
		finalURL := fmt.Sprintf("%s/%s/%s", baseURL, c.bucketName, info.Key)
		c.logger.Debug(fmt.Sprintf("public share url: %s", finalURL))
		return finalURL, nil
	}
	return c.getPrivateShareLink(ctx, fileName)
}

func (c *MinioClient) setBucketPublic(ctx context.Context, policy string) error {
	err := c.client.SetBucketPolicy(ctx, c.bucketName, policy)
	if err != nil {
		return fmt.Errorf("setting bucket policy: %w", err)
	}
	c.logger.Debug(fmt.Sprintf("set bucket %s policy to public", c.bucketName))
	return nil
}

func (c *MinioClient) createBucket(ctx context.Context, public bool) error {
	if !public {
		c.bucketName += "-private"
		c.logger.Debug("using private bucket")
	}
	exists, err := c.client.BucketExists(ctx, c.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check if bucket exists: %w", err)
	}
	if !exists {
		err = c.client.MakeBucket(ctx, c.bucketName, miniolib.MakeBucketOptions{
			Region:        c.bucketRegion,
			ObjectLocking: c.objectLocking,
		})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		c.logger.Debug(fmt.Sprintf("created bucket %s", c.bucketName))
	} else {
		c.logger.Warn(fmt.Sprintf("bucket %s already exists", c.bucketName))
	}
	if public {
		err := c.setBucketPublic(ctx, fmt.Sprintf(bucketPolicyPublic, c.bucketName))
		if err != nil {
			return fmt.Errorf("failed to set bucket policy: %w", err)
		}
		c.logger.Debug("set bucket policy to public")
	}
	return nil
}

func (c *MinioClient) getPrivateShareLink(ctx context.Context, obj string) (string, error) {
	presignedURL, err := c.client.PresignedGetObject(
		ctx,
		c.bucketName,
		obj,
		c.expiry,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to get presigned url: %w", err)
	}
	return presignedURL.String(), nil
}

// NewClient creates a new minio client
func NewClient(dir string, debug bool, cfg *environment.EnvConfig) (*MinioClient, error) {
	mClient, err := miniolib.New(cfg.MinioEndpoint, &miniolib.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioAccessSecret, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}
	return &MinioClient{
		logger: log.NewLogger().
			WithDirectory(dir).
			WithName("minio").
			WithDebug(debug).
			WithConsoleOutput(debug),
		client:        mClient,
		bucketName:    cfg.MinioBucketName,
		bucketRegion:  cfg.MinioRegion,
		objectLocking: cfg.MinioObjectLocking,
		expiry:        cfg.MinioDefaultExpiry,
	}, nil
}

const (
	bucketPolicyPublic string = `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Sid": "AddPerm",
				"Effect": "Allow",
				"Principal": {
					"AWS": [
						"*"
					]
				},
				"Action": [
					"s3:GetObject"
				],
				"Resource": [
					"arn:aws:s3:::%s/*"
				]
			}
		]
	}`
)

func findContentType(filePath string) (string, error) {
	mime, err := mimetype.DetectFile(filePath)
	if err != nil {
		return "", err
	}
	return mime.String(), nil
}

func randomiseFileName(filePath string) string {
	return uuid.New().String()
}
