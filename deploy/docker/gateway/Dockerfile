# 统一入口：对齐 example.github.io
# - 未设 SITE_DOMAIN：仅 :80
# - 设 SITE_DOMAIN：域名自动 HTTPS（Let's Encrypt），裸 IP 仍走 :80
FROM caddy:2.9-alpine

COPY Caddyfile.snippet /etc/caddy/Caddyfile.snippet
COPY Caddyfile.http /etc/caddy/Caddyfile.http
COPY Caddyfile.https /etc/caddy/Caddyfile.https
COPY Caddyfile.images /etc/caddy/Caddyfile.images
COPY Caddyfile.resume /etc/caddy/Caddyfile.resume
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

EXPOSE 80 443
VOLUME ["/data", "/config"]

ENTRYPOINT ["/docker-entrypoint.sh"]
