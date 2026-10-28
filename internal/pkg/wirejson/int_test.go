package wirejson

import (
	"encoding/json"
	"testing"
)

func TestIntUnmarshal(t *testing.T) {
	cases := []struct {
		raw string
		want int
	}{
		{`10`, 10},
		{`"10"`, 10},
		{`"10.0"`, 10},
		{`10.9`, 10},
		{`null`, 0},
		{`""`, 0},
	}
	for _, c := range cases {
		var n Int
		if err := json.Unmarshal([]byte(c.raw), &n); err != nil {
			t.Fatalf("%s err=%v", c.raw, err)
		}
		if n.Int() != c.want {
			t.Fatalf("%s got %d want %d", c.raw, n.Int(), c.want)
		}
	}
}

func TestSectionSaveSortCodeString(t *testing.T) {
	// 模拟 axios stringifyScalars 后的 section/save body
	raw := `{"resume_id":"r","id":"s","type":"education","title":"教育","icon":"lucide:x","sort_code":"10","meta_json":{},"body_html":"","entries_json":[]}`
	type body struct {
		SortCode Int `json:"sort_code"`
	}
	var b body
	if err := json.Unmarshal([]byte(raw), &b); err != nil {
		t.Fatal(err)
	}
	if b.SortCode.Int() != 10 {
		t.Fatalf("got %d", b.SortCode.Int())
	}
}
