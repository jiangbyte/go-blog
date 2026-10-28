package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

type frontMatter struct {
	Title       string
	Description string
	Date        string
	Cover       string
}

type parsedMD struct {
	Path string
	Meta frontMatter
	Body string
}

var (
	mdImageRe  = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	htmlImageRe = regexp.MustCompile(`(?i)<img\b[^>]*\bsrc\s*=\s*["']([^"']+)["'][^>]*>`)
)

// parseMarkdownFile 读取 md，拆 frontmatter 与正文。
func parseMarkdownFile(path string) (*parsedMD, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	text := string(raw)
	meta, body := splitFrontMatter(text)
	if strings.TrimSpace(meta.Title) == "" {
		meta.Title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	return &parsedMD{Path: path, Meta: meta, Body: body}, nil
}

func splitFrontMatter(text string) (frontMatter, string) {
	var fm frontMatter
	s := strings.TrimLeftFunc(text, unicode.IsSpace)
	if !strings.HasPrefix(s, "---\n") && !strings.HasPrefix(s, "---\r\n") {
		return fm, text
	}
	rest := s[3:]
	if strings.HasPrefix(rest, "\r\n") {
		rest = rest[2:]
	} else if strings.HasPrefix(rest, "\n") {
		rest = rest[1:]
	}
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return fm, text
	}
	block := rest[:end]
	after := rest[end+1:] // starts with ---
	after = strings.TrimPrefix(after, "---")
	after = strings.TrimPrefix(after, "\r")
	after = strings.TrimPrefix(after, "\n")

	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		val = strings.TrimSpace(val)
		val = strings.Trim(val, `"'`)
		switch key {
		case "title":
			fm.Title = val
		case "description", "summary":
			fm.Description = val
		case "date":
			fm.Date = val
		case "cover":
			fm.Cover = val
		}
	}
	return fm, after
}

func isRemoteOrDataURL(ref string) bool {
	ref = strings.TrimSpace(ref)
	lower := strings.ToLower(ref)
	return strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "//") ||
		strings.HasPrefix(lower, "data:")
}

func isAbsLocalPath(ref string) bool {
	ref = strings.TrimSpace(ref)
	if filepath.IsAbs(ref) {
		return true
	}
	// Windows: C:\...
	if len(ref) >= 3 && unicode.IsLetter(rune(ref[0])) && ref[1] == ':' && (ref[2] == '\\' || ref[2] == '/') {
		return true
	}
	return false
}

// resolveLocalImage 将 markdown 中的图片引用解析为本地绝对路径；非本地返回空。
func resolveLocalImage(mdDir, ref string) string {
	ref = strings.TrimSpace(ref)
	if ref == "" || isRemoteOrDataURL(ref) {
		return ""
	}
	// 去掉 title 部分：path "title"
	if i := strings.IndexAny(ref, " \t"); i > 0 {
		ref = strings.TrimSpace(ref[:i])
	}
	ref = strings.Trim(ref, `"'`)
	var abs string
	if isAbsLocalPath(ref) {
		abs = filepath.Clean(ref)
	} else {
		abs = filepath.Clean(filepath.Join(mdDir, ref))
	}
	st, err := os.Stat(abs)
	if err != nil || st.IsDir() {
		return ""
	}
	return abs
}

// rewriteImages 上传本地图并改写 body / cover 为 object_name。
func rewriteImages(client *APIClient, md *parsedMD, dryRun bool, cache map[string]string) (body, cover string, err error) {
	mdDir := filepath.Dir(md.Path)
	body = md.Body
	cover = strings.TrimSpace(md.Meta.Cover)

	// 1. 同路径去重上传；dry-run 只生成占位 object_name
	uploadOne := func(local string) (string, error) {
		if obj, ok := cache[local]; ok {
			return obj, nil
		}
		if dryRun {
			cache[local] = "uploads/(dry-run)/" + filepath.Base(local)
			return cache[local], nil
		}
		obj, err := client.uploadFile(local)
		if err != nil {
			return "", fmt.Errorf("上传图片 %s: %w", local, err)
		}
		fmt.Fprintf(os.Stderr, "  已上传图片 %s → %s\n", local, obj)
		cache[local] = obj
		return obj, nil
	}

	// 2. 改写 Markdown / HTML 图片引用中的本地路径
	body = mdImageRe.ReplaceAllStringFunc(body, func(m string) string {
		if err != nil {
			return m
		}
		sub := mdImageRe.FindStringSubmatch(m)
		if len(sub) < 3 {
			return m
		}
		alt, ref := sub[1], sub[2]
		local := resolveLocalImage(mdDir, ref)
		if local == "" {
			return m
		}
		obj, e := uploadOne(local)
		if e != nil {
			err = e
			return m
		}
		return fmt.Sprintf("![%s](%s)", alt, obj)
	})
	if err != nil {
		return "", "", err
	}

	body = htmlImageRe.ReplaceAllStringFunc(body, func(m string) string {
		if err != nil {
			return m
		}
		sub := htmlImageRe.FindStringSubmatch(m)
		if len(sub) < 2 {
			return m
		}
		ref := sub[1]
		local := resolveLocalImage(mdDir, ref)
		if local == "" {
			return m
		}
		obj, e := uploadOne(local)
		if e != nil {
			err = e
			return m
		}
		return strings.Replace(m, ref, obj, 1)
	})
	if err != nil {
		return "", "", err
	}

	// 3. 封面若为本地文件则同样上传改写
	if cover != "" && !isRemoteOrDataURL(cover) {
		local := resolveLocalImage(mdDir, cover)
		if local != "" {
			obj, e := uploadOne(local)
			if e != nil {
				return "", "", e
			}
			cover = obj
		}
	}
	return body, cover, nil
}

// collectMDFiles 收集单文件或目录下匹配 glob 的 md。
func collectMDFiles(files []string, dir, globPattern string) ([]string, error) {
	seen := map[string]struct{}{}
	var out []string
	add := func(p string) error {
		abs, err := filepath.Abs(p)
		if err != nil {
			return err
		}
		st, err := os.Stat(abs)
		if err != nil {
			return err
		}
		if st.IsDir() {
			return fmt.Errorf("%s 是目录，请用 --dir", abs)
		}
		if _, ok := seen[abs]; ok {
			return nil
		}
		seen[abs] = struct{}{}
		out = append(out, abs)
		return nil
	}

	for _, f := range files {
		if err := add(f); err != nil {
			return nil, err
		}
	}
	if strings.TrimSpace(dir) == "" {
		return out, nil
	}
	root, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if globPattern == "" {
		globPattern = "**/*.md"
	}
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		ok, err := matchGlob(globPattern, rel)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		return add(path)
	})
	return out, err
}

// matchGlob 支持常见 **/*.md 与 filepath.Match 模式（路径用 /）。
func matchGlob(pattern, name string) (bool, error) {
	pattern = filepath.ToSlash(strings.TrimSpace(pattern))
	name = filepath.ToSlash(name)
	if pattern == "**/*.md" {
		return strings.HasSuffix(strings.ToLower(name), ".md"), nil
	}
	if strings.HasPrefix(pattern, "**/") {
		suf := pattern[3:]
		if ok, err := filepath.Match(suf, name); err != nil || ok {
			return ok, err
		}
		return filepath.Match(suf, filepath.Base(name))
	}
	if strings.Contains(pattern, "**") {
		// 后缀匹配：取最后一个 ** 之后的模式，在任意深度上比对
		idx := strings.LastIndex(pattern, "**")
		suf := strings.TrimPrefix(pattern[idx+2:], "/")
		if suf == "" {
			return true, nil
		}
		for rest := name; ; {
			ok, err := filepath.Match(suf, rest)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
			slash := strings.Index(rest, "/")
			if slash < 0 {
				return filepath.Match(suf, filepath.Base(name))
			}
			rest = rest[slash+1:]
		}
	}
	return filepath.Match(pattern, name)
}
