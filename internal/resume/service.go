package resume

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/jiangbyte/go-blog/internal/file"
	"github.com/jiangbyte/go-blog/internal/model"
	"github.com/jiangbyte/go-blog/internal/pkg/idgen"
	"github.com/jiangbyte/go-blog/internal/pkg/page"
)

// Service 简历领域服务。
type Service struct {
	db    *gorm.DB
	files *file.Service
}

func NewService(db *gorm.DB, files *file.Service) *Service {
	return &Service{db: db, files: files}
}

// EnsureSchema 迁移简历表。
func (s *Service) EnsureSchema(ctx context.Context) error {
	return s.db.WithContext(ctx).AutoMigrate(&model.Resume{}, &model.ResumeSection{})
}

func (s *Service) resolveProfileMedia(ctx context.Context, profile map[string]any) map[string]any {
	// 1. 拷贝资料，避免改到缓存对象
	src := map[string]any{}
	for k, v := range profile {
		src[k] = v
	}
	raw := strings.TrimSpace(asString(src["avatar_url"]))
	if raw == "" {
		return src
	}
	// 2. 重新签发访问地址
	resolved := s.files.ResolveAccessURL(ctx, raw)
	src["avatar_url"] = SafeMediaURL(resolved)
	return src
}

func (s *Service) toResumeMap(ctx context.Context, po *model.Resume, sections []model.ResumeSection) map[string]any {
	theme := mustJSONMap(po.ThemeJSON, defaultTheme)
	profile := mustJSONMap(po.ProfileJSON, defaultProfile)
	custom := ""
	if po.CustomCSS != nil {
		custom = *po.CustomCSS
	}
	row := map[string]any{
		"id":           po.ID,
		"title":        po.Title,
		"template_id":  po.TemplateID,
		"theme_json":   theme,
		"custom_css":   custom,
		"profile_json": s.resolveProfileMedia(ctx, profile),
		"status":       po.Status,
		"published_at": fmtTime(po.PublishedAt),
		"pdf_file_id":  "",
		"sort_code":    itoa(po.SortCode),
		"created_at":   po.CreatedAt.UTC().Format(time.RFC3339Nano),
		"updated_at":   po.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
	if po.PdfFileID != nil {
		row["pdf_file_id"] = *po.PdfFileID
		if s.files != nil {
			if u := s.files.ResolveAccessURL(ctx, strings.TrimSpace(*po.PdfFileID)); u != "" {
				row["pdf_url"] = u
			}
		}
	}
	if sections != nil {
		out := make([]map[string]any, 0, len(sections))
		for i := range sections {
			out = append(out, toSectionMap(&sections[i]))
		}
		row["sections"] = out
	}
	return row
}

func toSectionMap(po *model.ResumeSection) map[string]any {
	meta := mustJSONMapPtr(po.MetaJSON)
	entries := mustJSONArrayPtr(po.EntriesJSON)
	icon := ""
	if po.Icon != nil {
		icon = *po.Icon
	}
	body := ""
	if po.BodyHTML != nil {
		body = *po.BodyHTML
	}
	return map[string]any{
		"id":           po.ID,
		"resume_id":    po.ResumeID,
		"type":         po.Type,
		"title":        po.Title,
		"icon":         icon,
		"sort_code":    itoa(po.SortCode),
		"meta_json":    normalizeMeta(meta),
		"body_html":    body,
		"entries_json": entries,
		"created_at":   po.CreatedAt.UTC().Format(time.RFC3339Nano),
		"updated_at":   po.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func itoa(n int) string {
	return strings.TrimSpace(asString(n))
}

func mustJSONMap(raw string, fallback map[string]any) map[string]any {
	if strings.TrimSpace(raw) == "" {
		return cloneMap(fallback)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil || m == nil {
		return cloneMap(fallback)
	}
	return m
}

func mustJSONMapPtr(raw *string) map[string]any {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return map[string]any{}
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(*raw), &m); err != nil || m == nil {
		return map[string]any{}
	}
	return m
}

func mustJSONArrayPtr(raw *string) []any {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return []any{}
	}
	var arr []any
	if err := json.Unmarshal([]byte(*raw), &arr); err != nil || arr == nil {
		return []any{}
	}
	return arr
}

func cloneMap(src map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range src {
		out[k] = v
	}
	return out
}

func encodeJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// Page 分页查询；附带本页 sections。
func (s *Service) Page(ctx context.Context, current, size int, status, title string) (map[string]any, error) {
	// 1. 查主表分页
	q := s.db.WithContext(ctx).Model(&model.Resume{})
	if st := strings.TrimSpace(status); st != "" {
		q = q.Where("status = ?", st)
	}
	if t := strings.TrimSpace(title); t != "" {
		q = q.Where("title LIKE ?", "%"+t+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	current, size = page.Normalize(current, size, 20)
	var rows []model.Resume
	if err := q.Select("id", "title", "template_id", "theme_json", "custom_css", "profile_json",
		"status", "published_at", "sort_code", "created_at", "updated_at").
		Order("updated_at DESC").Offset((current - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return nil, err
	}
	// 2. 按本页 id 一次 IN 拉分节
	ids := make([]string, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	byResume := map[string][]model.ResumeSection{}
	if len(ids) > 0 {
		var secs []model.ResumeSection
		if err := s.db.WithContext(ctx).
			Where("resume_id IN ?", ids).
			Order("sort_code ASC, created_at ASC").
			Find(&secs).Error; err != nil {
			return nil, err
		}
		for i := range secs {
			byResume[secs[i].ResumeID] = append(byResume[secs[i].ResumeID], secs[i])
		}
	}
	// 3. 组装 records
	records := make([]map[string]any, 0, len(rows))
	for i := range rows {
		records = append(records, s.toResumeMap(ctx, &rows[i], byResume[rows[i].ID]))
	}
	return map[string]any{
		"records": records,
		"total":   itoa(int(total)),
		"current": itoa(current),
		"size":    itoa(size),
	}, nil
}

func (s *Service) Detail(ctx context.Context, id string) (map[string]any, error) {
	po, err := s.getResume(ctx, id)
	if err != nil {
		return nil, err
	}
	secs, err := s.listSections(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toResumeMap(ctx, po, secs), nil
}

func (s *Service) getResume(ctx context.Context, id string) (*model.Resume, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrBadRequest("id required")
	}
	var po model.Resume
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return nil, ErrNotFound("简历不存在")
	}
	if err != nil {
		return nil, err
	}
	return &po, nil
}

func (s *Service) listSections(ctx context.Context, resumeID string) ([]model.ResumeSection, error) {
	var secs []model.ResumeSection
	err := s.db.WithContext(ctx).
		Where("resume_id = ?", resumeID).
		Order("sort_code ASC, created_at ASC").
		Find(&secs).Error
	return secs, err
}

func (s *Service) requireTemplate(templateID string) (string, error) {
	meta, err := ResolveTemplateForWrite(templateID)
	if err != nil {
		return "", err
	}
	return meta.ID, nil
}

// Create 新建未发布简历。
func (s *Service) Create(ctx context.Context, in ResumeWriteInput) (map[string]any, error) {
	// 1. 规范化入参
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, ErrBadRequest("标题不能为空")
	}
	tpl, err := s.requireTemplate(in.TemplateID)
	if err != nil {
		return nil, err
	}
	css, err := SanitizeCustomCSSOrError(in.CustomCSS)
	if err != nil {
		return nil, err
	}
	id, err := idgen.NewUID()
	if err != nil {
		return nil, err
	}
	// 2. 写入未发布简历
	po := model.Resume{
		ID:          id,
		Title:       truncate(title, 128, ""),
		TemplateID:  tpl,
		ThemeJSON:   encodeJSON(normalizeTheme(in.ThemeJSON)),
		CustomCSS:   css,
		ProfileJSON: encodeJSON(normalizeProfile(in.ProfileJSON)),
		Status:      model.ResumeStatusUnpublished,
		SortCode:    in.SortCode.Int(),
	}
	if err := s.db.WithContext(ctx).Create(&po).Error; err != nil {
		return nil, err
	}
	return s.toResumeMap(ctx, &po, []model.ResumeSection{}), nil
}

// Update 更新简历主表（发布状态走 Publish）。
func (s *Service) Update(ctx context.Context, in ResumeWriteInput) (map[string]any, error) {
	po, err := s.getResume(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, ErrBadRequest("标题不能为空")
	}
	tpl, err := s.requireTemplate(in.TemplateID)
	if err != nil {
		return nil, err
	}
	css, err := SanitizeCustomCSSOrError(in.CustomCSS)
	if err != nil {
		return nil, err
	}
	if in.Status != nil {
		st := strings.TrimSpace(*in.Status)
		if st != model.ResumeStatusUnpublished {
			return nil, ErrBadRequest("请使用发布接口变更发布状态")
		}
		po.Status = model.ResumeStatusUnpublished
		po.PublishedAt = nil
	}
	po.Title = truncate(title, 128, "")
	po.TemplateID = tpl
	po.ThemeJSON = encodeJSON(normalizeTheme(in.ThemeJSON))
	po.CustomCSS = css
	po.ProfileJSON = encodeJSON(normalizeProfile(in.ProfileJSON))
	po.SortCode = in.SortCode.Int()
	if err := s.db.WithContext(ctx).Save(po).Error; err != nil {
		return nil, err
	}
	secs, err := s.listSections(ctx, po.ID)
	if err != nil {
		return nil, err
	}
	return s.toResumeMap(ctx, po, secs), nil
}

// Remove 仅允许删除 unpublished。
func (s *Service) Remove(ctx context.Context, ids []string) error {
	unique := dedupeIDs(ids)
	if len(unique) == 0 {
		return nil
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var rows []model.Resume
		if err := tx.Select("id", "status").Where("id IN ?", unique).Find(&rows).Error; err != nil {
			return err
		}
		for _, r := range rows {
			if r.Status == model.ResumeStatusPublished {
				return ErrBadRequest("请先取消发布再删除")
			}
		}
		if err := tx.Where("resume_id IN ?", unique).Delete(&model.ResumeSection{}).Error; err != nil {
			return err
		}
		return tx.Where("id IN ?", unique).Delete(&model.Resume{}).Error
	})
}

// Publish 互斥发布：先全部撤回再发布当前篇，事务外生成公开 PDF。
func (s *Service) Publish(ctx context.Context, id string) (map[string]any, error) {
	po, err := s.getResume(ctx, id)
	if err != nil {
		return nil, err
	}
	profile := mustJSONMap(po.ProfileJSON, defaultProfile)
	if strings.TrimSpace(asString(profile["name"])) == "" {
		return nil, ErrBadRequest("发布前请填写姓名")
	}
	now := time.Now()
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 全部撤回
		if err := tx.Model(&model.Resume{}).
			Where("status = ?", model.ResumeStatusPublished).
			Updates(map[string]any{"status": model.ResumeStatusUnpublished, "published_at": nil}).Error; err != nil {
			return err
		}
		// 2. 发布当前
		po.Status = model.ResumeStatusPublished
		po.PublishedAt = &now
		return tx.Save(po).Error
	})
	if err != nil {
		return nil, err
	}
	// 3. 事务外生成 PDF 并上传 OSS（chromedp 长 IO）
	if err := s.RefreshPublishedPDF(ctx, id); err != nil {
		return nil, err
	}
	secs, err := s.listSections(ctx, po.ID)
	if err != nil {
		return nil, err
	}
	return s.toResumeMap(ctx, po, secs), nil
}

func (s *Service) Unpublish(ctx context.Context, id string) (map[string]any, error) {
	po, err := s.getResume(ctx, id)
	if err != nil {
		return nil, err
	}
	po.Status = model.ResumeStatusUnpublished
	po.PublishedAt = nil
	if err := s.db.WithContext(ctx).Save(po).Error; err != nil {
		return nil, err
	}
	// 撤回后删除 OSS 上的公开 PDF
	s.ClearPublishedPDF(ctx, id)
	secs, err := s.listSections(ctx, po.ID)
	if err != nil {
		return nil, err
	}
	return s.toResumeMap(ctx, po, secs), nil
}

// Copy 深拷贝简历（副本未发布）。
func (s *Service) Copy(ctx context.Context, id string) (map[string]any, error) {
	src, err := s.getResume(ctx, id)
	if err != nil {
		return nil, err
	}
	sections, err := s.listSections(ctx, id)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(src.Title)
	if title == "" {
		title = "简历"
	}
	if !strings.HasSuffix(title, "（副本）") {
		title = truncate(title+"（副本）", 128, "")
	}
	newID, err := idgen.NewUID()
	if err != nil {
		return nil, err
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		po := model.Resume{
			ID: newID, Title: title, TemplateID: src.TemplateID,
			ThemeJSON: src.ThemeJSON, CustomCSS: src.CustomCSS, ProfileJSON: src.ProfileJSON,
			Status: model.ResumeStatusUnpublished, SortCode: src.SortCode,
		}
		if err := tx.Create(&po).Error; err != nil {
			return err
		}
		for _, sec := range sections {
			sid, err := idgen.NewUID()
			if err != nil {
				return err
			}
			cp := model.ResumeSection{
				ID: sid, ResumeID: newID, Type: sec.Type, Title: sec.Title,
				Icon: sec.Icon, SortCode: sec.SortCode, MetaJSON: sec.MetaJSON,
				BodyHTML: sec.BodyHTML, EntriesJSON: sec.EntriesJSON,
			}
			if err := tx.Create(&cp).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.Detail(ctx, newID)
}

// SaveSection 新建或更新分节。
func (s *Service) SaveSection(ctx context.Context, in SectionSaveInput) (map[string]any, error) {
	if _, err := s.getResume(ctx, in.ResumeID); err != nil {
		return nil, err
	}
	secType := strings.TrimSpace(in.Type)
	if _, ok := sectionTypes[secType]; !ok {
		return nil, ErrBadRequest("未知分节类型: " + secType)
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, ErrBadRequest("分节标题不能为空")
	}
	meta := normalizeMeta(in.MetaJSON)
	entries := normalizeEntries(in.EntriesJSON)
	rawBody := ""
	if in.BodyHTML != nil {
		rawBody = *in.BodyHTML
	}
	body := truncate(SanitizeBodyHTML(rawBody), bodyHTMLMax, "")
	iconRaw := ""
	if in.Icon != nil {
		iconRaw = *in.Icon
	}
	iconNorm := normalizeIcon(iconRaw)
	var iconPtr *string
	if iconNorm != "" {
		iconPtr = &iconNorm
	}
	metaJSON := encodeJSON(meta)
	entriesJSON := encodeJSON(entries)
	bodyPtr := &body

	if sid := strings.TrimSpace(in.ID); sid != "" {
		var po model.ResumeSection
		err := s.db.WithContext(ctx).Where("id = ? AND resume_id = ?", sid, in.ResumeID).First(&po).Error
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound("分节不存在")
		}
		if err != nil {
			return nil, err
		}
		// 1. 只更新可变字段，Omit created_at，避免 Save 把零时间写回触发 MySQL 1292
		po.Type = secType
		po.Title = truncate(title, 128, "")
		po.Icon = iconPtr
		po.SortCode = in.SortCode.Int()
		po.MetaJSON = &metaJSON
		po.BodyHTML = bodyPtr
		po.EntriesJSON = &entriesJSON
		if err := s.db.WithContext(ctx).Omit("created_at").Save(&po).Error; err != nil {
			return nil, err
		}
		return toSectionMap(&po), nil
	}

	id, err := idgen.NewUID()
	if err != nil {
		return nil, err
	}
	po := model.ResumeSection{
		ID: id, ResumeID: in.ResumeID, Type: secType,
		Title: truncate(title, 128, ""), Icon: iconPtr, SortCode: in.SortCode.Int(),
		MetaJSON: &metaJSON, BodyHTML: bodyPtr, EntriesJSON: &entriesJSON,
	}
	if err := s.db.WithContext(ctx).Create(&po).Error; err != nil {
		return nil, err
	}
	return toSectionMap(&po), nil
}

func (s *Service) CopySection(ctx context.Context, resumeID, sectionID string) (map[string]any, error) {
	if _, err := s.getResume(ctx, resumeID); err != nil {
		return nil, err
	}
	var src model.ResumeSection
	err := s.db.WithContext(ctx).Where("id = ? AND resume_id = ?", sectionID, resumeID).First(&src).Error
	if err == gorm.ErrRecordNotFound {
		return nil, ErrNotFound("分节不存在")
	}
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(src.Title)
	if title == "" {
		title = "分节"
	}
	if !strings.HasSuffix(title, "（副本）") {
		title = truncate(title+"（副本）", 128, "")
	}
	insertAt := src.SortCode + 1
	var created model.ResumeSection
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.ResumeSection{}).
			Where("resume_id = ? AND sort_code >= ?", resumeID, insertAt).
			UpdateColumn("sort_code", gorm.Expr("sort_code + 1")).Error; err != nil {
			return err
		}
		id, err := idgen.NewUID()
		if err != nil {
			return err
		}
		created = model.ResumeSection{
			ID: id, ResumeID: resumeID, Type: src.Type, Title: title,
			Icon: src.Icon, SortCode: insertAt, MetaJSON: src.MetaJSON,
			BodyHTML: src.BodyHTML, EntriesJSON: src.EntriesJSON,
		}
		return tx.Create(&created).Error
	})
	if err != nil {
		return nil, err
	}
	return toSectionMap(&created), nil
}

func (s *Service) ReorderSections(ctx context.Context, resumeID string, ids []string) ([]map[string]any, error) {
	if _, err := s.getResume(ctx, resumeID); err != nil {
		return nil, err
	}
	unique := dedupeIDs(ids)
	secs, err := s.listSections(ctx, resumeID)
	if err != nil {
		return nil, err
	}
	byID := map[string]*model.ResumeSection{}
	for i := range secs {
		byID[secs[i].ID] = &secs[i]
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		order := 10
		seen := map[string]struct{}{}
		for _, sid := range unique {
			po := byID[sid]
			if po == nil {
				continue
			}
			if err := tx.Model(po).Update("sort_code", order).Error; err != nil {
				return err
			}
			seen[sid] = struct{}{}
			order += 10
		}
		for i := range secs {
			if _, ok := seen[secs[i].ID]; ok {
				continue
			}
			if err := tx.Model(&secs[i]).Update("sort_code", order).Error; err != nil {
				return err
			}
			order += 10
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	secs, err = s.listSections(ctx, resumeID)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(secs))
	for i := range secs {
		out = append(out, toSectionMap(&secs[i]))
	}
	return out, nil
}

func (s *Service) RemoveSections(ctx context.Context, resumeID string, ids []string) error {
	if _, err := s.getResume(ctx, resumeID); err != nil {
		return err
	}
	unique := dedupeIDs(ids)
	if len(unique) == 0 {
		return nil
	}
	return s.db.WithContext(ctx).
		Where("resume_id = ? AND id IN ?", resumeID, unique).
		Delete(&model.ResumeSection{}).Error
}

func (s *Service) PublicPublished(ctx context.Context) (map[string]any, error) {
	var po model.Resume
	err := s.db.WithContext(ctx).
		Where("status = ?", model.ResumeStatusPublished).
		Order("published_at DESC").
		First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return nil, ErrNotFound("暂无公开简历")
	}
	if err != nil {
		return nil, err
	}
	secs, err := s.listSections(ctx, po.ID)
	if err != nil {
		return nil, err
	}
	return s.toResumeMap(ctx, &po, secs), nil
}

func (s *Service) PageIcons(prefix string, current, size int, q string) (map[string]any, error) {
	pref := strings.ToLower(strings.TrimSpace(prefix))
	if pref == "" {
		pref = "lucide"
	}
	if pref != "lucide" {
		return nil, ErrBadRequest("仅支持 lucide 图标集")
	}
	names := listLucideIconNames(q)
	current, size = page.Normalize(current, size, 100)
	total := len(names)
	start := (current - 1) * size
	if start > total {
		start = total
	}
	end := start + size
	if end > total {
		end = total
	}
	records := make([]map[string]string, 0, end-start)
	for _, n := range names[start:end] {
		records = append(records, map[string]string{"name": n, "id": "lucide:" + n})
	}
	return map[string]any{
		"records": records,
		"total":   itoa(total),
		"current": itoa(current),
		"size":    itoa(size),
	}, nil
}

// RenderHTMLDoc 规范化入参后渲染纸面 HTML。
func (s *Service) RenderHTMLDoc(ctx context.Context, doc map[string]any) (string, error) {
	payload := map[string]any{}
	for k, v := range doc {
		payload[k] = v
	}
	theme, _ := payload["theme_json"].(map[string]any)
	payload["theme_json"] = normalizeTheme(theme)
	profile, _ := payload["profile_json"].(map[string]any)
	payload["profile_json"] = s.resolveProfileMedia(ctx, normalizeProfile(profile))
	css := SanitizeCustomCSS(asString(payload["custom_css"]))
	payload["custom_css"] = css
	tpl := strings.TrimSpace(asString(payload["template_id"]))
	if tpl == "" {
		tpl = "classic"
	}
	if meta, err := ResolveTemplateForWrite(tpl); err == nil {
		payload["template_id"] = meta.ID
	} else {
		payload["template_id"] = "classic"
	}
	sectionsOut := make([]any, 0)
	for _, m := range asMapSlice(payload["sections"]) {
		meta, _ := m["meta_json"].(map[string]any)
		entriesAny, _ := m["entries_json"].([]any)
		if entriesAny == nil {
			for _, em := range asMapSlice(m["entries_json"]) {
				entriesAny = append(entriesAny, em)
			}
		}
		sectionsOut = append(sectionsOut, map[string]any{
			"id":           truncate(asString(m["id"]), 64, ""),
			"type":         truncate(asString(m["type"]), 32, "custom"),
			"title":        truncate(coalesce(asString(m["title"]), "分节"), 128, "分节"),
			"icon":         normalizeIcon(asString(m["icon"])),
			"sort_code":    m["sort_code"],
			"meta_json":    normalizeMeta(meta),
			"body_html":    truncate(SanitizeBodyHTML(asString(m["body_html"])), bodyHTMLMax, ""),
			"entries_json": normalizeEntries(entriesAny),
		})
	}
	payload["sections"] = sectionsOut
	return RenderHTML(payload)
}

func (s *Service) ExportPDF(ctx context.Context, id string) ([]byte, string, error) {
	doc, err := s.Detail(ctx, id)
	if err != nil {
		return nil, "", err
	}
	html, err := s.RenderHTMLDoc(ctx, doc)
	if err != nil {
		return nil, "", err
	}
	pdf, err := ExportPDFBytes(html)
	if err != nil {
		return nil, "", err
	}
	return pdf, SafePDFFilename(doc), nil
}

// RenderPDFDoc 把当前文档打成 PDF，与 ExportPDF 同源，供未落库预览。
func (s *Service) RenderPDFDoc(ctx context.Context, doc map[string]any) ([]byte, error) {
	// 1. 先渲染与导出一致的纸面 HTML
	html, err := s.RenderHTMLDoc(ctx, doc)
	if err != nil {
		return nil, err
	}
	// 2. 走同一 chromedp PrintToPDF，保证分页与下载文件一致
	return ExportPDFBytes(html)
}

func dedupeIDs(ids []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
