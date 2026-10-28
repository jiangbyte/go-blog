package blog

import (
	"context"
	"encoding/xml"
	"strings"
	"time"

	"github.com/jiangbyte/go-blog/internal/model"
)

// SitemapURLSet 符合 sitemaps.org 协议的 urlset。
type SitemapURLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []SitemapURL `xml:"url"`
}

// SitemapURL 单条 URL。
type SitemapURL struct {
	Loc        string  `xml:"loc"`
	LastMod    string  `xml:"lastmod,omitempty"`
	ChangeFreq string  `xml:"changefreq,omitempty"`
	Priority   float64 `xml:"priority,omitempty"`
}

// BuildSitemapXML 生成已发布内容的 sitemap（供 Bing/Google 发现）。
func (s *Service) BuildSitemapXML(ctx context.Context, baseURL string) ([]byte, error) {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = "https://example.com"
	}

	urls := []SitemapURL{
		{Loc: base + "/", ChangeFreq: "daily", Priority: 1.0},
		{Loc: base + "/archives", ChangeFreq: "daily", Priority: 0.9},
		{Loc: base + "/projects", ChangeFreq: "weekly", Priority: 0.8},
		{Loc: base + "/categories", ChangeFreq: "weekly", Priority: 0.6},
		{Loc: base + "/tags", ChangeFreq: "weekly", Priority: 0.6},
		{Loc: base + "/about", ChangeFreq: "monthly", Priority: 0.5},
	}

	// 1. 已发布文章
	var posts []model.BlogPost
	if err := s.db.WithContext(ctx).
		Select("uid", "updated_at", "published_at", "created_at").
		Where("status = ?", StatusPublished).
		Order("id DESC").
		Find(&posts).Error; err != nil {
		return nil, err
	}
	for _, p := range posts {
		urls = append(urls, SitemapURL{
			Loc:        base + "/p/" + p.UID,
			LastMod:    sitemapLastMod(p.ContentUpdatedAt, p.PublishedAt, p.CreatedAt),
			ChangeFreq: "weekly",
			Priority:   0.8,
		})
	}

	// 2. 已发布项目
	var projects []model.BlogProject
	if err := s.db.WithContext(ctx).
		Select("uid", "published_at", "created_at").
		Where("status = ?", StatusPublished).
		Order("id DESC").
		Find(&projects).Error; err != nil {
		return nil, err
	}
	for _, p := range projects {
		urls = append(urls, SitemapURL{
			Loc:        base + "/projects/" + p.UID,
			LastMod:    sitemapLastMod(nil, p.PublishedAt, p.CreatedAt),
			ChangeFreq: "monthly",
			Priority:   0.7,
		})
	}

	// 3. 有内容的分类 / 标签 / 项目标签页，便于搜索引擎发现聚合页
	if cats, err := s.Categories(ctx); err == nil {
		for _, c := range cats {
			uid, _ := c["uid"].(string)
			if uid == "" {
				continue
			}
			urls = append(urls, SitemapURL{
				Loc: base + "/categories/" + uid, ChangeFreq: "weekly", Priority: 0.5,
			})
		}
	}
	if tags, err := s.Tags(ctx); err == nil {
		for _, t := range tags {
			uid, _ := t["uid"].(string)
			if uid == "" {
				continue
			}
			urls = append(urls, SitemapURL{
				Loc: base + "/tags/" + uid, ChangeFreq: "weekly", Priority: 0.4,
			})
		}
	}
	if ptags, err := s.ProjectTags(ctx); err == nil {
		for _, t := range ptags {
			uid, _ := t["uid"].(string)
			if uid == "" {
				continue
			}
			urls = append(urls, SitemapURL{
				Loc: base + "/project-tags/" + uid, ChangeFreq: "monthly", Priority: 0.4,
			})
		}
	}

	set := SitemapURLSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}
	out, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), out...), nil
}

func sitemapLastMod(updated *time.Time, published, created time.Time) string {
	t := created
	if !published.IsZero() {
		t = published
	}
	if updated != nil && !updated.IsZero() {
		t = *updated
	}
	return t.UTC().Format("2006-01-02")
}
