package resume

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/jiangbyte/go-blog/internal/resume/templates"
)

const structureSheetV1 = "sheet-v1"
const fallbackColor = "#3b82f6"

type templateMeta struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	DefaultColor string `json:"defaultColor"`
	Structure    string `json:"structure"`
	CSS          string `json:"css"`
	HTML         string `json:"html"`
}

type themeTokens struct {
	Colors        map[string]string `json:"colors"`
	FontStacks    map[string]string `json:"fontStacks"`
	FontCdns      map[string]string `json:"fontCdns"`
	Layouts       []string          `json:"layouts"`
	InfoDisplays  []string          `json:"infoDisplays"`
}

var (
	themeOnce     sync.Once
	cachedTokens  themeTokens
	manifestOnce  sync.Once
	cachedManifest []templateMeta
)

func loadThemeTokens() themeTokens {
	themeOnce.Do(func() {
		cachedTokens = themeTokens{
			Colors:       map[string]string{"blue": fallbackColor},
			FontStacks:   map[string]string{"default": "sans-serif"},
			FontCdns:     map[string]string{},
			Layouts:      []string{"left", "center", "tiled"},
			InfoDisplays: []string{"text", "icon", "plain"},
		}
		raw, err := templates.FS.ReadFile("themeTokens.json")
		if err != nil {
			return
		}
		var t themeTokens
		if err := json.Unmarshal(raw, &t); err != nil {
			return
		}
		if t.Colors != nil {
			cachedTokens.Colors = t.Colors
		}
		if t.FontStacks != nil {
			cachedTokens.FontStacks = t.FontStacks
		}
		if t.FontCdns != nil {
			cachedTokens.FontCdns = t.FontCdns
		}
		if len(t.Layouts) > 0 {
			cachedTokens.Layouts = t.Layouts
		}
		if len(t.InfoDisplays) > 0 {
			cachedTokens.InfoDisplays = t.InfoDisplays
		}
	})
	return cachedTokens
}

func colorPresets() map[string]string {
	return loadThemeTokens().Colors
}

func loadManifest() []templateMeta {
	manifestOnce.Do(func() {
		raw, err := templates.FS.ReadFile("manifest.json")
		if err != nil {
			cachedManifest = []templateMeta{{
				ID: "classic", Name: "经典蓝", Structure: structureSheetV1,
				CSS: "classic.css", HTML: "sheet.html", DefaultColor: "blue",
			}}
			return
		}
		var rows []templateMeta
		if err := json.Unmarshal(raw, &rows); err != nil {
			return
		}
		out := make([]templateMeta, 0, len(rows))
		for _, row := range rows {
			id := stringsTrim(row.ID)
			if id == "" {
				continue
			}
			css := stringsTrim(row.CSS)
			if css == "" {
				css = id + ".css"
			}
			if stringsContainsAny(css, "/", `\`, "..") {
				continue
			}
			htmlFile := stringsTrim(row.HTML)
			if htmlFile == "" {
				htmlFile = "sheet.html"
			}
			if stringsContainsAny(htmlFile, "/", `\`, "..") {
				continue
			}
			structure := stringsTrim(row.Structure)
			if structure == "" {
				structure = structureSheetV1
			}
			out = append(out, templateMeta{
				ID:           id,
				Name:         coalesce(row.Name, id),
				Description:  row.Description,
				DefaultColor: coalesce(row.DefaultColor, "blue"),
				Structure:    structure,
				CSS:          css,
				HTML:         htmlFile,
			})
		}
		cachedManifest = out
	})
	return cachedManifest
}

// ListTemplateMetas 管理端模板清单。
func ListTemplateMetas() []map[string]string {
	items := loadManifest()
	out := make([]map[string]string, 0, len(items))
	for _, t := range items {
		out = append(out, map[string]string{
			"id":           t.ID,
			"name":         t.Name,
			"description":  t.Description,
			"defaultColor": t.DefaultColor,
			"structure":    t.Structure,
		})
	}
	return out
}

func getTemplateMeta(templateID string) *templateMeta {
	key := stringsTrim(templateID)
	if key == "" {
		return nil
	}
	for _, item := range loadManifest() {
		if item.ID == key {
			cp := item
			return &cp
		}
	}
	return nil
}

// ResolveTemplateForWrite 写库校验模板白名单。
func ResolveTemplateForWrite(templateID string) (templateMeta, error) {
	key := stringsTrim(templateID)
	if key == "" {
		key = "classic"
	}
	meta := getTemplateMeta(key)
	if meta == nil {
		return templateMeta{}, ErrBadRequest(fmt.Sprintf("未知模板: %s", key))
	}
	if meta.Structure != structureSheetV1 {
		return templateMeta{}, ErrBadRequest(fmt.Sprintf("模板结构未实现: %s", meta.Structure))
	}
	return *meta, nil
}

func themeVarsCSS(theme map[string]any) string {
	tokens := loadThemeTokens()
	spacing, _ := theme["spacing"].(map[string]any)
	primary := SafeCSSColor(asString(theme["themeColor"]), tokens.Colors, fallbackColor)
	fontID := asString(theme["fontFamily"])
	if fontID == "" {
		fontID = "default"
	}
	lineHeight := asFloat(spacing["lineHeight"], 1.55)
	if lineHeight < 1 {
		lineHeight = 1.55
	}
	if lineHeight > 3 {
		lineHeight = 3
	}
	fontScale := asFloat(theme["fontSizeScale"], 1)
	if fontScale < 0.8 {
		fontScale = 0.8
	}
	if fontScale > 1.2 {
		fontScale = 1.2
	}
	moduleGap := asInt(spacing["moduleGap"], 12)
	padX := maxInt(0, asInt(spacing["padX"], 48))
	padTop := maxInt(0, asInt(spacing["padTop"], 48))
	padBottom := maxInt(0, asInt(spacing["padBottom"], 60))
	fontFamily := tokens.FontStacks[fontID]
	if fontFamily == "" {
		fontFamily = tokens.FontStacks["default"]
	}
	return fmt.Sprintf(
		`.resume-page[data-resume-root]{--color-primary:%s;--color-primary-soft:%s1a;--color-text:#111827;--color-text-secondary:#4b5563;--color-border:#e5e7eb;--space-module:%dpx;--line-height:%g;--pad-x:%dpx;--pad-top:%dpx;--pad-bottom:%dpx;--font-family:%s;--font-scale:%g;--entry-gap:10px;--section-content-gap:6px}`,
		primary, primary, moduleGap, lineHeight, padX, padTop, padBottom, fontFamily, fontScale,
	)
}

func layoutSet() map[string]struct{} {
	m := map[string]struct{}{}
	for _, v := range loadThemeTokens().Layouts {
		m[v] = struct{}{}
	}
	return m
}

func infoDisplaySet() map[string]struct{} {
	m := map[string]struct{}{}
	for _, v := range loadThemeTokens().InfoDisplays {
		m[v] = struct{}{}
	}
	return m
}

func fontFamilySet() map[string]struct{} {
	m := map[string]struct{}{}
	for k := range loadThemeTokens().FontStacks {
		m[k] = struct{}{}
	}
	return m
}
