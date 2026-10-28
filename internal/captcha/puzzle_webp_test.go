package captcha

import (
	"os"
	"testing"
)

func TestLooksLikeImageWebP(t *testing.T) {
	// RIFF....WEBP 最小头
	data := []byte("RIFF____WEBPXXXX")
	if !looksLikeImage(data) {
		t.Fatal("expected webp header to pass")
	}
	if looksLikeImage([]byte("<html>not image")) {
		t.Fatal("html should fail")
	}
}

func TestCreatePuzzleFromRemoteWebPFile(t *testing.T) {
	path := "/tmp/captcha-test2.bin"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Skip("no sample webp at", path)
	}
	puzzle, err := CreatePuzzleFromBytes(raw, 320, 160, 60)
	if err != nil {
		t.Fatal(err)
	}
	if puzzle.BGDataURL == "" || puzzle.PuzzleDataURL == "" {
		t.Fatal("empty data urls")
	}
	if puzzle.BGWidth != 320 || puzzle.BGHeight != 160 {
		t.Fatalf("size=%dx%d", puzzle.BGWidth, puzzle.BGHeight)
	}
}
