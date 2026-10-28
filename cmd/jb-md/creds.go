package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// 默认连正式站网关（/api 前缀由 Caddy handle_path 剥掉）；本机调试用 --base / JB_API_BASE 覆盖
	defaultAPIBase = "https://example.com/api"
	credsFileName  = "credentials.json"
)

// Credentials 本地落盘的 CLI 登录凭证。
type Credentials struct {
	Base     string `json:"base"`
	Token    string `json:"token"`
	Username string `json:"username"`
}

func configDir() (string, error) {
	if xdg := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); xdg != "" {
		return filepath.Join(xdg, "go-blog"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "go-blog"), nil
}

func credentialsPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, credsFileName), nil
}

func loadCredentials() (*Credentials, error) {
	path, err := credentialsPath()
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var c Credentials
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("解析凭证失败: %w", err)
	}
	return &c, nil
}

func saveCredentials(c Credentials) error {
	// 1. 确保配置目录存在
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	// 2. 序列化并以 0600 写入，避免他人读 token
	path := filepath.Join(dir, credsFileName)
	raw, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o600)
}

func removeCredentials() error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// resolveAuth 合并 flag/env/文件：token 与 base 的优先级见方案。
func resolveAuth(flagBase, flagToken string) (base, token string, err error) {
	creds, err := loadCredentials()
	if err != nil {
		return "", "", err
	}
	base = strings.TrimSpace(flagBase)
	if base == "" {
		base = strings.TrimSpace(os.Getenv("JB_API_BASE"))
	}
	if base == "" && creds != nil {
		base = strings.TrimSpace(creds.Base)
	}
	if base == "" {
		base = defaultAPIBase
	}
	base = strings.TrimRight(base, "/")

	token = strings.TrimSpace(flagToken)
	if token == "" {
		token = strings.TrimSpace(os.Getenv("JB_TOKEN"))
	}
	if token == "" && creds != nil {
		token = strings.TrimSpace(creds.Token)
	}
	return base, token, nil
}
