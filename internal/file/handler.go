package file

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/jiangbyte/go-blog/internal/pkg/page"
	"github.com/jiangbyte/go-blog/internal/pkg/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterAdmin(r gin.IRoutes) {
	r.GET("/admin/sys/file/storage/status", h.StorageStatus)
	r.POST("/admin/sys/file/upload", h.Upload)
	r.GET("/admin/sys/file/page", h.Page)
	r.GET("/admin/sys/file/detail", h.Detail)
	r.POST("/admin/sys/file/list_by_ids", h.ListByIDs)
	r.POST("/admin/sys/file/delete", h.Delete)
	r.POST("/admin/sys/file/url", h.URL)
	r.GET("/admin/sys/file/download", h.Download)
}

func (h *Handler) StorageStatus(c *gin.Context) {
	response.OK(c, h.svc.StorageStatus())
}

func (h *Handler) Upload(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请上传文件")
		return
	}
	f, err := fh.Open()
	if err != nil {
		response.BadRequest(c, "读取文件失败")
		return
	}
	defer f.Close()
	raw, err := io.ReadAll(f)
	if err != nil {
		response.BadRequest(c, "读取文件失败")
		return
	}
	out, err := h.svc.Upload(c.Request.Context(), UploadInput{
		Filename:        fh.Filename,
		Content:         raw,
		ContentType:     fh.Header.Get("Content-Type"),
		Category:        c.DefaultPostForm("category", "uploads"),
		StorageProvider: c.PostForm("storage_provider"),
		UploaderID:      strings.TrimSpace(c.GetHeader("X-Operator-Id")),
	})
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "上传失败") || strings.Contains(msg, "不允许") ||
			strings.Contains(msg, "为空") || strings.Contains(msg, "超过") {
			response.BadRequest(c, msg)
			return
		}
		response.BadRequestLog(c, err, "上传失败")
		return
	}
	response.OK(c, out)
}

func (h *Handler) Page(c *gin.Context) {
	current, size := page.Parse(c, 20)
	// 兼容 mypage page / page_size
	if v := c.Query("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			current = n
		}
	}
	if v := c.Query("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			size = page.ClampSize(n, 20)
		}
	}
	pg, err := h.svc.Page(c.Request.Context(), current, size, c.Query("storage_provider"), c.Query("keyword"))
	if err != nil {
		response.ServerErrorLog(c, err, "list files failed")
		return
	}
	response.OK(c, pg)
}

func (h *Handler) Detail(c *gin.Context) {
	id := strings.TrimSpace(c.Query("id"))
	if id == "" {
		response.BadRequest(c, "id required")
		return
	}
	out, err := h.svc.Detail(c.Request.Context(), id)
	if err != nil {
		if !replyKnownFileErr(c, err) {
			response.NotFoundLog(c, err, ErrNotFound.Error())
		}
		return
	}
	response.OK(c, out)
}

func (h *Handler) ListByIDs(c *gin.Context) {
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	out, err := h.svc.ListByIDs(c.Request.Context(), body.IDs)
	if err != nil {
		if !replyKnownFileErr(c, err) {
			response.ServerErrorLog(c, err, "list files by ids failed")
		}
		return
	}
	response.OK(c, out)
}

func (h *Handler) Delete(c *gin.Context) {
	var body struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), body.IDs); err != nil {
		if !replyKnownFileErr(c, err) {
			response.ServerErrorLog(c, err, "delete files failed")
		}
		return
	}
	response.OK(c, nil)
}

func (h *Handler) URL(c *gin.Context) {
	var body struct {
		ObjectName string `json:"object_name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.ObjectName) == "" {
		response.BadRequest(c, "object_name required")
		return
	}
	out, err := h.svc.URLByObjectName(c.Request.Context(), body.ObjectName)
	if err != nil {
		if !replyKnownFileErr(c, err) {
			response.NotFoundLog(c, err, ErrNotFound.Error())
		}
		return
	}
	response.OK(c, out)
}

func (h *Handler) Download(c *gin.Context) {
	id := strings.TrimSpace(c.Query("id"))
	if id == "" {
		response.BadRequest(c, "id required")
		return
	}
	data, row, err := h.svc.DownloadBytes(c.Request.Context(), id)
	if err != nil {
		if !replyKnownFileErr(c, err) {
			response.NotFoundLog(c, err, ErrNotFound.Error())
		}
		return
	}
	name := row.OriginalName
	if name == "" {
		name = path.Base(row.ObjectName)
	}
	if name == "" {
		name = "download"
	}
	c.Header("Content-Disposition", contentDispositionAttachment(name))
	c.Header("X-Content-Type-Options", "nosniff")
	ct := row.ContentType
	if ct == "" {
		ct = "application/octet-stream"
	}
	c.Data(http.StatusOK, ct, data)
}

// replyKnownFileErr 处理文件域已知业务错误；已写响应时返回 true。
func replyKnownFileErr(c *gin.Context, err error) bool {
	// 1. 未找到 → 404
	if errors.Is(err, ErrNotFound) {
		response.NotFound(c, ErrNotFound.Error())
		return true
	}
	// 2. ids 超限 → 400
	if errors.Is(err, ErrTooManyID) {
		response.BadRequest(c, ErrTooManyID.Error())
		return true
	}
	// 3. 其它错误交由调用方决定对外文案
	return false
}

func contentDispositionAttachment(filename string) string {
	escaped := strings.ReplaceAll(filename, `"`, `\`)
	encoded := url.PathEscape(filename)
	return fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, escaped, encoded)
}
