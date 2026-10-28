package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/jiangbyte/go-blog/internal/pkg/response"
)

// StaticMenu is the simplified admin menu (no RBAC tables).
func StaticMenu() []gin.H {
	return []gin.H{
		{
			"id": "1", "parent_id": nil, "code": "dashboard", "name": "概览",
			"resource_type": "MENU", "module_id": "1", "module_id_name": "管理",
			"path": "/dashboard", "component": "/dashboard/index.vue",
			"redirect": nil, "icon": "icon-park-outline:analysis",
			"sort": 1, "is_visible": true, "is_cache": false, "is_affix": true,
			"status": "ENABLED", "layout": nil,
		},
		{
			"id": "10", "parent_id": nil, "code": "blog", "name": "博客",
			"resource_type": "CATALOG", "module_id": "1", "module_id_name": "管理",
			"path": "/blog", "component": nil, "redirect": "/blog/posts",
			"icon": "icon-park-outline:notebook",
			"sort": 10, "is_visible": true, "is_cache": false, "is_affix": false,
			"status": "ENABLED", "layout": nil,
		},
		{
			"id": "11", "parent_id": "10", "code": "blog-posts", "name": "文章",
			"resource_type": "MENU", "module_id": "1", "module_id_name": "管理",
			"path": "/blog/posts", "component": "/blog/posts/index.vue",
			"redirect": nil, "icon": "icon-park-outline:doc-detail",
			"sort": 1, "is_visible": true, "is_cache": false, "is_affix": false,
			"status": "ENABLED", "layout": nil,
		},
		{
			"id": "12", "parent_id": "10", "code": "blog-projects", "name": "项目",
			"resource_type": "MENU", "module_id": "1", "module_id_name": "管理",
			"path": "/blog/projects", "component": "/blog/projects/index.vue",
			"redirect": nil, "icon": "icon-park-outline:application-menu",
			"sort": 2, "is_visible": true, "is_cache": false, "is_affix": false,
			"status": "ENABLED", "layout": nil,
		},
		{
			"id": "13", "parent_id": "10", "code": "blog-categories", "name": "分类",
			"resource_type": "MENU", "module_id": "1", "module_id_name": "管理",
			"path": "/blog/categories", "component": "/blog/categories/index.vue",
			"redirect": nil, "icon": "icon-park-outline:category-management",
			"sort": 3, "is_visible": true, "is_cache": false, "is_affix": false,
			"status": "ENABLED", "layout": nil,
		},
		{
			"id": "19", "parent_id": "10", "code": "blog-tags", "name": "标签",
			"resource_type": "MENU", "module_id": "1", "module_id_name": "管理",
			"path": "/blog/tags", "component": "/blog/tags/index.vue",
			"redirect": nil, "icon": "icon-park-outline:tag-one",
			"sort": 4, "is_visible": true, "is_cache": false, "is_affix": false,
			"status": "ENABLED", "layout": nil,
		},
		{
			"id": "20", "parent_id": "10", "code": "blog-project-tags", "name": "项目标签",
			"resource_type": "MENU", "module_id": "1", "module_id_name": "管理",
			"path": "/blog/project-tags", "component": "/blog/project-tags/index.vue",
			"redirect": nil, "icon": "icon-park-outline:label",
			"sort": 5, "is_visible": true, "is_cache": false, "is_affix": false,
			"status": "ENABLED", "layout": nil,
		},
		{
			"id": "18", "parent_id": "10", "code": "blog-site", "name": "站点",
			"resource_type": "MENU", "module_id": "1", "module_id_name": "管理",
			"path": "/blog/site", "component": "/blog/site/index.vue",
			"redirect": nil, "icon": "icon-park-outline:config",
			"sort": 6, "is_visible": true, "is_cache": false, "is_affix": false,
			"status": "ENABLED", "layout": nil,
		},
		{
			"id": "21", "parent_id": "10", "code": "blog-about", "name": "关于",
			"resource_type": "MENU", "module_id": "1", "module_id_name": "管理",
			"path": "/blog/about", "component": "/blog/about/index.vue",
			"redirect": nil, "icon": "icon-park-outline:info",
			"sort": 7, "is_visible": true, "is_cache": false, "is_affix": false,
			"status": "ENABLED", "layout": nil,
		},
		{
			"id": "40", "parent_id": nil, "code": "resume", "name": "简历",
			"resource_type": "CATALOG", "module_id": "1", "module_id_name": "管理",
			"path": "/resume", "component": nil, "redirect": "/resume/list",
			"icon": "icon-park-outline:id-card",
			"sort": 20, "is_visible": true, "is_cache": false, "is_affix": false,
			"status": "ENABLED", "layout": nil,
		},
		{
			"id": "41", "parent_id": "40", "code": "resume-list", "name": "简历列表",
			"resource_type": "MENU", "module_id": "1", "module_id_name": "管理",
			"path": "/resume/list", "component": "/resume/list/index.vue",
			"redirect": nil, "icon": "icon-park-outline:list",
			"sort": 1, "is_visible": true, "is_cache": false, "is_affix": false,
			"status": "ENABLED", "layout": nil,
		},
		{
			"id": "42", "parent_id": "41", "code": "resume-edit", "name": "编辑简历",
			"resource_type": "PAGE", "module_id": "1", "module_id_name": "管理",
			"path": "/resume/edit", "component": "/resume/edit/index.vue",
			"redirect": nil, "icon": nil,
			"sort": 1, "is_visible": false, "is_cache": false, "is_affix": false,
			"status": "ENABLED", "layout": nil,
		},
		{
			"id": "43", "parent_id": "41", "code": "resume-preview", "name": "预览简历",
			"resource_type": "PAGE", "module_id": "1", "module_id_name": "管理",
			"path": "/resume/preview", "component": "/resume/preview/index.vue",
			"redirect": nil, "icon": nil,
			"sort": 2, "is_visible": false, "is_cache": false, "is_affix": false,
			"status": "ENABLED", "layout": "fullscreen",
		},
		{
			"id": "30", "parent_id": nil, "code": "sys", "name": "系统",
			"resource_type": "CATALOG", "module_id": "1", "module_id_name": "管理",
			"path": "/sys", "component": nil, "redirect": "/sys/file",
			"icon": "icon-park-outline:setting-two",
			"sort": 30, "is_visible": true, "is_cache": false, "is_affix": false,
			"status": "ENABLED", "layout": nil,
		},
		{
			"id": "31", "parent_id": "30", "code": "sys-file", "name": "文件",
			"resource_type": "MENU", "module_id": "1", "module_id_name": "管理",
			"path": "/sys/file", "component": "/sys/file/index.vue",
			"redirect": nil, "icon": "icon-park-outline:folder-open",
			"sort": 1, "is_visible": true, "is_cache": false, "is_affix": false,
			"status": "ENABLED", "layout": nil,
		},
		// 隐藏表单页（菜单高亮仍挂在文章 / 项目下）
		{
			"id": "14", "parent_id": "11", "code": "blog-posts-create", "name": "新建文章",
			"resource_type": "PAGE", "module_id": "1", "module_id_name": "管理",
			"path": "/blog/posts/create", "component": "/blog/posts/form.vue",
			"redirect": nil, "icon": nil,
			"sort": 1, "is_visible": false, "is_cache": false, "is_affix": false,
			"status": "ENABLED", "layout": nil,
		},
		{
			"id": "15", "parent_id": "11", "code": "blog-posts-edit", "name": "编辑文章",
			"resource_type": "PAGE", "module_id": "1", "module_id_name": "管理",
			"path": "/blog/posts/edit", "component": "/blog/posts/form.vue",
			"redirect": nil, "icon": nil,
			"sort": 2, "is_visible": false, "is_cache": false, "is_affix": false,
			"status": "ENABLED", "layout": nil,
		},
		{
			"id": "16", "parent_id": "12", "code": "blog-projects-create", "name": "新建项目",
			"resource_type": "PAGE", "module_id": "1", "module_id_name": "管理",
			"path": "/blog/projects/create", "component": "/blog/projects/form.vue",
			"redirect": nil, "icon": nil,
			"sort": 1, "is_visible": false, "is_cache": false, "is_affix": false,
			"status": "ENABLED", "layout": nil,
		},
		{
			"id": "17", "parent_id": "12", "code": "blog-projects-edit", "name": "编辑项目",
			"resource_type": "PAGE", "module_id": "1", "module_id_name": "管理",
			"path": "/blog/projects/edit", "component": "/blog/projects/form.vue",
			"redirect": nil, "icon": nil,
			"sort": 2, "is_visible": false, "is_cache": false, "is_affix": false,
			"status": "ENABLED", "layout": nil,
		},
	}
}

func (h *Handler) RegisterAdmin(r gin.IRoutes) {
	r.GET("/admin/user/resource/current", h.CurrentResources)
}

func (h *Handler) CurrentResources(c *gin.Context) {
	response.OK(c, StaticMenu())
}
