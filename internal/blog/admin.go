package blog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/jiangbyte/go-blog/internal/file"
	"github.com/jiangbyte/go-blog/internal/model"
	"github.com/jiangbyte/go-blog/internal/pkg/batch"
	"github.com/jiangbyte/go-blog/internal/pkg/page"
	"github.com/jiangbyte/go-blog/internal/pkg/wirejson"
)

type PostWriteInput struct {
	UID          string        `json:"uid"`
	Title        string        `json:"title"`
	Description  string        `json:"description"`
	Date         string        `json:"date"`
	Status       string        `json:"status"`
	Cover        string        `json:"cover"`
	Locate       string        `json:"locate"`
	CategoryUIDs []string      `json:"category_uids"`
	TagUIDs      []string      `json:"tag_uids"`
	Reprint      wirejson.Bool `json:"reprint"`
	Sticky       wirejson.Bool `json:"sticky"`
	Home         *wirejson.Bool `json:"home"`
	Body         string        `json:"body"`
}

type ProjectWriteInput struct {
	UID         string         `json:"uid"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Date        string         `json:"date"`
	Status      string         `json:"status"`
	Progress    string         `json:"progress"`
	Weight      *wirejson.Int  `json:"weight"`
	Sticky      wirejson.Bool  `json:"sticky"`
	Cover       string         `json:"cover"`
	Git         string         `json:"git"`
	Site        string         `json:"site"`
	TagUIDs     []string       `json:"tag_uids"`
	Body        string         `json:"body"`
}

func (s *Service) AdminPagePosts(ctx context.Context, current, size int, title, status string) (page.IPage, error) {
	current, size = page.Normalize(current, size, 12)
	q := s.db.WithContext(ctx).Model(&model.BlogPost{})
	if t := strings.TrimSpace(title); t != "" {
		q = q.Where("title LIKE ?", "%"+t+"%")
	}
	if st := strings.ToLower(strings.TrimSpace(status)); st == StatusDraft || st == StatusPublished {
		q = q.Where("status = ?", st)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return page.IPage{}, err
	}
	var rows []model.BlogPost
	err := q.Omit("body").
		Order("sticky DESC").Order("published_at DESC").Order("id DESC").
		Offset((current - 1) * size).Limit(size).
		Find(&rows).Error
	if err != nil {
		return page.IPage{}, err
	}
	tax, err := s.taxonomyByUIDs(ctx, collectUIDsFromPosts(rows))
	if err != nil {
		return page.IPage{}, err
	}
	records := make([]Post, 0, len(rows))
	for _, r := range rows {
		records = append(records, postFromModel(r, false, tax))
	}
	return page.Of(records, total, current, size), nil
}

func (s *Service) AdminGetPost(ctx context.Context, uid string) (*Post, error) {
	// 管理端回显原始 object_name / 外链，避免把预签名 URL 写回库
	uid = strings.TrimSpace(uid)
	var row model.BlogPost
	if err := s.db.WithContext(ctx).Where("uid = ?", uid).First(&row).Error; err != nil {
		return nil, err
	}
	tax, err := s.taxonomyByUIDs(ctx, collectUIDsFromPosts([]model.BlogPost{row}))
	if err != nil {
		return nil, err
	}
	p := postFromModel(row, true, tax)
	return &p, nil
}

// AdminFindPostsByTitle 按标题精确匹配文章（供 CLI 重名检测），按更新日与 id 倒序。
func (s *Service) AdminFindPostsByTitle(ctx context.Context, title, status string) ([]Post, error) {
	// 1. 标题必填；可选 status 过滤草稿/已发布
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, gorm.ErrInvalidData
	}
	q := s.db.WithContext(ctx).Model(&model.BlogPost{}).Where("title = ?", title)
	if st := strings.ToLower(strings.TrimSpace(status)); st == StatusDraft || st == StatusPublished {
		q = q.Where("status = ?", st)
	}
	// 2. 省略正文，限制条数，取最近更新优先
	var rows []model.BlogPost
	if err := q.Omit("body").
		Order("updated_at DESC").Order("id DESC").
		Limit(20).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	// 3. 批量加载 taxonomy 并装配 DTO
	tax, err := s.taxonomyByUIDs(ctx, collectUIDsFromPosts(rows))
	if err != nil {
		return nil, err
	}
	out := make([]Post, 0, len(rows))
	for _, r := range rows {
		out = append(out, postFromModel(r, false, tax))
	}
	return out, nil
}

// AdminCreatePost 新建文章：校验并装配模型后插入，再回读详情。
func (s *Service) AdminCreatePost(ctx context.Context, in PostWriteInput) (*Post, error) {
	// 1. 由入参装配模型（标题/taxonomy 校验）
	row, err := s.buildPostModel(ctx, in, nil)
	if err != nil {
		return nil, err
	}
	// 2. 插入并回读，保证返回形态与详情接口一致
	if err := s.db.WithContext(ctx).Create(row).Error; err != nil {
		return nil, err
	}
	// 3. 已发布则异步 IndexNow（事务外远程 IO）
	if row.Status == StatusPublished {
		s.notifyIndexNowAsync("/p/" + row.UID)
	}
	return s.AdminGetPost(ctx, row.UID)
}

// AdminUpdatePost 更新文章：校验 taxonomy 后保留主键与创建时间再 Save。
func (s *Service) AdminUpdatePost(ctx context.Context, in PostWriteInput) (*Post, error) {
	// 1. uid 必填并加载原行
	uid := strings.TrimSpace(in.UID)
	if uid == "" {
		return nil, gorm.ErrInvalidData
	}
	var existing model.BlogPost
	if err := s.db.WithContext(ctx).Where("uid = ?", uid).First(&existing).Error; err != nil {
		return nil, err
	}
	// 2. 由入参装配模型（含分类/标签校验）
	row, err := s.buildPostModel(ctx, in, &existing)
	if err != nil {
		return nil, err
	}
	// 3. 保留不可变字段后落库，再回读详情
	row.ID = existing.ID
	row.UID = existing.UID
	row.CreatedAt = existing.CreatedAt
	if err := s.db.WithContext(ctx).Save(row).Error; err != nil {
		return nil, err
	}
	// 4. 已发布内容变更时通知搜索引擎尽快抓取
	if row.Status == StatusPublished {
		s.notifyIndexNowAsync("/p/" + uid)
	}
	return s.AdminGetPost(ctx, uid)
}

func (s *Service) AdminRemovePosts(ctx context.Context, uids []string) error {
	uids = batch.UniqueStrings(uids)
	if len(uids) == 0 {
		return nil
	}
	return s.deleteByUIDInBatches(ctx, &model.BlogPost{}, uids)
}

// AdminBatchSetPostStatus 批量改文章状态：一次 IN 加载、分批 UPDATE，禁止前端逐条 detail/update。
func (s *Service) AdminBatchSetPostStatus(ctx context.Context, uids []string, status string) (*BatchStatusResult, error) {
	return s.batchSetContentStatus(ctx, &model.BlogPost{}, uids, status)
}

// AdminBatchSetProjectStatus 批量改项目状态：一次 IN 加载、分批 UPDATE。
func (s *Service) AdminBatchSetProjectStatus(ctx context.Context, uids []string, status string) (*BatchStatusResult, error) {
	return s.batchSetContentStatus(ctx, &model.BlogProject{}, uids, status)
}

// batchSetContentStatus 对文章/项目按 uid 批量改 draft|published。
func (s *Service) batchSetContentStatus(ctx context.Context, modelPtr any, uids []string, status string) (*BatchStatusResult, error) {
	// 1. 规范化状态与 uid 列表
	status = strings.ToLower(strings.TrimSpace(status))
	if status != StatusDraft && status != StatusPublished {
		return nil, gorm.ErrInvalidData
	}
	uids = batch.UniqueStrings(uids)
	out := &BatchStatusResult{Failed: []BatchStatusFail{}}
	if len(uids) == 0 {
		return out, nil
	}

	// 2. 分批 IN 查出 uid/title/status/body（禁止循环单条）
	type statusRow struct {
		UID    string
		Title  string
		Status string
		Body   string
	}
	found := make(map[string]statusRow, len(uids))
	for _, chunk := range batch.ChunkStrings(uids, batch.MaxINSize) {
		var part []statusRow
		if err := s.db.WithContext(ctx).Model(modelPtr).
			Select("uid", "title", "status", "body").
			Where("uid IN ?", chunk).
			Find(&part).Error; err != nil {
			return nil, err
		}
		for _, r := range part {
			found[r.UID] = r
		}
	}

	// 3. 分类：已是目标状态 skip；发布时空正文 fail；其余待更新
	toUpdate := make([]string, 0, len(uids))
	for _, uid := range uids {
		r, ok := found[uid]
		if !ok {
			out.Failed = append(out.Failed, BatchStatusFail{UID: uid, Reason: "不存在"})
			continue
		}
		if r.Status == status {
			out.Skipped++
			continue
		}
		if status == StatusPublished && strings.TrimSpace(r.Body) == "" {
			out.Failed = append(out.Failed, BatchStatusFail{
				UID: uid, Title: r.Title, Reason: "正文为空，无法发布",
			})
			continue
		}
		toUpdate = append(toUpdate, uid)
	}

	// 4. 分批 UPDATE status；文章另写内容更新日期（项目表无该列）
	if len(toUpdate) > 0 {
		updates := map[string]any{"status": status}
		if _, ok := modelPtr.(*model.BlogPost); ok {
			now := time.Now()
			today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			updates["updated_at"] = today
		}
		for _, chunk := range batch.ChunkStrings(toUpdate, batch.MaxINSize) {
			if err := s.db.WithContext(ctx).Model(modelPtr).
				Where("uid IN ?", chunk).
				Updates(updates).Error; err != nil {
				return nil, err
			}
		}
	}
	out.Updated = len(toUpdate)
	// 5. 批量发布成功后异步 IndexNow（路径按实体类型区分）
	if status == StatusPublished && len(toUpdate) > 0 {
		prefix := "/p/"
		if _, ok := modelPtr.(*model.BlogProject); ok {
			prefix = "/projects/"
		}
		paths := make([]string, len(toUpdate))
		for i, uid := range toUpdate {
			paths[i] = prefix + uid
		}
		s.notifyIndexNowAsync(paths...)
	}
	return out, nil
}

func (s *Service) buildPostModel(ctx context.Context, in PostWriteInput, existing *model.BlogPost) (*model.BlogPost, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, gorm.ErrInvalidData
	}
	catUIDs, err := s.validateTaxonomyUIDs(ctx, "category", in.CategoryUIDs)
	if err != nil {
		return nil, err
	}
	tagUIDs, err := s.validateTaxonomyUIDs(ctx, "tag", in.TagUIDs)
	if err != nil {
		return nil, err
	}
	catJSON, _ := json.Marshal(catUIDs)
	tagJSON, _ := json.Marshal(tagUIDs)
	published := parseDateOrToday(in.Date)
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	home := true
	if in.Home != nil {
		home = in.Home.Bool()
	} else if existing != nil {
		home = existing.HomeVisible
	}
	words := utf8.RuneCountInString(stripMD(in.Body))
	reading := words / 400
	if reading < 1 && words > 0 {
		reading = 1
	}
	existingStatus := ""
	if existing != nil {
		existingStatus = existing.Status
	}
	row := &model.BlogPost{
		Title:              title,
		Description:        strings.TrimSpace(in.Description),
		PublishedAt:        published,
		Status:             normalizeContentStatus(in.Status, existingStatus, existing == nil),
		ContentUpdatedAt:   &today,
		Cover:              file.PersistRef(in.Cover),
		CategoriesJSON:     datatypes.JSON(catJSON),
		TagsJSON:           datatypes.JSON(tagJSON),
		Reprint:            in.Reprint.Bool(),
		Sticky:             in.Sticky.Bool(),
		HomeVisible:        home,
		ReadingTimeMinutes: reading,
		WordCount:          words,
		Body:               in.Body,
	}
	// 管理端已去掉归档路径编辑；更新时保留原值
	if existing != nil {
		row.UID = existing.UID
		row.Locate = existing.Locate
	} else {
		row.Locate = strings.TrimSpace(in.Locate)
	}
	return row, nil
}

func (s *Service) AdminPageProjects(ctx context.Context, current, size int, title, status, progress string) (page.IPage, error) {
	current, size = page.Normalize(current, size, 12)
	q := s.db.WithContext(ctx).Model(&model.BlogProject{})
	if t := strings.TrimSpace(title); t != "" {
		q = q.Where("title LIKE ?", "%"+t+"%")
	}
	if st := strings.ToLower(strings.TrimSpace(status)); st == StatusDraft || st == StatusPublished {
		q = q.Where("status = ?", st)
	}
	if strings.TrimSpace(progress) != "" {
		q = q.Where("progress = ?", normalizeProjectProgress(progress))
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return page.IPage{}, err
	}
	var rows []model.BlogProject
	err := q.Omit("body").
		Order("sticky DESC").Order("weight ASC").Order("published_at DESC").Order("id DESC").
		Offset((current - 1) * size).Limit(size).
		Find(&rows).Error
	if err != nil {
		return page.IPage{}, err
	}
	tax, err := s.taxonomyByUIDs(ctx, collectUIDsFromProjects(rows))
	if err != nil {
		return page.IPage{}, err
	}
	records := make([]Project, 0, len(rows))
	for _, r := range rows {
		records = append(records, projectFromModel(r, false, tax))
	}
	return page.Of(records, total, current, size), nil
}

func (s *Service) AdminGetProject(ctx context.Context, uid string) (*Project, error) {
	uid = strings.TrimSpace(uid)
	var row model.BlogProject
	if err := s.db.WithContext(ctx).Where("uid = ?", uid).First(&row).Error; err != nil {
		return nil, err
	}
	tax, err := s.taxonomyByUIDs(ctx, collectUIDsFromProjects([]model.BlogProject{row}))
	if err != nil {
		return nil, err
	}
	p := projectFromModel(row, true, tax)
	return &p, nil
}

func (s *Service) AdminCreateProject(ctx context.Context, in ProjectWriteInput) (*Project, error) {
	// 1. 装配并插入
	row, err := s.buildProjectModel(ctx, in, nil)
	if err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Create(row).Error; err != nil {
		return nil, err
	}
	// 2. 已发布则异步 IndexNow
	if row.Status == StatusPublished {
		s.notifyIndexNowAsync("/projects/" + row.UID)
	}
	return s.AdminGetProject(ctx, row.UID)
}

func (s *Service) AdminUpdateProject(ctx context.Context, in ProjectWriteInput) (*Project, error) {
	// 1. 加载原行并装配
	uid := strings.TrimSpace(in.UID)
	if uid == "" {
		return nil, gorm.ErrInvalidData
	}
	var existing model.BlogProject
	if err := s.db.WithContext(ctx).Where("uid = ?", uid).First(&existing).Error; err != nil {
		return nil, err
	}
	row, err := s.buildProjectModel(ctx, in, &existing)
	if err != nil {
		return nil, err
	}
	// 2. 落库
	row.ID = existing.ID
	row.UID = existing.UID
	row.CreatedAt = existing.CreatedAt
	if err := s.db.WithContext(ctx).Save(row).Error; err != nil {
		return nil, err
	}
	// 3. 已发布则通知搜索引擎
	if row.Status == StatusPublished {
		s.notifyIndexNowAsync("/projects/" + uid)
	}
	return s.AdminGetProject(ctx, uid)
}

func (s *Service) AdminRemoveProjects(ctx context.Context, uids []string) error {
	uids = batch.UniqueStrings(uids)
	if len(uids) == 0 {
		return nil
	}
	return s.deleteByUIDInBatches(ctx, &model.BlogProject{}, uids)
}

func (s *Service) buildProjectModel(ctx context.Context, in ProjectWriteInput, existing *model.BlogProject) (*model.BlogProject, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, gorm.ErrInvalidData
	}
	tagUIDs, err := s.validateTaxonomyUIDs(ctx, "project_tag", in.TagUIDs)
	if err != nil {
		return nil, err
	}
	tagJSON, _ := json.Marshal(tagUIDs)
	weight := 100
	if in.Weight != nil {
		weight = in.Weight.Int()
	} else if existing != nil {
		weight = existing.Weight
	}
	existingStatus := ""
	if existing != nil {
		existingStatus = existing.Status
	}
	progress := strings.TrimSpace(in.Progress)
	if progress == "" && existing != nil {
		progress = existing.Progress
	}
	row := &model.BlogProject{
		Title:       title,
		Description: strings.TrimSpace(in.Description),
		PublishedAt: parseDateOrToday(in.Date),
		Status:      normalizeContentStatus(in.Status, existingStatus, existing == nil),
		Progress:    normalizeProjectProgress(progress),
		Weight:      weight,
		Sticky:      in.Sticky.Bool(),
		Cover:       file.PersistRef(in.Cover),
		Git:         strings.TrimSpace(in.Git),
		Site:        strings.TrimSpace(in.Site),
		TagsJSON:    datatypes.JSON(tagJSON),
		Body:        in.Body,
	}
	if existing != nil {
		row.UID = existing.UID
	}
	return row, nil
}

func (s *Service) AdminGetSite(ctx context.Context) (map[string]any, error) {
	var row model.BlogSite
	if err := s.db.WithContext(ctx).First(&row, 1).Error; err != nil {
		return nil, err
	}
	cfg, err := model.JSONMap(row.ConfigJSON)
	if err != nil {
		return nil, err
	}
	stripAuthorIdentity(cfg)
	return cfg, nil
}

func (s *Service) AdminUpdateSite(ctx context.Context, cfg map[string]any) (map[string]any, error) {
	if cfg == nil {
		cfg = map[string]any{}
	}
	// 保留库里作者卡的展示壳（背景 / 俏皮话 / 链接），身份字段一律走个人中心
	var prev model.BlogSite
	if err := s.db.WithContext(ctx).First(&prev, 1).Error; err == nil {
		if prevCfg, err := model.JSONMap(prev.ConfigJSON); err == nil {
			if prevAuthor, ok := prevCfg["author"].(map[string]any); ok {
				author, _ := cfg["author"].(map[string]any)
				if author == nil {
					author = map[string]any{}
				}
				for _, key := range []string{"background_url", "witty_words", "links"} {
					if _, has := author[key]; !has {
						if v, ok := prevAuthor[key]; ok {
							author[key] = v
						}
					}
				}
				cfg["author"] = author
			}
		}
	}
	stripAuthorIdentity(cfg)
	sanitizeSiteMediaRefs(cfg)
	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	row := model.BlogSite{ID: 1, ConfigJSON: datatypes.JSON(raw)}
	if err := s.db.WithContext(ctx).Save(&row).Error; err != nil {
		return nil, err
	}
	out, err := model.JSONMap(row.ConfigJSON)
	if err != nil {
		return nil, err
	}
	stripAuthorIdentity(out)
	return out, nil
}

func sanitizeSiteMediaRefs(cfg map[string]any) {
	if cfg == nil {
		return
	}
	if raw, ok := cfg["default_covers"].([]any); ok {
		out := make([]any, 0, len(raw))
		for _, v := range raw {
			if str, ok := v.(string); ok {
				if ref := file.PersistRef(str); ref != "" {
					out = append(out, ref)
				}
			}
		}
		cfg["default_covers"] = out
	}
	if author, ok := cfg["author"].(map[string]any); ok && author != nil {
		if bg, ok := author["background_url"].(string); ok {
			author["background_url"] = file.PersistRef(bg)
		}
		cfg["author"] = author
	}
	if footer, ok := cfg["footer"].(map[string]any); ok && footer != nil {
		if qr, ok := footer["wechat_qr"].(string); ok {
			footer["wechat_qr"] = file.PersistRef(qr)
		}
		cfg["footer"] = footer
	}
}

func stripAuthorIdentity(cfg map[string]any) {
	if cfg == nil {
		return
	}
	author, ok := cfg["author"].(map[string]any)
	if !ok || author == nil {
		return
	}
	delete(author, "name")
	delete(author, "avatar_url")
	delete(author, "description")
	delete(author, "content")
	cfg["author"] = author
}

func (s *Service) AdminGetAbout(ctx context.Context) (map[string]any, error) {
	var row model.BlogAbout
	if err := s.db.WithContext(ctx).First(&row, 1).Error; err != nil {
		return nil, err
	}
	doc, err := model.JSONMap(row.DocJSON)
	if err != nil {
		return nil, err
	}
	return normalizeAboutDoc(doc), nil
}

func (s *Service) AdminUpdateAbout(ctx context.Context, doc map[string]any) (map[string]any, error) {
	if doc == nil {
		doc = map[string]any{}
	}
	out := normalizeAboutDoc(doc)
	raw, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	row := model.BlogAbout{ID: 1, DocJSON: datatypes.JSON(raw)}
	if err := s.db.WithContext(ctx).Save(&row).Error; err != nil {
		return nil, err
	}
	saved, err := model.JSONMap(row.DocJSON)
	if err != nil {
		return nil, err
	}
	return normalizeAboutDoc(saved), nil
}

// normalizeAboutDoc 统一为 { title, content }；旧 sections 结构会拼成一份 Markdown。
func normalizeAboutDoc(doc map[string]any) map[string]any {
	title := "关于"
	if doc != nil {
		if t, ok := doc["title"].(string); ok && strings.TrimSpace(t) != "" {
			title = strings.TrimSpace(t)
		}
		if c, ok := doc["content"].(string); ok && strings.TrimSpace(c) != "" {
			return map[string]any{"title": title, "content": c}
		}
	}
	return map[string]any{
		"title":   title,
		"content": aboutSectionsToMarkdown(doc),
	}
}

func aboutSectionsToMarkdown(doc map[string]any) string {
	if doc == nil {
		return ""
	}
	sections, _ := doc["sections"].([]any)
	if len(sections) == 0 {
		return ""
	}
	var parts []string
	for _, raw := range sections {
		sec, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		switch strings.TrimSpace(fmt.Sprint(sec["type"])) {
		case "intro":
			if md := aboutIntroToMarkdown(sec); md != "" {
				parts = append(parts, md)
			}
		case "values":
			if md := aboutValuesToMarkdown(sec); md != "" {
				parts = append(parts, md)
			}
		case "story":
			if md := aboutStoryToMarkdown(sec); md != "" {
				parts = append(parts, md)
			}
		}
	}
	return strings.Join(parts, "\n\n---\n\n")
}

func aboutIntroToMarkdown(sec map[string]any) string {
	if c, ok := sec["content"].(string); ok && strings.TrimSpace(c) != "" {
		return strings.TrimSpace(c)
	}
	var lines []string
	if g, ok := sec["greeting"].(string); ok && strings.TrimSpace(g) != "" {
		lines = append(lines, strings.TrimSpace(g))
	}
	prefix, _ := sec["name_prefix"].(string)
	name, _ := sec["name"].(string)
	prefix = strings.TrimSpace(prefix)
	name = strings.TrimSpace(name)
	if prefix != "" || name != "" {
		if name != "" {
			lines = append(lines, strings.TrimSpace(prefix+" **"+name+"**"))
		} else {
			lines = append(lines, prefix)
		}
	}
	if d, ok := sec["description"].(string); ok && strings.TrimSpace(d) != "" {
		lines = append(lines, strings.TrimSpace(d))
	}
	if f, ok := sec["focus"].(string); ok && strings.TrimSpace(f) != "" {
		lines = append(lines, strings.TrimSpace(f))
	} else if pursuit, ok := sec["pursuit"].(map[string]any); ok {
		label, _ := pursuit["label"].(string)
		var words []string
		if arr, ok := pursuit["lines"].([]any); ok {
			for _, x := range arr {
				if s := strings.TrimSpace(fmt.Sprint(x)); s != "" {
					words = append(words, s)
				}
			}
		}
		if arr, ok := pursuit["rotating_words"].([]any); ok {
			for _, x := range arr {
				if s := strings.TrimSpace(fmt.Sprint(x)); s != "" {
					words = append(words, s)
				}
			}
		}
		body := strings.Join(words, " · ")
		label = strings.TrimSpace(label)
		if label != "" && body != "" {
			lines = append(lines, label+" **"+body+"**")
		} else if body != "" {
			lines = append(lines, body)
		} else if label != "" {
			lines = append(lines, label)
		}
	}
	return strings.Join(lines, "\n\n")
}

func aboutValuesToMarkdown(sec map[string]any) string {
	var blocks []string
	if motto, ok := sec["motto"].(map[string]any); ok {
		label, _ := motto["label"].(string)
		text, _ := motto["text"].(string)
		if strings.TrimSpace(text) == "" {
			text = strings.TrimSpace(strings.Join([]string{
				strings.TrimSpace(fmt.Sprint(motto["prefix"])),
				strings.TrimSpace(fmt.Sprint(motto["content"])),
			}, "，"))
		}
		label = strings.TrimSpace(label)
		text = strings.TrimSpace(text)
		if label != "" && text != "" {
			blocks = append(blocks, "### "+label+"\n\n"+text)
		} else if text != "" {
			blocks = append(blocks, text)
		}
	}
	if expertise, ok := sec["expertise"].(map[string]any); ok {
		label, _ := expertise["label"].(string)
		text, _ := expertise["text"].(string)
		if strings.TrimSpace(text) == "" {
			head := strings.TrimSpace(fmt.Sprint(expertise["prefix"])) + strings.TrimSpace(fmt.Sprint(expertise["specialist"]))
			tail := strings.TrimSpace(strings.Join([]string{
				strings.TrimSpace(fmt.Sprint(expertise["content"])),
				strings.TrimSpace(fmt.Sprint(expertise["level"])),
			}, " · "))
			text = strings.TrimSpace(strings.Join([]string{head, tail}, "，"))
		}
		label = strings.TrimSpace(label)
		text = strings.TrimSpace(text)
		if label != "" && text != "" {
			blocks = append(blocks, "### "+label+"\n\n"+text)
		} else if text != "" {
			blocks = append(blocks, text)
		}
	}
	return strings.Join(blocks, "\n\n")
}

func aboutStoryToMarkdown(sec map[string]any) string {
	content, _ := sec["content"].(string)
	content = strings.TrimSpace(content)
	label, _ := sec["label"].(string)
	title, _ := sec["title"].(string)
	label = strings.TrimSpace(label)
	title = strings.TrimSpace(title)
	if content != "" {
		if strings.Contains(content, "#") || (label == "" && title == "") {
			return content
		}
		var parts []string
		if label != "" {
			parts = append(parts, "*"+label+"*")
		}
		if title != "" {
			parts = append(parts, "## "+title)
		}
		parts = append(parts, "", content)
		return strings.Join(parts, "\n")
	}
	var parts []string
	if label != "" {
		parts = append(parts, "*"+label+"*")
	}
	if title != "" {
		parts = append(parts, "## "+title)
	}
	return strings.Join(parts, "\n")
}

func (s *Service) AdminTaxonomyOptions(ctx context.Context, kind string) ([]map[string]any, error) {
	kind = strings.TrimSpace(kind)
	if kind == "" {
		return nil, gorm.ErrInvalidData
	}
	var rows []model.BlogTaxonomy
	if err := s.db.WithContext(ctx).
		Where("kind = ?", kind).
		Order("name ASC").
		Limit(page.MaxSize).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]any{"uid": r.UID, "name": r.Name})
	}
	return out, nil
}

type TaxonomyWriteInput struct {
	UID  string `json:"uid"`
	Kind string `json:"kind"`
	Name string `json:"name"`
}

func (s *Service) AdminPageTaxonomy(ctx context.Context, kind, name string, current, size int) (page.IPage, error) {
	// 1. 分页查出当前 kind 的 taxonomy
	current, size = page.Normalize(current, size, 12)
	kind = strings.TrimSpace(kind)
	if kind == "" {
		return page.IPage{}, gorm.ErrInvalidData
	}
	q := s.db.WithContext(ctx).Model(&model.BlogTaxonomy{}).Where("kind = ?", kind)
	if n := strings.TrimSpace(name); n != "" {
		q = q.Where("name LIKE ?", "%"+n+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return page.IPage{}, err
	}
	var rows []model.BlogTaxonomy
	if err := q.Order("name ASC").Offset((current - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return page.IPage{}, err
	}
	// 2. 批量统计本页 uid 被文章/项目引用次数（禁止按行 COUNT）
	pageUIDs := make([]string, 0, len(rows))
	for _, r := range rows {
		pageUIDs = append(pageUIDs, r.UID)
	}
	usage, err := s.taxonomyUsageCounts(ctx, kind, pageUIDs)
	if err != nil {
		return page.IPage{}, err
	}
	// 3. 装配列表：usage_count 为文章数（分类/标签）或项目数（项目标签）
	records := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		records = append(records, map[string]any{
			"uid":         r.UID,
			"kind":        r.Kind,
			"name":        r.Name,
			"usage_count": usage[r.UID],
			"created_at":  r.CreatedAt.Format(time.RFC3339),
		})
	}
	return page.Of(records, total, current, size), nil
}

// taxonomyUsageCounts 按 kind 扫对应 JSON 列，统计给定 uid 的引用次数。
func (s *Service) taxonomyUsageCounts(ctx context.Context, kind string, uids []string) (map[string]int, error) {
	out := make(map[string]int, len(uids))
	if len(uids) == 0 {
		return out, nil
	}
	want := make(map[string]struct{}, len(uids))
	for _, u := range uids {
		want[u] = struct{}{}
		out[u] = 0
	}
	inc := func(ids []string) {
		for _, id := range ids {
			if _, ok := want[id]; ok {
				out[id]++
			}
		}
	}

	switch kind {
	case "category":
		var values []datatypes.JSON
		if err := s.db.WithContext(ctx).Model(&model.BlogPost{}).
			Pluck("categories_json", &values).Error; err != nil {
			return nil, err
		}
		for _, raw := range values {
			inc(model.JSONStrings(raw))
		}
	case "tag":
		var values []datatypes.JSON
		if err := s.db.WithContext(ctx).Model(&model.BlogPost{}).
			Pluck("tags_json", &values).Error; err != nil {
			return nil, err
		}
		for _, raw := range values {
			inc(model.JSONStrings(raw))
		}
	case "project_tag":
		var values []datatypes.JSON
		if err := s.db.WithContext(ctx).Model(&model.BlogProject{}).
			Pluck("tags_json", &values).Error; err != nil {
			return nil, err
		}
		for _, raw := range values {
			inc(model.JSONStrings(raw))
		}
	default:
		return out, nil
	}
	return out, nil
}

// AdminCreateTaxonomy 创建 taxonomy；同 kind+name 已存在则返回已有行（输入创建 / ensure 语义）。
func (s *Service) AdminCreateTaxonomy(ctx context.Context, in TaxonomyWriteInput) (map[string]any, error) {
	// 1. 校验 kind / name
	kind := strings.TrimSpace(in.Kind)
	name := strings.TrimSpace(in.Name)
	if kind == "" || name == "" {
		return nil, gorm.ErrInvalidData
	}
	if !validTaxonomyKind(kind) {
		return nil, gorm.ErrInvalidData
	}
	// 2. 先查：同名直接复用，避免唯一键冲突
	var existing model.BlogTaxonomy
	err := s.db.WithContext(ctx).Where("kind = ? AND name = ?", kind, name).Take(&existing).Error
	if err == nil {
		return map[string]any{"uid": existing.UID, "kind": existing.Kind, "name": existing.Name}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	// 3. 插入；并发撞唯一键时再查一次返回
	row := model.BlogTaxonomy{Kind: kind, Name: name}
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		var again model.BlogTaxonomy
		if takeErr := s.db.WithContext(ctx).Where("kind = ? AND name = ?", kind, name).Take(&again).Error; takeErr == nil {
			return map[string]any{"uid": again.UID, "kind": again.Kind, "name": again.Name}, nil
		}
		return nil, err
	}
	return map[string]any{"uid": row.UID, "kind": row.Kind, "name": row.Name}, nil
}

func (s *Service) AdminUpdateTaxonomy(ctx context.Context, in TaxonomyWriteInput) (map[string]any, error) {
	uid := strings.TrimSpace(in.UID)
	name := strings.TrimSpace(in.Name)
	if uid == "" || name == "" {
		return nil, gorm.ErrInvalidData
	}
	var row model.BlogTaxonomy
	if err := s.db.WithContext(ctx).Where("uid = ?", uid).First(&row).Error; err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Model(&row).Update("name", name).Error; err != nil {
		return nil, err
	}
	row.Name = name
	return map[string]any{"uid": row.UID, "kind": row.Kind, "name": row.Name}, nil
}

func (s *Service) AdminRemoveTaxonomy(ctx context.Context, uids []string) error {
	uids = batch.UniqueStrings(uids)
	if len(uids) == 0 {
		return nil
	}
	var rows []model.BlogTaxonomy
	for _, chunk := range batch.ChunkStrings(uids, batch.DefaultSize) {
		var part []model.BlogTaxonomy
		if err := s.db.WithContext(ctx).Where("uid IN ?", chunk).Find(&part).Error; err != nil {
			return err
		}
		rows = append(rows, part...)
	}
	used, err := s.taxonomyUIDsInUse(ctx, uids)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if used[row.UID] {
			return fmt.Errorf("「%s」仍被文章或项目引用，无法删除", row.Name)
		}
	}
	return s.deleteByUIDInBatches(ctx, &model.BlogTaxonomy{}, uids)
}

// taxonomyUIDsInUse 两次全量扫 JSON 字段判定引用，避免按 uid 循环 COUNT（N+1）。
func (s *Service) taxonomyUIDsInUse(ctx context.Context, uids []string) (map[string]bool, error) {
	want := make(map[string]bool, len(uids))
	for _, u := range uids {
		want[u] = true
	}
	out := make(map[string]bool)
	mark := func(ids []string) {
		for _, id := range ids {
			if want[id] {
				out[id] = true
			}
		}
	}

	var posts []struct {
		CategoriesJSON datatypes.JSON
		TagsJSON       datatypes.JSON
	}
	if err := s.db.WithContext(ctx).Model(&model.BlogPost{}).
		Select("categories_json", "tags_json").
		Find(&posts).Error; err != nil {
		return nil, err
	}
	for _, p := range posts {
		mark(model.JSONStrings(p.CategoriesJSON))
		mark(model.JSONStrings(p.TagsJSON))
	}

	var projects []struct {
		TagsJSON datatypes.JSON
	}
	if err := s.db.WithContext(ctx).Model(&model.BlogProject{}).
		Select("tags_json").
		Find(&projects).Error; err != nil {
		return nil, err
	}
	for _, p := range projects {
		mark(model.JSONStrings(p.TagsJSON))
	}
	return out, nil
}

func (s *Service) deleteByUIDInBatches(ctx context.Context, modelPtr any, uids []string) error {
	for _, chunk := range batch.ChunkStrings(uids, batch.DefaultSize) {
		if err := s.db.WithContext(ctx).Where("uid IN ?", chunk).Delete(modelPtr).Error; err != nil {
			return err
		}
	}
	return nil
}

func validTaxonomyKind(kind string) bool {
	switch kind {
	case "category", "tag", "project_tag":
		return true
	default:
		return false
	}
}

func (s *Service) validateTaxonomyUIDs(ctx context.Context, kind string, uids []string) ([]string, error) {
	uids = batch.UniqueStrings(uids)
	if len(uids) == 0 {
		return []string{}, nil
	}
	ok := make(map[string]bool, len(uids))
	for _, chunk := range batch.ChunkStrings(uids, batch.DefaultSize) {
		var rows []model.BlogTaxonomy
		if err := s.db.WithContext(ctx).
			Where("kind = ? AND uid IN ?", kind, chunk).
			Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, r := range rows {
			ok[r.UID] = true
		}
	}
	out := make([]string, 0, len(uids))
	for _, uid := range uids {
		if ok[uid] {
			out = append(out, uid)
		}
	}
	return out, nil
}

func parseDateOrToday(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if raw == "" {
		return today
	}
	if t, err := time.ParseInLocation("2006-01-02", raw[:min(10, len(raw))], now.Location()); err == nil {
		return t
	}
	return today
}

func stripMD(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "```") || strings.HasPrefix(line, "---") {
			continue
		}
		line = strings.TrimLeft(line, "#>*+- ")
		b.WriteString(line)
		b.WriteByte(' ')
	}
	return strings.TrimSpace(b.String())
}
