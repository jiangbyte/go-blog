package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// APIClient 调用本站 Admin API（X-Token）。
type APIClient struct {
	Base   string
	Token  string
	HTTP   *http.Client
	stderr io.Writer
}

func newAPIClient(base, token string) *APIClient {
	return &APIClient{
		Base:  strings.TrimRight(base, "/"),
		Token: token,
		HTTP:  &http.Client{Timeout: 120 * time.Second},
		stderr: os.Stderr,
	}
}

type apiEnvelope struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (c *APIClient) doJSON(method, path string, query url.Values, body any, out any) error {
	var rdr io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(raw)
	}
	u := c.Base + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequest(method, u, rdr)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("X-Token", c.Token)
	}
	req.Header.Set("X-Go-Blog-CLI", "1")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var env apiEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return fmt.Errorf("非法响应 (%d): %s", resp.StatusCode, truncate(string(raw), 200))
	}
	if env.Code != "0000" {
		msg := env.Message
		if msg == "" {
			msg = "请求失败"
		}
		return fmt.Errorf("%s (%s)", msg, env.Code)
	}
	if out == nil || len(env.Data) == 0 || string(env.Data) == "null" {
		return nil
	}
	return json.Unmarshal(env.Data, out)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

type postDTO struct {
	UID         string `json:"uid"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Date        string `json:"date"`
	Status      string `json:"status"`
	Cover       string `json:"cover"`
	Body        string `json:"body"`
}

type taxonomyOpt struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
}

func (c *APIClient) findPostsByTitle(title, status string) ([]postDTO, error) {
	q := url.Values{}
	q.Set("title", title)
	if status != "" {
		q.Set("status", status)
	}
	var list []postDTO
	if err := c.doJSON(http.MethodGet, "/admin/blog/posts/by-title", q, nil, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *APIClient) getPost(uid string) (*postDTO, error) {
	q := url.Values{}
	q.Set("uid", uid)
	var p postDTO
	if err := c.doJSON(http.MethodGet, "/admin/blog/posts/detail", q, nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *APIClient) createPost(body map[string]any) (*postDTO, error) {
	var p postDTO
	if err := c.doJSON(http.MethodPost, "/admin/blog/posts/create", nil, body, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *APIClient) updatePost(body map[string]any) (*postDTO, error) {
	var p postDTO
	if err := c.doJSON(http.MethodPost, "/admin/blog/posts/update", nil, body, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *APIClient) taxonomyOptions(kind string) ([]taxonomyOpt, error) {
	q := url.Values{}
	q.Set("kind", kind)
	var list []taxonomyOpt
	if err := c.doJSON(http.MethodGet, "/admin/blog/taxonomy/options", q, nil, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (c *APIClient) createTaxonomy(kind, name string) (string, error) {
	var row taxonomyOpt
	if err := c.doJSON(http.MethodPost, "/admin/blog/taxonomy/create", nil, map[string]any{
		"kind": kind,
		"name": name,
	}, &row); err != nil {
		return "", err
	}
	return row.UID, nil
}

// uploadFile 上传本地文件，返回 object_name。
func (c *APIClient) uploadFile(localPath string) (string, error) {
	f, err := os.Open(localPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", filepath.Base(localPath))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, f); err != nil {
		return "", err
	}
	if err := w.WriteField("category", "uploads"); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, c.Base+"/admin/sys/file/upload", &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	if c.Token != "" {
		req.Header.Set("X-Token", c.Token)
	}
	req.Header.Set("X-Go-Blog-CLI", "1")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var env apiEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return "", fmt.Errorf("上传响应非法: %s", truncate(string(raw), 200))
	}
	if env.Code != "0000" {
		return "", fmt.Errorf("上传失败: %s", env.Message)
	}
	var data struct {
		ObjectName string `json:"object_name"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return "", err
	}
	if data.ObjectName == "" {
		return "", fmt.Errorf("上传成功但缺少 object_name")
	}
	return data.ObjectName, nil
}
