#include "Common/sms.h"
#include <arpa/inet.h>
#include <netinet/in.h>
#include <sys/select.h>
#include <sys/socket.h>
#include <unistd.h>
#include <cstring>
#include <chrono>

static int64_t now_ms() {
    return std::chrono::duration_cast<std::chrono::milliseconds>(
        std::chrono::steady_clock::now().time_since_epoch()).count();
}

Session::Session(std::string i) : id(std::move(i)) {}

Session::~Session() {
    running = false;
    if (udp_fd >= 0) {
        shutdown(udp_fd, SHUT_RDWR);
        close(udp_fd);
        udp_fd = -1;
    }
    if (tcp_fd >= 0) {
        shutdown(tcp_fd, SHUT_RDWR);
        close(tcp_fd);
        tcp_fd = -1;
    }
    if (relay_listen >= 0) {
        shutdown(relay_listen, SHUT_RDWR);
        close(relay_listen);
        relay_listen = -1;
    }
    {
        std::lock_guard<std::mutex> lk(dist_mu);
        for (auto& sk : sinks) if (sk.fd >= 0) close(sk.fd);
        sinks.clear();
        for (auto& r : rtsp) {
            if (r.tcp >= 0) close(r.tcp);
            if (r.udp >= 0) close(r.udp);
        }
        rtsp.clear();
        for (int fd : rtmp) if (fd >= 0) close(fd);
        rtmp.clear();
        for (auto& p : rtc) rtc_stop(p);
        rtc.clear();
    }
    if (worker.joinable()) worker.detach();
}

void Session::broadcast(const std::vector<uint8_t>& tag) {
    if (tag.empty()) return;
    {
        std::lock_guard<std::mutex> lk(mu);
        flv_tags.push_back(tag);
        while (flv_tags.size() > 400) flv_tags.pop_front();
        for (auto& s : subs) {
            if (!s->sent_header || s->fd < 0) continue;
            send(s->fd, tag.data(), tag.size(), MSG_NOSIGNAL);
        }
        int extra = 0;
        {
            std::lock_guard<std::mutex> dk(dist_mu);
            extra = (int)rtsp.size() + (int)rtmp.size() + (int)rtc.size() + (int)sinks.size();
        }
        stats.num_outputs = (int)subs.size() + extra;
    }
    fanout_flv(tag);
}

std::vector<uint8_t> Session::flv_header_and_gop() {
    std::vector<uint8_t> out = flv_file_header();
    if (!vps.empty() && !sps.empty() && !pps.empty()) {
        auto seq = hevc_seq_header(vps, sps, pps);
        out.insert(out.end(), seq.begin(), seq.end());
    } else if (!sps.empty() && !pps.empty()) {
        auto seq = avc_seq_header(sps, pps);
        out.insert(out.end(), seq.begin(), seq.end());
    }
    std::lock_guard<std::mutex> lk(mu);
    for (auto& t : flv_tags) out.insert(out.end(), t.begin(), t.end());
    return out;
}

void Session::add_nals(const std::vector<NAL>& nals) {
    uint32_t dts = (uint32_t)(stats.video_frames * 40);
    bool dummy = false;
    bool sent_hevc = !vps.empty() && !sps.empty() && !pps.empty();
    for (auto& n : nals) {
        auto raw = n.data;
        size_t off = 0;
        if (raw.size() >= 4 && raw[0] == 0 && raw[1] == 0 && raw[2] == 0 && raw[3] == 1) off = 4;
        else if (raw.size() >= 3 && raw[0] == 0 && raw[1] == 0 && raw[2] == 1) off = 3;
        if (off >= raw.size()) continue;
        if (n.h265) {
            int t = (raw[off] >> 1) & 0x3F;
            if (t == 32) vps.assign(raw.begin() + off, raw.end());
            if (t == 33) sps.assign(raw.begin() + off, raw.end());
            if (t == 34) pps.assign(raw.begin() + off, raw.end());
        } else {
            int t = raw[off] & 0x1F;
            if (t == 7) sps.assign(raw.begin() + off, raw.end());
            if (t == 8) pps.assign(raw.begin() + off, raw.end());
        }
        if (!sent_hevc && !vps.empty() && !sps.empty() && !pps.empty()) {
            broadcast(hevc_seq_header(vps, sps, pps));
            sent_hevc = true;
        }
        auto tag = annexb_to_flv_tag(n, dts, sps, pps, dummy);
        if (!tag.empty()) {
            stats.video_frames++;
            stats.ready = true;
            stats.codec = demux.codec();
            if (!demux.audio_codec().empty()) stats.audio_codec = demux.audio_codec();
            broadcast(tag);
        }
    }
    playout(nals);
}

static void pull_media(Session* s) {
    s->add_nals(s->demux.take());
    for (auto& a : s->demux.take_audio()) {
        s->stats.audio_codec = a.alaw ? "G711A" : "G711U";
        s->broadcast(g711_flv_tag(a.data.data(), a.data.size(), a.ts90 / 90, a.alaw));
    }
}

static void emit_annexb(Session* s, const uint8_t* nal, size_t n, uint32_t ts90, bool h265) {
    if (!nal || n == 0) return;
    NAL one;
    one.h265 = h265;
    one.ts90 = ts90;
    one.data = {0, 0, 0, 1};
    one.data.insert(one.data.end(), nal, nal + n);
    if (h265) {
        int t = (nal[0] >> 1) & 0x3F;
        one.key = t == 19 || t == 20 || t == 21 || t == 32 || t == 33 || t == 34;
        s->demux.hint_video("H265");
    } else {
        int t = nal[0] & 0x1F;
        one.key = t == 5 || t == 7 || t == 8;
        s->demux.hint_video("H264");
    }
    s->add_nals({one});
}

static void h264_rtp(Session* s, const uint8_t* p, size_t n, uint32_t ts) {
    if (n < 1) return;
    int typ = p[0] & 0x1F;
    if (typ >= 1 && typ <= 23) {
        s->fu_open = false;
        emit_annexb(s, p, n, ts, false);
        return;
    }
    if (typ == 24) {
        size_t i = 1;
        while (i + 2 <= n) {
            uint16_t len = (p[i] << 8) | p[i + 1];
            i += 2;
            if (!len || i + len > n) break;
            emit_annexb(s, p + i, len, ts, false);
            i += len;
        }
        return;
    }
    if (typ == 28 && n >= 2) {
        bool start = p[1] & 0x80;
        bool end = p[1] & 0x40;
        uint8_t nh = (p[0] & 0xE0) | (p[1] & 0x1F);
        if (start) {
            s->fu = {0, 0, 0, 1, nh};
            s->fu.insert(s->fu.end(), p + 2, p + n);
            s->fu_open = true;
        } else if (s->fu_open) {
            s->fu.insert(s->fu.end(), p + 2, p + n);
        }
        if (end && s->fu_open && s->fu.size() > 5) {
            NAL one;
            one.ts90 = ts;
            one.data = std::move(s->fu);
            one.key = (one.data[4] & 0x1F) == 5;
            s->fu.clear();
            s->fu_open = false;
            s->demux.hint_video("H264");
            s->add_nals({one});
        }
    }
}

static void h265_rtp(Session* s, const uint8_t* p, size_t n, uint32_t ts) {
    if (n < 2) return;
    int typ = (p[0] >> 1) & 0x3F;
    if (typ == 49 && n >= 3) {
        bool start = p[2] & 0x80;
        bool end = p[2] & 0x40;
        uint8_t nal_type = p[2] & 0x3F;
        uint8_t h0 = (p[0] & 0x81) | (nal_type << 1);
        if (start) {
            s->fu = {0, 0, 0, 1, h0, p[1]};
            s->fu.insert(s->fu.end(), p + 3, p + n);
            s->fu_open = true;
        } else if (s->fu_open) {
            s->fu.insert(s->fu.end(), p + 3, p + n);
        }
        if (end && s->fu_open && s->fu.size() > 6) {
            NAL one;
            one.h265 = true;
            one.ts90 = ts;
            one.data = std::move(s->fu);
            int t = (one.data[4] >> 1) & 0x3F;
            one.key = t == 19 || t == 20 || t == 21;
            s->fu.clear();
            s->fu_open = false;
            s->demux.hint_video("H265");
            s->add_nals({one});
        }
        return;
    }
    if (typ == 48 && n > 2) {
        size_t i = 2;
        while (i + 2 <= n) {
            uint16_t len = (p[i] << 8) | p[i + 1];
            i += 2;
            if (!len || i + len > n) break;
            emit_annexb(s, p + i, len, ts, true);
            i += len;
        }
        return;
    }
    s->fu_open = false;
    emit_annexb(s, p, n, ts, true);
}

static std::string ts_codec(uint8_t st) {
    if (st == 0x1B) return "H264";
    if (st == 0x24) return "H265";
    if (st == 0x90) return "G711A";
    if (st == 0x91) return "G711U";
    if (st == 0x0F || st == 0x11) return "AAC";
    return "";
}

static void ts_section(Session* s, uint16_t pid, const uint8_t* sec, size_t n) {
    if (n < 8) return;
    uint8_t table = sec[0];
    if (pid == 0 && table == 0x00) {
        for (size_t i = 8; i + 4 <= n && i < 8 + 32; i += 4) {
            uint16_t prog = (sec[i] << 8) | sec[i + 1];
            uint16_t pmt = ((sec[i + 2] & 0x1F) << 8) | sec[i + 3];
            if (prog != 0) s->ts_pmt = pmt;
        }
        return;
    }
    if (pid != s->ts_pmt || table != 0x02 || n < 12) return;
    uint16_t info = ((sec[10] & 0x0F) << 8) | sec[11];
    size_t i = 12 + info;
    while (i + 5 <= n) {
        uint8_t st = sec[i];
        uint16_t epid = ((sec[i + 1] & 0x1F) << 8) | sec[i + 2];
        uint16_t esl = ((sec[i + 3] & 0x0F) << 8) | sec[i + 4];
        auto c = ts_codec(st);
        if (!c.empty()) s->ts_type[epid] = c;
        if (i + 5 + esl > n) break;
        i += 5 + esl;
    }
}

static void feed_ts(Session* s, const uint8_t* data, size_t n) {
    for (size_t i = 0; i + 188 <= n; i += 188) {
        const uint8_t* p = data + i;
        if (p[0] != 0x47) continue;
        uint16_t pid = ((p[1] & 0x1F) << 8) | p[2];
        bool start = p[1] & 0x40;
        int afc = (p[3] >> 4) & 0x03;
        size_t off = 4;
        if (afc == 2 || afc == 3) {
            if (off >= 188) continue;
            off += 1u + p[4];
        }
        if ((afc & 1) == 0 || off >= 188) continue;
        if (pid == 0 || pid == s->ts_pmt) {
            size_t poff = off;
            if (start) {
                uint8_t ptr = p[off];
                poff = off + 1u + ptr;
            }
            if (poff < 188) ts_section(s, pid, p + poff, 188 - poff);
            continue;
        }
        auto& buf = s->ts_pes[pid];
        if (start && !buf.empty()) {
            auto it = s->ts_type.find(pid);
            if (it != s->ts_type.end()) {
                if (it->second == "H264" || it->second == "H265") s->demux.hint_video(it->second);
            }
            s->demux.push(buf.data(), buf.size(), 0);
            pull_media(s);
            buf.clear();
        }
        buf.insert(buf.end(), p + off, p + 188);
        if (buf.size() > 1024 * 1024) buf.clear();
    }
}

static void dispatch_frame(Session* s, uint8_t pt, uint32_t ts, const std::vector<uint8_t>& buf) {
    if (buf.empty()) return;
    if (buf.size() >= 188 && buf.size() % 188 == 0 && buf[0] == 0x47) {
        feed_ts(s, buf.data(), buf.size());
        return;
    }
    s->demux.push(buf.data(), buf.size(), ts);
    pull_media(s);
}

static void decode_rtp(Session* s, const uint8_t* b, size_t n) {
    if (n < 12) return;
    uint8_t cc = b[0] & 0x0F;
    bool ext = b[0] & 0x10;
    bool pad = b[0] & 0x20;
    bool mark = b[1] & 0x80;
    uint8_t pt = b[1] & 0x7F;
    uint16_t seq = (b[2] << 8) | b[3];
    uint32_t ts = (b[4] << 24) | (b[5] << 16) | (b[6] << 8) | b[7];
    size_t off = 12 + cc * 4;
    size_t end = n;
    if (pad && end > off) {
        uint8_t plen = b[n - 1];
        if (plen < end - off) end -= plen;
    }
    if (ext && off + 4 <= end) {
        uint16_t elen = (b[off + 2] << 8) | b[off + 3];
        off += 4 + elen * 4;
    }
    if (off > end) return;
    s->stats.rtp_count++;
    s->stats.in_bytes += (int64_t)(end - off);
    s->last_rtp_ms = now_ms();
    const uint8_t* payload = b + off;
    size_t plen = end - off;
    if (pt == 98) {
        h264_rtp(s, payload, plen, ts);
        return;
    }
    if (pt == 99) {
        h265_rtp(s, payload, plen, ts);
        return;
    }
    if (pt == 0 || pt == 8) {
        s->stats.audio_codec = pt == 8 ? "G711A" : "G711U";
        s->broadcast(g711_flv_tag(payload, plen, ts / 8, pt == 8));
        return;
    }
    if (!s->asm_has || ts != s->asm_stamp || s->asm_buf.size() > 256 * 1024) {
        if (s->asm_has && !s->asm_drop) dispatch_frame(s, s->asm_pt, s->asm_stamp, s->asm_buf);
        s->asm_buf.clear();
        s->asm_drop = false;
        s->asm_stamp = ts;
        s->asm_pt = pt;
        s->asm_has = true;
        s->asm_seq_ok = false;
    } else if (s->asm_seq_ok && (uint16_t)(s->asm_seq + 1) != seq) {
        s->asm_drop = true;
        s->asm_buf.clear();
    }
    if (!s->asm_drop && plen) s->asm_buf.insert(s->asm_buf.end(), payload, payload + plen);
    s->asm_seq = seq;
    s->asm_seq_ok = true;
    if (mark) {
        if (!s->asm_drop) dispatch_frame(s, pt, ts, s->asm_buf);
        s->asm_buf.clear();
        s->asm_drop = false;
    }
}

static void flush_rtp(Session* s) {
    while (!s->rtp_buf.empty()) {
        uint16_t expect = s->last_seq + 1;
        auto it = s->rtp_buf.find(expect);
        if (it == s->rtp_buf.end()) {
            if (s->rtp_buf.size() < 32) return;
            uint16_t best = 0;
            uint16_t best_dist = 65535;
            for (auto& kv : s->rtp_buf) {
                uint16_t dist = (uint16_t)(kv.first - expect);
                if (dist < best_dist) {
                    best = kv.first;
                    best_dist = dist;
                }
            }
            if (best_dist > 1) s->stats.rtp_lost += best_dist - 1;
            s->last_seq = (uint16_t)(best - 1);
            continue;
        }
        auto pkt = std::move(it->second);
        s->rtp_buf.erase(it);
        s->last_seq = expect;
        decode_rtp(s, pkt.data(), pkt.size());
    }
}

static void handle_rtp(Session* s, const uint8_t* b, size_t n) {
    if (n < 12) return;
    s->forward_rtp(b, n);
    uint16_t seq = (b[2] << 8) | b[3];
    if (!s->has_seq) {
        s->has_seq = true;
        s->last_seq = (uint16_t)(seq - 1);
    }
    s->rtp_buf[seq] = std::vector<uint8_t>(b, b + n);
    flush_rtp(s);
}

static void udp_loop(std::shared_ptr<Session> s) {
    uint8_t buf[65536];
    while (s->running) {
        fd_set fds;
        FD_ZERO(&fds);
        FD_SET(s->udp_fd, &fds);
        timeval tv{1, 0};
        int rc = select(s->udp_fd + 1, &fds, nullptr, nullptr, &tv);
        if (rc <= 0) continue;
        ssize_t n = recv(s->udp_fd, buf, sizeof(buf), 0);
        if (n > 0) handle_rtp(s.get(), buf, (size_t)n);
    }
}

void sms_tcp_loop(std::shared_ptr<Session> s) {
    int cfd = -1;
    if (s->mode == "active") {
        // peer 地址由 open 时已经 connect 到 tcp_fd
        cfd = s->tcp_fd;
        s->tcp_fd = -1;
    } else {
        while (s->running && cfd < 0) {
            fd_set fds;
            FD_ZERO(&fds);
            FD_SET(s->tcp_fd, &fds);
            timeval tv{1, 0};
            if (select(s->tcp_fd + 1, &fds, nullptr, nullptr, &tv) <= 0) continue;
            cfd = accept(s->tcp_fd, nullptr, nullptr);
        }
    }
    if (cfd < 0) return;
    std::vector<uint8_t> acc;
    uint8_t buf[65536];
    while (s->running) {
        ssize_t n = recv(cfd, buf, sizeof(buf), 0);
        if (n <= 0) break;
        acc.insert(acc.end(), buf, buf + n);
        while (acc.size() >= 2) {
            uint16_t len = (acc[0] << 8) | acc[1];
            if (acc.size() < (size_t)len + 2) break;
            handle_rtp(s.get(), acc.data() + 2, len);
            acc.erase(acc.begin(), acc.begin() + 2 + len);
        }
        if (acc.size() > 4 * 1024 * 1024) acc.clear();
    }
    close(cfd);
}

std::shared_ptr<Session> Hub::open(const std::string& id, const std::string& transport, const std::string& mode, const std::string& ssrc, const std::string& peer_ip, int peer_port) {
    std::lock_guard<std::mutex> lk(mu_);
    auto it = sessions_.find(id);
    if (it != sessions_.end()) return it->second;
    auto s = std::make_shared<Session>(id);
    s->transport = transport;
    s->mode = mode;
    s->ssrc = ssrc;
    if (next_port_ < udp_min || next_port_ > udp_max) next_port_ = udp_min;
    int port = next_port_++;
    if (next_port_ > udp_max) next_port_ = udp_min;
    s->port = port;
    if (transport == "TCP" && mode == "active") {
        // 端口留给 SDP。设备 200 OK 之后由 relay 连过去，这里不 accept。
        int fd = socket(AF_INET, SOCK_STREAM, 0);
        int opt = 1;
        setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt));
        sockaddr_in addr{};
        addr.sin_family = AF_INET;
        addr.sin_port = htons(port);
        addr.sin_addr.s_addr = INADDR_ANY;
        bind(fd, (sockaddr*)&addr, sizeof(addr));
        s->tcp_fd = fd;
        if (!peer_ip.empty() && peer_port > 0) s->start_recv_tcp(s, peer_ip, peer_port);
    } else if (transport == "TCP") {
        int fd = socket(AF_INET, SOCK_STREAM, 0);
        int opt = 1;
        setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt));
        sockaddr_in addr{};
        addr.sin_family = AF_INET;
        addr.sin_port = htons(port);
        addr.sin_addr.s_addr = INADDR_ANY;
        bind(fd, (sockaddr*)&addr, sizeof(addr));
        listen(fd, 4);
        s->tcp_fd = fd;
        s->worker = std::thread(sms_tcp_loop, s);
    } else {
        int fd = socket(AF_INET, SOCK_DGRAM, 0);
        int opt = 1;
        setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt));
        sockaddr_in addr{};
        addr.sin_family = AF_INET;
        addr.sin_port = htons(port);
        addr.sin_addr.s_addr = INADDR_ANY;
        bind(fd, (sockaddr*)&addr, sizeof(addr));
        s->udp_fd = fd;
        s->worker = std::thread(udp_loop, s);
    }
    sessions_[id] = s;
    return s;
}

int Hub::grab_media_port() {
    std::lock_guard<std::mutex> lk(mu_);
    int p = next_port_++;
    if (next_port_ > udp_max) next_port_ = udp_min;
    if (p < udp_min) p = udp_min;
    return p;
}

int Hub::grab_rtc_port() {
    std::lock_guard<std::mutex> lk(mu_);
    if (next_rtc_ < rtc_lo || next_rtc_ > rtc_hi) next_rtc_ = rtc_lo;
    int p = next_rtc_++;
    if (next_rtc_ > rtc_hi) next_rtc_ = rtc_lo;
    return p;
}

void Hub::close(const std::string& id) {
    std::shared_ptr<Session> s;
    {
        std::lock_guard<std::mutex> lk(mu_);
        auto it = sessions_.find(id);
        if (it == sessions_.end()) return;
        s = it->second;
        sessions_.erase(it);
    }
    s->running = false;
}

std::shared_ptr<Session> Hub::get(const std::string& id) {
    std::lock_guard<std::mutex> lk(mu_);
    auto it = sessions_.find(id);
    if (it == sessions_.end()) return {};
    return it->second;
}
