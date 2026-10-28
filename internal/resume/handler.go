package resume

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/jiangbyte/go-blog/internal/pkg/page"
	"github.com/jiangbyte/go-blog/internal/pkg/response"
)

// Handler 简历 HTTP 入口。
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register 公开 Portal 路由。
func (h *Handler) Register(r gin.IRoutes) {
	r.GET("/portal/resume/resume", h.PublicResume)
	r.GET("/portal/resume/resume/html", h.PublicResumeHTML)
	r.GET("/portal/resume/resume/pdf", h.PublicResumePDF)
}

// RegisterAdmin 管理端路由（需 RequireAdmin）。
func (h *Handler) RegisterAdmin(r gin.IRoutes) {
	r.GET("/admin/resume/resume/page", h.Page)
	r.GET("/admin/resume/resume/detail", h.Detail)
	r.POST("/admin/resume/resume/create", h.Create)
	r.POST("/admin/resume/resume/update", h.Update)
	r.POST("/admin/resume/resume/remove", h.Remove)
	r.POST("/admin/resume/resume/publish", h.Publish)
	r.POST("/admin/resume/resume/unpublish", h.Unpublish)
	r.POST("/admin/resume/resume/copy", h.Copy)
	r.GET("/admin/resume/resume/templates", h.Templates)
	r.GET("/admin/resume/resume/icons", h.Icons)
	r.POST("/admin/resume/resume/section/save", h.SaveSection)
	r.POST("/admin/resume/resume/section/copy", h.CopySection)
	r.POST("/admin/resume/resume/section/reorder", h.ReorderSections)
	r.POST("/admin/resume/resume/section/remove", h.RemoveSections)
	r.POST("/admin/resume/resume/render-html", h.RenderHTML)
	r.POST("/admin/resume/resume/render-pdf", h.RenderPDF)
	r.GET("/admin/resume/resume/export-pdf", h.ExportPDF)

	r.GET("/portal/resume/resume/preview", h.Preview)
	r.GET("/portal/resume/resume/preview/html", h.PreviewHTML)
	r.GET("/portal/resume/resume/preview/pdf", h.PreviewPDF)
}

func (h *Handler) writeErr(c *gin.Context, err error) {
	if IsNotFound(err) {
		response.NotFound(c, err.Error())
		return
	}
	if IsBadRequest(err) {
		response.BadRequest(c, err.Error())
		return
	}
	response.ServerErrorLog(c, err, "resume operation failed")
}

func (h *Handler) Page(c *gin.Context) {
	cur, size := page.Parse(c, 20)
	data, err := h.svc.Page(c.Request.Context(), cur, size, c.Query("status"), c.Query("title"))
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.OK(c, data)
}

func (h *Handler) Detail(c *gin.Context) {
	data, err := h.svc.Detail(c.Request.Context(), c.Query("id"))
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.OK(c, data)
}

func (h *Handler) Create(c *gin.Context) {
	var body ResumeWriteInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	data, err := h.svc.Create(c.Request.Context(), body)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.OK(c, data)
}

func (h *Handler) Update(c *gin.Context) {
	var body ResumeWriteInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	data, err := h.svc.Update(c.Request.Context(), body)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.OK(c, data)
}

func (h *Handler) Remove(c *gin.Context) {
	var body idsBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	if err := h.svc.Remove(c.Request.Context(), body.IDs); err != nil {
		h.writeErr(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *Handler) Publish(c *gin.Context) {
	var body idBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	data, err := h.svc.Publish(c.Request.Context(), body.ID)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.OK(c, data)
}

func (h *Handler) Unpublish(c *gin.Context) {
	var body idBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	data, err := h.svc.Unpublish(c.Request.Context(), body.ID)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.OK(c, data)
}

func (h *Handler) Copy(c *gin.Context) {
	var body idBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	data, err := h.svc.Copy(c.Request.Context(), body.ID)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.OK(c, data)
}

func (h *Handler) Templates(c *gin.Context) {
	response.OK(c, ListTemplateMetas())
}

func (h *Handler) Icons(c *gin.Context) {
	cur, size := page.Parse(c, 100)
	data, err := h.svc.PageIcons(c.Query("prefix"), cur, size, c.Query("q"))
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.OK(c, data)
}

func (h *Handler) SaveSection(c *gin.Context) {
	var body SectionSaveInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	data, err := h.svc.SaveSection(c.Request.Context(), body)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.OK(c, data)
}

func (h *Handler) CopySection(c *gin.Context) {
	var body sectionCopyBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	data, err := h.svc.CopySection(c.Request.Context(), body.ResumeID, body.SectionID)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.OK(c, data)
}

func (h *Handler) ReorderSections(c *gin.Context) {
	var body sectionReorderBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	data, err := h.svc.ReorderSections(c.Request.Context(), body.ResumeID, body.IDs)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.OK(c, data)
}

func (h *Handler) RemoveSections(c *gin.Context) {
	var body sectionRemoveBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	if err := h.svc.RemoveSections(c.Request.Context(), body.ResumeID, body.IDs); err != nil {
		h.writeErr(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *Handler) RenderHTML(c *gin.Context) {
	var body renderHTMLBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	css := ""
	if body.CustomCSS != nil {
		css = *body.CustomCSS
	}
	html, err := h.svc.RenderHTMLDoc(c.Request.Context(), map[string]any{
		"title":        body.Title,
		"template_id":  body.TemplateID,
		"theme_json":   body.ThemeJSON,
		"custom_css":   css,
		"profile_json": body.ProfileJSON,
		"sections":     body.Sections,
	})
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.OK(c, gin.H{"html": html})
}

// RenderPDF 用与导出相同的 chromedp 管线，供编辑态分页预览（与下载 PDF 分页一致）。
func (h *Handler) RenderPDF(c *gin.Context) {
	// 1. 解析与 render-html 相同的文档 body
	var body renderHTMLBody
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	css := ""
	if body.CustomCSS != nil {
		css = *body.CustomCSS
	}
	// 2. 渲染 PDF 并直接写二进制（非 JSON 信封，便于 iframe 打开）
	pdf, err := h.svc.RenderPDFDoc(c.Request.Context(), map[string]any{
		"title":        body.Title,
		"template_id":  body.TemplateID,
		"theme_json":   body.ThemeJSON,
		"custom_css":   css,
		"profile_json": body.ProfileJSON,
		"sections":     body.Sections,
	})
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, "application/pdf", pdf)
}

func (h *Handler) ExportPDF(c *gin.Context) {
	id := strings.TrimSpace(c.Query("id"))
	if id == "" {
		response.BadRequest(c, "id required")
		return
	}
	pdf, filename, err := h.svc.ExportPDF(c.Request.Context(), id)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	disp := "attachment; filename*=UTF-8''" + url.PathEscape(filename)
	c.Header("Content-Disposition", disp)
	c.Data(http.StatusOK, "application/pdf", pdf)
}

func (h *Handler) PublicResume(c *gin.Context) {
	data, err := h.svc.PublicPublished(c.Request.Context())
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.OK(c, data)
}

func (h *Handler) PublicResumeHTML(c *gin.Context) {
	doc, err := h.svc.PublicPublished(c.Request.Context())
	if err != nil {
		h.writeErr(c, err)
		return
	}
	html, err := h.svc.RenderHTMLDoc(c.Request.Context(), doc)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.OK(c, gin.H{"html": html})
}

// PublicResumePDF 公开简历 PDF。
// - 默认 302 → OSS（主站外链直达）
// - ?stream=1 同源流式输出（本地 vite / resume 子域预览，规避图床 CORS）
func (h *Handler) PublicResumePDF(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	if strings.TrimSpace(c.Query("stream")) == "1" {
		pdf, filename, err := h.svc.PublicPDFBytes(c.Request.Context())
		if err != nil {
			h.writeErr(c, err)
			return
		}
		disp := "inline; filename*=UTF-8''" + url.PathEscape(filename)
		c.Header("Content-Disposition", disp)
		c.Data(http.StatusOK, "application/pdf", pdf)
		return
	}
	u, _, err := h.svc.PublicPDFAccessURL(c.Request.Context())
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.Redirect(http.StatusFound, u)
}

func (h *Handler) Preview(c *gin.Context) {
	data, err := h.svc.Detail(c.Request.Context(), c.Query("id"))
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.OK(c, data)
}

func (h *Handler) PreviewHTML(c *gin.Context) {
	doc, err := h.svc.Detail(c.Request.Context(), c.Query("id"))
	if err != nil {
		h.writeErr(c, err)
		return
	}
	html, err := h.svc.RenderHTMLDoc(c.Request.Context(), doc)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	response.OK(c, gin.H{"html": html})
}

// PreviewPDF 管理端预览用：按 id 现场生成，不写入公开缓存。
func (h *Handler) PreviewPDF(c *gin.Context) {
	id := strings.TrimSpace(c.Query("id"))
	if id == "" {
		response.BadRequest(c, "id required")
		return
	}
	doc, err := h.svc.Detail(c.Request.Context(), id)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	pdf, err := h.svc.RenderPDFDoc(c.Request.Context(), doc)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	disp := "inline; filename*=UTF-8''" + url.PathEscape(SafePDFFilename(doc))
	c.Header("Content-Disposition", disp)
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, "application/pdf", pdf)
}
