package file

import (
	"net/url"
	"strings"
)

var mediaPrefixes = []string{
	"uploads/", "avatars/", "files/", "resume/", "content/",
}

func isExternalURL(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "data:") ||
		strings.HasPrefix(lower, "blob:")
}

func looksLikePresignedURL(value string) bool {
	if value == "" {
		return false
	}
	lower := strings.ToLower(value)
	return strings.Contains(lower, "x-amz-") ||
		strings.Contains(lower, "x-oss-") ||
		strings.Contains(lower, "signature=") ||
		strings.Contains(lower, "token=") ||
		strings.Contains(lower, "x-qiniu-") ||
		(strings.Contains(lower, "e=") && strings.Contains(lower, "token="))
}

func stripToObjectKey(pathOrKey string) string {
	normalized := strings.ReplaceAll(strings.TrimSpace(pathOrKey), "\\", "/")
	normalized = strings.TrimPrefix(normalized, "/")
	return normalized
}

// objectKeyCandidates 从 file id / object_name / 预签名 URL 推导可查库的对象键。
func objectKeyCandidates(value string) []string {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	add := func(item string) {
		key := stripToObjectKey(item)
		if key == "" || seen[key] {
			return
		}
		seen[key] = true
		out = append(out, key)
	}

	// 1. 外链：从 path 剥出 object_name，业务前缀优先
	if isExternalURL(raw) {
		if u, err := url.Parse(raw); err == nil {
			pathKey := stripToObjectKey(u.Path)
			add(pathKey)
			if pathKey != "" && strings.Contains(pathKey, "/") {
				parts := strings.SplitN(pathKey, "/", 2)
				if len(parts) == 2 {
					add(parts[1])
				}
			}
			preferred := make([]string, 0, len(out))
			rest := make([]string, 0, len(out))
			for _, c := range out {
				hit := false
				for _, p := range mediaPrefixes {
					if strings.HasPrefix(c, p) {
						hit = true
						break
					}
				}
				if hit {
					preferred = append(preferred, c)
				} else {
					rest = append(rest, c)
				}
			}
			return append(preferred, rest...)
		}
		return nil
	}

	// 2. 非 URL：原文与去掉首段前缀后的键都作为候选
	add(raw)
	if strings.Contains(raw, "/") {
		parts := strings.SplitN(stripToObjectKey(raw), "/", 2)
		if len(parts) == 2 {
			add(parts[1])
		}
	}
	return out
}

// PersistRef 写入库前规范化：预签名链 → object_name；稳定外链原样保留。
func PersistRef(ref string) string {
	// 1. 空串直接返回
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return ""
	}
	// 2. 外链：预签名剥成 object_name；稳定外链原样保留
	if isExternalURL(ref) {
		if looksLikePresignedURL(ref) {
			cands := objectKeyCandidates(ref)
			if len(cands) > 0 {
				return cands[0]
			}
			return ""
		}
		return ref
	}
	// 3. 站内相对路径保留；其余清洗为 object_name
	if strings.HasPrefix(ref, "/") {
		return ref
	}
	return stripToObjectKey(ref)
}
