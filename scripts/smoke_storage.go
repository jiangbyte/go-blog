//go:build ignore

package main

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/jiangbyte/go-blog/internal/config"
	"github.com/jiangbyte/go-blog/internal/file"
	"github.com/jiangbyte/go-blog/internal/infra"
	"github.com/jiangbyte/go-blog/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := infra.OpenMySQL(cfg.MySQL, false)
	if err != nil {
		log.Fatal(err)
	}
	reg := storage.NewRegistry(cfg.Storage)
	fmt.Printf("status default=%s engines=%d\n", reg.Status().DefaultEngine, len(reg.Status().Engines))
	svc := file.NewService(db, reg)
	_ = svc.EnsureSchema(context.Background())
	png := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xde, 0x00, 0x00, 0x00,
		0x0c, 0x49, 0x44, 0x41, 0x54, 0x08, 0xd7, 0x63, 0xf8, 0xff, 0xff, 0x3f,
		0x00, 0x05, 0xfe, 0x02, 0xfe, 0xa7, 0x35, 0x81, 0x84, 0x00, 0x00, 0x00,
		0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	}
	out, err := svc.Upload(context.Background(), file.UploadInput{
		Filename: "smoke.png", Content: png, Category: "uploads", UploaderID: "smoke",
	})
	if err != nil {
		log.Fatalf("upload: %v", err)
	}
	st, _, err := reg.Get("minio")
	if err != nil {
		log.Fatal(err)
	}
	data, err := st.GetBytes(context.Background(), out["object_name"].(string))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ok provider=%v bytes=%d equal=%v url_len=%d\n",
		out["storage_provider"], len(data), bytes.Equal(data, png), len(fmt.Sprint(out["url"])))
}
