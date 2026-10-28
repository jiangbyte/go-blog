package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultCategoryName = "其它"

type uploadOptions struct {
	Base     string
	Token    string
	Dir      string
	Glob     string
	DryRun   bool
	OnExists string // ask|overwrite|skip|create
	Files    []string
}

// runUpload 批量上传 md 为草稿：ensure 分类 → 图片改写 → 重名策略 → create/update。
func runUpload(opt uploadOptions) error {
	// 1. 解析凭证与目标文件列表
	base, token, err := resolveAuth(opt.Base, opt.Token)
	if err != nil {
		return err
	}
	if token == "" {
		return fmt.Errorf("未登录：请先执行 jb-md login，或设置 --token / JB_TOKEN")
	}
	files, err := collectMDFiles(opt.Files, opt.Dir, opt.Glob)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("未找到任何 markdown 文件")
	}
	onExists := strings.ToLower(strings.TrimSpace(opt.OnExists))
	if onExists == "" {
		onExists = "ask"
	}
	switch onExists {
	case "ask", "overwrite", "skip", "create":
	default:
		return fmt.Errorf("--on-exists 须为 ask|overwrite|skip|create")
	}
	if onExists == "ask" && !isTTY() {
		return fmt.Errorf("非交互终端不能使用 --on-exists=ask，请显式指定 overwrite|skip|create")
	}

	client := newAPIClient(base, token)
	taxCache := map[string]string{}
	imgCache := map[string]string{}

	// 2. 在处理任何 md 之前 resolve 默认分类（dry-run 允许暂无 uid）
	catUID, err := ensureCategory(client, defaultCategoryName, opt.DryRun, taxCache)
	if err != nil {
		return err
	}
	if !opt.DryRun && catUID == "" {
		return fmt.Errorf("无法解析默认分类「%s」", defaultCategoryName)
	}

	var failed int
	for _, path := range files {
		if err := uploadOne(client, path, catUID, onExists, opt.DryRun, imgCache); err != nil {
			fmt.Fprintf(os.Stderr, "失败 %s: %v\n", path, err)
			failed++
			continue
		}
	}
	if failed > 0 {
		return fmt.Errorf("%d/%d 个文件失败", failed, len(files))
	}
	fmt.Fprintf(os.Stderr, "完成：%d 个文件\n", len(files))
	return nil
}

func uploadOne(client *APIClient, path, catUID, onExists string, dryRun bool, imgCache map[string]string) error {
	fmt.Fprintf(os.Stderr, "处理 %s\n", path)
	// 1. 解析 frontmatter / 正文，并上传改写本地图片
	md, err := parseMarkdownFile(path)
	if err != nil {
		return err
	}
	body, cover, err := rewriteImages(client, md, dryRun, imgCache)
	if err != nil {
		return err
	}

	// 2. 按标题精确查重，决定 create / overwrite / skip
	title := strings.TrimSpace(md.Meta.Title)
	existing, err := client.findPostsByTitle(title, "")
	if err != nil {
		return err
	}

	action := "create"
	var targetUID string
	if len(existing) > 0 {
		if len(existing) > 1 {
			fmt.Fprintf(os.Stderr, "  警告: 标题「%s」匹配到 %d 篇，取最近更新 uid=%s\n", title, len(existing), existing[0].UID)
		}
		targetUID = existing[0].UID
		switch onExists {
		case "overwrite":
			action = "overwrite"
		case "skip":
			fmt.Fprintf(os.Stderr, "  跳过（已存在 uid=%s）\n", targetUID)
			return nil
		case "create":
			action = "create"
			title = nextAvailableTitle(client, title)
			targetUID = ""
		case "ask":
			fmt.Fprintf(os.Stderr, "  已存在「%s」uid=%s；覆盖将重置分类为「%s」、清空标签\n", title, targetUID, defaultCategoryName)
			choice, err := promptExists()
			if err != nil {
				return err
			}
			switch choice {
			case "o":
				action = "overwrite"
			case "s":
				fmt.Fprintf(os.Stderr, "  已跳过\n")
				return nil
			case "c":
				action = "create"
				title = nextAvailableTitle(client, title)
				targetUID = ""
			}
		}
	}

	// 3. 组装草稿 payload（默认分类 + 空标签）；dry-run 到此结束
	payload := map[string]any{
		"title":         title,
		"description":   md.Meta.Description,
		"date":          md.Meta.Date,
		"status":        "draft",
		"cover":         cover,
		"category_uids": []string{},
		"tag_uids":      []string{},
		"body":          body,
	}
	if catUID != "" {
		payload["category_uids"] = []string{catUID}
	}

	if dryRun {
		fmt.Fprintf(os.Stderr, "  [dry-run] %s title=%q cover=%q body=%d bytes\n", action, title, cover, len(body))
		return nil
	}

	// 4. 覆盖走 update，否则 create
	if action == "overwrite" {
		payload["uid"] = targetUID
		p, err := client.updatePost(payload)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "  已覆盖草稿 uid=%s title=%q\n", p.UID, p.Title)
		return nil
	}
	p, err := client.createPost(payload)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "  已创建草稿 uid=%s title=%q\n", p.UID, p.Title)
	return nil
}

func nextAvailableTitle(client *APIClient, base string) string {
	title := base
	for n := 2; n < 1000; n++ {
		list, err := client.findPostsByTitle(title, "")
		if err != nil || len(list) == 0 {
			return title
		}
		title = fmt.Sprintf("%s (%d)", base, n)
	}
	return fmt.Sprintf("%s (%d)", base, 1000)
}

func promptExists() (string, error) {
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, "  [o]verwrite / [s]kip / [c]reate 新标题? ")
		line, err := r.ReadString('\n')
		if err != nil {
			return "", err
		}
		line = strings.ToLower(strings.TrimSpace(line))
		if line == "" {
			continue
		}
		switch line[0] {
		case 'o', 's', 'c':
			return string(line[0]), nil
		}
		fmt.Fprintln(os.Stderr, "  请输入 o / s / c")
	}
}

func isTTY() bool {
	st, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (st.Mode() & os.ModeCharDevice) != 0
}

func usage() string {
	exe := "jb-md"
	if len(os.Args) > 0 {
		exe = filepath.Base(os.Args[0])
	}
	return strings.TrimSpace(fmt.Sprintf(`
用法:
  %[1]s login [--base URL]
  %[1]s logout
  %[1]s upload [flags] [file.md ...]

本机运行，默认上传到正式站 API（https://example.com/api）。
本机 API 调试：--base http://127.0.0.1:8000 或 JB_API_BASE。

upload flags:
  --base URL              API 地址（默认凭证或 https://example.com/api）
  --token TOKEN           覆盖凭证中的 token
  --dir DIR               扫描目录
  --glob PATTERN          默认 **/*.md
  --dry-run               只解析不写库（分类不存在时仅警告）
  --on-exists MODE        ask|overwrite|skip|create（默认 ask）

环境变量: JB_API_BASE, JB_TOKEN
凭证文件: ~/.config/go-blog/credentials.json
`, exe))
}
