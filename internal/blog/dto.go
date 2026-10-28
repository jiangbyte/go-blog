package blog

// API 出站 DTO：JSON 一律 snake_case；文章/项目/分类/标签对外主键均为 uid

type Taxon struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
}

type Post struct {
	UID                string  `json:"uid"`
	Title              string  `json:"title"`
	Description        string  `json:"description,omitempty"`
	Date               string  `json:"date"`
	Status             string  `json:"status"`
	Updated            string  `json:"updated,omitempty"`
	Cover              string  `json:"cover,omitempty"`
	Categories         []Taxon `json:"categories"`
	Tags               []Taxon `json:"tags"`
	Reprint            bool    `json:"reprint,omitempty"`
	Sticky             bool    `json:"sticky,omitempty"`
	Home               bool    `json:"home"`
	ReadingTimeMinutes int     `json:"reading_time_minutes,omitempty"`
	Locate             string  `json:"locate,omitempty"`
	ViewCount          int64   `json:"view_count"`
	Body               string  `json:"body,omitempty"`
}

type Project struct {
	UID         string  `json:"uid"`
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	Date        string  `json:"date"`
	Status      string  `json:"status"`
	Progress    string  `json:"progress"` // planned|active|done|maintained|paused
	Weight      int     `json:"weight,omitempty"`
	Sticky      bool    `json:"sticky,omitempty"`
	Cover       string  `json:"cover,omitempty"`
	Git         string  `json:"git,omitempty"`
	Site        string  `json:"site,omitempty"`
	ProjectTags []Taxon `json:"project_tags"`
	ViewCount   int64   `json:"view_count"`
	Body        string  `json:"body,omitempty"`
}

// BatchStatusFail 批量改状态失败项。
type BatchStatusFail struct {
	UID    string `json:"uid"`
	Title  string `json:"title,omitempty"`
	Reason string `json:"reason"`
}

// BatchStatusResult 批量发布/撤回汇总。
type BatchStatusResult struct {
	Updated int               `json:"updated"`
	Skipped int               `json:"skipped"`
	Failed  []BatchStatusFail `json:"failed"`
}
