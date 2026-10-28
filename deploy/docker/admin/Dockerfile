# Admin 静态镜像；base 与 API 前缀由 ARG 注入
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/node:22-alpine AS build

ARG VITE_BASE_URL=/admin/
ARG VITE_API_URL=/api

ENV NODE_OPTIONS="--dns-result-order=ipv4first" \
    VITE_BASE_URL=${VITE_BASE_URL} \
    VITE_API_URL=${VITE_API_URL}

WORKDIR /app
RUN corepack enable

COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
COPY web/apps/blog/package.json ./apps/blog/
COPY web/apps/admin/package.json ./apps/admin/
COPY web/apps/resume/package.json ./apps/resume/
COPY web/packages/shared/package.json ./packages/shared/
COPY web/packages/markdown/package.json ./packages/markdown/

RUN pnpm config set registry https://registry.npmmirror.com \
    && pnpm config set fetch-retries 5 \
    && pnpm config set fetch-retry-mintimeout 20000 \
    && pnpm config set fetch-retry-maxtimeout 120000 \
    && pnpm install --frozen-lockfile

COPY web/ .
RUN pnpm --filter @jiangbyte/admin run prebuild \
    && pnpm --filter @jiangbyte/admin exec vite build

FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/nginx:1.27-alpine

COPY --from=build /app/apps/admin/dist /usr/share/nginx/html
COPY web/apps/admin/nginx/spa.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
