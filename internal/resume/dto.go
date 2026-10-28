package resume

import "github.com/jiangbyte/go-blog/internal/pkg/wirejson"

// ResumeWriteInput 创建/更新入参。
type ResumeWriteInput struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	TemplateID  string         `json:"template_id"`
	ThemeJSON   map[string]any `json:"theme_json"`
	CustomCSS   *string        `json:"custom_css"`
	ProfileJSON map[string]any `json:"profile_json"`
	// SortCode 用 wirejson.Int：axios stringifyScalars 会把数字变成字符串
	SortCode wirejson.Int `json:"sort_code"`
	Status   *string      `json:"status"`
}

// SectionSaveInput 分节保存入参。
type SectionSaveInput struct {
	ResumeID string `json:"resume_id"`
	ID       string `json:"id"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	Icon     *string `json:"icon"`
	// SortCode 兼容 number / "10" 字符串，避免 invalid body
	SortCode    wirejson.Int   `json:"sort_code"`
	MetaJSON    map[string]any `json:"meta_json"`
	BodyHTML    *string        `json:"body_html"`
	EntriesJSON []any          `json:"entries_json"`
}

type idBody struct {
	ID string `json:"id"`
}

type idsBody struct {
	IDs []string `json:"ids"`
}

type sectionReorderBody struct {
	ResumeID string   `json:"resume_id"`
	IDs      []string `json:"ids"`
}

type sectionRemoveBody struct {
	ResumeID string   `json:"resume_id"`
	IDs      []string `json:"ids"`
}

type sectionCopyBody struct {
	ResumeID  string `json:"resume_id"`
	SectionID string `json:"section_id"`
}

type renderHTMLBody struct {
	Title       string         `json:"title"`
	TemplateID  string         `json:"template_id"`
	ThemeJSON   map[string]any `json:"theme_json"`
	CustomCSS   *string        `json:"custom_css"`
	ProfileJSON map[string]any `json:"profile_json"`
	Sections    []any          `json:"sections"`
}
