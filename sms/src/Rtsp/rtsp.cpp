#include "Common/sms.h"
#include "Common/net.h"

#include <arpa/inet.h>
#include <fcntl.h>
#include <netinet/in.h>
#include <netinet/tcp.h>
#include <sys/select.h>
#include <sys/socket.h>
#include <unistd.h>

#include <algorithm>
#include <cerrno>
#include <chrono>
#include <cstring>
#include <functional>
#include <iostream>
#include <map>
#include <sstream>
#include <thread>


static const uint8_t* nal_body(const NAL& n, size_t& len) {
    const uint8_t* p = n.data.data();
    size_t nlen = n.data.size();
    if (nlen >= 4 && p[0] == 0 && p[1] == 0 && p[2] == 0 && p[3] == 1) {
        len = nlen - 4;
        return p + 4;
    }
    if (nlen >= 3 && p[0] == 0 && p[1] == 0 && p[2] == 1) {
        len = nlen - 3;
        return p + 3;
    }
    len = nlen;
    return p;
}
static void rtp_header(uint8_t* h, uint16_t& seq, uint32_t ts, uint32_t ssrc, bool marker) {
    h[0] = 0x80;
    h[1] = (uint8_t)((marker ? 0x80 : 0) | 96);
    h[2] = (uint8_t)(seq >> 8);
    h[3] = (uint8_t)seq;
    seq++;
    h[4] = (uint8_t)(ts >> 24);
    h[5] = (uint8_t)(ts >> 16);
    h[6] = (uint8_t)(ts >> 8);
    h[7] = (uint8_t)ts;
    h[8] = (uint8_t)(ssrc >> 24);
    h[9] = (uint8_t)(ssrc >> 16);
    h[10] = (uint8_t)(ssrc >> 8);
    h[11] = (uint8_t)ssrc;
}

static void send_h264(uint16_t& seq, uint32_t ts, uint32_t ssrc, const uint8_t* nal, size_t n, bool marker,
                      const std::function<void(const uint8_t*, size_t)>& emit) {
    if (n == 0) return;
    if (n <= 1200) {
        std::vector<uint8_t> pkt(12 + n);
        rtp_header(pkt.data(), seq, ts, ssrc, marker);
        memcpy(pkt.data() + 12, nal, n);
        emit(pkt.data(), pkt.size());
        return;
    }
    uint8_t nalh = nal[0];
    size_t off = 1;
    while (off < n) {
        size_t chunk = std::min((size_t)1200, n - off);
        bool start = off == 1;
        bool end = off + chunk >= n;
        std::vector<uint8_t> pkt(14 + chunk);
        rtp_header(pkt.data(), seq, ts, ssrc, end && marker);
        pkt[12] = (uint8_t)((nalh & 0xE0) | 28);
        pkt[13] = (uint8_t)((start ? 0x80 : 0) | (end ? 0x40 : 0) | (nalh & 0x1F));
        memcpy(pkt.data() + 14, nal + off, chunk);
        emit(pkt.data(), pkt.size());
        off += chunk;
    }
}

static void rtsp_send(Session& s, const std::vector<NAL>& nals, uint32_t ts) {
    uint32_t ssrc = s.ssrc.empty() ? 1u : (uint32_t)strtoul(s.ssrc.c_str(), nullptr, 10);
    std::lock_guard<std::mutex> lk(s.dist_mu);
    if (s.rtsp.empty()) return;
    for (size_t i = 0; i < nals.size(); ++i) {
        size_t len = 0;
        const uint8_t* nal = nal_body(nals[i], len);
        if (!nal || len == 0 || nals[i].h265) continue;
        bool marker = i + 1 == nals.size();
        send_h264(s.rtsp_seq, ts, ssrc, nal, len, marker, [&](const uint8_t* p, size_t n) {
            uint8_t lead[4] = {'$', 0, (uint8_t)(n >> 8), (uint8_t)n};
            for (auto it = s.rtsp.begin(); it != s.rtsp.end();) {
                bool ok = true;
                if (it->tcp >= 0) ok = send_all_fd(it->tcp, lead, 4) && send_all_fd(it->tcp, p, n);
                if (it->udp >= 0) {
                    if (send(it->udp, p, n, MSG_NOSIGNAL) < 0) ok = false;
                }
                if (!ok) {
                    if (it->tcp >= 0) ::close(it->tcp);
                    if (it->udp >= 0) ::close(it->udp);
                    it = s.rtsp.erase(it);
                } else ++it;
            }
        });
    }
}

void Session::attach_rtsp(int tcp_fd, int udp_fd) {
    std::lock_guard<std::mutex> lk(dist_mu);
    RtspOut o;
    o.tcp = tcp_fd;
    o.udp = udp_fd;
    rtsp.push_back(o);
}

bool Session::release_rtsp(int tcp_fd, int udp_fd) {
    std::lock_guard<std::mutex> lk(dist_mu);
    for (auto it = rtsp.begin(); it != rtsp.end(); ++it) {
        if ((tcp_fd >= 0 && it->tcp == tcp_fd) || (udp_fd >= 0 && it->udp == udp_fd)) {
            rtsp.erase(it);
            return true;
        }
    }
    return false;
}
static std::string rtsp_header(const std::string& req, const std::string& key) {
    auto pos = req.find("\r\n" + key + ":");
    if (pos == std::string::npos) {
        if (req.rfind(key + ":", 0) == 0) pos = 0;
        else return "";
    }
    if (pos > 0) pos += 2;
    pos = req.find(':', pos);
    if (pos == std::string::npos) return "";
    pos++;
    while (pos < req.size() && req[pos] == ' ') pos++;
    auto end = req.find("\r\n", pos);
    return req.substr(pos, end == std::string::npos ? std::string::npos : end - pos);
}

static void rtsp_write(int fd, int code, const char* text, const std::string& cseq, const std::string& extra, const std::string& body) {
    std::ostringstream o;
    o << "RTSP/1.0 " << code << " " << text << "\r\nCSeq: " << (cseq.empty() ? "1" : cseq) << "\r\n" << extra;
    if (!body.empty()) o << "Content-Length: " << body.size() << "\r\n";
    o << "\r\n" << body;
    std::string s = o.str();
    send_all_fd(fd, (const uint8_t*)s.data(), s.size());
}

static std::string stream_of_url(std::string url) {
    auto q = url.find('?');
    if (q != std::string::npos) url = url.substr(0, q);
    std::string id;
    auto live = url.rfind("/live/");
    if (live != std::string::npos) id = url.substr(live + 6);
    else {
        auto slash = url.rfind('/');
        id = slash == std::string::npos ? url : url.substr(slash + 1);
    }
    auto slash = id.find('/');
    if (slash != std::string::npos) id = id.substr(0, slash);
    return id;
}

void rtsp_client(Hub* hub, int fd) {
    std::string stash;
    std::shared_ptr<Session> session;
    int udp_fd = -1;
    while (true) {
        if (!read_full(fd, stash, 1)) break;
        auto pos = stash.find("\r\n\r\n");
        while (pos == std::string::npos) {
            if (!read_full(fd, stash, stash.size() + 1)) {
                close(fd);
                if (udp_fd >= 0) close(udp_fd);
                return;
            }
            pos = stash.find("\r\n\r\n");
            if (stash.size() > 65536) {
                close(fd);
                return;
            }
        }
        std::string req = stash.substr(0, pos + 4);
        stash.erase(0, pos + 4);
        std::istringstream is(req);
        std::string method, url, ver;
        is >> method >> url >> ver;
        std::string cseq = rtsp_header(req, "CSeq");
        if (method == "OPTIONS") {
            rtsp_write(fd, 200, "OK", cseq, "Public: OPTIONS, DESCRIBE, SETUP, PLAY, TEARDOWN\r\n", "");
            continue;
        }
        std::string id = stream_of_url(url);
        auto s = hub->get(id);
        if (!s && method != "TEARDOWN") {
            rtsp_write(fd, 404, "Not Found", cseq, "", "");
            continue;
        }
        if (method == "DESCRIBE") {
            std::string sdp =
                "v=0\r\no=- 0 0 IN IP4 0.0.0.0\r\ns=ArgusSMS\r\nc=IN IP4 0.0.0.0\r\nt=0 0\r\n"
                "m=video 0 RTP/AVP 96\r\na=rtpmap:96 H264/90000\r\n"
                "a=fmtp:96 packetization-mode=1\r\na=control:track1\r\n";
            rtsp_write(fd, 200, "OK", cseq, "Content-Type: application/sdp\r\n", sdp);
        } else if (method == "SETUP") {
            std::string transport = rtsp_header(req, "Transport");
            if (transport.find("TCP") != std::string::npos) {
                rtsp_write(fd, 200, "OK", cseq,
                           "Transport: RTP/AVP/TCP;unicast;interleaved=0-1\r\nSession: 1\r\n", "");
            } else {
                int cport = 0;
                auto p = transport.find("client_port=");
                if (p != std::string::npos) cport = atoi(transport.c_str() + p + 12);
                sockaddr_in peer{};
                socklen_t plen = sizeof(peer);
                getpeername(fd, (sockaddr*)&peer, &plen);
                udp_fd = socket(AF_INET, SOCK_DGRAM, 0);
                sockaddr_in local{};
                local.sin_family = AF_INET;
                local.sin_addr.s_addr = INADDR_ANY;
                bind(udp_fd, (sockaddr*)&local, sizeof(local));
                sockaddr_in dst = peer;
                dst.sin_port = htons((uint16_t)cport);
                connect(udp_fd, (sockaddr*)&dst, sizeof(dst));
                sockaddr_in bound{};
                socklen_t bl = sizeof(bound);
                getsockname(udp_fd, (sockaddr*)&bound, &bl);
                int sport = ntohs(bound.sin_port);
                std::ostringstream extra;
                extra << "Transport: RTP/AVP;unicast;client_port=" << cport << "-" << (cport + 1)
                      << ";server_port=" << sport << "-" << (sport + 1) << "\r\nSession: 1\r\n";
                rtsp_write(fd, 200, "OK", cseq, extra.str(), "");
            }
            session = s;
        } else if (method == "PLAY") {
            int handed_tcp = -1;
            int handed_udp = -1;
            if (session) {
                if (udp_fd >= 0) handed_udp = udp_fd;
                else handed_tcp = fd;
                udp_fd = -1;
                session->attach_rtsp(handed_tcp, handed_udp);
            }
            rtsp_write(fd, 200, "OK", cseq, "Session: 1\r\nRange: npt=0.000-\r\n", "");
            if (handed_tcp >= 0 && session) {
                char sinkb[8];
                while (session->running && recv(fd, sinkb, sizeof(sinkb), 0) > 0) {
                }
                if (session->release_rtsp(handed_tcp, handed_udp)) close(fd);
                return;
            }
        } else if (method == "TEARDOWN") {
            rtsp_write(fd, 200, "OK", cseq, "", "");
            break;
        } else {
            rtsp_write(fd, 200, "OK", cseq, "", "");
        }
    }
    if (udp_fd >= 0) close(udp_fd);
    close(fd);
}

void rtsp_on_video(Session& s, const std::vector<NAL>& nals, uint32_t ts) {
    rtsp_send(s, nals, ts);
}
