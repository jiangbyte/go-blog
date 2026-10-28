package resume

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

const (
	customCSSMax = 50_000
	bodyHTMLMax  = 100_000
)

var (
	hexColorRE     = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)
	safeCSSURLRE   = regexp.MustCompile(`(?i)url\s*\(\s*['"]?\s*(?:javascript|vbscript|data)\s*:`)
	styleBreakRE   = regexp.MustCompile(`(?i)</\s*style`)
	scriptishCSSRE = regexp.MustCompile(`(?i)</?\s*script|@import|expression\s*\(|behavior\s*:|javascript\s*:|-moz-binding`)
	objectNameRE  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_./-]{0,510}$`)
	telRE          = regexp.MustCompile(`^\+?\d{5,20}$`)
	emailRE        = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	domainPathRE   = regexp.MustCompile(`(?i)^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+(/[\w\-./?%&=+#]*)?$`)
	bodyPolicy     = newBodyPolicy()
)

func newBodyPolicy() *bluemonday.Policy {
	// 1. 仅保留简历正文标签，去掉全部属性与事件
	p := bluemonday.NewPolicy()
	p.AllowElements("p", "br", "strong", "b", "em", "i", "ul", "ol", "li")
	return p
}

// SanitizeBodyHTML 消毒富文本正文（UGC 白名单）。
func SanitizeBodyHTML(raw string) string {
	// 1. 空串直接返回
	text := strings.TrimSpace(raw)
	if text == "" {
		return ""
	}
	// 2. 超长截断，避免异常 payload 拖垮渲染
	if len(text) > bodyHTMLMax {
		text = text[:bodyHTMLMax]
	}
	// 3. bluemonday 白名单清洗
	return bodyPolicy.Sanitize(text)
}

// SanitizeCustomCSS 渲染路径：非法 CSS 丢弃为空。
func SanitizeCustomCSS(raw string) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return ""
	}
	if len(text) > customCSSMax {
		text = text[:customCSSMax]
	}
	if styleBreakRE.MatchString(text) || scriptishCSSRE.MatchString(text) || safeCSSURLRE.MatchString(text) {
		return ""
	}
	return text
}

// SanitizeCustomCSSOrError 写库路径：非法 CSS 返回错误。
func SanitizeCustomCSSOrError(raw *string) (*string, error) {
	if raw == nil {
		return nil, nil
	}
	text := strings.TrimSpace(*raw)
	if text == "" {
		return nil, nil
	}
	if len(text) > customCSSMax {
		return nil, ErrBadRequest("自定义 CSS 过长")
	}
	if styleBreakRE.MatchString(text) || scriptishCSSRE.MatchString(text) || safeCSSURLRE.MatchString(text) {
		return nil, ErrBadRequest("自定义 CSS 含不允许的指令")
	}
	return &text, nil
}

// SafeCSSColor 主题色：#hex 或预设名。
func SafeCSSColor(raw string, presets map[string]string, fallback string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return fallback
	}
	if hexColorRE.MatchString(value) {
		return strings.ToLower(value)
	}
	if hex, ok := presets[strings.ToLower(value)]; ok {
		return hex
	}
	if hex, ok := presets[value]; ok {
		return hex
	}
	return fallback
}

// SafeHref 仅允许 http(s)/mailto/tel/站内相对路径。
func SafeHref(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	lower := strings.ToLower(value)
	if strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "data:") ||
		strings.HasPrefix(lower, "vbscript:") || strings.HasPrefix(lower, "file:") {
		return ""
	}
	if strings.HasPrefix(lower, "mailto:") {
		addr := strings.TrimSpace(value[7:])
		if addr == "" || strings.ContainsAny(addr, `<>"'`) {
			return ""
		}
		return "mailto:" + addr
	}
	if strings.HasPrefix(lower, "tel:") {
		num := strings.TrimSpace(value[4:])
		if !telRE.MatchString(num) {
			return ""
		}
		return "tel:" + num
	}
	if strings.HasPrefix(value, "/") {
		if strings.HasPrefix(value, "//") || rejectDotDot(strings.TrimPrefix(value, "/")) {
			return ""
		}
		return value
	}
	u, err := url.Parse(value)
	if err != nil || u.Host == "" {
		return ""
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ""
	}
	return value
}

// SafeMediaURL 媒体资源：仅 http(s) 或站内相对路径。
func SafeMediaURL(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	lower := strings.ToLower(value)
	if strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "data:") ||
		strings.HasPrefix(lower, "vbscript:") || strings.HasPrefix(lower, "file:") {
		return ""
	}
	if strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//") {
		if rejectDotDot(strings.TrimPrefix(value, "/")) {
			return ""
		}
		return value
	}
	u, err := url.Parse(value)
	if err != nil || u.Host == "" {
		return ""
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ""
	}
	return value
}

// NormalizeAvatarStorageRef 写库头像引用：优先稳定 object_name。
func NormalizeAvatarStorageRef(raw string) string {
	// 1. 空值清空
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	if len(value) > 512 {
		value = value[:512]
	}
	// 2. 外链若能抽出 uploads/... 等对象键则存键
	if isExternalURL(value) {
		if key := extractObjectKey(value); key != "" {
			return key
		}
		if looksLikePresignedURL(value) {
			return ""
		}
		return SafeMediaURL(value)
	}
	// 3. 站内相对路径 / object_name
	if strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//") {
		if rejectDotDot(strings.TrimPrefix(value, "/")) {
			return ""
		}
		return value
	}
	if objectNameRE.MatchString(value) && strings.Contains(value, "/") {
		if rejectDotDot(value) {
			return ""
		}
		return value
	}
	return ""
}

func rejectDotDot(path string) bool {
	for _, seg := range strings.Split(path, "/") {
		if seg == "" || seg == "." || seg == ".." {
			return true
		}
	}
	return false
}

func isExternalURL(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

func looksLikePresignedURL(value string) bool {
	lower := strings.ToLower(value)
	return strings.Contains(lower, "x-qiniu") || strings.Contains(lower, "token=") ||
		strings.Contains(lower, "signature=") || strings.Contains(lower, "x-amz-")
}

func extractObjectKey(value string) string {
	u, err := url.Parse(value)
	if err != nil {
		return ""
	}
	path := strings.TrimPrefix(u.Path, "/")
	prefixes := []string{"uploads/", "avatars/", "files/", "resume/", "content/"}
	for _, p := range prefixes {
		if idx := strings.Index(path, p); idx >= 0 {
			key := path[idx:]
			if !rejectDotDot(key) {
				if len(key) > 512 {
					key = key[:512]
				}
				return key
			}
		}
	}
	return ""
}

// PickAllowlisted 枚举字段回落。
func PickAllowlisted(raw string, allowed map[string]struct{}, fallback string) string {
	value := strings.TrimSpace(raw)
	if _, ok := allowed[value]; ok {
		return value
	}
	return fallback
}

// FieldHref 联系字段 value → 安全超链接。
func FieldHref(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return ""
	}
	if matched, _ := regexp.MatchString(`(?i)^https?://`, text); matched {
		return SafeHref(text)
	}
	if emailRE.MatchString(text) {
		return SafeHref("mailto:" + text)
	}
	digits := regexp.MustCompile(`[\s\-()]`).ReplaceAllString(text, "")
	digits = regexp.MustCompile(`(?i)^\+?86`).ReplaceAllString(digits, "")
	if matched, _ := regexp.MatchString(`^1\d{10}$`, digits); matched {
		return SafeHref("tel:" + digits)
	}
	if domainPathRE.MatchString(text) {
		return SafeHref("https://" + text)
	}
	return ""
}
