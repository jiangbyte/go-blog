package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/jiangbyte/go-blog/internal/auth"
	"github.com/jiangbyte/go-blog/internal/blog"
	"github.com/jiangbyte/go-blog/internal/captcha"
	"github.com/jiangbyte/go-blog/internal/config"
	"github.com/jiangbyte/go-blog/internal/file"
	"github.com/jiangbyte/go-blog/internal/pkg/response"
	"github.com/jiangbyte/go-blog/internal/resume"
	"github.com/jiangbyte/go-blog/internal/storage"
)

// New 组装 Gin 引擎：中间件、依赖注入与门户/管理路由注册。
func New(cfg *config.Config, db *gorm.DB, rdb *redis.Client) (*gin.Engine, error) {
	// 1. 基础引擎与 CORS / 可信代理
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
	_ = r.SetTrustedProxies([]string{"127.0.0.1", "::1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"})
	r.Use(corsMiddleware(cfg.CORSOrigins))

	// 2. 构造认证、验证码、对象存储与博客服务
	rsaCipher, err := auth.NewRSACipher(cfg.AuthRSAKeyPath)
	if err != nil {
		return nil, err
	}
	sess := auth.NewSessionStore(rdb, cfg.Token.TTLSeconds)
	captchaSvc := captcha.NewService(rdb, cfg.Captcha)

	storeReg := storage.NewRegistry(cfg.Storage)
	fileSvc := file.NewService(db, storeReg)
	if err := fileSvc.EnsureSchema(context.Background()); err != nil {
		log.Printf("sys_file migrate: %v", err)
	}

	authH := auth.NewHandler(db, sess, rsaCipher, cfg, captchaSvc, fileSvc)
	// PV/UV / 阅读：Redis 热写，异步落库 MySQL
	blogSvc := blog.NewService(db, rdb, fileSvc)
	blogSvc.SetPublicSiteURL(cfg.PublicSiteURL)
	blogSvc.SetIndexNowKey(cfg.IndexNowKey)
	blogSvc.StartStatsSync(context.Background())
	blogH := blog.NewHandler(blogSvc)
	fileH := file.NewHandler(fileSvc)
	resumeSvc := resume.NewService(db, fileSvc)
	if err := resumeSvc.EnsureSchema(context.Background()); err != nil {
		log.Printf("resume migrate: %v", err)
	}
	resumeH := resume.NewHandler(resumeSvc)

	// 3. 全局会话中间件 + 公开路由
	r.Use(authH.Middleware())

	r.GET("/health", func(c *gin.Context) {
		response.OK(c, gin.H{"status": "up", "app": cfg.AppName})
	})

	_ = os.MkdirAll("data/uploads", 0o755)
	r.Static("/uploads", "data/uploads")

	authH.Register(r)
	blogH.Register(r)
	resumeH.Register(r)

	// 4. 管理端路由组（RequireAdmin）
	admin := r.Group("/")
	admin.Use(auth.RequireAdmin())
	authH.RegisterAdmin(admin)
	blogH.RegisterAdmin(admin)
	fileH.RegisterAdmin(admin)
	resumeH.RegisterAdmin(admin)

	r.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "not found")
	})
	return r, nil
}

func corsMiddleware(allowed []string) gin.HandlerFunc {
	allow := make(map[string]struct{}, len(allowed))
	for _, o := range allowed {
		o = strings.TrimSpace(o)
		if o != "" {
			allow[o] = struct{}{}
		}
	}
	return func(c *gin.Context) {
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		if origin != "" {
			if _, ok := allow[origin]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Allow-Headers", "Content-Type, X-Token, Authorization")
				c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				c.Header("Vary", "Origin")
			}
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
