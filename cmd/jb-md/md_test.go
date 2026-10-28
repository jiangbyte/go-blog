package main

import "testing"

func TestSplitFrontMatter(t *testing.T) {
	text := "---\ntitle: Hello\ndescription: desc\ncover: ./a.png\n---\n# Body\n"
	fm, body := splitFrontMatter(text)
	if fm.Title != "Hello" || fm.Description != "desc" || fm.Cover != "./a.png" {
		t.Fatalf("meta=%+v", fm)
	}
	if body != "# Body\n" {
		t.Fatalf("body=%q", body)
	}
}

func TestMatchGlob(t *testing.T) {
	ok, err := matchGlob("**/*.md", "a/b/c.md")
	if err != nil || !ok {
		t.Fatalf("**/*.md: %v %v", ok, err)
	}
	ok, err = matchGlob("**/*.md", "a/b/c.txt")
	if err != nil || ok {
		t.Fatalf("should reject txt")
	}
	ok, err = matchGlob("notes/*.md", "notes/a.md")
	if err != nil || !ok {
		t.Fatalf("notes/*.md: %v %v", ok, err)
	}
}

func TestResolveLocalImageRemote(t *testing.T) {
	if resolveLocalImage("/tmp", "https://example.com/a.png") != "" {
		t.Fatal("remote should skip")
	}
	if resolveLocalImage("/tmp", "data:image/png;base64,xx") != "" {
		t.Fatal("data url should skip")
	}
}
