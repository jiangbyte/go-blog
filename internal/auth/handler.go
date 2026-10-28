package auth

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/jiangbyte/go-blog/internal/captcha"
	"github.com/jiangbyte/go-blog/internal/config"
	"github.com/jiangbyte/go-blog/internal/file"
	"github.com/jiangbyte/go-blog/internal/model"
	"github.com/jiangbyte/go-blog/internal/pkg/response"
)

type Handler struct {
	db      *gorm.DB
	sess    *SessionStore
	rsa     *RSACipher
	cfg     *config.Config
	captcha *captcha.Service
	files   *file.Service
}

func NewHandler(db *gorm.DB, sess *SessionStore, rsa *RSACipher, cfg *config.Config, captchaSvc *captcha.Service, files *file.Service) *Handler {
	return &Handler{db: db, sess: sess, rsa: rsa, cfg: cfg, captcha: captchaSvc, files: files}
}

func (h *Handler) Register(r gin.IRoutes) {
	r.GET("/auth/public-key", h.PublicKey)
	r.GET("/auth/captcha/create", h.CreateCaptcha)
	r.POST("/auth/captcha/verify", h.VerifyCaptcha)
	r.POST("/auth/login", h.Login)
	r.POST("/auth/logout", h.Logout)
	r.GET("/auth/me", h.Me)
	h.registerProfile(r)
}

func (h *Handler) PublicKey(c *gin.Context) {
	response.OK(c, gin.H{"public_key": h.rsa.PublicKeyPEM()})
}

// CreateCaptcha 对齐 mypage：远端底图切拼图，答案只存 Redis 不回传 x。
func (h *Handler) CreateCaptcha(c *gin.Context) {
	res, err := h.captcha.Create(c.Request.Context())
	if err != nil {
		response.ServerError(c, "captcha generate failed")
		return
	}
	response.OK(c, gin.H{
		"captcha_id":  res.CaptchaID,
		"bg_url":      res.BGURL,
		"puzzle_url":  res.PuzzleURL,
		"y":           res.Y,
		"bg_width":    res.BGWidth,
		"bg_height":   res.BGHeight,
		"puzzle_size": res.PuzzleSize,
	})
}

func (h *Handler) VerifyCaptcha(c *gin.Context) {
	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "验证码已失效，请刷新后重试")
		return
	}
	// captcha-pro 用 camelCase；本站约定 snake_case，两边都收
	id := anyString(body["captcha_id"])
	if id == "" {
		id = anyString(body["captchaId"])
	}
	x := anyFloat(body["x"])
	token, err := h.captcha.Verify(c.Request.Context(), id, x)
	if err != nil {
		// 验证码业务错误为明确中文
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"captcha_token": token})
}

func anyString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatInt(int64(t), 10)
	default:
		return ""
	}
}

func anyFloat(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case string:
		f, _ := strconv.ParseFloat(t, 64)
		return f
	default:
		return 0
	}
}

// Login 管理端登录：消费滑块 token、解密密码、校验账号并签发会话 Cookie。
func (h *Handler) Login(c *gin.Context) {
	// 1. 解析并校验必填字段
	var body struct {
		Username       string `json:"username"`
		PasswordCipher string `json:"password_cipher"`
		CaptchaToken   string `json:"captcha_token"`
		ClientType     string `json:"client_type"`
		ReturnToken    bool   `json:"return_token"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	if body.Username == "" || body.PasswordCipher == "" {
		response.BadRequest(c, "username/password required")
		return
	}
	// 2. 一次性消费滑块 token，并 RSA 解密密码密文
	if err := h.captcha.ConsumeToken(c.Request.Context(), body.CaptchaToken); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	plain, err := h.rsa.DecryptBase64(body.PasswordCipher)
	if err != nil {
		response.BadRequestLog(c, err, "密码解密失败")
		return
	}

	// 3. 查用户：不存在/禁用/类型不对/密码错统一模糊提示，避免枚举
	var user model.AdminUser
	err = h.db.WithContext(c.Request.Context()).
		Where("username = ?", body.Username).
		First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Unauthorized(c, "账号或密码错误")
		return
	}
	if err != nil {
		response.ServerError(c, "login failed")
		return
	}
	if user.Enabled != 1 || user.UserType != "ADMIN" {
		response.Unauthorized(c, "账号不可用")
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(plain)) != nil {
		response.Unauthorized(c, "账号或密码错误")
		return
	}

	// 4. 建会话并写 Cookie；CLI 标记时额外在 JSON 回传 token（网页默认仍不回传）
	sessUser := sessionUserFromModel(&user)
	sessUser.Avatar = h.resolveAvatar(c, sessUser.Avatar)
	token, err := h.sess.Create(c.Request.Context(), sessUser)
	if err != nil {
		response.ServerError(c, "session create failed")
		return
	}
	h.setCookies(c, token)
	data := gin.H{
		"profile": h.profileMap(sessUser),
	}
	if body.ReturnToken || strings.TrimSpace(c.GetHeader("X-Go-Blog-CLI")) == "1" {
		data["token"] = token
	}
	response.OK(c, data)
}

func (h *Handler) Logout(c *gin.Context) {
	token := h.extractToken(c)
	if token != "" {
		_ = h.sess.Delete(c.Request.Context(), token)
	}
	h.clearCookies(c)
	response.OK(c, nil)
}

func (h *Handler) Me(c *gin.Context) {
	sess, ok := UserFromContext(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	user, err := h.loadAdminUser(c, sess.ID)
	if err != nil {
		response.Unauthorized(c, "未登录")
		return
	}
	h.syncSession(c, user)
	response.OK(c, h.profileMapFromUser(c, user))
}

func (h *Handler) profileMap(u SessionUser) gin.H {
	gender := u.Gender
	if gender == "" {
		gender = "UNKNOWN"
	}
	var signature any
	if strings.TrimSpace(u.Signature) == "" {
		signature = nil
	} else {
		signature = u.Signature
	}
	avatar := u.Avatar
	return gin.H{
		"id":              u.ID,
		"account_id":      u.ID,
		"username":        u.Username,
		"account":         u.Username,
		"nickname":        u.Nickname,
		"avatar":          avatar,
		"user_type":       u.UserType,
		"account_type":    u.UserType,
		"gender":          gender,
		"signature":       signature,
		"permission_keys": []string{"admin:console:access"},
		"permissions":     []string{"admin:console:access"},
	}
}

func (h *Handler) resolveAvatar(c *gin.Context, raw string) string {
	if h.files == nil || strings.TrimSpace(raw) == "" {
		return raw
	}
	if u := h.files.ResolveAccessURL(c.Request.Context(), raw); u != "" {
		return u
	}
	return raw
}

func (h *Handler) extractToken(c *gin.Context) string {
	if t := c.GetHeader(h.cfg.Token.HeaderName); t != "" {
		return t
	}
	if t, err := c.Cookie(h.cfg.Token.CookieNameAdmin); err == nil && t != "" {
		return t
	}
	if t, err := c.Cookie(h.cfg.Token.CookieName); err == nil && t != "" {
		return t
	}
	return ""
}

func (h *Handler) cookieSecure(c *gin.Context) bool {
	switch strings.ToLower(strings.TrimSpace(h.cfg.Token.CookieSecure)) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default: // auto
		if c.Request.TLS != nil {
			return true
		}
		return strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https")
	}
}

func (h *Handler) setCookies(c *gin.Context, token string) {
	maxAge := h.cfg.Token.TTLSeconds
	secure := h.cookieSecure(c)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(h.cfg.Token.CookieName, token, maxAge, "/", "", secure, true)
	c.SetCookie(h.cfg.Token.CookieNameAdmin, token, maxAge, "/", "", secure, true)
}

func (h *Handler) clearCookies(c *gin.Context) {
	secure := h.cookieSecure(c)
	c.SetCookie(h.cfg.Token.CookieName, "", -1, "/", "", secure, true)
	c.SetCookie(h.cfg.Token.CookieNameAdmin, "", -1, "/", "", secure, true)
}

func (h *Handler) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := h.extractToken(c)
		if token == "" {
			c.Next()
			return
		}
		user, err := h.sess.Get(c.Request.Context(), token)
		if err == nil && user != nil {
			c.Set(ctxUserKey, user)
			c.Set(ctxTokenKey, token)
		}
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := UserFromContext(c)
		if !ok || user.UserType != "ADMIN" {
			response.Unauthorized(c, "需要管理员登录")
			c.Abort()
			return
		}
		c.Next()
	}
}

const ctxUserKey = "auth_user"
const ctxTokenKey = "auth_token"

func UserFromContext(c *gin.Context) (*SessionUser, bool) {
	v, ok := c.Get(ctxUserKey)
	if !ok {
		return nil, false
	}
	u, ok := v.(*SessionUser)
	return u, ok
}
