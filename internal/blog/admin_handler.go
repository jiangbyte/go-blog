package blog

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/jiangbyte/go-blog/internal/pkg/page"
	"github.com/jiangbyte/go-blog/internal/pkg/response"
)

// RegisterAdmin 管理端博客 CRUD（需 RequireAdmin）。
func (h *Handler) RegisterAdmin(r gin.IRoutes) {
	r.GET("/admin/sys/stats/overview", h.AdminStatsOverview)

	r.GET("/admin/blog/posts/page", h.AdminPostsPage)
	r.GET("/admin/blog/posts/detail", h.AdminPostDetail)
	r.GET("/admin/blog/posts/by-title", h.AdminPostsByTitle)
	r.POST("/admin/blog/posts/create", h.AdminPostCreate)
	r.POST("/admin/blog/posts/update", h.AdminPostUpdate)
	r.POST("/admin/blog/posts/batch-status", h.AdminPostBatchStatus)
	r.POST("/admin/blog/posts/remove", h.AdminPostRemove)

	r.GET("/admin/blog/projects/page", h.AdminProjectsPage)
	r.GET("/admin/blog/projects/detail", h.AdminProjectDetail)
	r.POST("/admin/blog/projects/create", h.AdminProjectCreate)
	r.POST("/admin/blog/projects/update", h.AdminProjectUpdate)
	r.POST("/admin/blog/projects/batch-status", h.AdminProjectBatchStatus)
	r.POST("/admin/blog/projects/remove", h.AdminProjectRemove)

	r.GET("/admin/blog/site", h.AdminSiteGet)
	r.POST("/admin/blog/site", h.AdminSiteUpdate)
	r.GET("/admin/blog/about", h.AdminAboutGet)
	r.POST("/admin/blog/about", h.AdminAboutUpdate)
	r.GET("/admin/blog/taxonomy/options", h.AdminTaxonomyOptions)
	r.GET("/admin/blog/taxonomy/page", h.AdminTaxonomyPage)
	r.POST("/admin/blog/taxonomy/create", h.AdminTaxonomyCreate)
	r.POST("/admin/blog/taxonomy/update", h.AdminTaxonomyUpdate)
	r.POST("/admin/blog/taxonomy/remove", h.AdminTaxonomyRemove)
}

func (h *Handler) AdminPostsPage(c *gin.Context) {
	current, size := page.Parse(c, 20)
	data, err := h.svc.AdminPagePosts(c.Request.Context(), current, size, c.Query("title"), c.Query("status"))
	if err != nil {
		response.ServerError(c, "list posts failed")
		return
	}
	response.OK(c, data)
}

func (h *Handler) AdminPostDetail(c *gin.Context) {
	uid := c.Query("uid")
	if uid == "" {
		response.BadRequest(c, "uid required")
		return
	}
	post, err := h.svc.AdminGetPost(c.Request.Context(), uid)
	if IsNotFound(err) {
		response.NotFound(c, "post not found")
		return
	}
	if err != nil {
		response.ServerError(c, "get post failed")
		return
	}
	response.OK(c, post)
}

func (h *Handler) AdminPostsByTitle(c *gin.Context) {
	title := c.Query("title")
	list, err := h.svc.AdminFindPostsByTitle(c.Request.Context(), title, c.Query("status"))
	if errors.Is(err, gorm.ErrInvalidData) {
		response.BadRequest(c, "title required")
		return
	}
	if err != nil {
		response.ServerError(c, "list posts by title failed")
		return
	}
	response.OK(c, list)
}

func (h *Handler) AdminPostCreate(c *gin.Context) {
	var body PostWriteInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	post, err := h.svc.AdminCreatePost(c.Request.Context(), body)
	if errors.Is(err, gorm.ErrInvalidData) {
		response.BadRequest(c, "标题不能为空")
		return
	}
	if err != nil {
		response.ServerError(c, "create post failed")
		return
	}
	response.OK(c, post)
}

func (h *Handler) AdminPostUpdate(c *gin.Context) {
	var body PostWriteInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	post, err := h.svc.AdminUpdatePost(c.Request.Context(), body)
	if IsNotFound(err) {
		response.NotFound(c, "post not found")
		return
	}
	if errors.Is(err, gorm.ErrInvalidData) {
		response.BadRequest(c, "参数无效")
		return
	}
	if err != nil {
		response.ServerError(c, "update post failed")
		return
	}
	response.OK(c, post)
}

func (h *Handler) AdminPostRemove(c *gin.Context) {
	var body struct {
		UIDs []string `json:"uids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	if err := h.svc.AdminRemovePosts(c.Request.Context(), body.UIDs); err != nil {
		response.ServerError(c, "remove posts failed")
		return
	}
	response.OK(c, nil)
}

func (h *Handler) AdminPostBatchStatus(c *gin.Context) {
	var body struct {
		UIDs   []string `json:"uids"`
		Status string   `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	data, err := h.svc.AdminBatchSetPostStatus(c.Request.Context(), body.UIDs, body.Status)
	if errors.Is(err, gorm.ErrInvalidData) {
		response.BadRequest(c, "status 须为 draft 或 published")
		return
	}
	if err != nil {
		response.ServerError(c, "batch update post status failed")
		return
	}
	response.OK(c, data)
}

func (h *Handler) AdminProjectsPage(c *gin.Context) {
	current, size := page.Parse(c, 20)
	data, err := h.svc.AdminPageProjects(
		c.Request.Context(), current, size, c.Query("title"), c.Query("status"), c.Query("progress"),
	)
	if err != nil {
		response.ServerError(c, "list projects failed")
		return
	}
	response.OK(c, data)
}

func (h *Handler) AdminProjectDetail(c *gin.Context) {
	uid := c.Query("uid")
	if uid == "" {
		response.BadRequest(c, "uid required")
		return
	}
	project, err := h.svc.AdminGetProject(c.Request.Context(), uid)
	if IsNotFound(err) {
		response.NotFound(c, "project not found")
		return
	}
	if err != nil {
		response.ServerError(c, "get project failed")
		return
	}
	response.OK(c, project)
}

func (h *Handler) AdminProjectCreate(c *gin.Context) {
	var body ProjectWriteInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	project, err := h.svc.AdminCreateProject(c.Request.Context(), body)
	if errors.Is(err, gorm.ErrInvalidData) {
		response.BadRequest(c, "名称不能为空")
		return
	}
	if err != nil {
		response.ServerError(c, "create project failed")
		return
	}
	response.OK(c, project)
}

func (h *Handler) AdminProjectUpdate(c *gin.Context) {
	var body ProjectWriteInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	project, err := h.svc.AdminUpdateProject(c.Request.Context(), body)
	if IsNotFound(err) {
		response.NotFound(c, "project not found")
		return
	}
	if errors.Is(err, gorm.ErrInvalidData) {
		response.BadRequest(c, "参数无效")
		return
	}
	if err != nil {
		response.ServerError(c, "update project failed")
		return
	}
	response.OK(c, project)
}

func (h *Handler) AdminProjectRemove(c *gin.Context) {
	var body struct {
		UIDs []string `json:"uids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	if err := h.svc.AdminRemoveProjects(c.Request.Context(), body.UIDs); err != nil {
		response.ServerError(c, "remove projects failed")
		return
	}
	response.OK(c, nil)
}

func (h *Handler) AdminProjectBatchStatus(c *gin.Context) {
	var body struct {
		UIDs   []string `json:"uids"`
		Status string   `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	data, err := h.svc.AdminBatchSetProjectStatus(c.Request.Context(), body.UIDs, body.Status)
	if errors.Is(err, gorm.ErrInvalidData) {
		response.BadRequest(c, "status 须为 draft 或 published")
		return
	}
	if err != nil {
		response.ServerError(c, "batch update project status failed")
		return
	}
	response.OK(c, data)
}

func (h *Handler) AdminSiteGet(c *gin.Context) {
	data, err := h.svc.AdminGetSite(c.Request.Context())
	if err != nil {
		response.ServerError(c, "get site failed")
		return
	}
	response.OK(c, data)
}

func (h *Handler) AdminSiteUpdate(c *gin.Context) {
	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	data, err := h.svc.AdminUpdateSite(c.Request.Context(), body)
	if err != nil {
		response.ServerError(c, "update site failed")
		return
	}
	response.OK(c, data)
}

func (h *Handler) AdminAboutGet(c *gin.Context) {
	data, err := h.svc.AdminGetAbout(c.Request.Context())
	if err != nil {
		response.ServerError(c, "get about failed")
		return
	}
	response.OK(c, data)
}

func (h *Handler) AdminAboutUpdate(c *gin.Context) {
	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	data, err := h.svc.AdminUpdateAbout(c.Request.Context(), body)
	if err != nil {
		response.ServerError(c, "update about failed")
		return
	}
	response.OK(c, data)
}

func (h *Handler) AdminTaxonomyOptions(c *gin.Context) {
	list, err := h.svc.AdminTaxonomyOptions(c.Request.Context(), c.Query("kind"))
	if errors.Is(err, gorm.ErrInvalidData) {
		response.BadRequest(c, "kind required")
		return
	}
	if err != nil {
		response.ServerError(c, "list taxonomy failed")
		return
	}
	response.OK(c, list)
}

func (h *Handler) AdminTaxonomyPage(c *gin.Context) {
	current, size := page.Parse(c, 20)
	data, err := h.svc.AdminPageTaxonomy(c.Request.Context(), c.Query("kind"), c.Query("name"), current, size)
	if errors.Is(err, gorm.ErrInvalidData) {
		response.BadRequest(c, "kind required")
		return
	}
	if err != nil {
		response.ServerError(c, "list taxonomy failed")
		return
	}
	response.OK(c, data)
}

func (h *Handler) AdminTaxonomyCreate(c *gin.Context) {
	var body TaxonomyWriteInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	row, err := h.svc.AdminCreateTaxonomy(c.Request.Context(), body)
	if errors.Is(err, gorm.ErrInvalidData) {
		response.BadRequest(c, "名称或类型无效")
		return
	}
	if err != nil {
		response.BadRequest(c, "创建失败，名称可能已存在")
		return
	}
	response.OK(c, row)
}

func (h *Handler) AdminTaxonomyUpdate(c *gin.Context) {
	var body TaxonomyWriteInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	row, err := h.svc.AdminUpdateTaxonomy(c.Request.Context(), body)
	if IsNotFound(err) {
		response.NotFound(c, "记录不存在")
		return
	}
	if errors.Is(err, gorm.ErrInvalidData) {
		response.BadRequest(c, "参数无效")
		return
	}
	if err != nil {
		response.BadRequest(c, "更新失败，名称可能已存在")
		return
	}
	response.OK(c, row)
}

func (h *Handler) AdminTaxonomyRemove(c *gin.Context) {
	var body struct {
		UIDs []string `json:"uids"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	if err := h.svc.AdminRemoveTaxonomy(c.Request.Context(), body.UIDs); err != nil {
		// 业务拒绝（仍被引用）为明确中文，可直接返回
		msg := err.Error()
		if strings.Contains(msg, "无法删除") {
			response.BadRequest(c, msg)
			return
		}
		response.BadRequestLog(c, err, "删除失败")
		return
	}
	response.OK(c, nil)
}

func (h *Handler) AdminStatsOverview(c *gin.Context) {
	days := 7
	if v := c.Query("days"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			days = n
		}
	}
	data, err := h.svc.AdminStatsOverview(c.Request.Context(), days)
	if err != nil {
		response.ServerErrorLog(c, err, "stats overview failed")
		return
	}
	response.OK(c, data)
}
