package blog

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/jiangbyte/go-blog/internal/model"
	"github.com/jiangbyte/go-blog/internal/pkg/page"
)

// jsonContainsUID 构造 JSON 数组包含某 uid 的绑定条件（避免字符串拼接注入）。
func jsonContainsUID(column, uid string) (cond string, arg string) {
	return fmt.Sprintf("JSON_CONTAINS(%s, JSON_QUOTE(?))", column), uid
}

// postFilter 文章列表过滤条件。
type postFilter struct {
	HomeOnly    bool
	CategoryUID string
	TagUID      string
	Year        string // YYYY；空或 all 表示不限
	KeywordLike string // 小写已 trim；匹配标题或摘要；空不限
}

// postsQuery 按过滤条件构建文章查询（不含排序与分页）；前台仅已发布。
func (s *Service) postsQuery(ctx context.Context, f postFilter) *gorm.DB {
	q := s.db.WithContext(ctx).Model(&model.BlogPost{}).Where("status = ?", StatusPublished)
	if f.HomeOnly {
		q = q.Where("home_visible = ?", true)
	}
	if uid := strings.TrimSpace(f.CategoryUID); uid != "" {
		cond, arg := jsonContainsUID("categories_json", uid)
		q = q.Where(cond, arg)
	}
	if uid := strings.TrimSpace(f.TagUID); uid != "" {
		cond, arg := jsonContainsUID("tags_json", uid)
		q = q.Where(cond, arg)
	}
	if y := strings.TrimSpace(f.Year); y != "" && y != "all" {
		q = q.Where("YEAR(published_at) = ?", y)
	}
	if kw := sanitizeLikeKeyword(f.KeywordLike); kw != "" {
		pat := "%" + kw + "%"
		q = q.Where("(LOWER(title) LIKE ? OR LOWER(IFNULL(description, '')) LIKE ?)", pat, pat)
	}
	return q
}

// orderPosts 首页/列表统一排序：置顶优先，再按发布日与 id。
func (s *Service) orderPosts(q *gorm.DB) *gorm.DB {
	return q.Order("sticky DESC").Order("published_at DESC").Order("id DESC")
}

// fetchPosts 执行查询并装配 taxonomy、批量签发封面。
func (s *Service) fetchPosts(ctx context.Context, q *gorm.DB, includeBody bool) ([]Post, error) {
	// 1. 列表场景省略正文，降低 IO
	if !includeBody {
		q = q.Omit("body")
	}
	var rows []model.BlogPost
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	// 2. 一次加载本页涉及的分类/标签元数据
	tax, err := s.taxonomyByUIDs(ctx, collectUIDsFromPosts(rows))
	if err != nil {
		return nil, err
	}
	// 3. 转 DTO 并收集封面 ref，批量签发访问 URL
	list := make([]Post, 0, len(rows))
	covers := make([]string, 0, len(rows))
	for _, r := range rows {
		p := postFromModel(r, includeBody, tax)
		covers = append(covers, p.Cover)
		list = append(list, p)
	}
	urls := s.resolveMediaBatch(ctx, covers)
	for i := range list {
		ref := strings.TrimSpace(list[i].Cover)
		if u, ok := urls[ref]; ok {
			list[i].Cover = u
		}
	}
	return list, nil
}

// pagePosts 文章标准分页：Count + Offset/Limit + fetchPosts。
func (s *Service) pagePosts(ctx context.Context, f postFilter, current, size int, includeBody bool) (page.IPage, error) {
	// 1. 规范化页码/页大小
	current, size = page.Normalize(current, size, 12)
	// 2. 统计总数（独立查询，避免带 Order/Limit 干扰）
	base := s.postsQuery(ctx, f)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return page.IPage{}, err
	}
	// 3. 取当前页并装配
	q := s.orderPosts(s.postsQuery(ctx, f)).Offset((current - 1) * size).Limit(size)
	list, err := s.fetchPosts(ctx, q, includeBody)
	if err != nil {
		return page.IPage{}, err
	}
	return page.Of(list, total, current, size), nil
}

// projectFilter 项目列表过滤条件。
type projectFilter struct {
	TagUID      string
	KeywordLike string
}

// projectsQuery 按过滤条件构建项目查询；前台仅已发布。
func (s *Service) projectsQuery(ctx context.Context, f projectFilter) *gorm.DB {
	q := s.db.WithContext(ctx).Model(&model.BlogProject{}).Where("status = ?", StatusPublished)
	if uid := strings.TrimSpace(f.TagUID); uid != "" {
		cond, arg := jsonContainsUID("tags_json", uid)
		q = q.Where(cond, arg)
	}
	if kw := sanitizeLikeKeyword(f.KeywordLike); kw != "" {
		pat := "%" + kw + "%"
		q = q.Where("(LOWER(title) LIKE ? OR LOWER(IFNULL(description, '')) LIKE ?)", pat, pat)
	}
	return q
}

// sanitizeLikeKeyword 去掉 LIKE 通配符，避免用户输入 %/_ 扩大命中面。
func sanitizeLikeKeyword(kw string) string {
	kw = strings.TrimSpace(strings.ToLower(kw))
	if kw == "" {
		return ""
	}
	return strings.Map(func(r rune) rune {
		if r == '%' || r == '_' {
			return -1
		}
		return r
	}, kw)
}

// orderProjects 项目列表排序：置顶优先，再权重升序、发布日。
func (s *Service) orderProjects(q *gorm.DB) *gorm.DB {
	return q.Order("sticky DESC").Order("weight ASC").Order("published_at DESC").Order("id DESC")
}

// fetchProjects 执行查询并装配项目标签、批量签发封面。
func (s *Service) fetchProjects(ctx context.Context, q *gorm.DB, includeBody bool) ([]Project, error) {
	// 1. 列表省略正文
	if !includeBody {
		q = q.Omit("body")
	}
	var rows []model.BlogProject
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	// 2. 批量加载项目标签
	tax, err := s.taxonomyByUIDs(ctx, collectUIDsFromProjects(rows))
	if err != nil {
		return nil, err
	}
	// 3. 转 DTO 并批量签发封面 URL
	list := make([]Project, 0, len(rows))
	covers := make([]string, 0, len(rows))
	for _, r := range rows {
		p := projectFromModel(r, includeBody, tax)
		covers = append(covers, p.Cover)
		list = append(list, p)
	}
	urls := s.resolveMediaBatch(ctx, covers)
	for i := range list {
		ref := strings.TrimSpace(list[i].Cover)
		if u, ok := urls[ref]; ok {
			list[i].Cover = u
		}
	}
	return list, nil
}

// pageProjects 项目标准分页。
func (s *Service) pageProjects(ctx context.Context, f projectFilter, current, size int, includeBody bool) (page.IPage, error) {
	// 1. 规范化分页
	current, size = page.Normalize(current, size, 12)
	// 2. Count 总数
	base := s.projectsQuery(ctx, f)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return page.IPage{}, err
	}
	// 3. 取页并装配
	q := s.orderProjects(s.projectsQuery(ctx, f)).Offset((current - 1) * size).Limit(size)
	list, err := s.fetchProjects(ctx, q, includeBody)
	if err != nil {
		return page.IPage{}, err
	}
	return page.Of(list, total, current, size), nil
}

func archiveEntryFromPost(p Post, cover string) map[string]any {
	primaryName := "未分类"
	if len(p.Categories) > 0 {
		primaryName = p.Categories[0].Name
	}
	year := ""
	if len(p.Date) >= 4 {
		year = p.Date[:4]
	}
	return map[string]any{
		"uid":              p.UID,
		"title":            p.Title,
		"url":              "/p/" + p.UID,
		"cover":            cover,
		"primary_category": primaryName,
		"date_label":       p.Date,
		"year":             year,
		"description":      p.Description,
		"date":             p.Date,
		"tags":             p.Tags,
		"categories":       p.Categories,
	}
}

func stickyInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func (s *Service) neighborBrief(ctx context.Context, p model.BlogPost) map[string]any {
	return map[string]any{
		"uid":   p.UID,
		"title": p.Title,
		"cover": s.resolveMedia(ctx, p.Cover),
	}
}

// neighborRow 按列表序找更新 / 更旧一篇（列表：sticky DESC, published_at DESC, id DESC）。
func (s *Service) neighborRow(ctx context.Context, cur model.BlogPost, newer bool) (*model.BlogPost, error) {
	st := stickyInt(cur.Sticky)
	q := s.db.WithContext(ctx).Model(&model.BlogPost{}).
		Select("uid", "title", "cover").
		Where("status = ?", StatusPublished)
	if newer {
		q = q.Where(
			"(sticky > ?) OR (sticky = ? AND published_at > ?) OR (sticky = ? AND published_at = ? AND id > ?)",
			st, st, cur.PublishedAt, st, cur.PublishedAt, cur.ID,
		).Order("sticky ASC").Order("published_at ASC").Order("id ASC")
	} else {
		q = q.Where(
			"(sticky < ?) OR (sticky = ? AND published_at < ?) OR (sticky = ? AND published_at = ? AND id < ?)",
			st, st, cur.PublishedAt, st, cur.PublishedAt, cur.ID,
		).Order("sticky DESC").Order("published_at DESC").Order("id DESC")
	}
	var row model.BlogPost
	err := q.Limit(1).Take(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}
