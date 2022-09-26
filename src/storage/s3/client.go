package s3

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/traPtitech/trap-collection-server/src/config"
	"github.com/traPtitech/trap-collection-server/src/storage"
)

type Client struct {
	client *s3.Client
	bucket string
}

func NewClient(conf config.StorageS3) (*Client, error) {
	ctx := context.Background()

	accessKeyID, err := conf.AccessKeyID()
	if err != nil {
		return nil, fmt.Errorf("failed to get access key id: %w", err)
	}

	secretAccessKey, err := conf.SecretAccessKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get secret access key: %w", err)
	}

	region, err := conf.Region()
	if err != nil {
		return nil, fmt.Errorf("failed to get region: %w", err)
	}

	bucket, err := conf.Bucket()
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket: %w", err)
	}

	endpoint, err := conf.Endpoint()
	if err != nil {
		return nil, fmt.Errorf("failed to get endpoint: %w", err)
	}

	client, err := setupS3(ctx, accessKeyID, secretAccessKey, region, endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to setup s3: %w", err)
	}

	return &Client{
		client: client,
		bucket: bucket,
	}, nil
}

func setupS3(
	ctx context.Context,
	accessKeyID string,
	secretAccessKey string,
	region string,
	endpoint string,
) (*s3.Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(
		ctx,
		awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
		),
		awsConfig.WithRegion(region),
		awsConfig.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(
				func(service string, region string, options ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{
						URL:           endpoint,
						SigningRegion: region,
					}, nil
				},
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return client, nil
}

var (
	ErrAlreadyExists = fmt.Errorf("already exists")
)

func (c *Client) saveFile(
	ctx context.Context,
	name string,
	content io.Reader,
) error {
	// オブジェクトの存在確認
	objects, err := c.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: &c.bucket,
		Prefix: &name,
	})
	if err != nil {
		return fmt.Errorf("failed to list objects: %w", err)
	}

	for _, object := range objects.Contents {
		if *object.Key == name {
			return ErrAlreadyExists
		}
	}

	uploader := manager.NewUploader(c.client, func(u *manager.Uploader) {
		u.Concurrency = 5
	})
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: &c.bucket,
		Key:    &name,
		Body:   content,
	})
	if err != nil {
		return fmt.Errorf("failed to create object: %w", err)
	}

	return nil
}

func (c *Client) createTempURL(ctx context.Context, name string, expires time.Duration) (*url.URL, error) {
	objects, err := c.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: &c.bucket,
		Prefix: &name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	exist := false
	for _, object := range objects.Contents {
		if object.Key != nil && *object.Key == name {
			exist = true
			break
		}
	}
	if !exist {
		return nil, storage.ErrNotFound
	}

	presignClient := s3.NewPresignClient(c.client)
	result, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &c.bucket,
		Key:    &name,
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return nil, fmt.Errorf("failed to presign get object: %w", err)
	}

	tmpURL, err := url.Parse(result.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	return tmpURL, nil
}
