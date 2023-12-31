package minio

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
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
	c.logger.Debug(fmt.Sprintf("trying to upload file: %s (public: %t)", filePath, public))
	if err := c.createBucket(ctx, public); err != nil {
		return "", err
	}
	contentType, err := findContentType(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get mime type: %w", err)
	}
	c.logger.Debug(fmt.Sprintf("got content type: %s", contentType))
	fileName := randomiseFileName(filePath) + filepath.Ext(filePath)
	c.logger.Debug(fmt.Sprintf("generated file name: %s", fileName))
	info, err := c.client.FPutObject(
		ctx,
		c.bucketName,
		fileName,
		filePath,
		miniolib.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}
	if public {
		baseURL := c.client.EndpointURL().String()
		finalURL := fmt.Sprintf("%s/%s/%s", baseURL, c.bucketName, info.Key)
		c.logger.Debug(fmt.Sprintf("public link: %s", finalURL))
		return finalURL, nil
	}
	link, err := c.getPrivateShareLink(ctx, fileName)
	if err != nil {
		return "", fmt.Errorf("failed to get private share link: %w", err)
	}
	c.logger.Debug(fmt.Sprintf("private link: %s", link))
	return link, nil
}

// DownloadFile downloads a file from minio by the given input url,
// if customPath = "" file path will be the same as the URL object path
func (c *MinioClient) DownloadFile(ctx context.Context, input string, customPath string) error {
	c.logger.Debug(fmt.Sprintf("trying to download file: %s", input))
	u, err := url.Parse(input)
	if err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	}
	c.logger.Debug(fmt.Sprintf("parsed url: %s", u.String()))
	path := u.EscapedPath()
	pathSplit := strings.Split(path, "/")
	if len(pathSplit) < 2 {
		return fmt.Errorf("invalid url, could not fetch bucket or object name")
	}
	bucketName := pathSplit[1]
	objectName := pathSplit[2]
	if bucketName == "" || objectName == "" {
		return fmt.Errorf("invalid url, could not fetch bucket or object name")
	}
	c.logger.Debug(fmt.Sprintf("bucket name: %s, object name: %s", bucketName, objectName))
	if customPath == "" {
		customPath = "./files/" + objectName
		c.logger.Debug(fmt.Sprintf("custom path not provided, using default: %s", customPath))
	}
	err = c.client.FGetObject(ctx, bucketName, objectName, customPath, miniolib.GetObjectOptions{})
	return err
}

// Gets objects from urls and checks if they still exist
func (c *MinioClient) GetObjects(ctx context.Context, objectLinks []string) error {
	fmt.Println("Objects uploaded using minio-link:")
	fmt.Println("==================================")
	for _, link := range objectLinks {
		u, err := url.Parse(link)
		if err != nil {
			return fmt.Errorf("failed to parse url: %w", err)
		}
		path := u.EscapedPath()
		pathSplit := strings.Split(path, "/")
		if len(pathSplit) < 2 {
			return fmt.Errorf("invalid url, could not fetch bucket or object name")
		}
		bucketName := pathSplit[1]
		objectName := pathSplit[2]
		if bucketName == "" || objectName == "" {
			return fmt.Errorf("invalid url, could not fetch bucket or object name")
		}
		obj, err := c.client.StatObject(ctx, bucketName, objectName, miniolib.StatObjectOptions{})
		if err != nil {
			return fmt.Errorf("failed to get object: %w", err)
		}
		// TODO: improve design, add option to regen links
		fmt.Printf(
			"Object: %s - Bucket: %s - Expiry: %v - Last Modified: %v\n",
			obj.Key,
			bucketName,
			obj.Expires,
			obj.LastModified,
		)
	}
	return nil
}

func (c *MinioClient) setBucketPublic(ctx context.Context, policy string) error {
	err := c.client.SetBucketPolicy(ctx, c.bucketName, policy)
	if err != nil {
		return fmt.Errorf("setting bucket policy: %w", err)
	}
	return nil
}

func (c *MinioClient) createBucket(ctx context.Context, public bool) error {
	if !public {
		c.bucketName += "-private"
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
	}
	if public {
		err := c.setBucketPublic(ctx, fmt.Sprintf(bucketPolicyPublic, c.bucketName))
		if err != nil {
			return fmt.Errorf("failed to set bucket policy: %w", err)
		}
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
	logger := log.NewLogger().
		WithDirectory(dir).
		WithName("minio").
		WithDebug(debug).
		WithConsoleOutput(debug)
	mClient, err := miniolib.New(cfg.MinioEndpoint, &miniolib.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioAccessSecret, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}
	return &MinioClient{
		logger:        logger,
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
