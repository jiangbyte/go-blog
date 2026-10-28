#!/usr/bin/env bash
# 构建并推送 api / blog / admin / resume / gateway 到阿里云 ACR。
# 每个镜像同时打 :VERSION 与 :latest；线上 compose 只拉 :latest。
# ACR 不接受 BuildKit provenance/SBOM，构建时关闭。
#
# Usage:
#   ./deploy/scripts/docker-build-push.sh           # 默认 1.0.1 + latest
#   ./deploy/scripts/docker-build-push.sh 1.0.2
#   VERSION=1.0.2 ./deploy/scripts/docker-build-push.sh
#   make docker-push IMAGE_VERSION=1.0.2
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

REGISTRY="${REGISTRY:-your-registry.example.com}"
NAMESPACE="${NAMESPACE:-your-namespace}"
VERSION="${1:-${VERSION:-1.0.1}}"
PREFIX="${REGISTRY}/${NAMESPACE}"

# 镜像名 → Dockerfile 相对仓库根路径 → build context → 额外 build-arg
# gateway 的 context 为 deploy/docker/gateway
build_one() {
  local name="$1"
  local dockerfile="$2"
  local context="$3"
  shift 3
  local full="${PREFIX}/${name}"
  echo "==> building ${full}:${VERSION} and ${full}:latest"
  docker build \
    --provenance=false \
    --sbom=false \
    -f "${dockerfile}" \
    -t "${full}:${VERSION}" \
    -t "${full}:latest" \
    "$@" \
    "${context}"
  echo "==> pushing ${full}:${VERSION}"
  docker push "${full}:${VERSION}"
  echo "==> pushing ${full}:latest"
  docker push "${full}:latest"
}

echo "registry=${PREFIX}  version=${VERSION} (+ latest)"

build_one go-blog-api deploy/docker/api/Dockerfile .
build_one go-blog-blog deploy/docker/blog/Dockerfile . --network=host
build_one go-blog-admin deploy/docker/admin/Dockerfile . --network=host \
  --build-arg VITE_BASE_URL=/admin/ \
  --build-arg VITE_API_URL=/api
build_one go-blog-resume deploy/docker/resume/Dockerfile . --network=host
build_one go-blog-gateway deploy/docker/gateway/Dockerfile deploy/docker/gateway

echo "done (version + latest)"
for n in api blog admin resume gateway; do
  echo "  ${PREFIX}/go-blog-${n}:${VERSION}"
  echo "  ${PREFIX}/go-blog-${n}:latest"
done
