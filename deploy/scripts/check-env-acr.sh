#!/usr/bin/env bash
# 检查 ACR 部署所需 .env 变量
set -euo pipefail
DIR="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="${DIR}/.env"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "ERROR: 缺少 .env，请执行: cp .env.example .env" >&2
  exit 1
fi

set -a
# shellcheck disable=SC1090
source <(sed 's/\r$//' "$ENV_FILE")
set +a

missing=()
for key in ALIYUN_ACR_REGISTRY ALIYUN_ACR_NAMESPACE ALIYUN_ACR_USER ALIYUN_ACR_PASSWORD ACR_IMAGE_PREFIX ADVERTISE_IP; do
  if [[ -z "${!key:-}" ]]; then
    missing+=("$key")
  fi
done

if ((${#missing[@]} > 0)); then
  echo "ERROR: .env 缺少或未填写: ${missing[*]}" >&2
  echo "参考 .env.example 填写 ACR 账号、ACR_IMAGE_PREFIX 与 ADVERTISE_IP" >&2
  exit 1
fi

if [[ "$ALIYUN_ACR_USER" == "你的ACR用户名" || "$ALIYUN_ACR_PASSWORD" == "你的ACR密码" ]]; then
  echo "ERROR: 请把 ALIYUN_ACR_USER / ALIYUN_ACR_PASSWORD 改成实际 ACR 账号" >&2
  exit 1
fi

if [[ "$ACR_IMAGE_PREFIX" == *"你的命名空间"* || "$ACR_IMAGE_PREFIX" == *"your-namespace"* ]]; then
  echo "ERROR: 请将 ACR_IMAGE_PREFIX 改为实际地址，例如 registry.cn-hangzhou.aliyuncs.com/osms" >&2
  exit 1
fi

expected="${ALIYUN_ACR_REGISTRY}/${ALIYUN_ACR_NAMESPACE}"
if [[ "$ACR_IMAGE_PREFIX" != "$expected" ]]; then
  echo "WARN: ACR_IMAGE_PREFIX=${ACR_IMAGE_PREFIX}，按 REGISTRY/NAMESPACE 应为 ${expected}" >&2
fi

echo "ACR env OK (IMAGE_TAG=${IMAGE_TAG:-latest}, ADVERTISE_IP=${ADVERTISE_IP}, prefix=${ACR_IMAGE_PREFIX})"
