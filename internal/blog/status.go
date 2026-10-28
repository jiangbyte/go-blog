package blog

import "strings"

const (
	StatusDraft     = "draft"
	StatusPublished = "published"

	// 项目进度（与发布 status 无关）
	ProgressPlanned    = "planned"
	ProgressActive     = "active"
	ProgressDone       = "done"
	ProgressMaintained = "maintained"
	ProgressPaused     = "paused"
)

// normalizeContentStatus 规范化内容状态：合法值透传；新建默认草稿；更新保留原状态；否则已发布。
func normalizeContentStatus(raw, existing string, isCreate bool) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case StatusDraft:
		return StatusDraft
	case StatusPublished:
		return StatusPublished
	}
	if existing == StatusDraft || existing == StatusPublished {
		return existing
	}
	if isCreate {
		return StatusDraft
	}
	return StatusPublished
}

// normalizeProjectProgress 规范化项目进度；非法值回落 active。
func normalizeProjectProgress(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case ProgressPlanned:
		return ProgressPlanned
	case ProgressActive:
		return ProgressActive
	case ProgressDone:
		return ProgressDone
	case ProgressMaintained:
		return ProgressMaintained
	case ProgressPaused:
		return ProgressPaused
	default:
		return ProgressActive
	}
}
