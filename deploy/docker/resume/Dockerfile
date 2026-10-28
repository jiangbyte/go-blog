# Resume 静态镜像（构建上下文为仓库根目录）
FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/node:22-alpine AS build

ENV NODE_OPTIONS="--dns-result-order=ipv4first" \
    COREPACK_NPM_REGISTRY=https://registry.npmmirror.com

WORKDIR /app
RUN corepack enable && corepack prepare pnpm@9.15.9 --activate

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
ENV VITE_API_BASE=/api \
    VITE_BASE_URL=/
RUN pnpm --filter @jiangbyte/resume exec vite build

FROM swr.cn-north-4.myhuaweicloud.com/ddn-k8s/docker.io/nginx:1.27-alpine

COPY --from=build /app/apps/resume/dist /usr/share/nginx/html
COPY deploy/docker/resume/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
