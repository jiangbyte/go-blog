// Author: go-blog — 分批处理 ID 列表，避免超大 IN 子句

package batch

import "strings"

const DefaultSize = 200

// MaxINSize 单次 IN 查询 ID 数量上限。
const MaxINSize = 500

// ChunkStrings 按 size 切分；size<=0 时用 DefaultSize。
func ChunkStrings(ids []string, size int) [][]string {
	if len(ids) == 0 {
		return nil
	}
	if size <= 0 {
		size = DefaultSize
	}
	out := make([][]string, 0, (len(ids)+size-1)/size)
	for i := 0; i < len(ids); i += size {
		end := i + size
		if end > len(ids) {
			end = len(ids)
		}
		out = append(out, ids[i:end])
	}
	return out
}

// UniqueStrings trim、去空、去重，保持首次出现顺序。
func UniqueStrings(ids []string) []string {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
