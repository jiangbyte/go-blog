package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	_ "embed"
)

//go:embed embed/login.html
var loginHTML []byte

var allowedProxyPaths = map[string]bool{
	"/auth/public-key":     true,
	"/auth/captcha/create": true,
	"/auth/captcha/verify": true,
	"/auth/login":          true,
}

// runLogin 起本地回调服务、打开内嵌登录页，完成后写入凭证。
func runLogin(base string) error {
	// 1. 规范化 API base，并解析为反代目标
	base = strings.TrimSpace(base)
	if base == "" {
		if c, err := loadCredentials(); err == nil && c != nil && c.Base != "" {
			base = c.Base
		}
	}
	if env := strings.TrimSpace(os.Getenv("JB_API_BASE")); base == "" && env != "" {
		base = env
	}
	if base == "" {
		base = defaultAPIBase
	}
	base = strings.TrimRight(base, "/")
	target, err := url.Parse(base)
	if err != nil || target.Scheme == "" || target.Host == "" {
		return fmt.Errorf("无效 API base: %s", base)
	}

	// 2. 监听 127.0.0.1 随机端口
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	port := ln.Addr().(*net.TCPAddr).Port
	localURL := fmt.Sprintf("http://127.0.0.1:%d/login?base=%s", port, url.QueryEscape(base))

	doneCh := make(chan Credentials, 1)
	errCh := make(chan error, 1)
	var once sync.Once
	finish := func(c Credentials, e error) {
		once.Do(func() {
			if e != nil {
				errCh <- e
				return
			}
			doneCh <- c
		})
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	origDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		origDirector(req)
		req.Host = target.Host
		req.Header.Set("X-Go-Blog-CLI", "1")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login?base="+url.QueryEscape(base), http.StatusFound)
	})
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(loginHTML)
	})
	mux.HandleFunc("/proxy/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/proxy")
		if !allowedProxyPaths[path] {
			http.Error(w, "proxy path not allowed", http.StatusForbidden)
			return
		}
		r.URL.Path = path
		r.URL.RawPath = path
		proxy.ServeHTTP(w, r)
	})
	mux.HandleFunc("/done", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			http.Error(w, "read body failed", http.StatusBadRequest)
			return
		}
		var body struct {
			Token    string `json:"token"`
			Username string `json:"username"`
		}
		if err := json.Unmarshal(raw, &body); err != nil || strings.TrimSpace(body.Token) == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"code":"4000","message":"token required","data":null}`))
			return
		}
		creds := Credentials{
			Base:     base,
			Token:    strings.TrimSpace(body.Token),
			Username: strings.TrimSpace(body.Username),
		}
		if err := saveCredentials(creds); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"code":"5000","message":"save credentials failed","data":null}`))
			finish(Credentials{}, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"0000","message":"ok","data":null}`))
		finish(creds, nil)
	})

	srv := &http.Server{Handler: mux}
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			finish(Credentials{}, err)
		}
	}()

	fmt.Fprintf(os.Stderr, "打开浏览器完成登录: %s\n", localURL)
	fmt.Fprintf(os.Stderr, "若未自动打开，请手动访问上述地址（5 分钟超时）\n")
	_ = openBrowser(localURL)

	// 3. 等待完成或超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	select {
	case creds := <-doneCh:
		_ = srv.Shutdown(context.Background())
		path, _ := credentialsPath()
		fmt.Fprintf(os.Stderr, "已写入凭证: %s (user=%s)\n", path, creds.Username)
		return nil
	case err := <-errCh:
		_ = srv.Shutdown(context.Background())
		return err
	case <-ctx.Done():
		_ = srv.Shutdown(context.Background())
		return fmt.Errorf("登录超时，请重试")
	}
}

func openBrowser(u string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", u)
	case "darwin":
		cmd = exec.Command("open", u)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", u)
	default:
		return fmt.Errorf("unsupported OS")
	}
	return cmd.Start()
}

func runLogout() error {
	if err := removeCredentials(); err != nil {
		return err
	}
	path, _ := credentialsPath()
	fmt.Fprintf(os.Stderr, "已删除凭证: %s\n", path)
	return nil
}
