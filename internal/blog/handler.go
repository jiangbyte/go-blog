package blog

import (
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

func (h *Handler) Register(r gin.IRoutes) {
	r.GET("/portal/blog/sitemap.xml", h.Sitemap)
	r.GET("/portal/blog/site", h.Site)
	r.GET("/portal/blog/site/stats", h.SiteStats)
	r.POST("/portal/blog/visit", h.TrackVisit)
	// Admin 登录页品牌/页脚（沿用 mypage 路径）
	r.GET("/portal/sys/site/footer", h.SysSiteFooter)
	r.GET("/portal/sys/site/legal", h.SysSiteLegal)
	r.GET("/portal/blog/about", h.About)
	r.GET("/portal/blog/posts", h.Posts)
	r.GET("/portal/blog/posts/home", h.HomePosts)
	r.GET("/portal/blog/posts/detail", h.PostDetail)
	r.GET("/portal/blog/posts/related", h.Related)
	r.GET("/portal/blog/posts/neighbors", h.Neighbors)
	r.GET("/portal/blog/posts/page", h.PostsPage)
	r.GET("/portal/blog/posts/by-category", h.ByCategory)
	r.GET("/portal/blog/posts/by-tag", h.ByTag)
	r.GET("/portal/blog/projects", h.Projects)
	r.GET("/portal/blog/projects/page", h.ProjectsPage)
	r.GET("/portal/blog/projects/detail", h.ProjectDetail)
	r.GET("/portal/blog/projects/by-tag", h.ProjectsByTag)
	r.GET("/portal/blog/archive/entries", h.ArchiveEntries)
	r.GET("/portal/blog/archive/years", h.ArchiveYears)
	r.GET("/portal/blog/archive/page", h.ArchivePage)
	r.GET("/portal/blog/taxonomy/categories", h.Categories)
	r.GET("/portal/blog/taxonomy/tags", h.Tags)
	r.GET("/portal/blog/taxonomy/project-tags", h.ProjectTags)
	r.GET("/portal/blog/search", h.Search)
}

func (h *Handler) Sitemap(c *gin.Context) {
	raw, err := h.svc.BuildSitemapXML(c.Request.Context(), h.svc.publicSiteURL)
	if err != nil {
		response.ServerErrorLog(c, err, "sitemap failed")
		return
	}
	c.Data(200, "application/xml; charset=utf-8", raw)
}

func (h *Handler) SysSiteFooter(c *gin.Context) {
	data, err := h.svc.SysSiteFooter(c.Request.Context())
	if err != nil {
		response.ServerErrorLog(c, err, "load site footer failed")
		return
	}
	response.OK(c, data)
}

func (h *Handler) SysSiteLegal(c *gin.Context) {
	data, err := h.svc.SysSiteLegal(c.Request.Context())
	if err != nil {
		response.ServerErrorLog(c, err, "load site legal failed")
		return
	}
	response.OK(c, data)
}

func (h *Handler) Site(c *gin.Context) {
	site, err := h.svc.Site(c.Request.Context())
	if err != nil {
		response.ServerErrorLog(c, err, "load site failed")
		return
	}
	response.OK(c, gin.H{"site": site})
}

func (h *Handler) SiteStats(c *gin.Context) {
	stats, err := h.svc.SiteStats(c.Request.Context())
	if err != nil {
		response.ServerErrorLog(c, err, "load site stats failed")
		return
	}
	response.OK(c, stats)
}

func (h *Handler) TrackVisit(c *gin.Context) {
	var body VisitInput
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	if err := h.svc.AllowVisitRate(c.Request.Context(), c.ClientIP()); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	out, err := h.svc.TrackVisit(c.Request.Context(), body)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, out)
}

func (h *Handler) About(c *gin.Context) {
	doc, err := h.svc.About(c.Request.Context())
	if err != nil {
		response.ServerErrorLog(c, err, "load about failed")
		return
	}
	response.OK(c, gin.H{"doc": doc})
}

func (h *Handler) Posts(c *gin.Context) {
	cur, size := page.Parse(c, 12)
	pg, err := h.svc.PagePosts(c.Request.Context(), cur, size, false)
	if err != nil {
		response.ServerErrorLog(c, err, "list posts failed")
		return
	}
	response.OK(c, pg)
}

func (h *Handler) HomePosts(c *gin.Context) {
	list, err := h.svc.HomePosts(c.Request.Context())
	if err != nil {
		response.ServerErrorLog(c, err, "list home posts failed")
		return
	}
	response.OK(c, gin.H{"list": list})
}

func (h *Handler) PostDetail(c *gin.Context) {
	uid := strings.TrimSpace(c.Query("uid"))
	if uid == "" {
		response.BadRequest(c, "uid required")
		return
	}
	post, err := h.svc.GetPost(c.Request.Context(), uid)
	if IsNotFound(err) {
		response.NotFound(c, "post not found")
		return
	}
	if err != nil {
		response.ServerErrorLog(c, err, "get post failed")
		return
	}
	response.OK(c, gin.H{"post": post})
}

func (h *Handler) Related(c *gin.Context) {
	uid := strings.TrimSpace(c.Query("uid"))
	if uid == "" {
		response.BadRequest(c, "uid required")
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "4"))
	if limit < 1 {
		limit = 1
	}
	if limit > 20 {
		limit = 20
	}
	list, err := h.svc.RelatedPosts(c.Request.Context(), uid, limit)
	if err != nil {
		response.ServerErrorLog(c, err, "list related posts failed")
		return
	}
	response.OK(c, gin.H{"list": list})
}

func (h *Handler) Neighbors(c *gin.Context) {
	uid := strings.TrimSpace(c.Query("uid"))
	if uid == "" {
		response.BadRequest(c, "uid required")
		return
	}
	data, err := h.svc.Neighbors(c.Request.Context(), uid)
	if IsNotFound(err) {
		response.NotFound(c, "post not found")
		return
	}
	if err != nil {
		response.ServerErrorLog(c, err, "get neighbors failed")
		return
	}
	response.OK(c, data)
}

func (h *Handler) PostsPage(c *gin.Context) {
	cur, size := page.Parse(c, 12)
	pg, err := h.svc.PagePosts(c.Request.Context(), cur, size, true)
	if err != nil {
		response.ServerErrorLog(c, err, "page posts failed")
		return
	}
	response.OK(c, pg)
}

func (h *Handler) ByCategory(c *gin.Context) {
	uid := strings.TrimSpace(c.Query("uid"))
	if uid == "" {
		response.BadRequest(c, "uid required")
		return
	}
	cur, size := page.Parse(c, 12)
	pg, err := h.svc.PagePostsByCategory(c.Request.Context(), cur, size, uid)
	if err != nil {
		response.ServerErrorLog(c, err, "list posts by category failed")
		return
	}
	response.OK(c, pg)
}

func (h *Handler) ByTag(c *gin.Context) {
	uid := strings.TrimSpace(c.Query("uid"))
	if uid == "" {
		response.BadRequest(c, "uid required")
		return
	}
	cur, size := page.Parse(c, 12)
	pg, err := h.svc.PagePostsByTag(c.Request.Context(), cur, size, uid)
	if err != nil {
		response.ServerErrorLog(c, err, "list posts by tag failed")
		return
	}
	response.OK(c, pg)
}

func (h *Handler) Projects(c *gin.Context) {
	cur, size := page.Parse(c, 12)
	pg, err := h.svc.PageProjects(c.Request.Context(), cur, size, "")
	if err != nil {
		response.ServerErrorLog(c, err, "list projects failed")
		return
	}
	response.OK(c, pg)
}

func (h *Handler) ProjectsPage(c *gin.Context) {
	cur, size := page.Parse(c, 12)
	pg, err := h.svc.PageProjects(c.Request.Context(), cur, size, c.Query("uid"))
	if err != nil {
		response.ServerErrorLog(c, err, "page projects failed")
		return
	}
	response.OK(c, pg)
}

func (h *Handler) ProjectDetail(c *gin.Context) {
	uid := strings.TrimSpace(c.Query("uid"))
	if uid == "" {
		response.BadRequest(c, "uid required")
		return
	}
	p, err := h.svc.GetProject(c.Request.Context(), uid)
	if IsNotFound(err) {
		response.NotFound(c, "project not found")
		return
	}
	if err != nil {
		response.ServerErrorLog(c, err, "get project failed")
		return
	}
	response.OK(c, gin.H{"project": p})
}

func (h *Handler) ProjectsByTag(c *gin.Context) {
	uid := strings.TrimSpace(c.Query("uid"))
	if uid == "" {
		response.BadRequest(c, "uid required")
		return
	}
	cur, size := page.Parse(c, 12)
	pg, err := h.svc.PageProjects(c.Request.Context(), cur, size, uid)
	if err != nil {
		response.ServerErrorLog(c, err, "list projects by tag failed")
		return
	}
	response.OK(c, pg)
}

func (h *Handler) ArchiveEntries(c *gin.Context) {
	cur, size := page.Parse(c, 12)
	pg, err := h.svc.PageArchive(c.Request.Context(), cur, size, "", c.Query("by"), c.Query("uid"))
	if err != nil {
		response.ServerErrorLog(c, err, "list archive entries failed")
		return
	}
	response.OK(c, pg)
}

func (h *Handler) ArchiveYears(c *gin.Context) {
	list, err := h.svc.ArchiveYears(c.Request.Context())
	if err != nil {
		response.ServerErrorLog(c, err, "list archive years failed")
		return
	}
	response.OK(c, gin.H{"list": list})
}

func (h *Handler) ArchivePage(c *gin.Context) {
	cur, size := page.Parse(c, 12)
	pg, err := h.svc.PageArchive(c.Request.Context(), cur, size, c.Query("year"), c.Query("by"), c.Query("uid"))
	if err != nil {
		response.ServerErrorLog(c, err, "page archive failed")
		return
	}
	response.OK(c, pg)
}

func (h *Handler) Categories(c *gin.Context) {
	list, err := h.svc.Categories(c.Request.Context())
	if err != nil {
		response.ServerErrorLog(c, err, "list categories failed")
		return
	}
	response.OK(c, gin.H{"list": list})
}

func (h *Handler) Tags(c *gin.Context) {
	list, err := h.svc.Tags(c.Request.Context())
	if err != nil {
		response.ServerErrorLog(c, err, "list tags failed")
		return
	}
	response.OK(c, gin.H{"list": list})
}

func (h *Handler) ProjectTags(c *gin.Context) {
	list, err := h.svc.ProjectTags(c.Request.Context())
	if err != nil {
		response.ServerErrorLog(c, err, "list project tags failed")
		return
	}
	response.OK(c, gin.H{"list": list})
}

func (h *Handler) Search(c *gin.Context) {
	cur, size := page.Parse(c, 12)
	data, err := h.svc.Search(c.Request.Context(), c.Query("q"), cur, size)
	if err != nil {
		response.ServerErrorLog(c, err, "search failed")
		return
	}
	response.OK(c, data)
}
