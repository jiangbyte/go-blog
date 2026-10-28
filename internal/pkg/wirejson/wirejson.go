package wirejson

import (
	"encoding/json"
	"strconv"
	"strings"
)

// Bool 兼容 JSON bool / "true"|"false" / 1|0（对齐 admin FE stringifyScalars）。
type Bool bool

func (b *Bool) UnmarshalJSON(data []byte) error {
	data = trimSpaceBytes(data)
	if string(data) == "null" {
		*b = false
		return nil
	}
	var v bool
	if err := json.Unmarshal(data, &v); err == nil {
		*b = Bool(v)
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*b = Bool(s == "true" || s == "1")
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		i, _ := n.Int64()
		*b = Bool(i != 0)
		return nil
	}
	*b = false
	return nil
}

func (b Bool) Bool() bool { return bool(b) }

// Int 兼容 JSON number / 数字字符串（admin axios stringifyScalars 会把 number 变成 string）。
type Int int

func (n *Int) UnmarshalJSON(data []byte) error {
	data = trimSpaceBytes(data)
	if string(data) == "null" || len(data) == 0 {
		*n = 0
		return nil
	}
	// 1. 纯整型
	var v int
	if err := json.Unmarshal(data, &v); err == nil {
		*n = Int(v)
		return nil
	}
	// 2. 数字字符串（含 "10" / "10.0"）
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		s = strings.TrimSpace(s)
		if s == "" {
			*n = 0
			return nil
		}
		if i, err := strconv.Atoi(s); err == nil {
			*n = Int(i)
			return nil
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil && !isNaN(f) {
			*n = Int(int(f))
			return nil
		}
		*n = 0
		return nil
	}
	// 3. JSON number（含 10.5）
	var num json.Number
	if err := json.Unmarshal(data, &num); err == nil {
		if i, err := num.Int64(); err == nil {
			*n = Int(i)
			return nil
		}
		if f, err := num.Float64(); err == nil && !isNaN(f) {
			*n = Int(int(f))
			return nil
		}
	}
	*n = 0
	return nil
}

func (n Int) Int() int { return int(n) }

func isNaN(f float64) bool { return f != f }

func trimSpaceBytes(data []byte) []byte {
	return []byte(strings.TrimSpace(string(data)))
}
