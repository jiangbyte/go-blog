package file

import "testing"

func TestCollectMarkdownMediaRefs(t *testing.T) {
	raw := `
![a](uploads/2026/10/13/a.png)
<img src="uploads/2026/10/13/b.png" alt="x" />
![ext](https://cdn.example.com/x.png)
![root](/uploads/2026/10/13/c.png)
`
	got := CollectMarkdownMediaRefs(raw)
	want := map[string]bool{
		"uploads/2026/10/13/a.png":  true,
		"uploads/2026/10/13/b.png":  true,
		"/uploads/2026/10/13/c.png": true,
	}
	if len(got) != len(want) {
		t.Fatalf("len=%d got=%v", len(got), got)
	}
	for _, g := range got {
		if !want[g] {
			t.Fatalf("unexpected ref %q in %v", g, got)
		}
	}
}

func TestIsMediaObjectRef(t *testing.T) {
	if !isMediaObjectRef("uploads/a.png") {
		t.Fatal("uploads should match")
	}
	if isMediaObjectRef("https://images.example.com/uploads/a.png") {
		t.Fatal("stable https should not need rewrite")
	}
}
