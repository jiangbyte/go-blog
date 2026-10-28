package main

import (
	"fmt"
	"log"

	"github.com/jiangbyte/go-blog/internal/config"
	"github.com/jiangbyte/go-blog/internal/infra"
	"github.com/jiangbyte/go-blog/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := infra.OpenMySQL(cfg.MySQL, cfg.Debug)
	if err != nil {
		log.Fatalf("mysql: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	rdb, err := infra.OpenRedis(cfg.Redis)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer rdb.Close()

	engine, err := server.New(cfg, db, rdb)
	if err != nil {
		log.Fatal(err)
	}
	addr := fmt.Sprintf("%s:%d", cfg.AppHost, cfg.AppPort)
	log.Printf("%s listening on http://%s (gorm)", cfg.AppName, addr)
	if err := engine.Run(addr); err != nil {
		log.Fatal(err)
	}
}
