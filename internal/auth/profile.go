package auth

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/jiangbyte/go-blog/internal/file"
	"github.com/jiangbyte/go-blog/internal/model"
	"github.com/jiangbyte/go-blog/internal/pkg/response"
)

const avatarMaxBytes = 2 << 20 // 2 MiB

var allowedGenders = map[string]bool{
	"UNKNOWN": true,
	"MALE":    true,
	"FEMALE":  true,
}

func (h *Handler) registerProfile(r gin.IRoutes) {
	r.POST("/auth/profile/update", h.UpdateProfile)
	r.POST("/auth/profile/password/update", h.UpdatePassword)
	r.POST("/auth/profile/avatar/upload", h.UploadAvatar)
}

// UpdateProfile updates nickname / gender / signature for the current admin.
func (h *Handler) UpdateProfile(c *gin.Context) {
	sess, ok := UserFromContext(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	var body struct {
		Nickname  string  `json:"nickname"`
		Gender    string  `json:"gender"`
		Signature *string `json:"signature"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	nickname := strings.TrimSpace(body.Nickname)
	if utf8.RuneCountInString(nickname) > 64 {
		response.BadRequest(c, "昵称过长")
		return
	}
	gender := strings.ToUpper(strings.TrimSpace(body.Gender))
	if gender == "" {
		gender = "UNKNOWN"
	}
	if !allowedGenders[gender] {
		response.BadRequest(c, "无效的性别")
		return
	}
	updates := map[string]any{
		"nickname": nickname,
		"gender":   gender,
	}
	if body.Signature != nil {
		sig := strings.TrimSpace(*body.Signature)
		if utf8.RuneCountInString(sig) > 255 {
			response.BadRequest(c, "签名过长")
			return
		}
		updates["signature"] = sig
	}
	if err := h.db.WithContext(c.Request.Context()).
		Model(&model.AdminUser{}).
		Where("id = ?", sess.ID).
		Updates(updates).Error; err != nil {
		response.ServerError(c, "更新失败")
		return
	}
	user, err := h.loadAdminUser(c, sess.ID)
	if err != nil {
		response.ServerError(c, "更新失败")
		return
	}
	h.syncSession(c, user)
	response.OK(c, h.profileMapFromUser(c, user))
}

func (h *Handler) UpdatePassword(c *gin.Context) {
	sess, ok := UserFromContext(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	var body struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	if strings.TrimSpace(body.OldPassword) == "" || strings.TrimSpace(body.NewPassword) == "" {
		response.BadRequest(c, "请填写旧密码与新密码")
		return
	}
	oldPwd, err := h.rsa.DecryptBase64(body.OldPassword)
	if err != nil {
		response.BadRequestLog(c, err, "密码解密失败")
		return
	}
	newPwd, err := h.rsa.DecryptBase64(body.NewPassword)
	if err != nil {
		response.BadRequestLog(c, err, "密码解密失败")
		return
	}
	if len(newPwd) < 6 {
		response.BadRequest(c, "新密码至少 6 位")
		return
	}
	if len(newPwd) > 72 {
		response.BadRequest(c, "新密码过长")
		return
	}
	user, err := h.loadAdminUser(c, sess.ID)
	if err != nil {
		response.ServerError(c, "改密失败")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPwd)); err != nil {
		response.BadRequest(c, "旧密码不正确")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
	if err != nil {
		response.ServerError(c, "改密失败")
		return
	}
	if err := h.db.WithContext(c.Request.Context()).
		Model(&model.AdminUser{}).
		Where("id = ?", sess.ID).
		Update("password_hash", string(hash)).Error; err != nil {
		response.ServerError(c, "改密失败")
		return
	}
	response.OK(c, nil)
}

// UploadAvatar 裁剪后的头像走对象存储（默认引擎），avatar 字段存 object_name。
func (h *Handler) UploadAvatar(c *gin.Context) {
	sess, ok := UserFromContext(c)
	if !ok {
		response.Unauthorized(c, "未登录")
		return
	}
	if h.files == nil {
		response.ServerError(c, "对象存储未就绪")
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传头像文件")
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		response.BadRequest(c, "读取头像失败")
		return
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, avatarMaxBytes+1))
	if err != nil {
		response.BadRequest(c, "读取头像失败")
		return
	}
	if len(raw) == 0 {
		response.BadRequest(c, "空文件")
		return
	}
	if len(raw) > avatarMaxBytes {
		response.BadRequest(c, "头像不能超过 2MB")
		return
	}
	ct := http.DetectContentType(raw)
	ext := ""
	switch ct {
	case "image/png":
		ext = ".png"
	case "image/jpeg":
		ext = ".jpg"
	case "image/webp":
		ext = ".webp"
	default:
		response.BadRequest(c, "仅支持 JPG、PNG、WebP")
		return
	}
	filename := fmt.Sprintf("avatar%s", ext)
	out, err := h.files.Upload(c.Request.Context(), file.UploadInput{
		Filename:    filename,
		Content:     raw,
		ContentType: ct,
		Category:    "avatars",
		UploaderID:  strconv.FormatInt(sess.ID, 10),
	})
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "不允许") || strings.Contains(msg, "为空") || strings.Contains(msg, "超过") {
			response.BadRequest(c, msg)
			return
		}
		response.BadRequestLog(c, err, "上传头像失败")
		return
	}
	objectName, _ := out["object_name"].(string)
	if objectName == "" {
		response.ServerError(c, "保存头像失败")
		return
	}
	if err := h.db.WithContext(c.Request.Context()).
		Model(&model.AdminUser{}).
		Where("id = ?", sess.ID).
		Update("avatar", objectName).Error; err != nil {
		response.ServerError(c, "保存头像失败")
		return
	}
	user, err := h.loadAdminUser(c, sess.ID)
	if err != nil {
		response.ServerError(c, "保存头像失败")
		return
	}
	h.syncSession(c, user)
	response.OK(c, h.profileMapFromUser(c, user))
}

func (h *Handler) loadAdminUser(c *gin.Context, id int64) (*model.AdminUser, error) {
	var user model.AdminUser
	err := h.db.WithContext(c.Request.Context()).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (h *Handler) syncSession(c *gin.Context, user *model.AdminUser) {
	token := h.extractToken(c)
	if token == "" {
		return
	}
	su := sessionUserFromModel(user)
	su.Avatar = h.resolveAvatar(c, su.Avatar)
	_ = h.sess.Update(c.Request.Context(), token, su)
}

func sessionUserFromModel(u *model.AdminUser) SessionUser {
	gender := u.Gender
	if gender == "" {
		gender = "UNKNOWN"
	}
	return SessionUser{
		ID: u.ID, Username: u.Username, Nickname: u.Nickname,
		UserType: u.UserType, Avatar: u.Avatar,
		Gender: gender, Signature: u.Signature,
	}
}

func (h *Handler) profileMapFromUser(c *gin.Context, u *model.AdminUser) gin.H {
	su := sessionUserFromModel(u)
	su.Avatar = h.resolveAvatar(c, su.Avatar)
	return h.profileMap(su)
}
