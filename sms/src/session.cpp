#include "sms.h"
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
    if (worker.joinable()) worker.detach();
}

void Session::broadcast(const std::vector<uint8_t>& tag) {
    if (tag.empty()) return;
    std::lock_guard<std::mutex> lk(mu);
    flv_tags.push_back(tag);
    while (flv_tags.size() > 400) flv_tags.pop_front();
    for (auto& s : subs) {
        if (!s->sent_header || s->fd < 0) continue;
        send(s->fd, tag.data(), tag.size(), MSG_NOSIGNAL);
    }
    stats.num_outputs = (int)subs.size();
}

std::vector<uint8_t> Session::flv_header_and_gop() {
    std::vector<uint8_t> out = flv_file_header();
    if (!sps.empty() && !pps.empty()) {
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
    for (auto& n : nals) {
        auto raw = n.data;
        size_t off = 0;
        if (raw.size() >= 4 && raw[0] == 0 && raw[1] == 0 && raw[2] == 0 && raw[3] == 1) off = 4;
        else if (raw.size() >= 3 && raw[0] == 0 && raw[1] == 0 && raw[2] == 1) off = 3;
        if (off >= raw.size()) continue;
        int t = raw[off] & 0x1F;
        if (t == 7) sps.assign(raw.begin() + off, raw.end());
        if (t == 8) pps.assign(raw.begin() + off, raw.end());
        auto tag = annexb_to_flv_tag(n, dts, sps, pps, dummy);
        if (!tag.empty()) {
            stats.video_frames++;
            stats.ready = true;
            stats.codec = demux.codec();
            broadcast(tag);
        }
    }
}

static void handle_rtp(Session* s, const uint8_t* b, size_t n) {
    if (n < 12) return;
    uint8_t cc = b[0] & 0x0F;
    bool ext = b[0] & 0x10;
    uint16_t seq = (b[2] << 8) | b[3];
    uint32_t ts = (b[4] << 24) | (b[5] << 16) | (b[6] << 8) | b[7];
    size_t off = 12 + cc * 4;
    if (ext && off + 4 <= n) {
        uint16_t elen = (b[off + 2] << 8) | b[off + 3];
        off += 4 + elen * 4;
    }
    if (off >= n) return;
    s->stats.rtp_count++;
    s->stats.in_bytes += (int64_t)n;
    if (s->has_seq) {
        uint16_t expect = s->last_seq + 1;
        if (seq != expect && seq != s->last_seq) {
            int gap = (int)seq - (int)expect;
            if (gap < 0) gap += 65536;
            if (gap < 1000) s->stats.rtp_lost += gap;
        }
    }
    s->has_seq = true;
    s->last_seq = seq;
    s->last_rtp_ms = now_ms();
    s->demux.push(b + off, n - off, ts);
    auto nals = s->demux.take();
    s->add_nals(nals);
    int64_t dt = now_ms() - (s->last_rtp_ms > 1000 ? s->last_rtp_ms - 1000 : s->last_rtp_ms);
    if (dt > 0) s->stats.in_bitrate = (int)(s->stats.in_bytes * 8 / (dt > 0 ? 1 : 1) / 1000);
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

static void tcp_loop(std::shared_ptr<Session> s) {
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
    int port = next_port_++;
    if (next_port_ > udp_max) next_port_ = udp_min;
    s->port = port;
    if (transport == "TCP" && mode == "active" && !peer_ip.empty() && peer_port > 0) {
        int fd = socket(AF_INET, SOCK_STREAM, 0);
        sockaddr_in addr{};
        addr.sin_family = AF_INET;
        addr.sin_port = htons(peer_port);
        inet_pton(AF_INET, peer_ip.c_str(), &addr.sin_addr);
        if (connect(fd, (sockaddr*)&addr, sizeof(addr)) == 0) {
            s->tcp_fd = fd;
            s->worker = std::thread(tcp_loop, s);
        } else {
            ::close(fd);
        }
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
        s->worker = std::thread(tcp_loop, s);
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
