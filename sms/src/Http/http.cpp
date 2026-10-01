#include "Common/sms.h"
#include <arpa/inet.h>
#include <netinet/in.h>
#include <sys/socket.h>
#include <unistd.h>
#include <openssl/sha.h>
#include <sstream>
#include <cstring>
#include <cstdlib>
#include <iostream>
#include <vector>

std::string host_ip() {
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

static std::string b64_bytes(const uint8_t* p, size_t n) {
    static const char* t = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
    std::string o;
    for (size_t i = 0; i < n; i += 3) {
        uint32_t v = (uint32_t)p[i] << 16;
        if (i + 1 < n) v |= (uint32_t)p[i + 1] << 8;
        if (i + 2 < n) v |= p[i + 2];
        o.push_back(t[(v >> 18) & 63]);
        o.push_back(t[(v >> 12) & 63]);
        o.push_back(i + 1 < n ? t[(v >> 6) & 63] : '=');
        o.push_back(i + 2 < n ? t[v & 63] : '=');
    }
    return o;
}

void ws_send(int fd, const uint8_t* data, size_t n) {
    uint8_t hdr[10];
    size_t h = 0;
    hdr[h++] = 0x82;
    if (n < 126) {
        hdr[h++] = (uint8_t)n;
    } else if (n <= 0xFFFF) {
        hdr[h++] = 126;
        hdr[h++] = (uint8_t)(n >> 8);
        hdr[h++] = (uint8_t)n;
    } else {
        hdr[h++] = 127;
        for (int i = 7; i >= 0; --i) hdr[h++] = (uint8_t)((uint64_t)n >> (8 * i));
    }
    send(fd, hdr, h, MSG_NOSIGNAL);
    if (n) send(fd, data, n, MSG_NOSIGNAL);
}

static std::string ws_accept_key(const std::string& key) {
    std::string src = key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11";
    unsigned char dig[SHA_DIGEST_LENGTH];
    SHA1(reinterpret_cast<const unsigned char*>(src.data()), src.size(), dig);
    return b64_bytes(dig, SHA_DIGEST_LENGTH);
}

static bool ws_read_frame(int fd) {
    uint8_t hdr[2];
    if (recv(fd, hdr, 2, MSG_WAITALL) != 2) return false;
    uint8_t opcode = hdr[0] & 0x0f;
    uint64_t len = hdr[1] & 0x7f;
    bool mask = hdr[1] & 0x80;
    if (len == 126) {
        uint8_t ext[2];
        if (recv(fd, ext, 2, MSG_WAITALL) != 2) return false;
        len = ((uint64_t)ext[0] << 8) | ext[1];
    } else if (len == 127) {
        uint8_t ext[8];
        if (recv(fd, ext, 8, MSG_WAITALL) != 8) return false;
        len = 0;
        for (int i = 0; i < 8; ++i) len = (len << 8) | ext[i];
    }
    uint8_t mkey[4] = {};
    if (mask && recv(fd, mkey, 4, MSG_WAITALL) != 4) return false;
    std::vector<uint8_t> payload(len);
    size_t got = 0;
    while (got < payload.size()) {
        ssize_t n = recv(fd, payload.data() + got, payload.size() - got, 0);
        if (n <= 0) return false;
        got += (size_t)n;
    }
    if (opcode == 0x8) return false;
    if (opcode == 0x9) {
        uint8_t ph[2] = {0x8A, (uint8_t)(len < 126 ? len : 0)};
        send(fd, ph, 2, MSG_NOSIGNAL);
        if (len && len < 126) send(fd, payload.data(), payload.size(), MSG_NOSIGNAL);
    }
    return true;
}

static void serve_flv(int fd, std::shared_ptr<Session> s, bool websocket) {
    if (!websocket) {
        std::string hdr =
            "HTTP/1.1 200 OK\r\n"
            "Content-Type: video/x-flv\r\n"
            "Access-Control-Allow-Origin: *\r\n"
            "Cache-Control: no-cache\r\n"
            "Connection: keep-alive\r\n\r\n";
        send_all(fd, hdr);
    }
    auto init = s->flv_header_and_gop();
    int64_t base = init.size() > 13 ? flv_first_media_ts(init.data() + 13, init.size() - 13) : 0;
    if (init.size() > 13) flv_shift_tags(init.data() + 13, init.size() - 13, base);
    if (websocket) ws_send(fd, init.data(), init.size());
    else send_all(fd, init.data(), init.size());
    auto sub = std::make_shared<Subscriber>();
    sub->fd = fd;
    sub->websocket = websocket;
    sub->sent_header = true;
    sub->stamp_base = base;
    {
        std::lock_guard<std::mutex> lk(s->mu);
        s->subs.push_back(sub);
    }
    // 连接由 broadcast 写；这里阻塞读到对端关闭
    if (websocket) {
        while (s->running && ws_read_frame(fd)) {}
    } else {
        char tmp[64];
        while (s->running) {
            ssize_t n = recv(fd, tmp, sizeof(tmp), 0);
            if (n <= 0) break;
        }
    }
    std::lock_guard<std::mutex> lk(s->mu);
    for (auto it = s->subs.begin(); it != s->subs.end(); ++it) {
        if (*it == sub) { s->subs.erase(it); break; }
    }
}

static void http_raw(int fd, int code, const char* reason, const std::string& ctype, const std::string& extra, const uint8_t* body, size_t n) {
    std::ostringstream o;
    o << "HTTP/1.1 " << code << " " << reason << "\r\n"
      << "Content-Type: " << ctype << "\r\n"
      << "Access-Control-Allow-Origin: *\r\n"
      << "Access-Control-Expose-Headers: Location\r\n"
      << extra
      << "Content-Length: " << n << "\r\n"
      << "Connection: close\r\n\r\n";
    send_all(fd, o.str());
    if (n && body) send_all(fd, body, n);
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
        std::string ip = hub->public_ip();
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
        std::string id = json_get(body, "stream_id");
        std::string tr = json_get(body, "transport");
        std::string mode = json_get(body, "mode");
        std::string peer = json_get(body, "peer_ip");
        std::string dir = json_get(body, "direction");
        int peer_port = 0;
        try { peer_port = std::stoi(json_get(body, "peer_port")); } catch (...) {}
        if (tr.empty()) tr = "UDP";
        int listen_port = 0;
        if (!hub->relay(id, tr, mode, peer, peer_port, dir, listen_port)) {
            http_json(fd, 404, "{\"error\":\"no stream\"}");
            close(fd);
            return;
        }
        http_json(fd, 200, "{\"port\":" + std::to_string(listen_port) + "}");
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
        std::string upgrade = header_value(req, "Upgrade");
        if (upgrade.empty()) upgrade = header_value(req, "upgrade");
        bool websocket = upgrade.find("websocket") != std::string::npos || upgrade.find("WebSocket") != std::string::npos;
        if (websocket) {
            std::string key = header_value(req, "Sec-WebSocket-Key");
            if (key.empty()) key = header_value(req, "sec-websocket-key");
            std::string accept = ws_accept_key(key);
            std::string hs =
                "HTTP/1.1 101 Switching Protocols\r\n"
                "Upgrade: websocket\r\n"
                "Connection: Upgrade\r\n"
                "Sec-WebSocket-Accept: " + accept + "\r\n"
                "Access-Control-Allow-Origin: *\r\n\r\n";
            send_all(fd, hs);
        }
        serve_flv(fd, s, websocket);
        close(fd);
        return;
    }
    if (path.rfind("/live/", 0) == 0 && path.find('/') != std::string::npos) {
        std::string rest = path.substr(6);
        auto slash = rest.find('/');
        if (slash != std::string::npos) {
            std::string id = rest.substr(0, slash);
            std::string file = rest.substr(slash + 1);
            auto s = hub->get(id);
            if (!s) { http_json(fd, 404, "{\"error\":\"no stream\"}"); close(fd); return; }
            if (file == "index.m3u8") {
                std::string pl = s->hls_m3u8();
                http_raw(fd, 200, "OK", "application/vnd.apple.mpegurl", "Cache-Control: no-cache\r\n",
                         (const uint8_t*)pl.data(), pl.size());
                close(fd);
                return;
            }
            if (file.rfind("seg", 0) == 0 && file.find(".ts") != std::string::npos) {
                int seq = std::atoi(file.c_str() + 3);
                std::vector<uint8_t> ts;
                if (!s->hls_seg(seq, ts)) { http_json(fd, 404, "{\"error\":\"no segment\"}"); close(fd); return; }
                http_raw(fd, 200, "OK", "video/mp2t", "Cache-Control: no-cache\r\n", ts.data(), ts.size());
                close(fd);
                return;
            }
        }
    }
    if ((path.rfind("/whep/", 0) == 0 || path.rfind("/webrtc/", 0) == 0) && method == "POST") {
        std::string id;
        if (path.rfind("/whep/", 0) == 0) id = path.substr(6);
        else {
            auto p = query.find("stream=");
            if (p != std::string::npos) {
                id = query.substr(p + 7);
                auto amp = id.find('&');
                if (amp != std::string::npos) id = id.substr(0, amp);
            }
        }
        while (!id.empty() && id.back() == '/') id.pop_back();
        bool as_json = !body.empty() && body[0] == '{';
        std::string offer = body;
        if (as_json) {
            offer = json_get(body, "sdp");
            std::string plain;
            for (size_t i = 0; i < offer.size(); ++i) {
                if (offer[i] == '\\' && i + 1 < offer.size() && offer[i + 1] == 'n') {
                    plain.push_back('\n');
                    ++i;
                } else plain.push_back(offer[i]);
            }
            offer.swap(plain);
        }
        int code = 500;
        std::string ctype = "application/json", resp = "{\"error\":\"webrtc\"}";
        hub->whep(id, offer, as_json, code, ctype, resp);
        std::string extra;
        if (code == 201) extra = "Location: /whep/" + id + "\r\n";
        const char* reason = code == 201 ? "Created" : code == 404 ? "Not Found" : code == 400 ? "Bad Request" : "OK";
        http_raw(fd, code, reason, ctype, extra, (const uint8_t*)resp.data(), resp.size());
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
    start_extra();
    std::cerr << "ArgusSMS HTTP :" << http_port << " RTSP :" << rtsp_port << " RTMP :" << rtmp_port
              << " WebRTC UDP " << rtc_lo << "-" << rtc_hi << std::endl;
    while (true) {
        int c = accept(fd, nullptr, nullptr);
        if (c < 0) continue;
        std::thread(handle_client, this, c).detach();
    }
}
