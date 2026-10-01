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
echo ""
echo "=== 部署完成 ==="
make prod-ps
echo ""
echo "访问地址："
echo "  管理页   http://${HOST}:10000/   账号 admin / admin"
echo "  信令 API http://${HOST}:10000/api/v1/getserverinfo  （进程端口 10002）"
echo "  流媒体   http://${HOST}:10001/api/v1/serverinfo"
echo "  设备 SIP ${HOST}:15060  域 3402000000  密码见 configs/arguscms.ini"
echo "  Redis    ${HOST}:26380"
