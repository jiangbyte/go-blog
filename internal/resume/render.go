package resume

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/jiangbyte/go-blog/internal/resume/templates"
)

const resumeCSP = "default-src 'none'; img-src https: http: data: blob:; style-src 'unsafe-inline' https://cdn.jsdelivr.net; font-src https://cdn.jsdelivr.net data:; connect-src 'none'; script-src 'none'; base-uri 'none'; form-action 'none'; frame-ancestors 'self'"

type sheetView struct {
	CSP          string
	Name         string
	FontLink     string
	BaseCSS      template.CSS
	ThemeVarsCSS template.CSS
	CustomCSS    template.CSS
	Layout       string
	TemplateID   string
	AvatarURL    string
	ContactsHTML template.HTML
	Sections     []sectionView
}

type sectionView struct {
	Title     string
	IconSVG   template.HTML
	EduHeader template.HTML
	Entries   []entryView
	BodyHTML  template.HTML
}

type entryView struct {
	HeaderHTML template.HTML
	TagsHTML   template.HTML
	BodyHTML   template.HTML
}

var (
	tplOnce sync.Once
	tplErr  error
	tplCache map[string]*template.Template
)

func loadLayout(name string) (*template.Template, error) {
	tplOnce.Do(func() {
		tplCache = map[string]*template.Template{}
		entries, err := templates.FS.ReadDir("layouts")
		if err != nil {
			tplErr = err
			return
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".html") {
				continue
			}
			raw, err := templates.FS.ReadFile(path.Join("layouts", e.Name()))
			if err != nil {
				tplErr = err
				return
			}
			t, err := template.New(e.Name()).Parse(string(raw))
			if err != nil {
				tplErr = err
				return
			}
			tplCache[e.Name()] = t
		}
	})
	if tplErr != nil {
		return nil, tplErr
	}
	t := tplCache[name]
	if t == nil {
		t = tplCache["sheet.html"]
	}
	if t == nil {
		return nil, fmt.Errorf("resume layout not found: %s", name)
	}
	return t, nil
}

func loadStyles(meta templateMeta) (template.CSS, error) {
	var chunks []string
	sheet, err := templates.FS.ReadFile("base/sheet.css")
	if err != nil {
		return "", err
	}
	chunks = append(chunks, string(sheet))
	skin, err := templates.FS.ReadFile(path.Join("skins", meta.CSS))
	if err == nil {
		chunks = append(chunks, string(skin))
	}
	patch, err := templates.FS.ReadFile("base/print-patch.css")
	if err == nil {
		chunks = append(chunks, string(patch))
	}
	return template.CSS(strings.Join(chunks, "\n")), nil
}

// RenderHTML 选模板 → 注入 theme vars → 输出完整 HTML。
func RenderHTML(doc map[string]any) (string, error) {
	// 1. 解析模板元数据；未知 id 回落 classic
	requested := strings.TrimSpace(asString(doc["template_id"]))
	if requested == "" {
		requested = "classic"
	}
	meta := getTemplateMeta(requested)
	if meta == nil {
		meta = getTemplateMeta("classic")
	}
	if meta == nil {
		return "", ErrBadRequest("无可用简历模板")
	}

	theme, _ := doc["theme_json"].(map[string]any)
	if theme == nil {
		theme = map[string]any{}
	}
	profile, _ := doc["profile_json"].(map[string]any)
	if profile == nil {
		profile = map[string]any{}
	}

	// 2. 白名单枚举 + 样式装载
	layout := PickAllowlisted(asString(theme["layout"]), layoutSet(), "left")
	infoDisplay := PickAllowlisted(asString(theme["infoDisplay"]), infoDisplaySet(), "text")
	primary := SafeCSSColor(asString(theme["themeColor"]), colorPresets(), fallbackColor)
	tokens := loadThemeTokens()
	fontID := PickAllowlisted(asString(theme["fontFamily"]), fontFamilySet(), "default")
	fontLink := tokens.FontCdns[fontID]

	baseCSS, err := loadStyles(*meta)
	if err != nil {
		return "", err
	}
	customCSS := SanitizeCustomCSS(asString(doc["custom_css"]))

	name := strings.TrimSpace(asString(profile["name"]))
	if name == "" {
		name = "简历"
	}
	avatar := SafeMediaURL(asString(profile["avatar_url"]))

	// 3. 组装分节视图（富文本仅消毒后注入）
	sections := buildSectionViews(doc, primary)
	contacts := renderContactsHTML(profile, infoDisplay, primary)

	view := sheetView{
		CSP:          resumeCSP,
		Name:         name,
		FontLink:     fontLink,
		BaseCSS:      baseCSS,
		ThemeVarsCSS: template.CSS(themeVarsCSS(theme)),
		CustomCSS:    template.CSS(customCSS),
		Layout:       layout,
		TemplateID:   meta.ID,
		AvatarURL:    avatar,
		ContactsHTML: contacts,
		Sections:     sections,
	}

	tpl, err := loadLayout(meta.HTML)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, view); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func buildSectionViews(doc map[string]any, primary string) []sectionView {
	_ = primary
	raw := asMapSlice(doc["sections"])
	type sortable struct {
		sort int
		sec  map[string]any
	}
	list := make([]sortable, 0, len(raw))
	for _, m := range raw {
		meta, _ := m["meta_json"].(map[string]any)
		if meta != nil && meta["visible"] == false {
			continue
		}
		list = append(list, sortable{sort: asInt(m["sort_code"], 0), sec: m})
	}
	sort.SliceStable(list, func(i, j int) bool { return list[i].sort < list[j].sort })

	out := make([]sectionView, 0, len(list))
	for _, item := range list {
		sec := item.sec
		meta, _ := sec["meta_json"].(map[string]any)
		sv := sectionView{
			Title:   coalesce(asString(sec["title"]), "分节"),
			IconSVG: template.HTML(iconSVG(asString(sec["icon"]), 11, "", "#ffffff")),
		}
		stype := asString(sec["type"])
		if stype == "education" && meta != nil {
			if asString(meta["school"]) != "" || asString(meta["major"]) != "" || asString(meta["period"]) != "" {
				sv.EduHeader = template.HTML(entryHeaderHTML(
					html.EscapeString(asString(meta["school"])),
					html.EscapeString(asString(meta["major"])),
					html.EscapeString(asString(meta["period"])),
				))
			}
		}
		for _, em := range asMapSlice(sec["entries_json"]) {
			sv.Entries = append(sv.Entries, buildEntryView(em))
		}
		body := SanitizeBodyHTML(asString(sec["body_html"]))
		if body != "" {
			sv.BodyHTML = template.HTML(body)
		}
		out = append(out, sv)
	}
	return out
}

func buildEntryView(entry map[string]any) entryView {
	ev := entryView{
		HeaderHTML: template.HTML(entryHeaderHTML(
			html.EscapeString(asString(entry["title"])),
			html.EscapeString(asString(entry["role"])),
			html.EscapeString(asString(entry["period"])),
		)),
	}
	if tags, ok := entry["tags"].([]string); ok && len(tags) > 0 {
		parts := make([]string, 0, len(tags))
		for _, t := range tags {
			if s := strings.TrimSpace(t); s != "" {
				parts = append(parts, html.EscapeString(s))
			}
		}
		if len(parts) > 0 {
			ev.TagsHTML = template.HTML(`<div class="resume-entry__tags" style="margin-bottom:4px;opacity:.75">` +
				strings.Join(parts, " · ") + `</div>`)
		}
	} else if tagsAny, ok := entry["tags"].([]any); ok && len(tagsAny) > 0 {
		parts := make([]string, 0, len(tagsAny))
		for _, t := range tagsAny {
			if s := strings.TrimSpace(asString(t)); s != "" {
				parts = append(parts, html.EscapeString(s))
			}
		}
		if len(parts) > 0 {
			ev.TagsHTML = template.HTML(`<div class="resume-entry__tags" style="margin-bottom:4px;opacity:.75">` +
				strings.Join(parts, " · ") + `</div>`)
		}
	}
	body := SanitizeBodyHTML(asString(entry["body_html"]))
	if body != "" {
		ev.BodyHTML = template.HTML(body)
	}
	return ev
}

func entryHeaderHTML(title, role, period string) string {
	return `<table class="entry-header" role="presentation"><colgroup>` +
		`<col class="entry-header__col-title" /><col class="entry-header__col-role" /><col class="entry-header__col-period" />` +
		`</colgroup><tr>` +
		`<td class="entry-header__title">` + title + `</td>` +
		`<td class="entry-header__role">` + role + `</td>` +
		`<td class="entry-header__period">` + period + `</td>` +
		`</tr></table>`
}

func renderContactsHTML(profile map[string]any, infoDisplay, primary string) template.HTML {
	fields := make([]map[string]any, 0)
	for _, m := range asMapSlice(profile["fields"]) {
		if strings.TrimSpace(asString(m["value"])) == "" {
			continue
		}
		fields = append(fields, m)
	}
	rows := balanceContactRows(fields, infoDisplay)
	if len(rows) == 0 {
		return ""
	}
	cls := "resume-contacts"
	if infoDisplay == "plain" {
		cls += " resume-contacts--plain"
	} else if infoDisplay == "icon" {
		cls += " resume-contacts--icon"
	}
	var rowHTML strings.Builder
	for _, row := range rows {
		var items strings.Builder
		for fi, field := range row {
			if fi > 0 {
				items.WriteString(`<span class="sep">|</span>`)
			}
			label := strings.TrimSpace(asString(field["label"]))
			value := strings.TrimSpace(asString(field["value"]))
			iconHTML := ""
			if infoDisplay == "icon" {
				iconHTML = iconSVG(asString(field["icon"]), 13, "resume-contacts__icon", primary)
			}
			labelHTML := ""
			if infoDisplay == "text" && label != "" {
				labelHTML = `<span class="label">` + html.EscapeString(label) + `：</span>`
			}
			inner := iconHTML + labelHTML + `<span>` + html.EscapeString(value) + `</span>`
			if href := FieldHref(value); href != "" {
				items.WriteString(`<span class="resume-contacts__item"><a href="` +
					html.EscapeString(href) + `" target="_blank" rel="noopener noreferrer">` +
					inner + `</a></span>`)
			} else {
				items.WriteString(`<span class="resume-contacts__item">` + inner + `</span>`)
			}
		}
		rowHTML.WriteString(`<div class="resume-contacts__row">` + items.String() + `</div>`)
	}
	return template.HTML(`<div class="` + cls + `">` + rowHTML.String() + `</div>`)
}

func balanceContactRows(fields []map[string]any, infoDisplay string) [][]map[string]any {
	n := len(fields)
	if n == 0 {
		return nil
	}
	if infoDisplay == "icon" {
		if n <= 3 {
			return [][]map[string]any{fields}
		}
		var rows [][]map[string]any
		for i := 0; i < n; i += 3 {
			end := i + 3
			if end > n {
				end = n
			}
			rows = append(rows, fields[i:end])
		}
		return rows
	}
	if n <= 4 {
		return [][]map[string]any{fields}
	}
	rowCount := 2
	if n > 6 {
		rowCount = 3
	}
	if n > 9 {
		rowCount = (n + 3) / 4
	}
	size := (n + rowCount - 1) / rowCount
	var rows [][]map[string]any
	for i := 0; i < n; i += size {
		end := i + size
		if end > n {
			end = n
		}
		rows = append(rows, fields[i:end])
	}
	return rows
}
