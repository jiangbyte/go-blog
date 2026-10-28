// Package file 管理端对象存储文件：上传、分页、批量查询与访问 URL 签发。
package file

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jiangbyte/go-blog/internal/model"
	"github.com/jiangbyte/go-blog/internal/pkg/batch"
	"github.com/jiangbyte/go-blog/internal/pkg/page"
	"github.com/jiangbyte/go-blog/internal/storage"
)

const maxIDsPerRequest = 100

var (
	// ErrNotFound 文件记录不存在（可对外返回）。
	ErrNotFound = errors.New("文件不存在")
	// ErrTooManyID 单次 ids 超过上限。
	ErrTooManyID = fmt.Errorf("ids 最多 %d 个", maxIDsPerRequest)
)

var (
	categoryRe = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
	deniedExt  = map[string]bool{".svg": true, ".html": true, ".htm": true, ".xhtml": true, ".xml": true, ".js": true, ".mjs": true}
	extMIME    = map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".pdf":  "application/pdf",
		".txt":  "text/plain; charset=utf-8",
		".md":   "text/markdown; charset=utf-8",
		".zip":  "application/zip",
	}
)

// UploadInput 上传入参（内容 + 分类/引擎选择）。
type UploadInput struct {
	Filename        string
	Content         []byte
	ContentType     string
	Category        string
	StorageProvider string
	UploaderID      string
}

// Service 文件业务服务。
type Service struct {
	db  *gorm.DB
	reg *storage.Registry
}

func NewService(db *gorm.DB, reg *storage.Registry) *Service {
	return &Service{db: db, reg: reg}
}

// EnsureSchema 确保 sys_file 表结构存在。
func (s *Service) EnsureSchema(ctx context.Context) error {
	return s.db.WithContext(ctx).AutoMigrate(&model.SysFile{})
}

// StorageStatus 返回各存储引擎可用性摘要。
func (s *Service) StorageStatus() storage.Status {
	return s.reg.Status()
}

// BuildObjectName 生成按日分目录的对象键，避免文件名冲突与路径穿越。
func (s *Service) BuildObjectName(filename, category string) string {
	// 1. 只取 basename，杜绝目录穿越
	safe := path.Base(strings.TrimSpace(filename))
	if safe == "." || safe == "/" || safe == "" {
		safe = "unnamed"
	}
	suffix := strings.ToLower(path.Ext(safe))
	// 2. 清洗 category，缺省 uploads
	cat := categoryRe.ReplaceAllString(strings.TrimSpace(category), "")
	if cat == "" {
		cat = "uploads"
	}
	// 3. category/年月日/uuid+扩展名
	now := time.Now().UTC()
	return fmt.Sprintf("%s/%04d/%02d/%02d/%s%s", cat, now.Year(), int(now.Month()), now.Day(),
		strings.ReplaceAll(uuid.NewString(), "-", ""), suffix)
}

// Upload 校验后写入对象存储并落库；落库失败则回滚远端对象。
func (s *Service) Upload(ctx context.Context, in UploadInput) (map[string]any, error) {
	// 1. 校验大小/扩展名
	if err := s.validateUpload(in); err != nil {
		return nil, err
	}
	// 2. 解析存储引擎与桶配置
	st, provider, err := s.reg.Get(in.StorageProvider)
	if err != nil {
		return nil, err
	}
	cfg := s.reg.Settings()
	var engCfg = cfg.MinIO
	if provider == storage.ProviderQiniu {
		engCfg = cfg.Qiniu
	}
	// 3. 生成对象键并上传字节
	objectName := s.BuildObjectName(in.Filename, in.Category)
	suffix := strings.ToLower(path.Ext(path.Base(in.Filename)))
	contentType := extMIME[suffix]
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if err := st.Upload(ctx, objectName, bytes.NewReader(in.Content), int64(len(in.Content)), contentType); err != nil {
		return nil, fmt.Errorf("上传失败: %w", err)
	}

	// 4. 写入 sys_file；失败则删除已上传对象，避免孤儿文件
	fileID := strings.ReplaceAll(uuid.NewString(), "-", "")
	var bucket *string
	if engCfg.Bucket != "" {
		b := engCfg.Bucket
		bucket = &b
	}
	var createdBy *string
	if in.UploaderID != "" {
		u := in.UploaderID
		createdBy = &u
	}
	row := model.SysFile{
		ID:              fileID,
		ObjectName:      objectName,
		OriginalName:    path.Base(strings.TrimSpace(in.Filename)),
		StorageProvider: provider,
		Bucket:          bucket,
		ContentType:     contentType,
		Size:            int64(len(in.Content)),
		URL:             objectName,
		CreatedBy:       createdBy,
		UpdatedBy:       createdBy,
	}
	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		if delErr := st.Delete(ctx, objectName); delErr != nil {
			log.Printf("upload rollback delete object failed: %s: %v", objectName, delErr)
		}
		return nil, err
	}
	// 5. 返回含可访问 URL 的字典
	return s.toDict(ctx, &row, true)
}

func (s *Service) Detail(ctx context.Context, id string) (map[string]any, error) {
	row, err := s.require(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toDict(ctx, row, true)
}

// ListByIDs 按 id 批量查询文件详情，禁止循环单条 Detail。
func (s *Service) ListByIDs(ctx context.Context, ids []string) ([]map[string]any, error) {
	// 1. 去空去重；空列表直接返回；超上限拒绝
	unique := batch.UniqueStrings(ids)
	if len(unique) == 0 {
		return []map[string]any{}, nil
	}
	if len(unique) > maxIDsPerRequest {
		return nil, ErrTooManyID
	}
	// 2. 分批 IN 查库，避免超大 IN 与 N+1
	byID := make(map[string]*model.SysFile, len(unique))
	for _, chunk := range batch.ChunkStrings(unique, batch.MaxINSize) {
		var rows []model.SysFile
		if err := s.db.WithContext(ctx).Where("id IN ?", chunk).Find(&rows).Error; err != nil {
			return nil, err
		}
		for i := range rows {
			byID[rows[i].ID] = &rows[i]
		}
	}
	// 3. 按请求顺序组装字典；缺失 id 跳过，单条 toDict 失败只记日志
	out := make([]map[string]any, 0, len(unique))
	for _, id := range unique {
		row := byID[id]
		if row == nil {
			continue
		}
		m, err := s.toDict(ctx, row, true)
		if err != nil {
			log.Printf("file toDict failed id=%s: %v", id, err)
			continue
		}
		out = append(out, m)
	}
	return out, nil
}

// Page 分页列出文件，可按存储引擎与关键字过滤。
func (s *Service) Page(ctx context.Context, current, size int, provider, keyword string) (page.IPage, error) {
	// 1. 规范化分页参数
	current, size = page.Normalize(current, size, 20)
	q := s.db.WithContext(ctx).Model(&model.SysFile{})
	// 2. 可选过滤：引擎（兼容别名）与文件名/对象键模糊搜
	if p := strings.TrimSpace(provider); p != "" {
		if resolved, err := storage.EngineToProvider(p); err == nil {
			q = q.Where("storage_provider = ?", resolved)
		} else {
			q = q.Where("storage_provider = ?", p)
		}
	}
	if kw := strings.TrimSpace(keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("original_name LIKE ? OR object_name LIKE ?", like, like)
	}
	// 3. 先 Count 再 Offset/Limit 取页
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return page.IPage{}, err
	}
	var rows []model.SysFile
	if err := q.Order("created_at DESC").Offset((current - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		return page.IPage{}, err
	}
	// 4. 转字典并签发 URL
	out := make([]map[string]any, 0, len(rows))
	for i := range rows {
		m, err := s.toDict(ctx, &rows[i], true)
		if err != nil {
			log.Printf("file toDict failed id=%s: %v", rows[i].ID, err)
			continue
		}
		out = append(out, m)
	}
	return page.Of(out, total, current, size), nil
}

// Delete 按 id 批量删除：先删远端对象再删库记录。
func (s *Service) Delete(ctx context.Context, ids []string) error {
	// 1. 去重与上限校验
	unique := batch.UniqueStrings(ids)
	if len(unique) == 0 {
		return nil
	}
	if len(unique) > maxIDsPerRequest {
		return ErrTooManyID
	}
	// 2. 分批查出待删行
	var rows []model.SysFile
	for _, chunk := range batch.ChunkStrings(unique, batch.MaxINSize) {
		var part []model.SysFile
		if err := s.db.WithContext(ctx).Where("id IN ?", chunk).Find(&part).Error; err != nil {
			return err
		}
		rows = append(rows, part...)
	}
	// 3. 逐条删存储（失败只打日志）再删库；库删除失败中断返回
	for _, row := range rows {
		st, _, err := s.reg.Get(row.StorageProvider)
		if err == nil {
			if err := st.Delete(ctx, row.ObjectName); err != nil {
				log.Printf("delete storage object failed: %s: %v", row.ObjectName, err)
			}
		}
		if err := s.db.WithContext(ctx).Delete(&model.SysFile{}, "id = ?", row.ID).Error; err != nil {
			log.Printf("delete sys_file failed id=%s: %v", row.ID, err)
			return err
		}
	}
	return nil
}

// URLByObjectName 按对象键/id/预签名链解析并签发可访问 URL。
func (s *Service) URLByObjectName(ctx context.Context, objectName string) (map[string]any, error) {
	// 1. 校验入参
	ref := strings.TrimSpace(objectName)
	if ref == "" {
		return nil, fmt.Errorf("object_name required")
	}
	// 2. 优先按 id/object_name 查库，再尝试候选键
	var row *model.SysFile
	if found, err := s.findByIDOrObject(ctx, ref); err == nil && found != nil {
		row = found
	} else {
		for _, candidate := range objectKeyCandidates(ref) {
			found, err := s.findByObjectName(ctx, candidate)
			if err == nil && found != nil {
				row = found
				break
			}
		}
	}
	// 3. 无库记录时：用规范化 object_name + 默认引擎签发；否则找不到
	if row == nil {
		key := PersistRef(ref)
		if key == "" || isExternalURL(key) || strings.HasPrefix(key, "/") {
			return nil, ErrNotFound
		}
		st, _, err := s.reg.Get("")
		if err != nil {
			return nil, ErrNotFound
		}
		u, err := st.URL(ctx, key)
		if err != nil {
			return nil, err
		}
		return map[string]any{"object_name": key, "url": u}, nil
	}
	// 4. 有记录则按所属引擎签发
	u, err := s.objectURL(ctx, row)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"object_name":   row.ObjectName,
		"url":           u,
		"original_name": row.OriginalName,
		"size":          row.Size,
		"content_type":  row.ContentType,
	}, nil
}

// DownloadBytes 按文件 id 拉取原始字节，供管理端下载。
func (s *Service) DownloadBytes(ctx context.Context, id string) ([]byte, *model.SysFile, error) {
	// 1. 查库确认文件存在
	row, err := s.require(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	// 2. 取对应引擎并读对象内容
	st, _, err := s.reg.Get(row.StorageProvider)
	if err != nil {
		return nil, nil, err
	}
	data, err := st.GetBytes(ctx, row.ObjectName)
	if err != nil {
		return nil, nil, err
	}
	return data, row, nil
}

// ResolveAccessURL 按 file id / object_name / 预签名 URL 重新签发可访问地址。
// 稳定外链（非预签名）原样返回；过期预签名链会剥键后重签。
func (s *Service) ResolveAccessURL(ctx context.Context, ref string) string {
	m := s.ResolveAccessURLs(ctx, []string{ref})
	return m[strings.TrimSpace(ref)]
}

// ResolveAccessURLs 批量签发访问地址，分批 IN 查库，避免列表 N+1。
func (s *Service) ResolveAccessURLs(ctx context.Context, refs []string) map[string]string {
	out := make(map[string]string, len(refs))
	// 1. 去重；相对路径与稳定外链无需查库
	uniq := batch.UniqueStrings(refs)
	if len(uniq) == 0 {
		return out
	}

	needLookup := make([]string, 0, len(uniq))
	for _, ref := range uniq {
		if strings.HasPrefix(ref, "/") && !strings.HasPrefix(ref, "//") {
			out[ref] = ref
			continue
		}
		if isExternalURL(ref) && !looksLikePresignedURL(ref) {
			out[ref] = ref
			continue
		}
		needLookup = append(needLookup, ref)
	}
	if len(needLookup) == 0 {
		return out
	}

	// 2. 收集每个 ref 的候选 id/object_name（含预签名剥键）
	refCandidates := make(map[string][]string, len(needLookup))
	var allIDs, allObjects []string
	for _, ref := range needLookup {
		cands := batch.UniqueStrings(append([]string{ref}, objectKeyCandidates(ref)...))
		if key := PersistRef(ref); key != "" && !isExternalURL(key) && !strings.HasPrefix(key, "/") {
			cands = batch.UniqueStrings(append(cands, key))
		}
		refCandidates[ref] = cands
		for _, c := range cands {
			if strings.Contains(c, "/") {
				allObjects = append(allObjects, c)
			} else {
				allIDs = append(allIDs, c)
				allObjects = append(allObjects, c)
			}
		}
	}
	allIDs = batch.UniqueStrings(allIDs)
	allObjects = batch.UniqueStrings(allObjects)

	// 3. 分批按 id / object_name IN 加载 sys_file
	byID := make(map[string]*model.SysFile)
	byObject := make(map[string]*model.SysFile)
	loadRows := func(rows []model.SysFile) {
		for i := range rows {
			r := &rows[i]
			byID[r.ID] = r
			byObject[r.ObjectName] = r
		}
	}
	for _, chunk := range batch.ChunkStrings(allIDs, batch.MaxINSize) {
		var rows []model.SysFile
		if err := s.db.WithContext(ctx).Where("id IN ?", chunk).Find(&rows).Error; err != nil {
			log.Printf("ResolveAccessURLs by id: %v", err)
			continue
		}
		loadRows(rows)
	}
	for _, chunk := range batch.ChunkStrings(allObjects, batch.MaxINSize) {
		var rows []model.SysFile
		if err := s.db.WithContext(ctx).Where("object_name IN ?", chunk).Find(&rows).Error; err != nil {
			log.Printf("ResolveAccessURLs by object_name: %v", err)
			continue
		}
		loadRows(rows)
	}

	// 4. 按候选匹配签发；无记录则默认引擎签发 object_name；失败记空串
	for _, ref := range needLookup {
		var row *model.SysFile
		for _, c := range refCandidates[ref] {
			if r := byID[c]; r != nil {
				row = r
				break
			}
			if r := byObject[c]; r != nil {
				row = r
				break
			}
		}
		if row != nil {
			if u, err := s.objectURL(ctx, row); err == nil {
				out[ref] = u
				continue
			}
		}
		key := PersistRef(ref)
		if key != "" && !isExternalURL(key) && !strings.HasPrefix(key, "/") {
			if st, _, err := s.reg.Get(""); err == nil {
				if u, err := st.URL(ctx, key); err == nil {
					out[ref] = u
					continue
				}
			}
		}
		out[ref] = ""
	}
	return out
}

func (s *Service) findByObjectName(ctx context.Context, objectName string) (*model.SysFile, error) {
	var row model.SysFile
	if err := s.db.WithContext(ctx).Where("object_name = ?", strings.TrimSpace(objectName)).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Service) findByIDOrObject(ctx context.Context, ref string) (*model.SysFile, error) {
	var row model.SysFile
	if err := s.db.WithContext(ctx).Where("id = ?", ref).First(&row).Error; err == nil {
		return &row, nil
	}
	return s.findByObjectName(ctx, ref)
}

func (s *Service) require(ctx context.Context, id string) (*model.SysFile, error) {
	var row model.SysFile
	if err := s.db.WithContext(ctx).Where("id = ?", strings.TrimSpace(id)).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &row, nil
}

func (s *Service) objectURL(ctx context.Context, row *model.SysFile) (string, error) {
	st, _, err := s.reg.Get(row.StorageProvider)
	if err != nil {
		return "", err
	}
	return st.URL(ctx, row.ObjectName)
}

func (s *Service) toDict(ctx context.Context, row *model.SysFile, resolveURL bool) (map[string]any, error) {
	data := map[string]any{
		"id":               row.ID,
		"object_name":      row.ObjectName,
		"original_name":    row.OriginalName,
		"storage_provider": row.StorageProvider,
		"bucket":           row.Bucket,
		"content_type":     row.ContentType,
		"size":             row.Size,
		"url":              row.URL,
		"created_at":       row.CreatedAt.Format(time.RFC3339),
		"created_by":       row.CreatedBy,
		"updated_at":       row.UpdatedAt.Format(time.RFC3339),
		"updated_by":       row.UpdatedBy,
	}
	if resolveURL {
		if u, err := s.objectURL(ctx, row); err == nil {
			data["url"] = u
		}
	}
	return data, nil
}

func (s *Service) validateUpload(in UploadInput) error {
	if len(in.Content) == 0 {
		return fmt.Errorf("文件内容为空")
	}
	max := s.reg.Settings().UploadMaxBytes
	if max > 0 && int64(len(in.Content)) > max {
		return fmt.Errorf("文件超过大小限制（%d 字节）", max)
	}
	suffix := strings.ToLower(path.Ext(path.Base(in.Filename)))
	if deniedExt[suffix] {
		return fmt.Errorf("不允许的文件类型: %s", suffix)
	}
	allowed := s.reg.Settings().AllowedExtSet()
	if len(allowed) > 0 && !allowed[suffix] {
		return fmt.Errorf("不允许的文件类型: %s", suffix)
	}
	return nil
}
