#!/usr/bin/env bash
# 生产部署：同步配置，从阿里云 ACR 拉取镜像并启动
set -euo pipefail
DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$DIR"

if [[ ! -f .env ]]; then
  cp .env.example .env
  echo "已创建 .env，请先填写 ALIYUN_ACR_*、ACR_IMAGE_PREFIX、ADVERTISE_IP 后重新运行"
  exit 1
fi

set -a
# shellcheck disable=SC1090
source <(sed 's/\r$//' .env)
set +a

chmod +x scripts/check-env-acr.sh scripts/sync-configs.sh
./scripts/check-env-acr.sh
./scripts/sync-configs.sh

make login-acr
make pull-images
docker compose -f docker-compose.yml -f docker-compose.acr.yml up -d --no-build --remove-orphans

HOST="${ADVERTISE_IP:-localhost}"
WEB_PORT="${WEB_PORT:-10000}"
SMS_PORT="${SMS_HTTP_PORT:-10001}"
SIP_PORT="${SIP_PORT:-15060}"
echo ""
echo "=== 部署完成 ==="
make prod-ps
echo ""
echo "访问地址："
echo "  管理页   http://${HOST}:${WEB_PORT}/   账号 admin / admin"
echo "  信令 API http://${HOST}:${WEB_PORT}/api/v1/getserverinfo"
echo "  流媒体   http://${HOST}:${SMS_PORT}/api/v1/serverinfo"
echo "  设备 SIP ${HOST}:${SIP_PORT}  域 3402000000  密码见 configs/arguscms.ini"
