package storage

import (
	"context"
	"fmt"
	"time"

	"cnmt/internal/common/env"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)


type ContentType string

const (
	ContentTypeImage = "image/png"
	ContentTypeVideo = "video/mp4"
	ContentTypeAudio = "audio/mpeg"
	ContentTypeDocument = "application/pdf"
	ContentTypeOther = "application/octet-stream"
)

type ObjStorage struct {
	client *s3.Client
	bucket string
	urlTTL time.Duration
}

func NewObjStorage(urlTTL time.Duration) (*ObjStorage, error) {
	accountID := env.GetString("OBJECT_STORAGE_ACCOUNT_ID", "r2")
	accessKeyID := env.GetString("OBJECT_STORAGE_ACCESS_KEY_ID", "r2")
	secretAccessKey := env.GetString("OBJECT_STORAGE_SECRET_ACCESS_KEY", "r2")
	bucket := env.GetString("OBJECT_STORAGE_BUCKET_NAME", "cnmt-stg")

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID))
		o.UsePathStyle = true
	})

	return &ObjStorage{client: client, bucket: bucket, urlTTL: urlTTL}, nil
}

func (o *ObjStorage) GetPresignedUrl(ctx context.Context, key, contentType string) (string, error) {
	
	presignClient := s3.NewPresignClient(o.client)
	presignedResult, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(o.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(o.urlTTL))
	if err != nil {
		return "", err
	}
	return presignedResult.URL, nil
}
