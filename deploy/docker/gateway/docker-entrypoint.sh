#!/bin/sh
set -eu

SNIPPET=/etc/caddy/Caddyfile.snippet

# 1. 按是否配置 SITE_DOMAIN 选择 HTTP-only 或自动 HTTPS
# 2. 可选 ACME_EMAIL 写入全局块
# 3. 可选 IMAGE_DOMAIN：Caddy 出图床证，HTTP 回源七牛（七牛 HTTPS 保持关）
# 4. 可选 RESUME_DOMAIN：简历子域名 → resume 静态容器 + 同源 /api
build_config() {
  if [ -n "${ACME_EMAIL:-}" ]; then
    printf '{\n\temail %s\n}\n\n' "$ACME_EMAIL"
  fi
  cat "$SNIPPET"
  echo
  cat "$1"
  if [ -n "${IMAGE_DOMAIN:-}" ] && [ -n "${IMAGE_UPSTREAM:-}" ]; then
    echo
    echo "# image CDN via Caddy → Qiniu HTTP upstream"
    cat /etc/caddy/Caddyfile.images
  fi
  if [ -n "${RESUME_DOMAIN:-}" ]; then
    echo
    echo "# resume subdomain → resume:80 + /api → api"
    cat /etc/caddy/Caddyfile.resume
  fi
}

if [ -n "${SITE_DOMAIN:-}" ]; then
  echo "Caddy: HTTP (:80 for IP) + HTTPS for ${SITE_DOMAIN}"
  if [ -n "${IMAGE_DOMAIN:-}" ]; then
    echo "Caddy: image CDN HTTPS ${IMAGE_DOMAIN} → http://${IMAGE_UPSTREAM} (Qiniu HTTPS off)"
  fi
  if [ -n "${RESUME_DOMAIN:-}" ]; then
    echo "Caddy: resume HTTPS ${RESUME_DOMAIN} → resume:80"
  fi
  build_config /etc/caddy/Caddyfile.https >/etc/caddy/Caddyfile
else
  echo "Caddy: HTTP only (:80). Set SITE_DOMAIN for HTTPS."
  if [ -n "${RESUME_DOMAIN:-}" ]; then
    echo "Caddy: resume ${RESUME_DOMAIN} → resume:80"
  fi
  build_config /etc/caddy/Caddyfile.http >/etc/caddy/Caddyfile
fi

exec caddy run --config /etc/caddy/Caddyfile --adapter caddyfile
