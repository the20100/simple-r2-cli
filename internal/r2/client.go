package r2

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	s3     *s3.Client
	ctx    context.Context
	cancel context.CancelFunc
}

type BucketInfo struct {
	Name         string `json:"name"`
	CreationDate string `json:"creation_date"`
}

type ObjectInfo struct {
	Key          string `json:"key"`
	Size         int64  `json:"size"`
	LastModified string `json:"last_modified"`
	ETag         string `json:"etag,omitempty"`
	StorageClass string `json:"storage_class,omitempty"`
}

type ObjectDetail struct {
	Key          string `json:"key"`
	Size         int64  `json:"size"`
	LastModified string `json:"last_modified"`
	ContentType  string `json:"content_type"`
	ETag         string `json:"etag"`
}

func NewClient(accountID, accessKeyID, secretAccessKey string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			"",
		)),
		config.WithRegion("auto"),
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(
			fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID),
		)
	})

	return &Client{s3: s3Client, ctx: ctx, cancel: cancel}, nil
}

func (c *Client) Close() {
	c.cancel()
}

func (c *Client) ListBuckets() ([]BucketInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	out, err := c.s3.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("listing buckets: %w", err)
	}

	buckets := make([]BucketInfo, len(out.Buckets))
	for i, b := range out.Buckets {
		buckets[i] = BucketInfo{
			Name:         aws.ToString(b.Name),
			CreationDate: b.CreationDate.UTC().Format(time.RFC3339),
		}
	}
	return buckets, nil
}

func (c *Client) ListObjects(bucket, prefix, continuationToken string, maxKeys int32) ([]ObjectInfo, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		MaxKeys: aws.Int32(maxKeys),
	}
	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}
	if continuationToken != "" {
		input.ContinuationToken = aws.String(continuationToken)
	}

	out, err := c.s3.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, "", fmt.Errorf("listing objects: %w", err)
	}

	objects := make([]ObjectInfo, len(out.Contents))
	for i, obj := range out.Contents {
		objects[i] = ObjectInfo{
			Key:          aws.ToString(obj.Key),
			Size:         aws.ToInt64(obj.Size),
			LastModified: obj.LastModified.UTC().Format(time.RFC3339),
			ETag:         aws.ToString(obj.ETag),
		}
	}

	nextToken := ""
	if out.NextContinuationToken != nil {
		nextToken = *out.NextContinuationToken
	}

	return objects, nextToken, nil
}

func (c *Client) HeadObject(bucket, key string) (*ObjectDetail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	out, err := c.s3.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("head object: %w", err)
	}

	return &ObjectDetail{
		Key:          key,
		Size:         aws.ToInt64(out.ContentLength),
		LastModified: out.LastModified.UTC().Format(time.RFC3339),
		ContentType:  aws.ToString(out.ContentType),
		ETag:         aws.ToString(out.ETag),
	}, nil
}

func (c *Client) GetObject(bucket, key string, w io.Writer) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	out, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return 0, fmt.Errorf("getting object: %w", err)
	}
	defer out.Body.Close()

	n, err := io.Copy(w, out.Body)
	if err != nil {
		return n, fmt.Errorf("reading object body: %w", err)
	}
	return n, nil
}

func (c *Client) PutObject(bucket, key, filePath, contentType string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   f,
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	_, err = c.s3.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("putting object: %w", err)
	}
	return nil
}

func (c *Client) DeleteObject(bucket, key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := c.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("deleting object: %w", err)
	}
	return nil
}
