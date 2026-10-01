#!/usr/bin/env bash
# 把 .env 里的对外地址写进 configs/*.ini。
# 容器内监听端口保持 10000 / 10001 / 15060 / 30000-30249，宿主机映射由 compose 读 .env。
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

HOST="${ADVERTISE_IP:-}"

set_ini() {
  local file="$1" section="$2" key="$3" value="$4"
  awk -v section="$section" -v key="$key" -v value="$value" '
    BEGIN { in_sec = 0; done = 0 }
    /^\[/ {
      if (in_sec && !done) {
        print key "=" value
        done = 1
      }
      in_sec = ($0 == "[" section "]")
    }
    in_sec && $0 ~ ("^" key "=") {
      print key "=" value
      done = 1
      next
    }
    { print }
    END {
      if (!done) {
        if (!in_sec) print "[" section "]"
        print key "=" value
      }
    }
  ' "$file" > "${file}.tmp"
  mv "${file}.tmp" "$file"
}

CMS="${DIR}/configs/arguscms.ini"
SMS="${DIR}/configs/argussms.ini"

set_ini "$CMS" sip host "$HOST"
set_ini "$SMS" rtp advertise_ip "$HOST"

echo "已同步配置:"
echo "  configs/arguscms.ini  [sip] host=${HOST}"
echo "  configs/argussms.ini  [rtp] advertise_ip=${HOST}"
echo "  对外端口 WEB=${WEB_PORT:-10000} SMS=${SMS_HTTP_PORT:-10001} SIP=${SIP_PORT:-15060} RTP=${RTP_PORT_RANGE:-30000-30249}"
