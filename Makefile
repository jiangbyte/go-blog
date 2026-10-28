# go-blog 本地开发
ROOT        := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
WEB         := $(ROOT)/web
SQL_DIR     := $(ROOT)/docs/sql

API_HOST    ?= 127.0.0.1
API_PORT    ?= 8000
BLOG_URL    := http://127.0.0.1:5173
ADMIN_URL   := http://127.0.0.1:5174
RESUME_URL  := http://127.0.0.1:5175
API_URL     := http://$(API_HOST):$(API_PORT)

MYSQL_HOST     ?= $(shell sed -n 's/^MYSQL_HOST=//p' $(ROOT)/.env 2>/dev/null | head -1)
MYSQL_PORT     ?= $(shell sed -n 's/^MYSQL_PORT=//p' $(ROOT)/.env 2>/dev/null | head -1)
MYSQL_USER     ?= $(shell sed -n 's/^MYSQL_USER=//p' $(ROOT)/.env 2>/dev/null | head -1)
MYSQL_PASSWORD ?= $(shell sed -n 's/^MYSQL_PASSWORD=//p' $(ROOT)/.env 2>/dev/null | head -1)
MYSQL_DATABASE ?= $(shell sed -n 's/^MYSQL_DATABASE=//p' $(ROOT)/.env 2>/dev/null | head -1)
MYSQL_DOCKER   ?= $(shell sed -n 's/^MYSQL_DOCKER=//p' $(ROOT)/.env 2>/dev/null | head -1)

MYSQL_HOST     := $(if $(MYSQL_HOST),$(MYSQL_HOST),127.0.0.1)
MYSQL_PORT     := $(if $(MYSQL_PORT),$(MYSQL_PORT),3306)
MYSQL_USER     := $(if $(MYSQL_USER),$(MYSQL_USER),root)
MYSQL_DATABASE := $(if $(MYSQL_DATABASE),$(MYSQL_DATABASE),go_blog)
MYSQL_DOCKER   := $(if $(MYSQL_DOCKER),$(MYSQL_DOCKER),mysql)

MYSQL_DOCKER_EXEC := docker exec -i $(MYSQL_DOCKER) mysql -u$(MYSQL_USER) --password=$(MYSQL_PASSWORD) --default-character-set=utf8mb4 $(MYSQL_DATABASE)

export GOPROXY ?= https://goproxy.cn,direct

.DEFAULT_GOAL := help

.PHONY: help env install install-be install-fe \
	dev api blog admin resume \
	build build-api build-blog build-admin build-resume jb-md \
	docker-up docker-down docker-logs docker-ps docker-push \
	sql sql-list sql-one open

DEPLOY := $(ROOT)/deploy
# 推送时打 :IMAGE_VERSION 与 :latest；线上 compose 只拉 :latest
IMAGE_VERSION ?= 1.0.0

help: ## 显示命令
	@echo "go-blog 开发命令"
	@echo ""
	@echo "  make env / install / install-be / install-fe"
	@echo "  make sql                 执行 docs/sql（需本机 Docker mysql）"
	@echo "  make dev                 api + blog + admin + resume"
	@echo "  make api / blog / admin / resume"
	@echo "  make build"
	@echo "  make jb-md               构建本机 Markdown 上传 CLI（默认连正式站 API）→ bin/jb-md"
	@echo "  make docker-up / docker-down / docker-logs / docker-ps"
	@echo "  make docker-push         推送 ACR（含 go-blog-resume）"
	@echo ""
	@echo "依赖本机容器: mysql / redis；对象存储默认七牛（见 .env.example）"
	@echo "Docker 部署见 deploy/README.md"
	@echo "jb-md 用法见 cmd/jb-md/README.md"

env: ## 复制 .env.example
	@test -f $(ROOT)/.env || cp $(ROOT)/.env.example $(ROOT)/.env
	@echo ".env ready"

install: install-be install-fe ## 安装前后端依赖

install-be: ## Go 依赖
	cd $(ROOT) && go mod download

install-fe: ## pnpm install
	cd $(WEB) && pnpm install

sql: ## 按文件名顺序执行 docs/sql/*.sql，并写入 Markdown 厨房水槽正文
	@for f in $$(ls $(SQL_DIR)/*.sql | sort); do \
	  echo "==> $$f"; \
	  $(MYSQL_DOCKER_EXEC) < $$f; \
	done
	@MYSQL_DOCKER=$(MYSQL_DOCKER) MYSQL_USER=$(MYSQL_USER) MYSQL_PASSWORD=$(MYSQL_PASSWORD) MYSQL_DATABASE=$(MYSQL_DATABASE) \
	  python3 $(ROOT)/scripts/seed-markdown-body.py
	@echo "sql done → $(MYSQL_DATABASE)"

sql-list: ## 列出 SQL
	@ls -1 $(SQL_DIR)/*.sql | sort

sql-one: ## 执行单个 SQL：make sql-one F=01-schema.sql
	@test -n "$(F)" || (echo "用法: make sql-one F=01-schema.sql" && exit 1)
	$(MYSQL_DOCKER_EXEC) < $(SQL_DIR)/$(F)

api: env ## 启动 Gin API
	cd $(ROOT) && go run ./cmd/server

blog: ## 启动 blog (5173)
	cd $(WEB) && pnpm --filter @jiangbyte/blog dev

admin: ## 启动 admin (5174)
	cd $(WEB) && pnpm --filter @jiangbyte/admin dev

resume: ## 启动 resume (5175)
	cd $(WEB) && pnpm --filter @jiangbyte/resume dev

dev: env ## 一键 api + blog + admin + resume
	@echo "API $(API_URL) | Blog $(BLOG_URL) | Admin $(ADMIN_URL) | Resume $(RESUME_URL)"
	@bash -c 'set -e; \
	  (cd $(ROOT) && go run ./cmd/server) & \
	  API_PID=$$!; \
	  (cd $(WEB) && pnpm --filter @jiangbyte/blog dev) & \
	  BLOG_PID=$$!; \
	  (cd $(WEB) && pnpm --filter @jiangbyte/admin dev) & \
	  ADMIN_PID=$$!; \
	  (cd $(WEB) && pnpm --filter @jiangbyte/resume dev) & \
	  RESUME_PID=$$!; \
	  trap "kill $$API_PID $$BLOG_PID $$ADMIN_PID $$RESUME_PID 2>/dev/null" INT TERM EXIT; \
	  wait'

build: build-api build-blog build-admin build-resume

build-api:
	cd $(ROOT) && go build -o bin/server ./cmd/server

jb-md: ## 构建本地 Markdown 上传 CLI → bin/jb-md
	cd $(ROOT) && go build -o bin/jb-md ./cmd/jb-md

build-blog:
	cd $(WEB) && pnpm --filter @jiangbyte/blog build

build-admin:
	cd $(WEB) && pnpm --filter @jiangbyte/admin build

build-resume:
	cd $(WEB) && pnpm --filter @jiangbyte/resume build

docker-up: ## 构建并启动 deploy/compose（入口 :8088）
	@test -f $(DEPLOY)/.env || cp $(DEPLOY)/.env.example $(DEPLOY)/.env
	cd $(DEPLOY) && docker compose up -d --build

docker-down: ## 停止 deploy/compose
	cd $(DEPLOY) && docker compose down

docker-logs: ## 跟踪 deploy/compose 日志
	cd $(DEPLOY) && docker compose logs -f --tail=200

docker-ps: ## 查看 deploy/compose 状态
	cd $(DEPLOY) && docker compose ps

docker-push: ## 构建并推送 ACR（:IMAGE_VERSION + :latest；默认 1.0.1）
	$(DEPLOY)/scripts/docker-build-push.sh $(IMAGE_VERSION)

open:
	@echo $(API_URL) $(BLOG_URL) $(ADMIN_URL) $(RESUME_URL)
