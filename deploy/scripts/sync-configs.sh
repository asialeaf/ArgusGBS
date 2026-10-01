#!/usr/bin/env bash
# 把 .env 里的对外地址写进 configs/*.ini。
# 宿主机网络，监听端口即对外端口：管理页 10000，信令 HTTP 10002，流媒体 HTTP 10001。
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
echo "  端口 管理页=10000 信令HTTP=10002 流媒体HTTP=10001 SIP=15060 RTP=30000-30249"
