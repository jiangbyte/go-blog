package resume

import (
	"context"
	"strings"

	"github.com/jiangbyte/go-blog/internal/file"
	"github.com/jiangbyte/go-blog/internal/model"
)

// RefreshPublishedPDF 事务外：chromedp 生成 PDF → 上传 OSS → 回写 pdf_file_id。
// 仅由 Publish /「更新并发布」触发，保存草稿不上传。
func (s *Service) RefreshPublishedPDF(ctx context.Context, id string) error {
	if s.files == nil {
		return ErrBadRequest("文件服务未就绪")
	}
	// 1. 渲染与导出同源的 PDF
	doc, err := s.Detail(ctx, id)
	if err != nil {
		return err
	}
	pdf, err := s.RenderPDFDoc(ctx, doc)
	if err != nil {
		return err
	}
	filename := SafePDFFilename(doc)

	// 2. 上传对象存储并落 sys_file
	uploaded, err := s.files.Upload(ctx, file.UploadInput{
		Filename:    filename,
		Content:     pdf,
		ContentType: "application/pdf",
		Category:    "resume",
	})
	if err != nil {
		return err
	}
	fileID := strings.TrimSpace(asString(uploaded["id"]))
	if fileID == "" {
		return ErrBadRequest("上传 PDF 未返回文件 id")
	}

	// 3. 读出旧 pdf_file_id，更新主表后再删旧对象
	po, err := s.getResume(ctx, id)
	if err != nil {
		_ = s.files.Delete(ctx, []string{fileID})
		return err
	}
	oldID := ""
	if po.PdfFileID != nil {
		oldID = strings.TrimSpace(*po.PdfFileID)
	}
	// UpdateColumn：不碰 updated_at，避免与内容时间混淆
	if err := s.db.WithContext(ctx).
		Model(&model.Resume{}).
		Where("id = ?", id).
		UpdateColumn("pdf_file_id", fileID).Error; err != nil {
		_ = s.files.Delete(ctx, []string{fileID})
		return err
	}
	if oldID != "" && oldID != fileID {
		_ = s.files.Delete(ctx, []string{oldID})
	}
	return nil
}

// ClearPublishedPDF 撤回公开：清空 pdf_file_id 并删除 OSS 对象。
func (s *Service) ClearPublishedPDF(ctx context.Context, id string) {
	po, err := s.getResume(ctx, id)
	if err != nil {
		return
	}
	oldID := ""
	if po.PdfFileID != nil {
		oldID = strings.TrimSpace(*po.PdfFileID)
	}
	_ = s.db.WithContext(ctx).
		Model(&model.Resume{}).
		Where("id = ?", id).
		UpdateColumn("pdf_file_id", nil).Error
	if oldID != "" && s.files != nil {
		_ = s.files.Delete(ctx, []string{oldID})
	}
}

// PublicPDFAccessURL 返回当前公开简历 PDF 的可访问 OSS URL（不自动重传，需「更新并发布」）。
func (s *Service) PublicPDFAccessURL(ctx context.Context) (string, string, error) {
	po, filename, err := s.ensurePublicPDF(ctx)
	if err != nil {
		return "", "", err
	}
	u := s.files.ResolveAccessURL(ctx, strings.TrimSpace(*po.PdfFileID))
	if u == "" {
		return "", "", ErrNotFound("公开 PDF 不可用")
	}
	return u, filename, nil
}

// PublicPDFBytes 服务端读出公开 PDF 字节（供前端同源预览，规避图床 CORS）。
func (s *Service) PublicPDFBytes(ctx context.Context) ([]byte, string, error) {
	po, filename, err := s.ensurePublicPDF(ctx)
	if err != nil {
		return nil, "", err
	}
	data, _, err := s.files.DownloadBytes(ctx, strings.TrimSpace(*po.PdfFileID))
	if err != nil {
		return nil, "", err
	}
	if len(data) == 0 {
		return nil, "", ErrNotFound("公开 PDF 不可用")
	}
	return data, filename, nil
}

// ensurePublicPDF 确保已发布简历有可用 pdf_file_id（缺失则补生成）。
func (s *Service) ensurePublicPDF(ctx context.Context) (*model.Resume, string, error) {
	// 1. 取公开简历
	var po model.Resume
	err := s.db.WithContext(ctx).
		Where("status = ?", model.ResumeStatusPublished).
		Order("published_at DESC").
		First(&po).Error
	if err != nil {
		return nil, "", ErrNotFound("暂无公开简历")
	}
	secs, err := s.listSections(ctx, po.ID)
	if err != nil {
		return nil, "", err
	}
	filename := SafePDFFilename(s.toResumeMap(ctx, &po, secs))

	// 2. 已有 OSS 文件则直接用；缺失则补生成（兼容历史发布）
	if po.PdfFileID != nil && strings.TrimSpace(*po.PdfFileID) != "" && s.files != nil {
		return &po, filename, nil
	}
	if err := s.RefreshPublishedPDF(ctx, po.ID); err != nil {
		return nil, "", err
	}
	po2, err := s.getResume(ctx, po.ID)
	if err != nil {
		return nil, "", err
	}
	if po2.PdfFileID == nil || strings.TrimSpace(*po2.PdfFileID) == "" || s.files == nil {
		return nil, "", ErrNotFound("公开 PDF 不可用")
	}
	return po2, filename, nil
}
