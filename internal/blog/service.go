// Package blog 门户与管理端博客业务：文章/项目/站点/访问统计。
package blog

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/jiangbyte/go-blog/internal/file"
	"github.com/jiangbyte/go-blog/internal/model"
	"github.com/jiangbyte/go-blog/internal/pkg/page"
)

// Service 博客领域服务。
type Service struct {
	db            *gorm.DB
	rdb           *redis.Client
	files         *file.Service
	publicSiteURL string // 对外站点根，如 https://example.com
	indexNowKey   string // IndexNow 密钥；空则不推送
}

func NewService(db *gorm.DB, rdb *redis.Client, files *file.Service) *Service {
	return &Service{db: db, rdb: rdb, files: files}
}

// SetPublicSiteURL 设置 sitemap 等使用的对外站点根地址。
func (s *Service) SetPublicSiteURL(u string) {
	s.publicSiteURL = strings.TrimRight(strings.TrimSpace(u), "/")
}

// Site 读取站点配置，并用个人中心资料覆盖作者展示、签发媒体 URL。
func (s *Service) Site(ctx context.Context) (map[string]any, error) {
	// 1. 读单行站点配置
	var row model.BlogSite
	if err := s.db.WithContext(ctx).First(&row, 1).Error; err != nil {
		return nil, err
	}
	cfg, err := model.JSONMap(row.ConfigJSON)
	if err != nil {
		return nil, err
	}
	// 2. 作者字段以 admin 资料为准，避免站点 JSON 与个人中心漂移
	s.applyAuthorFromProfile(ctx, cfg)
	// 3. 默认封面/背景等媒体 ref 签发为可访问 URL（nav_group / links 原样返回，由后台配置）
	s.resolveSiteMedia(ctx, cfg)
	return cfg, nil
}

func (s *Service) resolveMedia(ctx context.Context, ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" || s.files == nil {
		return ref
	}
	if u := s.files.ResolveAccessURL(ctx, ref); u != "" {
		return u
	}
	return ref
}

// resolveMediaBatch 批量签发封面等媒体地址，避免列表 N+1。
func (s *Service) resolveMediaBatch(ctx context.Context, refs []string) map[string]string {
	out := make(map[string]string, len(refs))
	if s.files == nil {
		for _, r := range refs {
			r = strings.TrimSpace(r)
			if r != "" {
				out[r] = r
			}
		}
		return out
	}
	resolved := s.files.ResolveAccessURLs(ctx, refs)
	for _, r := range refs {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		if u := resolved[r]; u != "" {
			out[r] = u
		} else {
			out[r] = r
		}
	}
	return out
}

func (s *Service) resolveSiteMedia(ctx context.Context, site map[string]any) {
	if site == nil {
		return
	}
	if raw, ok := site["default_covers"].([]any); ok {
		out := make([]any, 0, len(raw))
		for _, v := range raw {
			if str, ok := v.(string); ok {
				out = append(out, s.resolveMedia(ctx, str))
			} else {
				out = append(out, v)
			}
		}
		site["default_covers"] = out
	}
	if author, ok := site["author"].(map[string]any); ok && author != nil {
		if bg, ok := author["background_url"].(string); ok {
			author["background_url"] = s.resolveMedia(ctx, bg)
		}
		site["author"] = author
	}
	if footer, ok := site["footer"].(map[string]any); ok && footer != nil {
		if qr, ok := footer["wechat_qr"].(string); ok {
			footer["wechat_qr"] = s.resolveMedia(ctx, qr)
		}
		site["footer"] = footer
	}
}

// applyAuthorFromProfile 用个人中心（admin_user）覆盖站点 author 的身份字段。
func (s *Service) applyAuthorFromProfile(ctx context.Context, site map[string]any) {
	if site == nil {
		return
	}
	u, err := s.primaryAdmin(ctx)
	if err != nil || u == nil {
		return
	}
	author, _ := site["author"].(map[string]any)
	if author == nil {
		author = map[string]any{}
	}
	name := strings.TrimSpace(u.Nickname)
	if name == "" {
		name = u.Username
	}
	author["name"] = name
	author["avatar_url"] = s.resolveMedia(ctx, u.Avatar)
	author["description"] = u.Signature
	if content, _ := author["content"].(string); strings.TrimSpace(content) == "" {
		author["content"] = u.Signature
	}
	if _, ok := author["witty_words"]; !ok {
		author["witty_words"] = []string{}
	}
	if _, ok := author["links"]; !ok {
		if ng, ok := site["nav_group"].([]any); ok {
			author["links"] = ng
		} else {
			author["links"] = []any{}
		}
	}
	site["author"] = author
}

func (s *Service) primaryAdmin(ctx context.Context) (*model.AdminUser, error) {
	var u model.AdminUser
	err := s.db.WithContext(ctx).
		Where("enabled = ?", 1).
		Order("id ASC").
		First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// SysSiteFooter Admin 登录页品牌/备案（snake_case）。
func (s *Service) SysSiteFooter(ctx context.Context) (map[string]any, error) {
	site, err := s.Site(ctx)
	if err != nil {
		return nil, err
	}
	name, _ := site["site_name"].(string)
	if name == "" {
		name, _ = site["title"].(string)
	}
	if name == "" {
		name = "Demo Blog"
	}
	copyright := ""
	icpNumber := ""
	icpURL := "https://beian.miit.gov.cn/"
	if footer, ok := site["footer"].(map[string]any); ok {
		if c, ok := footer["copyright"].(string); ok {
			copyright = c
		}
		if beian, ok := footer["beian"].([]any); ok && len(beian) > 0 {
			if item, ok := beian[0].(map[string]any); ok {
				if n, ok := item["name"].(string); ok {
					icpNumber = n
				}
				if u, ok := item["url"].(string); ok && u != "" {
					icpURL = u
				}
			}
		}
	}
	authorName := ""
	avatar := ""
	signature := ""
	if author, ok := site["author"].(map[string]any); ok {
		authorName, _ = author["name"].(string)
		avatar, _ = author["avatar_url"].(string)
		signature, _ = author["description"].(string)
	}
	return map[string]any{
		"app_name":               name,
		"app_subtitle":           "管理后台",
		"copyright_text":         copyright,
		"copyright_url":          "",
		"icp_number":             icpNumber,
		"icp_url":                icpURL,
		"psb_number":             "",
		"psb_url":                "",
		"home_profile_avatar":    avatar,
		"home_profile_name":      authorName,
		"home_profile_tags":      []string{},
		"home_profile_signature": signature,
	}, nil
}

func (s *Service) SysSiteLegal(ctx context.Context) (map[string]any, error) {
	_ = ctx
	return map[string]any{
		"terms_of_service": "",
		"privacy_policy":   "",
	}, nil
}

func (s *Service) SiteStats(ctx context.Context) (map[string]any, error) {
	site, err := s.Site(ctx)
	if err != nil {
		return nil, err
	}
	var postCount int64
	_ = s.db.WithContext(ctx).Model(&model.BlogPost{}).Where("status = ?", StatusPublished).Count(&postCount).Error
	var projectCount int64
	_ = s.db.WithContext(ctx).Model(&model.BlogProject{}).Where("status = ?", StatusPublished).Count(&projectCount).Error
	startedAt, _ := site["started_at"].(string)
	visitCount, _, readCount, err := s.statsTotals(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"visit_count":   visitCount,
		"read_count":    readCount,
		"post_count":    postCount,
		"project_count": projectCount,
		"started_at":    startedAt,
	}, nil
}

func (s *Service) About(ctx context.Context) (map[string]any, error) {
	var row model.BlogAbout
	if err := s.db.WithContext(ctx).First(&row, 1).Error; err != nil {
		return nil, err
	}
	doc, err := model.JSONMap(row.DocJSON)
	if err != nil {
		return nil, err
	}
	out := normalizeAboutDoc(doc)
	// 门户关于页正文同样签发站内图片，避免相对路径裂图
	if c, ok := out["content"].(string); ok && strings.TrimSpace(c) != "" {
		out["content"] = s.resolveBodyMedia(ctx, c)
	}
	return out, nil
}

func postFromModel(r model.BlogPost, includeBody bool, tax map[string]model.BlogTaxonomy) Post {
	status := r.Status
	if status != StatusDraft && status != StatusPublished {
		status = StatusPublished
	}
	p := Post{
		UID:                r.UID,
		Title:              r.Title,
		Description:        r.Description,
		Date:               model.FormatDate(r.PublishedAt),
		Status:             status,
		Updated:            model.FormatDatePtr(r.ContentUpdatedAt),
		Cover:              r.Cover,
		Categories:         resolveTaxa(model.JSONStrings(r.CategoriesJSON), tax),
		Tags:               resolveTaxa(model.JSONStrings(r.TagsJSON), tax),
		Reprint:            r.Reprint,
		Sticky:             r.Sticky,
		Home:               r.HomeVisible,
		ReadingTimeMinutes: r.ReadingTimeMinutes,
		Locate:             r.Locate,
	}
	if includeBody {
		p.Body = r.Body
	}
	return p
}

// GetPost 按 uid 取文章详情（含正文、taxonomy、阅读量）；仅已发布。
func (s *Service) GetPost(ctx context.Context, uid string) (*Post, error) {
	// 1. 按 uid 查主表（草稿对前台不可见）
	var row model.BlogPost
	if err := s.db.WithContext(ctx).
		Where("uid = ? AND status = ?", uid, StatusPublished).
		First(&row).Error; err != nil {
		return nil, err
	}
	// 2. 解析分类/标签并装配 DTO
	tax, err := s.taxonomyByUIDs(ctx, collectUIDsFromPosts([]model.BlogPost{row}))
	if err != nil {
		return nil, err
	}
	p := postFromModel(row, true, tax)
	// 3. 签发封面与正文内图片，并从 Redis 读阅读量
	p.Cover = s.resolveMedia(ctx, p.Cover)
	p.Body = s.resolveBodyMedia(ctx, p.Body)
	p.ViewCount = s.entityViewCount(ctx, redisKeyPostViewsZ, p.UID)
	return &p, nil
}

// HomePosts 首页推荐流，仅 home_visible，最多 20 条。
func (s *Service) HomePosts(ctx context.Context) ([]Post, error) {
	q := s.orderPosts(s.postsQuery(ctx, postFilter{HomeOnly: true})).Limit(20)
	return s.fetchPosts(ctx, q, false)
}

func (s *Service) PagePosts(ctx context.Context, current, size int, homeOnly bool) (page.IPage, error) {
	return s.pagePosts(ctx, postFilter{HomeOnly: homeOnly}, current, size, false)
}

func (s *Service) PagePostsByCategory(ctx context.Context, current, size int, uid string) (page.IPage, error) {
	return s.pagePosts(ctx, postFilter{CategoryUID: uid}, current, size, false)
}

func (s *Service) PagePostsByTag(ctx context.Context, current, size int, uid string) (page.IPage, error) {
	return s.pagePosts(ctx, postFilter{TagUID: uid}, current, size, false)
}

// RelatedPosts 按标签重叠度推荐相关文章。
func (s *Service) RelatedPosts(ctx context.Context, uid string, limit int) ([]Post, error) {
	// 1. 取当前文并收集标签；无标签则无法推荐
	post, err := s.GetPost(ctx, uid)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 4
	}
	tagUIDs := make([]string, 0, len(post.Tags))
	tagSet := map[string]struct{}{}
	for _, t := range post.Tags {
		tagUIDs = append(tagUIDs, t.UID)
		tagSet[t.UID] = struct{}{}
	}
	q := s.db.WithContext(ctx).Model(&model.BlogPost{}).
		Where("uid <> ?", uid).
		Where("status = ?", StatusPublished)
	if len(tagUIDs) > 0 {
		conds := make([]string, 0, len(tagUIDs))
		args := make([]any, 0, len(tagUIDs))
		for _, t := range tagUIDs {
			cond, arg := jsonContainsUID("tags_json", t)
			conds = append(conds, cond)
			args = append(args, arg)
		}
		q = q.Where("("+strings.Join(conds, " OR ")+")", args...)
	} else {
		return []Post{}, nil
	}
	// 2. 多取候选再在内存按重叠标签数打分
	candidates, err := s.fetchPosts(ctx, s.orderPosts(q).Limit(limit*5), false)
	if err != nil {
		return nil, err
	}
	type scored struct {
		p Post
		s int
	}
	var scoredList []scored
	for _, p := range candidates {
		score := 0
		for _, t := range p.Tags {
			if _, ok := tagSet[t.UID]; ok {
				score++
			}
		}
		if score > 0 {
			scoredList = append(scoredList, scored{p: p, s: score})
		}
	}
	// 3. 按分数降序截取 limit 条
	sort.Slice(scoredList, func(i, j int) bool { return scoredList[i].s > scoredList[j].s })
	var out []Post
	for i := 0; i < len(scoredList) && i < limit; i++ {
		out = append(out, scoredList[i].p)
	}
	return out, nil
}

// Neighbors 按列表序返回上一篇/下一篇摘要。
func (s *Service) Neighbors(ctx context.Context, uid string) (map[string]any, error) {
	// 1. 定位当前已发布文章
	var cur model.BlogPost
	if err := s.db.WithContext(ctx).
		Where("uid = ? AND status = ?", strings.TrimSpace(uid), StatusPublished).
		First(&cur).Error; err != nil {
		return nil, err
	}
	// 2. 分别查更旧（prev）与更新（next）；找不到记 nil
	var prev, next any
	if older, err := s.neighborRow(ctx, cur, false); err == nil {
		prev = s.neighborBrief(ctx, *older)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if newer, err := s.neighborRow(ctx, cur, true); err == nil {
		next = s.neighborBrief(ctx, *newer)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return map[string]any{"prev": prev, "next": next}, nil
}

func projectFromModel(r model.BlogProject, includeBody bool, tax map[string]model.BlogTaxonomy) Project {
	status := r.Status
	if status != StatusDraft && status != StatusPublished {
		status = StatusPublished
	}
	p := Project{
		UID:         r.UID,
		Title:       r.Title,
		Description: r.Description,
		Date:        model.FormatDate(r.PublishedAt),
		Status:      status,
		Progress:    normalizeProjectProgress(r.Progress),
		Weight:      r.Weight,
		Sticky:      r.Sticky,
		Cover:       r.Cover,
		Git:         r.Git,
		Site:        r.Site,
		ProjectTags: resolveTaxa(model.JSONStrings(r.TagsJSON), tax),
	}
	if includeBody {
		p.Body = r.Body
	}
	return p
}

func (s *Service) GetProject(ctx context.Context, uid string) (*Project, error) {
	var row model.BlogProject
	if err := s.db.WithContext(ctx).
		Where("uid = ? AND status = ?", uid, StatusPublished).
		First(&row).Error; err != nil {
		return nil, err
	}
	tax, err := s.taxonomyByUIDs(ctx, collectUIDsFromProjects([]model.BlogProject{row}))
	if err != nil {
		return nil, err
	}
	p := projectFromModel(row, true, tax)
	// 签发封面与正文内 uploads/ 等对象键，避免前台相对路径落到 /projects/uploads/...
	p.Cover = s.resolveMedia(ctx, p.Cover)
	p.Body = s.resolveBodyMedia(ctx, p.Body)
	p.ViewCount = s.entityViewCount(ctx, redisKeyProjectViewsZ, p.UID)
	return &p, nil
}

// resolveBodyMedia 门户出站改写 Markdown 内站内图片为可访问 URL；管理端勿用。
func (s *Service) resolveBodyMedia(ctx context.Context, body string) string {
	body = strings.TrimSpace(body)
	if body == "" || s.files == nil {
		return body
	}
	return s.files.RewriteMarkdownMediaURLs(ctx, body)
}

func (s *Service) PageProjects(ctx context.Context, current, size int, tagUID string) (page.IPage, error) {
	return s.pageProjects(ctx, projectFilter{TagUID: tagUID}, current, size, false)
}

func (s *Service) defaultCovers(ctx context.Context) []string {
	site, err := s.Site(ctx)
	if err != nil {
		return nil
	}
	raw, _ := site["default_covers"].([]any)
	var covers []string
	for _, v := range raw {
		if s, ok := v.(string); ok {
			covers = append(covers, s)
		}
	}
	return covers
}

func coverOf(cover string, defaults []string, index int) string {
	if strings.TrimSpace(cover) != "" {
		return cover
	}
	if len(defaults) == 0 {
		return ""
	}
	return defaults[index%len(defaults)]
}

func (s *Service) archiveFilter(by, uid string) postFilter {
	f := postFilter{}
	switch strings.TrimSpace(by) {
	case "category":
		f.CategoryUID = uid
	case "tag":
		f.TagUID = uid
	}
	return f
}

func (s *Service) ArchiveYears(ctx context.Context) ([]string, error) {
	var years []int
	err := s.db.WithContext(ctx).Model(&model.BlogPost{}).
		Where("status = ?", StatusPublished).
		Distinct("YEAR(published_at)").
		Order("YEAR(published_at) DESC").
		Limit(page.MaxSize).
		Pluck("YEAR(published_at)", &years).Error
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(years))
	for _, y := range years {
		if y > 0 {
			out = append(out, strconv.Itoa(y))
		}
	}
	return out, nil
}

// PageArchive 归档分页：复用文章分页，再转成归档条目并补默认封面。
func (s *Service) PageArchive(ctx context.Context, current, size int, year, by, uid string) (page.IPage, error) {
	// 1. 按 year/by/uid 过滤后分页取文章
	f := s.archiveFilter(by, uid)
	f.Year = year
	pg, err := s.pagePosts(ctx, f, current, size, false)
	if err != nil {
		return page.IPage{}, err
	}
	posts, _ := pg.Records.([]Post)
	if posts == nil {
		posts = []Post{}
	}
	// 2. 映射为归档条目；无封面时用站点默认封面轮换
	covers := s.defaultCovers(ctx)
	entries := make([]map[string]any, 0, len(posts))
	base := (pg.Current - 1) * pg.Size
	for i, p := range posts {
		entries = append(entries, archiveEntryFromPost(p, coverOf(p.Cover, covers, base+i)))
	}
	return page.Of(entries, pg.Total, pg.Current, pg.Size), nil
}

func (s *Service) Categories(ctx context.Context) ([]map[string]any, error) {
	return s.taxonomyCloud(ctx, &model.BlogPost{}, "categories_json")
}

func (s *Service) Tags(ctx context.Context) ([]map[string]any, error) {
	return s.taxonomyCloud(ctx, &model.BlogPost{}, "tags_json")
}

func (s *Service) ProjectTags(ctx context.Context) ([]map[string]any, error) {
	return s.taxonomyCloud(ctx, &model.BlogProject{}, "tags_json")
}

// taxonomyCloud 从实体 JSON 字段聚合 taxonomy 计数（分类/标签共用）；仅统计已发布内容。
func (s *Service) taxonomyCloud(ctx context.Context, entity any, column string) ([]map[string]any, error) {
	// 1. 只拉取已发布行的 JSON 列，避免扫正文与草稿污染云图
	var values []datatypes.JSON
	if err := s.db.WithContext(ctx).Model(entity).
		Where("status = ?", StatusPublished).
		Pluck(column, &values).Error; err != nil {
		return nil, err
	}
	// 2. 统计各 uid 出现次数
	counts := map[string]int{}
	uids := make([]string, 0, len(values))
	for _, raw := range values {
		for _, uid := range model.JSONStrings(raw) {
			counts[uid]++
			uids = append(uids, uid)
		}
	}
	// 3. 批量解析名称并按 count 排序截断
	tax, err := s.taxonomyByUIDs(ctx, uids)
	if err != nil {
		return nil, err
	}
	names := map[string]string{}
	for uid, t := range tax {
		names[uid] = t.Name
	}
	return taxonomyListFromTaxons(counts, names), nil
}

func (s *Service) TaxonomyName(ctx context.Context, uid string) (string, error) {
	uid = strings.TrimSpace(uid)
	if uid == "" {
		return "", gorm.ErrRecordNotFound
	}
	var row model.BlogTaxonomy
	if err := s.db.WithContext(ctx).Where("uid = ?", uid).First(&row).Error; err != nil {
		return "", err
	}
	return row.Name, nil
}

const searchFetchCap = 300 // 相关性排序时每类最多拉多少条，避免大站全表进内存

// Search 站内关键词搜索：标题+摘要命中，按相关性排序后分页，并附带热门分类/标签。
func (s *Service) Search(ctx context.Context, q string, current, size int) (map[string]any, error) {
	start := time.Now()
	// 1. 规范化关键字与分页（去掉 LIKE 通配符）
	q = sanitizeLikeKeyword(q)
	current, size = page.Normalize(current, size, 12)
	covers := s.defaultCovers(ctx)
	var hits []map[string]any
	var total int64
	if q != "" {
		// 2. 分别 Count 文章/项目总数
		postQ := s.postsQuery(ctx, postFilter{KeywordLike: q})
		var postTotal int64
		if err := postQ.Count(&postTotal).Error; err != nil {
			return nil, err
		}
		projQ := s.projectsQuery(ctx, projectFilter{KeywordLike: q})
		var projTotal int64
		if err := projQ.Count(&projTotal).Error; err != nil {
			return nil, err
		}
		total = postTotal + projTotal

		// 3. 各类拉取上限条，合并后按相关性打分排序再切片（个人站体量足够）
		capN := searchFetchCap
		posts, err := s.fetchPosts(ctx, s.orderPosts(s.postsQuery(ctx, postFilter{KeywordLike: q})).Limit(capN), false)
		if err != nil {
			return nil, err
		}
		projects, err := s.fetchProjects(ctx, s.orderProjects(s.projectsQuery(ctx, projectFilter{KeywordLike: q})).Limit(capN), false)
		if err != nil {
			return nil, err
		}
		type scoredHit struct {
			score int
			hit   map[string]any
		}
		scored := make([]scoredHit, 0, len(posts)+len(projects))
		for i, p := range posts {
			scored = append(scored, scoredHit{
				score: searchRelevance(p.Title, p.Description, q),
				hit: map[string]any{
					"type": "post", "uid": p.UID, "title": p.Title,
					"description": p.Description, "snippet": searchSnippet(p.Description, q),
					"cover": coverOf(p.Cover, covers, i), "url": "/p/" + p.UID,
					"tags": taxonNames(p.Tags),
				},
			})
		}
		for i, p := range projects {
			scored = append(scored, scoredHit{
				score: searchRelevance(p.Title, p.Description, q),
				hit: map[string]any{
					"type": "project", "uid": p.UID, "title": p.Title,
					"description": p.Description, "snippet": searchSnippet(p.Description, q),
					"cover": coverOf(p.Cover, covers, i), "url": "/projects/" + p.UID,
					"tags": taxonNames(p.ProjectTags),
				},
			})
		}
		sort.SliceStable(scored, func(i, j int) bool {
			return scored[i].score > scored[j].score
		})
		hits = make([]map[string]any, len(scored))
		for i, sh := range scored {
			hits[i] = sh.hit
		}
		// 命中超过拉取上限时，总数与列表以实际可排序集合为准，避免深分页空页
		if int64(len(hits)) < total {
			total = int64(len(hits))
		}
	}
	// 4. 对合并结果做页切片，并附侧面 taxonomy 摘要
	pg := slicePage(hits, current, size)
	pg.Total = total
	if size > 0 && total > 0 {
		pg.Pages = int((total + int64(size) - 1) / int64(size))
	}
	cats, _ := s.Categories(ctx)
	tags, _ := s.Tags(ctx)
	if len(cats) > 8 {
		cats = cats[:8]
	}
	if len(tags) > 12 {
		tags = tags[:12]
	}
	return map[string]any{
		"records": pg.Records,
		"total":   pg.Total,
		"size":    pg.Size,
		"current": pg.Current,
		"pages":   pg.Pages,
		"took_ms": time.Since(start).Milliseconds(),
		"suggestions": map[string]any{
			"categories": cats,
			"tags":       tags,
		},
	}, nil
}

// searchRelevance 标题命中优先于摘要；前缀/全等再加权。
func searchRelevance(title, desc, kw string) int {
	if kw == "" {
		return 0
	}
	t := strings.ToLower(title)
	d := strings.ToLower(desc)
	score := 0
	if t == kw {
		score += 200
	} else if strings.HasPrefix(t, kw) {
		score += 150
	} else if strings.Contains(t, kw) {
		score += 100
	}
	if strings.Contains(d, kw) {
		score += 30
	}
	return score
}

// searchSnippet 从摘要中截取含关键词的短片段，便于结果列表预览。
func searchSnippet(desc, kw string) string {
	desc = strings.TrimSpace(desc)
	if desc == "" {
		return ""
	}
	runes := []rune(desc)
	if kw == "" {
		if len(runes) > 120 {
			return string(runes[:120]) + "…"
		}
		return desc
	}
	lower := strings.ToLower(desc)
	idx := strings.Index(lower, kw)
	if idx < 0 {
		if len(runes) > 120 {
			return string(runes[:120]) + "…"
		}
		return desc
	}
	// 按字节位置换算到 rune，再取前后文
	prefix := []rune(desc[:idx])
	start := len(prefix) - 40
	if start < 0 {
		start = 0
	}
	end := len(prefix) + len([]rune(kw)) + 80
	if end > len(runes) {
		end = len(runes)
	}
	out := string(runes[start:end])
	if start > 0 {
		out = "…" + out
	}
	if end < len(runes) {
		out += "…"
	}
	return out
}

func taxonNames(list []Taxon) []string {
	out := make([]string, 0, len(list))
	for _, t := range list {
		if n := strings.TrimSpace(t.Name); n != "" {
			out = append(out, n)
		}
	}
	return out
}

func slicePage[T any](items []T, current, size int) page.IPage {
	total := int64(len(items))
	if size <= 0 {
		size = 12
	}
	if current <= 0 {
		current = 1
	}
	pages := 0
	if total > 0 {
		pages = int((total + int64(size) - 1) / int64(size))
	}
	if pages > 0 && current > pages {
		current = pages
	}
	start := (current - 1) * size
	if start > len(items) {
		start = len(items)
	}
	end := start + size
	if end > len(items) {
		end = len(items)
	}
	var records any = items[start:end]
	if total == 0 {
		records = []T{}
	}
	return page.IPage{Records: records, Total: total, Size: size, Current: current, Pages: pages}
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
