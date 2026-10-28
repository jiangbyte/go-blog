package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/jiangbyte/go-blog/internal/config"
)

type MinIOStorage struct {
	cfg           config.EngineConfig
	presignExpire time.Duration
	client        *minio.Client
	presignClient *minio.Client
	bucket        string
}

func NewMinIO(cfg config.EngineConfig, presignExpireSeconds int) (*MinIOStorage, error) {
	if !cfg.Configured() {
		return nil, fmt.Errorf("minio 未配置完整凭据")
	}
	endpoint := strings.TrimSpace(cfg.Endpoint)
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimRight(endpoint, "/")
	if endpoint == "" {
		return nil, fmt.Errorf("minio endpoint 为空")
	}
	opts := &minio.Options{
		Creds:        credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure:       cfg.UseSSL,
		Region:       cfg.Region,
		BucketLookup: minio.BucketLookupPath,
	}
	client, err := minio.New(endpoint, opts)
	if err != nil {
		return nil, err
	}
	presignClient := client
	base := strings.TrimSpace(strings.TrimRight(cfg.BaseURL, "/"))
	apiURL := ensureScheme(endpoint, cfg.UseSSL)
	if base != "" && base != apiURL {
		baseHost := strings.TrimPrefix(base, "https://")
		baseHost = strings.TrimPrefix(baseHost, "http://")
		baseHost = strings.TrimRight(baseHost, "/")
		baseSecure := strings.HasPrefix(base, "https://") || (!strings.HasPrefix(base, "http://") && cfg.UseSSL)
		pc, err := minio.New(baseHost, &minio.Options{
			Creds:        credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
			Secure:       baseSecure,
			Region:       cfg.Region,
			BucketLookup: minio.BucketLookupPath,
		})
		if err == nil {
			presignClient = pc
		}
	}
	expire := time.Duration(presignExpireSeconds) * time.Second
	if expire <= 0 {
		expire = time.Hour
	}
	return &MinIOStorage{
		cfg:           cfg,
		presignExpire: expire,
		client:        client,
		presignClient: presignClient,
		bucket:        cfg.Bucket,
	}, nil
}

func (s *MinIOStorage) Upload(ctx context.Context, objectName string, r io.Reader, size int64, contentType string) error {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	_, err := s.client.PutObject(ctx, s.bucket, objectName, r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *MinIOStorage) Delete(ctx context.Context, objectName string) error {
	return s.client.RemoveObject(ctx, s.bucket, objectName, minio.RemoveObjectOptions{})
}

func (s *MinIOStorage) URL(ctx context.Context, objectName string) (string, error) {
	if s.cfg.BucketPublic {
		return s.publicURL(objectName), nil
	}
	u, err := s.presignClient.PresignedGetObject(ctx, s.bucket, objectName, s.presignExpire, url.Values{})
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (s *MinIOStorage) GetBytes(ctx context.Context, objectName string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()
	return io.ReadAll(obj)
}

func (s *MinIOStorage) publicURL(objectName string) string {
	encoded := quoteObjectName(objectName)
	base := strings.TrimSpace(strings.TrimRight(s.cfg.BaseURL, "/"))
	if base != "" {
		return base + "/" + s.bucket + "/" + encoded
	}
	return ensureScheme(s.cfg.Endpoint, s.cfg.UseSSL) + "/" + s.bucket + "/" + encoded
}
