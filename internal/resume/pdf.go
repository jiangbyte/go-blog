package resume

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

var unsafeFilenameRE = regexp.MustCompile(`[^\p{L}\p{N}._\-\s]+`)

// ExportPDFBytes 用 chromedp 将受控 HTML 打成 A4 PDF（禁外网导航到任意 URL）。
func ExportPDFBytes(html string) ([]byte, error) {
	if strings.TrimSpace(html) == "" {
		return nil, ErrBadRequest("empty html")
	}

	// 1. 写临时文件，避免超长 data URL；仅本进程可读
	dir, err := os.MkdirTemp("", "resume-pdf-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	htmlPath := filepath.Join(dir, "resume.html")
	if err := os.WriteFile(htmlPath, []byte(html), 0o600); err != nil {
		return nil, err
	}
	fileURL := "file://" + htmlPath

	// 2. 容器友好启动参数：无沙箱、禁用 crashpad、独立 user-data（避免 Alpine 无 HOME 时崩）
	userData := filepath.Join(dir, "chrome-user-data")
	if err := os.MkdirAll(userData, 0o700); err != nil {
		return nil, err
	}
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("disable-crash-reporter", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("no-crashpad", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.UserDataDir(userData),
	)
	if chromePath := strings.TrimSpace(os.Getenv("CHROME_PATH")); chromePath != "" {
		opts = append(opts, chromedp.ExecPath(chromePath))
	}
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	var pdf []byte
	err = chromedp.Run(ctx,
		chromedp.Navigate(fileURL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// 3. 仅打印当前文档，不跟随外链导航
			buf, _, err := page.PrintToPDF().
				WithPrintBackground(true).
				WithPreferCSSPageSize(true).
				WithPaperWidth(8.27).
				WithPaperHeight(11.69).
				WithMarginTop(0).
				WithMarginBottom(0).
				WithMarginLeft(0).
				WithMarginRight(0).
				Do(ctx)
			if err != nil {
				return err
			}
			pdf = buf
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("chromedp pdf: %w", err)
	}
	if len(pdf) == 0 {
		return nil, ErrBadRequest("pdf empty")
	}
	return pdf, nil
}

// SafePDFFilename 导出文件名。
func SafePDFFilename(doc map[string]any) string {
	profile, _ := doc["profile_json"].(map[string]any)
	name := ""
	if profile != nil {
		name = strings.TrimSpace(asString(profile["name"]))
	}
	if name == "" {
		name = strings.TrimSpace(asString(doc["title"]))
	}
	if name == "" {
		name = "resume"
	}
	name = unsafeFilenameRE.ReplaceAllString(name, "")
	name = strings.TrimSpace(name)
	if name == "" {
		name = "resume"
	}
	if len(name) > 80 {
		name = name[:80]
	}
	return name + ".pdf"
}
