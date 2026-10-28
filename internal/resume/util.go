package resume

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func stringsTrim(s string) string { return strings.TrimSpace(s) }

func stringsContainsAny(s string, parts ...string) bool {
	for _, p := range parts {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}

func coalesce(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func asString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	default:
		if v == nil {
			return ""
		}
		return fmt.Sprint(v)
	}
}

func asFloat(v any, def float64) float64 {
	switch t := v.(type) {
	case float64:
		if t != t {
			return def
		}
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case string:
		n, err := strconv.ParseFloat(strings.TrimSpace(t), 64)
		if err != nil || n != n {
			return def
		}
		return n
	default:
		return def
	}
}

func asInt(v any, def int) int {
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(t))
		if err != nil {
			return def
		}
		return n
	default:
		return def
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clampFloat(v any, def, lo, hi float64) float64 {
	n := asFloat(v, def)
	if n < lo {
		return lo
	}
	if n > hi {
		return hi
	}
	return n
}

func clampInt(v any, def, lo, hi int) int {
	n := asInt(v, def)
	if n < lo {
		return lo
	}
	if n > hi {
		return hi
	}
	return n
}

func fmtTime(t *time.Time) any {
	if t == nil || t.IsZero() {
		return nil
	}
	return t.UTC().Format(time.RFC3339Nano)
}

/** asMapSlice 将 sections / fields 等异构切片统一为 []map[string]any。 */
func asMapSlice(v any) []map[string]any {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case []map[string]any:
		return t
	case []any:
		out := make([]map[string]any, 0, len(t))
		for _, item := range t {
			if m, ok := item.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	case []map[string]string:
		out := make([]map[string]any, 0, len(t))
		for _, item := range t {
			m := make(map[string]any, len(item))
			for k, val := range item {
				m[k] = val
			}
			out = append(out, m)
		}
		return out
	default:
		return nil
	}
}
