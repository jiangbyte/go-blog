package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/jiangbyte/go-blog/internal/config"
)

const (
	ProviderMinIO = "minio"
	ProviderQiniu = "qiniu"

	EngineMinIO = "MINIO"
	EngineQiniu = "QINIU"
)

// ObjectStorage 对象存储统一接口。
type ObjectStorage interface {
	Upload(ctx context.Context, objectName string, r io.Reader, size int64, contentType string) error
	Delete(ctx context.Context, objectName string) error
	URL(ctx context.Context, objectName string) (string, error)
	GetBytes(ctx context.Context, objectName string) ([]byte, error)
}

type EngineInfo struct {
	Engine       string `json:"engine"`
	Provider     string `json:"provider"`
	DisplayName  string `json:"display_name"`
	Configured   bool   `json:"configured"`
	IsDefault    bool   `json:"is_default"`
	Bucket       any    `json:"bucket"`
	Endpoint     any    `json:"endpoint"`
	BucketPublic bool   `json:"bucket_public"`
}

type Status struct {
	DefaultEngine        string       `json:"default_engine"`
	UploadMaxBytes       int64        `json:"upload_max_bytes"`
	PresignExpireSeconds int          `json:"presign_expire_seconds"`
	Engines              []EngineInfo `json:"engines"`
	Hint                 string       `json:"hint"`
}

func EngineToProvider(engine string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(engine)) {
	case EngineMinIO, ProviderMinIO:
		return ProviderMinIO, nil
	case EngineQiniu, ProviderQiniu:
		return ProviderQiniu, nil
	default:
		return "", fmt.Errorf("未知存储引擎: %s", engine)
	}
}

func ProviderToEngine(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case ProviderMinIO:
		return EngineMinIO
	case ProviderQiniu:
		return EngineQiniu
	default:
		return strings.ToUpper(provider)
	}
}

func DisplayName(provider string) string {
	switch strings.ToLower(provider) {
	case ProviderMinIO:
		return "MinIO"
	case ProviderQiniu:
		return "七牛云"
	default:
		return provider
	}
}

func quoteObjectName(objectName string) string {
	parts := strings.Split(objectName, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return strings.Join(parts, "/")
}

// ensureScheme 补全协议。已带 http:// 或 https:// 时原样返回，绝不自动升 HTTPS。
func ensureScheme(endpoint string, useSSL bool) string {
	ep := strings.TrimSpace(strings.TrimRight(endpoint, "/"))
	if ep == "" {
		return ""
	}
	lower := strings.ToLower(ep)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return ep
	}
	if useSSL {
		return "https://" + ep
	}
	return "http://" + ep
}

func engineConfig(settings config.StorageSettings, provider string) (config.EngineConfig, error) {
	switch strings.ToLower(provider) {
	case ProviderMinIO:
		return settings.MinIO, nil
	case ProviderQiniu:
		return settings.Qiniu, nil
	default:
		return config.EngineConfig{}, fmt.Errorf("不支持的存储: %s", provider)
	}
}
