package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	qiniuStorage "github.com/qiniu/go-sdk/v7/storage"

	"github.com/jiangbyte/go-blog/internal/config"
)

type QiniuStorage struct {
	cfg           config.EngineConfig
	presignExpire int64
	mac           *qbox.Mac
	bucket        string
	uploader      *qiniuStorage.FormUploader
	bucketMgr     *qiniuStorage.BucketManager
}

func NewQiniu(cfg config.EngineConfig, presignExpireSeconds int) (*QiniuStorage, error) {
	if !cfg.Configured() {
		return nil, fmt.Errorf("七牛云未配置完整凭据")
	}
	expire := int64(presignExpireSeconds)
	if expire <= 0 {
		expire = 3600
	}
	mac := qbox.NewMac(cfg.AccessKey, cfg.SecretKey)
	zone, err := resolveQiniuZone(cfg.AccessKey, cfg.Bucket, cfg.Region)
	if err != nil {
		return nil, err
	}
	cfgUp := &qiniuStorage.Config{Zone: zone, UseHTTPS: cfg.UseSSL, UseCdnDomains: true}
	return &QiniuStorage{
		cfg:           cfg,
		presignExpire: expire,
		mac:           mac,
		bucket:        cfg.Bucket,
		uploader:      qiniuStorage.NewFormUploader(cfgUp),
		bucketMgr:     qiniuStorage.NewBucketManager(mac, cfgUp),
	}, nil
}

// resolveQiniuZone 解析机房：支持 z2 / up-z2.qiniup.com；未配或无法识别时按桶自动探测。
func resolveQiniuZone(accessKey, bucket, region string) (*qiniuStorage.Zone, error) {
	r := strings.ToLower(strings.TrimSpace(region))
	switch {
	case r == "" || r == "auto":
		// fallthrough to auto below
	case r == "z0" || r == "huadong" || strings.Contains(r, "up-z0"):
		return &qiniuStorage.ZoneHuadong, nil
	case r == "z1" || r == "huabei" || r == "north" || strings.Contains(r, "up-z1"):
		return &qiniuStorage.ZoneHuabei, nil
	case r == "z2" || r == "huanan" || r == "south" || strings.Contains(r, "up-z2"):
		return &qiniuStorage.ZoneHuanan, nil
	case r == "cn-east-2" || r == "zhejiang2" || strings.Contains(r, "up-cn-east-2"):
		return &qiniuStorage.ZoneHuadongZheJiang2, nil
	case r == "na0" || r == "northamerica" || strings.Contains(r, "up-na0"):
		return &qiniuStorage.ZoneBeimei, nil
	case r == "as0" || r == "southeastasia" || strings.Contains(r, "up-as0"):
		return &qiniuStorage.ZoneXinjiapo, nil
	default:
		// 未知写法：继续走 UC 自动探测
	}
	zone, err := qiniuStorage.GetRegion(accessKey, bucket)
	if err != nil {
		return nil, fmt.Errorf("七牛区域探测失败(可设 STORAGE_QINIU_REGION=z2): %w", err)
	}
	return zone, nil
}

func (s *QiniuStorage) Upload(ctx context.Context, objectName string, r io.Reader, size int64, contentType string) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	putPolicy := qiniuStorage.PutPolicy{Scope: s.bucket + ":" + objectName}
	putPolicy.Expires = 3600
	token := putPolicy.UploadToken(s.mac)
	ret := qiniuStorage.PutRet{}
	extra := &qiniuStorage.PutExtra{}
	if contentType != "" {
		extra.MimeType = contentType
	}
	return s.uploader.Put(ctx, &ret, token, objectName, bytes.NewReader(data), int64(len(data)), extra)
}

func (s *QiniuStorage) Delete(ctx context.Context, objectName string) error {
	_ = ctx
	return s.bucketMgr.Delete(s.bucket, objectName)
}

func (s *QiniuStorage) URL(ctx context.Context, objectName string) (string, error) {
	_ = ctx
	base := s.publicURL(objectName)
	if s.cfg.BucketPublic {
		return base, nil
	}
	deadline := time.Now().Add(time.Duration(s.presignExpire) * time.Second).Unix()
	return qiniuStorage.MakePrivateURL(s.mac, strings.TrimRight(s.domain(), "/"), objectName, deadline), nil
}

func (s *QiniuStorage) GetBytes(ctx context.Context, objectName string) ([]byte, error) {
	u, err := s.URL(ctx, objectName)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("qiniu download status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func (s *QiniuStorage) domain() string {
	base := strings.TrimSpace(s.cfg.BaseURL)
	if base != "" {
		return ensureScheme(base, s.cfg.UseSSL)
	}
	ep := strings.TrimSpace(s.cfg.Endpoint)
	if ep == "" {
		return ""
	}
	return ensureScheme(ep, s.cfg.UseSSL)
}

func (s *QiniuStorage) publicURL(objectName string) string {
	domain := strings.TrimRight(s.domain(), "/")
	if domain == "" {
		return quoteObjectName(objectName)
	}
	return domain + "/" + quoteObjectName(objectName)
}
