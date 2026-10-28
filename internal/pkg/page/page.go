// Package page 提供统一分页解析与响应结构，强制页大小上限。
package page

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

// MaxSize 列表页大小上限，禁止变相全量拉取。
const MaxSize = 100

// IPage 标准分页响应（records + 总数与页码信息）。
type IPage struct {
	Records interface{} `json:"records"`
	Total   int64       `json:"total"`
	Size    int         `json:"size"`
	Current int         `json:"current"`
	Pages   int         `json:"pages"`
}

// Parse 从 query 读取 current/size；size 强制 ≤ MaxSize。
func Parse(c *gin.Context, defaultSize int) (current, size int) {
	// 1. 默认第 1 页，页大小先按调用方缺省值钳制
	current = 1
	size = ClampSize(defaultSize, 12)
	// 2. 解析合法 current（>0）
	if v := c.Query("current"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			current = n
		}
	}
	// 3. 解析 size 并再次钳制到 MaxSize
	if v := c.Query("size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			size = ClampSize(n, defaultSize)
		}
	}
	return current, size
}

// ClampSize 校正页大小：无效时用 defaultSize，且不超过 MaxSize。
func ClampSize(size, defaultSize int) int {
	if defaultSize <= 0 {
		defaultSize = 12
	}
	if defaultSize > MaxSize {
		defaultSize = MaxSize
	}
	if size <= 0 {
		return defaultSize
	}
	if size > MaxSize {
		return MaxSize
	}
	return size
}

// Normalize 校正 current/size（current 默认 1）。
func Normalize(current, size, defaultSize int) (int, int) {
	if current <= 0 {
		current = 1
	}
	return current, ClampSize(size, defaultSize)
}

// Of 组装分页结果；current 超出总页数时回退到最后一页。
func Of(records interface{}, total int64, current, size int) IPage {
	// 1. 按 total/size 计算总页数
	pages := 0
	if size > 0 && total > 0 {
		pages = int(math.Ceil(float64(total) / float64(size)))
	}
	// 2. 当前页越界时钳到最后一页，避免空页误导
	if pages > 0 && current > pages {
		current = pages
	}
	// 3. records 为 nil 时返回空数组，保持 JSON 形态稳定
	if records == nil {
		records = []any{}
	}
	return IPage{
		Records: records,
		Total:   total,
		Size:    size,
		Current: current,
		Pages:   pages,
	}
}
