package resume

import (
	"strings"
	"testing"
)

func TestRenderHTMLSanitizesAndCSP(t *testing.T) {
	html, err := RenderHTML(map[string]any{
		"template_id": "classic",
		"theme_json": map[string]any{
			"themeColor": "blue", "layout": "left", "infoDisplay": "text",
			"fontFamily": "default", "fontSizeScale": 1,
			"spacing": map[string]any{"moduleGap": 12, "lineHeight": 1.55, "padX": 48, "padTop": 48, "padBottom": 60},
		},
		"profile_json": map[string]any{
			"name": "测试", "fields": []any{
				map[string]any{"key": "email", "label": "邮箱", "value": "a@b.com"},
			},
		},
		"sections": []any{
			map[string]any{
				"type": "experience", "title": "经历", "icon": "lucide:briefcase",
				"sort_code": 10, "meta_json": map[string]any{"visible": true},
				"body_html": "<p>hello <script>alert(1)</script></p>",
				"entries_json": []any{
					map[string]any{"title": "公司", "role": "工程师", "period": "2020", "body_html": "<ul><li>item</li></ul>"},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(html) < 500 {
		t.Fatalf("html too short: %d", len(html))
	}
	if strings.Contains(html, "<script>alert") {
		t.Fatal("script not sanitized")
	}
	if !strings.Contains(html, "Content-Security-Policy") || !strings.Contains(html, "测试") {
		t.Fatal("missing csp or name")
	}
}
