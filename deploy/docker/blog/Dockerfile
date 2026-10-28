# Blog 静态镜像（构建上下文为仓库根目录）
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/node:22-alpine AS build

ENV NODE_OPTIONS="--dns-result-order=ipv4first"

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
# 跳过 vue-tsc，仅 vite 打包
ENV VITE_API_BASE=/api \
    VITE_BASE_URL=/
RUN pnpm --filter @jiangbyte/blog exec vite build

FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/nginx:1.27-alpine

COPY --from=build /app/apps/blog/dist /usr/share/nginx/html
COPY deploy/docker/blog/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
