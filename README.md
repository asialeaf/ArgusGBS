# ArgusGBS

GB28181 流媒体服务平台。信令服务用 Go，流媒体服务用 C++。管理界面直接使用现有的 [GB28181-Server](../GB28181-Server) 前端，HTTP 接口对齐 LiveGBS `/api/v1`（见 LiveCMS 安装包里的 `apidoc`，线上文档 <https://gbs.liveqing.com:10010/apidoc/>）。

```
浏览器 ── HTTP /api/v1 ── ArgusCMS（信令、设备、用户、级联）
                              │ SIP UDP/TCP
                              │ GB/T 28181-2016 与 2022
                         设备 / 下级平台 / 上级平台
                              │
                         ArgusSMS（RTP/PS 收流，HTTP-FLV / HLS 分发）
```

## 启动

```bash
make all
./bin/argussms -config configs/argussms.ini
./bin/arguscms -config configs/arguscms.ini
```

- 管理页：<http://127.0.0.1:10000/>
- 默认账号：`admin` / `admin`
- 信令：`34020000002000000001`，域 `3402000000`，端口 `15060`，设备密码 `gbs12345`
- 流媒体控制面：<http://127.0.0.1:10001/>

前端静态文件按下面顺序自动查找：`www/`、`../GB28181-Server/www`、`../LiveCMS-linux-3.6.9-26093013/www`。用源码编译前端时：

```bash
cd ../GB28181-Server/web_src
npm install
npm run build
```

`vue.config.js` 的开发代理已经指向 `127.0.0.1:10000`。

## 设备侧要填的参数

| 项 | 值 |
| --- | --- |
| SIP 服务器 ID | 34020000002000000001 |
| SIP 域 | 3402000000 |
| SIP 主机 | 本机对设备可达的 IP |
| SIP 端口 | 15060 |
| 密码 | gbs12345 |
| 信令 | UDP 或 TCP |
| 流 | UDP、TCP 被动、TCP 主动 |

注册成功后平台会发 Catalog 查询。2016 与 2022 都走 MANSCDP XML；报文里出现 `DownloadSpeed`、`SecurityLevelCode`、`StreamNumber` 等 2022 字段，或 User-Agent 带 2022 时，设备记为 `GB2022`。云台精确控制、巡航轨迹、存储卡、目标跟踪这些 2022 指令走同一套 DeviceControl / Query。

## 播放地址

`/api/v1/stream/start` 返回的地址指向 ArgusSMS：

- HTTP-FLV：`http://{host}:10001/live/{ssrc}.flv`
- HLS：`http://{host}:10001/live/{ssrc}/index.m3u8`
- RTSP / RTMP / WebRTC / WHEP 字段已按 LiveGBS 返回，便于前端切换协议

当前 SMS 已经完成的是国标 RTP（UDP、TCP 被动、TCP 主动）收包、PS 解复用、H264 FLV 输出。HLS 播放列表、RTSP、RTMP、WebRTC 的 URL 和端口在接口里，分发实现还在继续补。

## 镜像与部署

推送到 `main` 或 `master` 时，GitHub Actions 构建 `arguscms`、`argussms` 并推到阿里云 ACR。仓库需要这些 Secrets，和 ProductCore 相同：

- `ALIYUN_ACR_REGISTRY`
- `ALIYUN_ACR_NAMESPACE`
- `ALIYUN_ACR_USER`
- `ALIYUN_ACR_PASSWORD`

镜像名是 `arguscms`、`argussms`，标签为 `sha-<短提交>`、同一次构建的版本标签和 `latest`。

本地用 ACR 镜像启动：

```bash
cd deploy
cp .env.example .env
# 填写 ADVERTISE_IP（摄像头能访问的本机 IP）、ACR_IMAGE_PREFIX、WWW_DIR
docker compose -f docker-compose.yml -f docker-compose.acr.yml pull
docker compose -f docker-compose.yml -f docker-compose.acr.yml up -d --no-build
```

不拉镜像、在本机编译：

```bash
cd deploy
cp .env.example .env
docker compose up -d --build
```

`ADVERTISE_IP` 会写进 SIP Contact 和收流 SDP。管理页默认 `http://<ADVERTISE_IP>:10000/`，账号 `admin` / `admin`。前端目录用 `WWW_DIR` 挂进信令容器的 `/app/www`。

## 目录

| 路径 | 内容 |
| --- | --- |
| `cmd/arguscms` | 信令服务进程 |
| `internal/sip` | SIP UDP/TCP、Digest 鉴权、INVITE/BYE/INFO |
| `internal/manscdp` | 2016/2022 MANSCDP XML、云台指令 |
| `internal/httpapi` | 与前端对齐的 `/api/v1` |
| `internal/store` | SQLite |
| `sms/` | 流媒体服务 |
| `configs/` | 本机直接运行的配置 |
| `docker/` | 信令、流媒体镜像 |
| `deploy/` | docker compose，本地构建或拉取 ACR 镜像 |
