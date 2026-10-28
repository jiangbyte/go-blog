package resume

import "strings"

var (
	sectionTypes = map[string]struct{}{
		"education": {}, "experience": {}, "skills": {},
		"projects": {}, "evaluation": {}, "custom": {},
	}
	languages = map[string]struct{}{"zh": {}, "en": {}}
)

var defaultTheme = map[string]any{
	"themeColor":    "blue",
	"layout":        "left",
	"infoDisplay":   "text",
	"language":      "zh",
	"fontFamily":    "default",
	"fontSizeScale": 1.0,
	"spacing": map[string]any{
		"moduleGap":  12,
		"lineHeight": 1.55,
		"padX":       48,
		"padTop":     48,
		"padBottom":  60,
	},
}

var defaultProfile = map[string]any{
	"name": "", "tagline": "", "avatar_url": "", "github_url": "", "fields": []any{},
}

func normalizeTheme(theme map[string]any) map[string]any {
	// 1. 取源对象与间距子对象
	src := theme
	if src == nil {
		src = map[string]any{}
	}
	spacingSrc, _ := src["spacing"].(map[string]any)
	if spacingSrc == nil {
		spacingSrc = map[string]any{}
	}
	// 2. 枚举类字段回落到默认
	layout := PickAllowlisted(asString(src["layout"]), layoutSet(), "left")
	info := PickAllowlisted(asString(src["infoDisplay"]), infoDisplaySet(), "text")
	language := PickAllowlisted(asString(src["language"]), languages, "zh")
	fontFamily := PickAllowlisted(asString(src["fontFamily"]), fontFamilySet(), "default")
	// 3. 数值钳制
	lineHeight := asFloat(spacingSrc["lineHeight"], 1.55)
	if lineHeight != lineHeight || lineHeight < 1 {
		lineHeight = 1.55
	} else if lineHeight > 3 {
		lineHeight = 3
	}
	return map[string]any{
		"themeColor":    truncate(asString(src["themeColor"]), 32, "blue"),
		"layout":        layout,
		"infoDisplay":   info,
		"language":      language,
		"fontFamily":    fontFamily,
		"fontSizeScale": clampFloat(src["fontSizeScale"], 1, 0.8, 1.2),
		"spacing": map[string]any{
			"moduleGap":  clampInt(spacingSrc["moduleGap"], 12, 4, 40),
			"lineHeight": lineHeight,
			"padX":       clampInt(spacingSrc["padX"], 48, 24, 80),
			"padTop":     clampInt(spacingSrc["padTop"], 48, 24, 80),
			"padBottom":  clampInt(spacingSrc["padBottom"], 60, 24, 100),
		},
	}
}

func normalizeProfile(profile map[string]any) map[string]any {
	src := profile
	if src == nil {
		src = map[string]any{}
	}
	fields := []map[string]string{}
	if raw, ok := src["fields"].([]any); ok {
		for _, item := range raw {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			key := strings.TrimSpace(asString(m["key"]))
			value := strings.TrimSpace(asString(m["value"]))
			if key == "" || value == "" {
				continue
			}
			label := strings.TrimSpace(asString(m["label"]))
			if label == "" {
				label = key
			}
			icon := ""
			if ic := strings.TrimSpace(asString(m["icon"])); ic != "" {
				icon = NormalizeResumeIcon(ic)
			}
			fields = append(fields, map[string]string{
				"key": truncate(key, 64, ""), "label": truncate(label, 64, ""),
				"value": truncate(value, 256, ""), "icon": icon,
			})
		}
	}
	github := SafeHref(truncate(strings.TrimSpace(asString(src["github_url"])), 512, ""))
	return map[string]any{
		"name":       truncate(strings.TrimSpace(asString(src["name"])), 64, ""),
		"tagline":    truncate(strings.TrimSpace(asString(src["tagline"])), 256, ""),
		"avatar_url": NormalizeAvatarStorageRef(asString(src["avatar_url"])),
		"github_url": github,
		"fields":     fields,
	}
}

func normalizeEntries(entries []any) []map[string]any {
	if entries == nil {
		return []map[string]any{}
	}
	out := make([]map[string]any, 0, len(entries))
	for _, item := range entries {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		tags := []string{}
		if raw, ok := m["tags"].([]any); ok {
			for _, t := range raw {
				s := strings.TrimSpace(asString(t))
				if s == "" {
					continue
				}
				tags = append(tags, truncate(s, 32, ""))
				if len(tags) >= 20 {
					break
				}
			}
		}
		out = append(out, map[string]any{
			"title":     truncate(strings.TrimSpace(asString(m["title"])), 128, ""),
			"role":      truncate(strings.TrimSpace(asString(m["role"])), 128, ""),
			"period":    truncate(strings.TrimSpace(asString(m["period"])), 64, ""),
			"body_html": truncate(SanitizeBodyHTML(asString(m["body_html"])), bodyHTMLMax, ""),
			"tags":      tags,
		})
	}
	return out
}

func normalizeMeta(meta map[string]any) map[string]any {
	out := map[string]any{}
	if meta != nil {
		for k, v := range meta {
			out[k] = v
		}
	}
	out["visible"] = coerceVisible(out["visible"])
	return out
}

func coerceVisible(raw any) bool {
	if raw == nil {
		return true
	}
	switch t := raw.(type) {
	case bool:
		return t
	case float64:
		return t != 0
	case int:
		return t != 0
	case string:
		text := strings.ToLower(strings.TrimSpace(t))
		switch text {
		case "false", "0", "no", "off":
			return false
		default:
			return true
		}
	default:
		return true
	}
}

func truncate(s string, max int, fallback string) string {
	if s == "" {
		return fallback
	}
	if len(s) <= max {
		return s
	}
	return s[:max]
}

func normalizeIcon(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	return NormalizeResumeIcon(value)
}
