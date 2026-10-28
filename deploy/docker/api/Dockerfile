# Go API 生产镜像（构建上下文：仓库根；GOTOOLCHAIN=auto 可拉起 go.mod 工具链）
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/golang:1.24-alpine AS build

WORKDIR /src
ENV GOPROXY=https://goproxy.cn,direct \
    GOTOOLCHAIN=auto \
    CGO_ENABLED=0 \
    GOOS=linux

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
RUN go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/alpine:3.21

# Chromium：简历 PDF（chromedp）；需可写 HOME，否则 crashpad 报 --database is required
RUN apk add --no-cache ca-certificates tzdata chromium nss freetype harfbuzz ttf-dejavu \
        font-noto-cjk \
    && adduser -D -u 10001 -h /home/app app \
    && mkdir -p /home/app /tmp/chrome \
    && chown -R app:app /home/app /tmp/chrome

WORKDIR /app
COPY --from=build /out/server /app/server
RUN mkdir -p /app/data/uploads && chown -R app:app /app

USER app
EXPOSE 8000
ENV APP_HOST=0.0.0.0 \
    APP_PORT=8000 \
    TZ=Asia/Shanghai \
    HOME=/home/app \
    XDG_CONFIG_HOME=/home/app/.config \
    XDG_CACHE_HOME=/home/app/.cache \
    AUTH_RSA_KEY_PATH=/app/data/rsa_private.pem \
    CHROME_PATH=/usr/bin/chromium

CMD ["/app/server"]
