package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName        string
	AppHost        string
	AppPort        int
	Debug          bool
	CORSOrigins    []string
	AuthRSAKeyPath string
	// PublicSiteURL：对外站点根地址（sitemap 等），如 https://example.com
	PublicSiteURL string
	// IndexNowKey：Bing IndexNow 密钥；须与站点根目录 {key}.txt 一致；空则不推送
	IndexNowKey string
	MySQL       MySQLConfig
	Redis       RedisConfig
	Token       TokenConfig
	Storage     StorageSettings
	Captcha     CaptchaConfig
	MySQLDocker string
}

// CaptchaConfig 对齐 mypage：底图来源 + 拼图尺寸 + Redis 容差。
type CaptchaConfig struct {
	TTLSeconds      int
	TokenTTLSeconds int
	TolerancePx     int
	MaxFailures     int
	ImageSource     string // remote | local | mixed
	RemoteURLs      string
	BGWidth         int
	BGHeight        int
	PuzzleSize      int
}

func (c CaptchaConfig) RemoteURLList() []string {
	parts := strings.Split(c.RemoteURLs, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

type MySQLConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

func (m MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		m.User, m.Password, m.Host, m.Port, m.Database)
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type TokenConfig struct {
	TTLSeconds      int
	CookieName      string
	CookieNameAdmin string
	HeaderName      string
	// CookieSecure: true | false | auto（按 TLS / X-Forwarded-Proto）
	CookieSecure string
}

// EngineConfig 单个对象存储引擎凭据。
type EngineConfig struct {
	AccessKey    string
	SecretKey    string
	Endpoint     string
	Bucket       string
	Region       string
	UseSSL       bool
	BaseURL      string
	BucketPublic bool
}

func (e EngineConfig) Configured() bool {
	return strings.TrimSpace(e.AccessKey) != "" &&
		strings.TrimSpace(e.SecretKey) != "" &&
		strings.TrimSpace(e.Bucket) != ""
}

// StorageSettings 默认引擎 + MinIO / 七牛双段配置。
type StorageSettings struct {
	DefaultEngine        string
	UploadMaxBytes       int64
	PresignExpireSeconds int
	AllowedExtensions    string
	MinIO                EngineConfig
	Qiniu                EngineConfig
}

func (s StorageSettings) AllowedExtSet() map[string]bool {
	parts := strings.Split(s.AllowedExtensions, ",")
	out := make(map[string]bool, len(parts))
	for _, p := range parts {
		p = strings.ToLower(strings.TrimSpace(p))
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, ".") {
			p = "." + p
		}
		out[p] = true
	}
	return out
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	cfg := &Config{
		AppName:        getenv("APP_NAME", "go-blog"),
		AppHost:        getenv("APP_HOST", "127.0.0.1"),
		AppPort:        getenvInt("APP_PORT", 8000),
		Debug:          getenvBool("DEBUG", false),
		CORSOrigins:    splitCSV(getenv("CORS_ALLOWED_ORIGINS", "http://127.0.0.1:8088,http://localhost:8088,http://127.0.0.1:5173,http://localhost:5173,http://127.0.0.1:5174,http://localhost:5174")),
		AuthRSAKeyPath: getenv("AUTH_RSA_KEY_PATH", "data/rsa_private.pem"),
		PublicSiteURL:  strings.TrimRight(getenv("PUBLIC_SITE_URL", "https://example.com"), "/"),
		IndexNowKey:    strings.TrimSpace(getenv("INDEXNOW_KEY", "")),
		MySQLDocker:    getenv("MYSQL_DOCKER", "mysql"),
		MySQL: MySQLConfig{
			Host:     getenv("MYSQL_HOST", "127.0.0.1"),
			Port:     getenvInt("MYSQL_PORT", 3306),
			User:     getenv("MYSQL_USER", "root"),
			Password: getenv("MYSQL_PASSWORD", "infra123!"),
			Database: getenv("MYSQL_DATABASE", "go_blog"),
		},
		Redis: RedisConfig{
			Host:     getenv("REDIS_HOST", "127.0.0.1"),
			Port:     getenvInt("REDIS_PORT", 6379),
			Password: getenv("REDIS_PASSWORD", "infra123!"),
			DB:       getenvInt("REDIS_DB", 0),
		},
		Token: TokenConfig{
			TTLSeconds:      getenvInt("TOKEN_TTL_SECONDS", 604800),
			CookieName:      getenv("TOKEN_COOKIE_NAME", "token"),
			CookieNameAdmin: getenv("TOKEN_COOKIE_NAME_ADMIN", "token_admin"),
			HeaderName:      getenv("TOKEN_HEADER_NAME", "X-Token"),
			CookieSecure:    strings.ToLower(getenv("TOKEN_COOKIE_SECURE", "auto")),
		},
		Storage: StorageSettings{
			DefaultEngine:        strings.ToUpper(getenv("DEFAULT_FILE_ENGINE", "MINIO")),
			UploadMaxBytes:       getenvInt64("STORAGE_UPLOAD_MAX_BYTES", 10<<20),
			PresignExpireSeconds: getenvInt("STORAGE_PRESIGN_EXPIRE_SECONDS", 3600),
			AllowedExtensions:    getenv("STORAGE_UPLOAD_ALLOWED_EXTENSIONS", ".jpg,.jpeg,.png,.gif,.webp,.pdf,.txt,.md,.zip"),
			MinIO: EngineConfig{
				AccessKey:    getenv("STORAGE_MINIO_ACCESS_KEY", "admin"),
				SecretKey:    getenv("STORAGE_MINIO_SECRET_KEY", "infra123!"),
				Endpoint:     getenv("STORAGE_MINIO_ENDPOINT", "127.0.0.1:9000"),
				Bucket:       getenv("STORAGE_MINIO_BUCKET", "go-blog"),
				Region:       getenv("STORAGE_MINIO_REGION", ""),
				UseSSL:       getenvBool("STORAGE_MINIO_USE_SSL", false),
				BaseURL:      getenv("STORAGE_MINIO_BASE_URL", "http://127.0.0.1:9000"),
				BucketPublic: getenvBool("STORAGE_MINIO_BUCKET_PUBLIC", false),
			},
			Qiniu: EngineConfig{
				AccessKey:    getenv("STORAGE_QINIU_ACCESS_KEY", ""),
				SecretKey:    getenv("STORAGE_QINIU_SECRET_KEY", ""),
				Endpoint:     getenv("STORAGE_QINIU_ENDPOINT", ""),
				Bucket:       getenv("STORAGE_QINIU_BUCKET", ""),
				Region:       getenv("STORAGE_QINIU_REGION", ""),
				// 默认 false：图床可走 HTTP CDN；有明确 scheme 的 BASE_URL 不会被改写
				UseSSL:       getenvBool("STORAGE_QINIU_USE_SSL", false),
				BaseURL:      getenv("STORAGE_QINIU_BASE_URL", ""),
				BucketPublic: getenvBool("STORAGE_QINIU_BUCKET_PUBLIC", true),
			},
		},
		Captcha: CaptchaConfig{
			TTLSeconds:      getenvInt("CAPTCHA_TTL_SECONDS", 120),
			TokenTTLSeconds: getenvInt("CAPTCHA_TOKEN_TTL_SECONDS", 300),
			TolerancePx:     getenvInt("CAPTCHA_TOLERANCE_PX", 5),
			MaxFailures:     getenvInt("CAPTCHA_MAX_FAILURES", 5),
			ImageSource:     getenv("CAPTCHA_IMAGE_SOURCE", "remote"),
			RemoteURLs:      getenv("CAPTCHA_REMOTE_URLS", "https://t.alcy.cc/ycy,https://t.alcy.cc/fj"),
			BGWidth:         getenvInt("CAPTCHA_BG_WIDTH", 320),
			BGHeight:        getenvInt("CAPTCHA_BG_HEIGHT", 160),
			PuzzleSize:      getenvInt("CAPTCHA_PUZZLE_SIZE", 60),
		},
	}
	return cfg, nil
}

func getenv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getenvInt64(key string, def int64) int64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return n
}

func getenvBool(key string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
