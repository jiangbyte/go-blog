package captcha

import (
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"github.com/jiangbyte/go-blog/internal/config"
)

// ImageSource 对齐 mypage CaptchaImageSource：remote / local / mixed。
type ImageSource struct {
	cfg         config.CaptchaConfig
	localLoader func() ([]byte, error) // 可选本地图库；nil 则跳过
	client      *http.Client
}

func NewImageSource(cfg config.CaptchaConfig, localLoader func() ([]byte, error)) *ImageSource {
	return &ImageSource{
		cfg:         cfg,
		localLoader: localLoader,
		client:      &http.Client{Timeout: 8 * time.Second},
	}
}

func (s *ImageSource) FetchBytes() []byte {
	mode := strings.ToLower(strings.TrimSpace(s.cfg.ImageSource))
	if mode == "" {
		mode = "remote"
	}
	bgW, bgH := s.cfg.BGWidth, s.cfg.BGHeight

	switch mode {
	case "remote":
		if data := s.fetchRemote(); len(data) > 0 {
			return data
		}
	case "local":
		if data := s.fetchLocal(); len(data) > 0 {
			return data
		}
	default: // mixed：本地优先，否则远端
		if data := s.fetchLocal(); len(data) > 0 {
			return data
		}
		if data := s.fetchRemote(); len(data) > 0 {
			return data
		}
	}
	return SolidFallbackImage(bgW, bgH)
}

func (s *ImageSource) fetchLocal() []byte {
	if s.localLoader == nil {
		return nil
	}
	data, err := s.localLoader()
	if err != nil {
		log.Printf("load local captcha image failed: %v", err)
		return nil
	}
	return data
}

func (s *ImageSource) fetchRemote() []byte {
	urls := append([]string(nil), s.cfg.RemoteURLList()...)
	if len(urls) == 0 {
		return nil
	}
	rand.Shuffle(len(urls), func(i, j int) { urls[i], urls[j] = urls[j], urls[i] })
	for _, url := range urls {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			continue
		}
		// 对齐 mypage：带 UA；额外声明可接受常见图片类型（含 webp）
		req.Header.Set("User-Agent", "go-blog-captcha/1.0")
		req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/*,*/*;q=0.8")
		resp, err := s.client.Do(req)
		if err != nil {
			log.Printf("fetch remote captcha image failed: %s: %v", url, err)
			continue
		}
		data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
		resp.Body.Close()
		if err != nil || resp.StatusCode != http.StatusOK || len(data) == 0 {
			log.Printf("fetch remote captcha image failed: %s status=%d", url, resp.StatusCode)
			continue
		}
		if !looksLikeImage(data) {
			log.Printf("fetch remote captcha image failed: %s not an image (ct=%s len=%d)",
				url, resp.Header.Get("Content-Type"), len(data))
			continue
		}
		return data
	}
	return nil
}

// looksLikeImage 粗判 JPEG/PNG/GIF/WebP/RIFF，避免把 HTML 错误页当底图。
func looksLikeImage(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	if data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff {
		return true // jpeg
	}
	if len(data) >= 8 && string(data[:8]) == "\x89PNG\r\n\x1a\n" {
		return true
	}
	if string(data[:3]) == "GIF" {
		return true
	}
	// WebP: RIFF....WEBP
	if string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		return true
	}
	return false
}
