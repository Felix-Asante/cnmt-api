package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"cnmt/internal/common/env"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

const MaxPaymentProofBytes int64 = 10 << 20 // 10 MiB

var AllowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
}

var (
	ErrObjectNotFound   = errors.New("payment proof has not been uploaded")
	ErrObjectTooLarge   = errors.New("payment proof exceeds the 10MB limit")
	ErrInvalidImageType = errors.New("payment proof must be a jpeg or png image")
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
	returnValue, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(o.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(o.urlTTL))
	if err != nil {
		return "", err
	}
	return returnValue.URL, nil
}

func (o *ObjStorage) ValidatePaymentProof(ctx context.Context, key string) error {
	head, err := o.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(o.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFound(err) {
			return ErrObjectNotFound
		}
		return err
	}
	if head.ContentLength == nil || *head.ContentLength == 0 {
		return ErrObjectNotFound
	}
	if *head.ContentLength > MaxPaymentProofBytes {
		return ErrObjectTooLarge
	}

	obj, err := o.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(o.bucket),
		Key:    aws.String(key),
		Range:  aws.String("bytes=0-15"),
	})
	if err != nil {
		if isNotFound(err) {
			return ErrObjectNotFound
		}
		return err
	}
	defer obj.Body.Close()

	header, err := io.ReadAll(io.LimitReader(obj.Body, 16))
	if err != nil {
		return err
	}
	if !isJPEGOrPNG(header) {
		return ErrInvalidImageType
	}
	return nil
}

func isNotFound(err error) bool {
	var nsk *types.NotFound
	if errors.As(err, &nsk) {
		return true
	}
	var apiErr smithy.APIError
	return errors.As(err, &apiErr) && (apiErr.ErrorCode() == "NotFound" || apiErr.ErrorCode() == "NoSuchKey")
}

func isJPEGOrPNG(b []byte) bool {
	if len(b) >= 3 && b[0] == 0xFF && b[1] == 0xD8 && b[2] == 0xFF {
		return true
	}
	return bytes.HasPrefix(b, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
}
