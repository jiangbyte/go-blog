package resume

import (
	"encoding/json"
	"fmt"
	"html"
	"sort"
	"strings"
	"sync"

	"github.com/jiangbyte/go-blog/internal/resume/templates"
)

const fallbackIcon = "circle"

var (
	lucideOnce  sync.Once
	lucideIcons map[string]lucideIcon
	lucideNames []string
)

type lucideIcon struct {
	Body string `json:"body"`
}

// listLucideIconNames 返回可用图标名（不含前缀）。
func listLucideIconNames(q string) []string {
	ensureLucide()
	keyword := strings.ToLower(strings.TrimSpace(q))
	if keyword == "" {
		out := make([]string, len(lucideNames))
		copy(out, lucideNames)
		return out
	}
	var filtered []string
	for _, n := range lucideNames {
		if strings.Contains(n, keyword) {
			filtered = append(filtered, n)
		}
	}
	return filtered
}

func ensureLucide() {
	lucideOnce.Do(func() {
		lucideIcons = map[string]lucideIcon{}
		raw, err := templates.FS.ReadFile("lucide-icons.json")
		if err != nil {
			return
		}
		var root struct {
			Icons map[string]lucideIcon `json:"icons"`
		}
		if err := json.Unmarshal(raw, &root); err != nil || root.Icons == nil {
			return
		}
		lucideIcons = root.Icons
		names := make([]string, 0, len(lucideIcons))
		for k := range lucideIcons {
			names = append(names, k)
		}
		sort.Strings(names)
		lucideNames = names
	})
}

func iconName(raw string) string {
	ensureLucide()
	value := strings.TrimSpace(raw)
	if value == "" {
		return fallbackIcon
	}
	if i := strings.Index(value, ":"); i >= 0 {
		prefix := strings.ToLower(strings.TrimSpace(value[:i]))
		if prefix != "lucide" {
			return fallbackIcon
		}
		value = strings.TrimSpace(value[i+1:])
	}
	if value == "" {
		return fallbackIcon
	}
	if _, ok := lucideIcons[value]; !ok && value != fallbackIcon {
		return fallbackIcon
	}
	return value
}

// NormalizeResumeIcon 规范化为 lucide:name。
func NormalizeResumeIcon(raw string) string {
	return "lucide:" + iconName(raw)
}

func iconSVG(raw string, size int, cssClass, stroke string) string {
	ensureLucide()
	name := iconName(raw)
	meta, ok := lucideIcons[name]
	if !ok {
		meta = lucideIcons[fallbackIcon]
	}
	color := SafeCSSColor(stroke, colorPresets(), "#111827")
	body := meta.Body
	if body == "" {
		body = fmt.Sprintf(`<circle cx="12" cy="12" r="4" fill="%s"/>`, html.EscapeString(color))
	} else {
		body = strings.ReplaceAll(body, "currentColor", color)
	}
	cls := ""
	if cssClass != "" {
		cls = fmt.Sprintf(` class="%s"`, html.EscapeString(cssClass))
	}
	sw := "1.75"
	if size > 14 {
		sw = "2"
	}
	// 不写 inline display，便于皮肤 CSS（如 classic 竖条）覆盖显隐
	return fmt.Sprintf(
		`<svg%s xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 24 24" fill="none" stroke="%s" stroke-width="%s" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">%s</svg>`,
		cls, size, size, html.EscapeString(color), sw, body,
	)
}
