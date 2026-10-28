package blog

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

const indexNowEndpoint = "https://api.indexnow.org/indexnow"

// SetIndexNowKey 设置 IndexNow 密钥；空则跳过推送。
func (s *Service) SetIndexNowKey(key string) {
	s.indexNowKey = strings.TrimSpace(key)
}

// IndexNowKey 返回已配置的密钥（供对外托管校验文件）。
func (s *Service) IndexNowKey() string {
	return s.indexNowKey
}

// notifyIndexNowAsync 发布后异步通知 Bing/IndexNow；事务外远程 IO，失败只打日志。
func (s *Service) notifyIndexNowAsync(paths ...string) {
	key := s.indexNowKey
	base := s.publicSiteURL
	if key == "" || base == "" {
		return
	}
	urls := make([]string, 0, len(paths))
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
			urls = append(urls, p)
			continue
		}
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		urls = append(urls, base+p)
	}
	if len(urls) == 0 {
		return
	}
	go s.postIndexNow(key, base, urls)
}

func (s *Service) postIndexNow(key, host string, urlList []string) {
	// 1. 组装 IndexNow JSON（host 不含协议）
	hostName := strings.TrimPrefix(strings.TrimPrefix(host, "https://"), "http://")
	hostName = strings.TrimRight(hostName, "/")
	body := map[string]any{
		"host":        hostName,
		"key":         key,
		"keyLocation": host + "/" + key + ".txt",
		"urlList":     urlList,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return
	}
	// 2. 短超时 POST；不阻塞发布主路径
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, indexNowEndpoint, bytes.NewReader(raw))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("indexnow: notify failed: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Printf("indexnow: unexpected status %d for %d urls", resp.StatusCode, len(urlList))
	}
}
