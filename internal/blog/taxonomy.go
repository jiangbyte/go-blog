package blog

import (
	"context"
	"sort"

	"github.com/jiangbyte/go-blog/internal/model"
	"github.com/jiangbyte/go-blog/internal/pkg/batch"
	"github.com/jiangbyte/go-blog/internal/pkg/page"
)

// taxonomyByUIDs 按需分批 IN 加载，避免全表扫与超大 IN。
func (s *Service) taxonomyByUIDs(ctx context.Context, uids []string) (map[string]model.BlogTaxonomy, error) {
	// 1. 去重；空集合直接返回
	uids = batch.UniqueStrings(uids)
	m := make(map[string]model.BlogTaxonomy, len(uids))
	if len(uids) == 0 {
		return m, nil
	}
	// 2. 分批 IN 查询并合并到 map
	for _, chunk := range batch.ChunkStrings(uids, batch.DefaultSize) {
		var rows []model.BlogTaxonomy
		if err := s.db.WithContext(ctx).Where("uid IN ?", chunk).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, r := range rows {
			m[r.UID] = r
		}
	}
	return m, nil
}

func resolveTaxa(uids []string, m map[string]model.BlogTaxonomy) []Taxon {
	out := make([]Taxon, 0, len(uids))
	for _, uid := range uids {
		if t, ok := m[uid]; ok {
			out = append(out, Taxon{UID: t.UID, Name: t.Name})
		}
	}
	return out
}

func collectUIDsFromPosts(rows []model.BlogPost) []string {
	out := make([]string, 0, len(rows)*4)
	for _, r := range rows {
		out = append(out, model.JSONStrings(r.CategoriesJSON)...)
		out = append(out, model.JSONStrings(r.TagsJSON)...)
	}
	return batch.UniqueStrings(out)
}

func collectUIDsFromProjects(rows []model.BlogProject) []string {
	out := make([]string, 0, len(rows)*2)
	for _, r := range rows {
		out = append(out, model.JSONStrings(r.TagsJSON)...)
	}
	return batch.UniqueStrings(out)
}

func taxonomyListFromTaxons(counts map[string]int, names map[string]string) []map[string]any {
	type item struct {
		uid   string
		name  string
		count int
	}
	var items []item
	for uid, n := range counts {
		name := names[uid]
		if name == "" {
			name = uid
		}
		items = append(items, item{uid: uid, name: name, count: n})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].count != items[j].count {
			return items[i].count > items[j].count
		}
		return items[i].name < items[j].name
	})
	// 参考列表硬上限，避免无界返回
	if len(items) > page.MaxSize {
		items = items[:page.MaxSize]
	}
	out := make([]map[string]any, 0, len(items))
	for _, it := range items {
		out = append(out, map[string]any{
			"uid":   it.uid,
			"name":  it.name,
			"count": it.count,
		})
	}
	return out
}
