package storage

import (
	"context"
	"os"

	middleware "blog_server/share/middelware"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type UploadedFileResponse struct {
	Key    string
	Bucket string
}

type Storage struct {
	client *s3.Client
}

func NewStorage() (*Storage, error) {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(os.Getenv("B2_REGION")),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				os.Getenv("B2_ACCESS_KEY_ID"),
				os.Getenv("B2_SECRET_ACCESS_KEY"),
				"",
			),
		),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = &[]string{os.Getenv("B2_ENDPOINT")}[0]
	})

	return &Storage{client: client}, nil
}

func (s *Storage) UploadFile(ctx context.Context, bucket, key, path, mine string) (*UploadedFileResponse, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	defer middleware.CleanupTempFile(path)

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: &[]string{mine}[0],
	})
	if err != nil {
		return nil, err
	}

	return &UploadedFileResponse{Key: key, Bucket: bucket}, nil
}

func (s *Storage) GetUrl(ctx context.Context, key, bucket string) (string, error) {
	pClient := s3.NewPresignClient(s.client)

	req, err := pClient.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		},
	)
	if err != nil {
		return "", err
	}

	return req.URL, nil
}

func (s *Storage) DeleteFile(ctx context.Context, bucket, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}

	return nil
}
