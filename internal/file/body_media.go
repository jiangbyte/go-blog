package file

import (
	"context"
	"regexp"
	"strings"
)

var (
	// Markdown 图片：![alt](url) / ![alt](<url> "title")
	mdImageURLRe = regexp.MustCompile(`!\[([^\]]*)\]\(\s*<?([^)\s>]+)>?(?:\s+(?:"[^"]*"|'[^']*'))?\s*\)`)
	// HTML img：双引号与单引号分组（Go regexp 无反向引用）
	htmlImgSrcRe = regexp.MustCompile(`(?i)\bsrc=(?:"([^"]+)"|'([^']+)')`)
)

// isMediaObjectRef 判断是否为站内对象键或可剥键的预签名链（需签发后才能在门户展示）。
func isMediaObjectRef(src string) bool {
	raw := strings.TrimSpace(src)
	if raw == "" {
		return false
	}
	if looksLikePresignedURL(raw) {
		return true
	}
	if isExternalURL(raw) {
		return false
	}
	key := strings.TrimPrefix(raw, "/")
	for _, p := range mediaPrefixes {
		if strings.HasPrefix(key, p) {
			return true
		}
	}
	return false
}

// CollectMarkdownMediaRefs 从 Markdown / 内嵌 HTML 收集需签发的媒体引用。
func CollectMarkdownMediaRefs(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	add := func(src string) {
		src = strings.TrimSpace(src)
		if src == "" || !isMediaObjectRef(src) || seen[src] {
			return
		}
		seen[src] = true
		out = append(out, src)
	}
	for _, m := range mdImageURLRe.FindAllStringSubmatch(text, -1) {
		if len(m) >= 3 {
			add(m[2])
		}
	}
	for _, m := range htmlImgSrcRe.FindAllStringSubmatch(text, -1) {
		if len(m) < 3 {
			continue
		}
		if m[1] != "" {
			add(m[1])
		} else {
			add(m[2])
		}
	}
	return out
}

// RewriteMarkdownMediaURLs 门户出站：把正文中的 object_name / 过期预签名链换成当前可访问 URL。
// 库内仍存稳定键；管理端回显勿调用，以免编辑器把短链写回。
func (s *Service) RewriteMarkdownMediaURLs(ctx context.Context, text string) string {
	// 1. 收集引用；无媒体则原样返回
	refs := CollectMarkdownMediaRefs(text)
	if len(refs) == 0 || s == nil {
		return text
	}
	// 2. 批量签发（含 PersistRef 剥键）
	lookup := make([]string, 0, len(refs)*2)
	lookup = append(lookup, refs...)
	for _, r := range refs {
		if k := PersistRef(r); k != "" && k != r {
			lookup = append(lookup, k)
		}
		// 带首 / 的 uploads/... 在 ResolveAccessURLs 会当站内相对路径原样返回，这里剥掉再查库
		if strings.HasPrefix(r, "/") {
			if k := strings.TrimPrefix(r, "/"); isMediaObjectRef(k) {
				lookup = append(lookup, k)
			}
		}
	}
	resolved := s.ResolveAccessURLs(ctx, lookup)

	urlOf := func(src string) string {
		src = strings.TrimSpace(src)
		if u := resolved[src]; u != "" && (isExternalURL(u) || u != src) {
			// 相对路径原样回填不算签发成功
			if !(strings.HasPrefix(u, "/") && !isExternalURL(u)) {
				return u
			}
		}
		if k := PersistRef(src); k != "" {
			if u := resolved[k]; u != "" && isExternalURL(u) {
				return u
			}
		}
		if strings.HasPrefix(src, "/") {
			k := strings.TrimPrefix(src, "/")
			if u := resolved[k]; u != "" && isExternalURL(u) {
				return u
			}
		}
		return ""
	}

	// 3. 替换 Markdown 图片 URL
	out := mdImageURLRe.ReplaceAllStringFunc(text, func(full string) string {
		m := mdImageURLRe.FindStringSubmatch(full)
		if len(m) < 3 {
			return full
		}
		src := m[2]
		u := urlOf(src)
		if u == "" || u == src {
			return full
		}
		return strings.Replace(full, src, u, 1)
	})
	// 4. 替换 HTML img src
	out = htmlImgSrcRe.ReplaceAllStringFunc(out, func(full string) string {
		m := htmlImgSrcRe.FindStringSubmatch(full)
		if len(m) < 3 {
			return full
		}
		quote, src := `"`, m[1]
		if src == "" {
			quote, src = `'`, m[2]
		}
		u := urlOf(src)
		if u == "" || u == src {
			return full
		}
		return "src=" + quote + u + quote
	})
	return out
}
