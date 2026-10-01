#include "sms.h"
#include <arpa/inet.h>
#include <netinet/in.h>
#include <sys/socket.h>
#include <unistd.h>
#include <sstream>
#include <cstring>
#include <iostream>

static std::string local_ip() {
    int fd = socket(AF_INET, SOCK_DGRAM, 0);
    sockaddr_in addr{};
    addr.sin_family = AF_INET;
    addr.sin_port = htons(80);
    inet_pton(AF_INET, "8.8.8.8", &addr.sin_addr);
    std::string ip = "127.0.0.1";
    if (connect(fd, (sockaddr*)&addr, sizeof(addr)) == 0) {
        sockaddr_in local{};
        socklen_t len = sizeof(local);
        if (getsockname(fd, (sockaddr*)&local, &len) == 0) {
            char buf[64];
            inet_ntop(AF_INET, &local.sin_addr, buf, sizeof(buf));
            ip = buf;
        }
    }
    close(fd);
    return ip;
}

static std::string read_headers(int fd, std::string& extra) {
    std::string data;
    char buf[2048];
    while (data.find("\r\n\r\n") == std::string::npos) {
        ssize_t n = recv(fd, buf, sizeof(buf), 0);
        if (n <= 0) break;
        data.append(buf, buf + n);
        if (data.size() > 1024 * 1024) break;
    }
    auto pos = data.find("\r\n\r\n");
    if (pos != std::string::npos) extra = data.substr(pos + 4);
    return data;
}

static void send_all(int fd, const std::string& s) {
    size_t off = 0;
    while (off < s.size()) {
        ssize_t n = send(fd, s.data() + off, s.size() - off, MSG_NOSIGNAL);
        if (n <= 0) return;
        off += n;
    }
}

static void send_all(int fd, const uint8_t* p, size_t n) {
    size_t off = 0;
    while (off < n) {
        ssize_t w = send(fd, p + off, n - off, MSG_NOSIGNAL);
        if (w <= 0) return;
        off += w;
    }
}

static std::string header_value(const std::string& h, const std::string& k) {
    auto pos = h.find(k + ":");
    if (pos == std::string::npos) return "";
    pos += k.size() + 1;
    while (pos < h.size() && h[pos] == ' ') pos++;
    auto end = h.find("\r\n", pos);
    if (end == std::string::npos) return h.substr(pos);
    return h.substr(pos, end - pos);
}

static std::string json_get(const std::string& body, const std::string& key) {
    auto p = body.find("\"" + key + "\"");
    if (p == std::string::npos) return "";
    p = body.find(':', p);
    if (p == std::string::npos) return "";
    p++;
    while (p < body.size() && (body[p] == ' ' || body[p] == '"')) p++;
    size_t e = p;
    while (e < body.size() && body[e] != '"' && body[e] != ',' && body[e] != '}') e++;
    return body.substr(p, e - p);
}

static void http_json(int fd, int code, const std::string& body) {
    std::ostringstream o;
    o << "HTTP/1.1 " << code << " OK\r\n"
      << "Content-Type: application/json\r\n"
      << "Access-Control-Allow-Origin: *\r\n"
      << "Content-Length: " << body.size() << "\r\n"
      << "Connection: close\r\n\r\n" << body;
    send_all(fd, o.str());
}

static void serve_flv(int fd, std::shared_ptr<Session> s) {
    std::string hdr =
        "HTTP/1.1 200 OK\r\n"
        "Content-Type: video/x-flv\r\n"
        "Access-Control-Allow-Origin: *\r\n"
        "Cache-Control: no-cache\r\n"
        "Connection: keep-alive\r\n\r\n";
    send_all(fd, hdr);
    auto init = s->flv_header_and_gop();
    send_all(fd, init.data(), init.size());
    auto sub = std::make_shared<Subscriber>();
    sub->fd = fd;
    sub->sent_header = true;
    {
        std::lock_guard<std::mutex> lk(s->mu);
        s->subs.push_back(sub);
    }
    // 连接由 broadcast 写；这里阻塞读到对端关闭
    char tmp[64];
    while (s->running) {
        ssize_t n = recv(fd, tmp, sizeof(tmp), 0);
        if (n <= 0) break;
    }
    std::lock_guard<std::mutex> lk(s->mu);
    for (auto it = s->subs.begin(); it != s->subs.end(); ++it) {
        if (*it == sub) { s->subs.erase(it); break; }
    }
}

static std::string hls_playlist(const std::string& id) {
    return "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:2\n#EXT-X-MEDIA-SEQUENCE:0\n"
           "#EXTINF:2.0,\nseg.ts\n";
}

void handle_client(Hub* hub, int fd) {
    std::string extra;
    std::string req = read_headers(fd, extra);
    if (req.empty()) { close(fd); return; }
    std::istringstream is(req);
    std::string method, path, ver;
    is >> method >> path >> ver;
    auto qpos = path.find('?');
    std::string query;
    if (qpos != std::string::npos) {
        query = path.substr(qpos + 1);
        path = path.substr(0, qpos);
    }
    int content_len = 0;
    try { content_len = std::stoi(header_value(req, "Content-Length")); } catch (...) {}
    std::string body = extra;
    while ((int)body.size() < content_len) {
        char buf[4096];
        ssize_t n = recv(fd, buf, sizeof(buf), 0);
        if (n <= 0) break;
        body.append(buf, buf + n);
    }
    if (method == "OPTIONS") {
        send_all(fd, "HTTP/1.1 204 No Content\r\nAccess-Control-Allow-Origin: *\r\nAccess-Control-Allow-Headers: *\r\nAccess-Control-Allow-Methods: GET,POST,OPTIONS\r\nContent-Length: 0\r\n\r\n");
        close(fd);
        return;
    }
    if (path == "/api/v1/serverinfo") {
        http_json(fd, 200, "{\"Server\":\"ArgusSMS\",\"Version\":\"0.1.0\"}");
        close(fd);
        return;
    }
    if (path == "/api/v1/rtp/open" && method == "POST") {
        std::string id = json_get(body, "stream_id");
        std::string tr = json_get(body, "transport");
        std::string mode = json_get(body, "mode");
        std::string ssrc = json_get(body, "ssrc");
        std::string peer = json_get(body, "peer_ip");
        int peer_port = 0;
        try { peer_port = std::stoi(json_get(body, "peer_port")); } catch (...) {}
        if (tr.empty()) tr = "UDP";
        if (mode.empty()) mode = "passive";
        if (id.empty()) { http_json(fd, 400, "{\"error\":\"stream_id\"}"); close(fd); return; }
        auto s = hub->open(id, tr, mode, ssrc, peer, peer_port);
        std::string ip = hub->advertise_ip.empty() ? local_ip() : hub->advertise_ip;
        http_json(fd, 200, "{\"ip\":\"" + ip + "\",\"public_ip\":\"" + ip + "\",\"port\":" + std::to_string(s->port) + "}");
        close(fd);
        return;
    }
    if (path == "/api/v1/rtp/close" && method == "POST") {
        hub->close(json_get(body, "stream_id"));
        http_json(fd, 200, "{}");
        close(fd);
        return;
    }
    if (path == "/api/v1/rtp/relay" && method == "POST") {
        // 级联转发在后续版本按 RTP 原包复制；当前确认会话存在即可。
        auto s = hub->get(json_get(body, "stream_id"));
        if (!s) { http_json(fd, 404, "{\"error\":\"no stream\"}"); close(fd); return; }
        http_json(fd, 200, "{}");
        close(fd);
        return;
    }
    if (path == "/api/v1/rtp/stats") {
        auto idpos = query.find("stream_id=");
        std::string id = idpos == std::string::npos ? "" : query.substr(idpos + 10);
        auto amp = id.find('&');
        if (amp != std::string::npos) id = id.substr(0, amp);
        auto s = hub->get(id);
        if (!s) { http_json(fd, 404, "{\"error\":\"no stream\"}"); close(fd); return; }
        auto& st = s->stats;
        std::ostringstream o;
        o << "{\"stream_id\":\"" << id << "\",\"codec\":\"" << st.codec << "\",\"width\":" << st.width
          << ",\"height\":" << st.height << ",\"fps\":" << st.fps << ",\"rtp_count\":" << st.rtp_count
          << ",\"rtp_lost\":" << st.rtp_lost << ",\"in_bytes\":" << st.in_bytes
          << ",\"in_bitrate\":" << st.in_bitrate << ",\"video_frames\":" << st.video_frames
          << ",\"num_outputs\":" << st.num_outputs << ",\"audio_codec\":\"" << st.audio_codec
          << "\",\"ready\":" << (st.ready ? "true" : "false") << "}";
        http_json(fd, 200, o.str());
        close(fd);
        return;
    }
    // /live/{id}.flv
    if (path.rfind("/live/", 0) == 0 && path.size() > 8 && path.find(".flv") != std::string::npos) {
        std::string id = path.substr(6);
        auto dot = id.find('.');
        if (dot != std::string::npos) id = id.substr(0, dot);
        auto s = hub->get(id);
        if (!s) { http_json(fd, 404, "{\"error\":\"no stream\"}"); close(fd); return; }
        serve_flv(fd, s);
        close(fd);
        return;
    }
    if (path.rfind("/live/", 0) == 0 && path.find("index.m3u8") != std::string::npos) {
        std::string pl = hls_playlist(path);
        std::ostringstream o;
        o << "HTTP/1.1 200 OK\r\nContent-Type: application/vnd.apple.mpegurl\r\nAccess-Control-Allow-Origin: *\r\nContent-Length: "
          << pl.size() << "\r\n\r\n" << pl;
        send_all(fd, o.str());
        close(fd);
        return;
    }
    if (path.rfind("/whep/", 0) == 0 || path.rfind("/webrtc/", 0) == 0) {
        http_json(fd, 501, "{\"error\":\"WebRTC 信令已预留，当前请使用 HTTP-FLV / WS-FLV / HLS / RTSP\"}");
        close(fd);
        return;
    }
    http_json(fd, 404, "{\"error\":\"not found\"}");
    close(fd);
}

void Hub::start() {
    int fd = socket(AF_INET, SOCK_STREAM, 0);
    int opt = 1;
    setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt));
    sockaddr_in addr{};
    addr.sin_family = AF_INET;
    addr.sin_port = htons(http_port);
    addr.sin_addr.s_addr = INADDR_ANY;
    if (bind(fd, (sockaddr*)&addr, sizeof(addr)) != 0) {
        std::perror("bind http");
        return;
    }
    listen(fd, 128);
    std::cerr << "ArgusSMS HTTP :" << http_port << " RTSP :" << rtsp_port << " RTMP :" << rtmp_port << std::endl;
    while (true) {
        int c = accept(fd, nullptr, nullptr);
        if (c < 0) continue;
        std::thread(handle_client, this, c).detach();
    }
}
