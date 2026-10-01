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


static int tcp_connect(const std::string& ip, int port, int timeout_ms) {
    int fd = socket(AF_INET, SOCK_STREAM, 0);
    if (fd < 0) return -1;
    int flags = fcntl(fd, F_GETFL, 0);
    fcntl(fd, F_SETFL, flags | O_NONBLOCK);
    sockaddr_in addr{};
    addr.sin_family = AF_INET;
    addr.sin_port = htons((uint16_t)port);
    if (inet_pton(AF_INET, ip.c_str(), &addr.sin_addr) != 1) {
        close(fd);
        return -1;
    }
    int rc = connect(fd, (sockaddr*)&addr, sizeof(addr));
    if (rc != 0 && errno != EINPROGRESS) {
        close(fd);
        return -1;
    }
    if (rc != 0) {
        fd_set w;
        FD_ZERO(&w);
        FD_SET(fd, &w);
        timeval tv{timeout_ms / 1000, (timeout_ms % 1000) * 1000};
        if (select(fd + 1, nullptr, &w, nullptr, &tv) <= 0) {
            close(fd);
            return -1;
        }
        int err = 0;
        socklen_t len = sizeof(err);
        getsockopt(fd, SOL_SOCKET, SO_ERROR, &err, &len);
        if (err != 0) {
            close(fd);
            return -1;
        }
    }
    fcntl(fd, F_SETFL, flags);
    timeval snd{0, 200000};
    setsockopt(fd, SOL_SOCKET, SO_SNDTIMEO, &snd, sizeof(snd));
    int one = 1;
    setsockopt(fd, IPPROTO_TCP, TCP_NODELAY, &one, sizeof(one));
    return fd;
}
static void relay_accept(std::shared_ptr<Session> s) {
    while (s->running && s->relay_listen >= 0) {
        fd_set fds;
        FD_ZERO(&fds);
        FD_SET(s->relay_listen, &fds);
        timeval tv{1, 0};
        if (select(s->relay_listen + 1, &fds, nullptr, nullptr, &tv) <= 0) continue;
        int c = accept(s->relay_listen, nullptr, nullptr);
        if (c < 0) continue;
        timeval snd{0, 200000};
        setsockopt(c, SOL_SOCKET, SO_SNDTIMEO, &snd, sizeof(snd));
        int one = 1;
        setsockopt(c, IPPROTO_TCP, TCP_NODELAY, &one, sizeof(one));
        std::lock_guard<std::mutex> lk(s->dist_mu);
        RtpSink sk;
        sk.fd = c;
        sk.tcp = true;
        s->sinks.push_back(std::move(sk));
        std::cerr << "级联 TCP 已接入 " << s->id << std::endl;
    }
}

bool Session::start_recv_tcp(std::shared_ptr<Session> self, const std::string& ip, int port) {
    if (ip.empty() || port <= 0) return false;
    bool expected = false;
    if (!recv_started.compare_exchange_strong(expected, true)) return true;
    std::thread([self, ip, port] {
        for (int i = 0; i < 40 && self->running; ++i) {
            if (self->worker.joinable()) return;
            int fd = tcp_connect(ip, port, 500);
            if (fd >= 0) {
                if (self->tcp_fd >= 0) {
                    ::close(self->tcp_fd);
                    self->tcp_fd = -1;
                }
                self->mode = "active";
                self->transport = "TCP";
                self->tcp_fd = fd;
                self->worker = std::thread(sms_tcp_loop, self);
                std::cerr << "TCP 主动已连接 " << ip << ":" << port << " 流 " << self->id << std::endl;
                return;
            }
            std::this_thread::sleep_for(std::chrono::milliseconds(250));
        }
        self->recv_started = false;
        std::cerr << "TCP 主动连接失败 " << ip << ":" << port << std::endl;
    }).detach();
    return true;
}

int Session::start_send_relay(std::shared_ptr<Session> self, const std::string& transport, const std::string& mode, const std::string& ip, int port, int listen_port) {
    bool tcp = transport == "TCP";
    bool peer_connects = tcp && mode == "active";
    if (peer_connects) {
        int fd = socket(AF_INET, SOCK_STREAM, 0);
        if (fd < 0) return -1;
        int opt = 1;
        setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt));
        sockaddr_in addr{};
        addr.sin_family = AF_INET;
        addr.sin_port = htons((uint16_t)listen_port);
        addr.sin_addr.s_addr = INADDR_ANY;
        if (bind(fd, (sockaddr*)&addr, sizeof(addr)) != 0 || listen(fd, 4) != 0) {
            ::close(fd);
            return -1;
        }
        relay_listen = fd;
        std::thread(relay_accept, self).detach();
        return listen_port;
    }
    if (ip.empty() || port <= 0) return -1;
    if (!tcp) {
        int fd = socket(AF_INET, SOCK_DGRAM, 0);
        if (fd < 0) return -1;
        std::lock_guard<std::mutex> lk(dist_mu);
        RtpSink sk;
        sk.fd = fd;
        sk.udp = true;
        sk.ip = ip;
        sk.port = port;
        sinks.push_back(std::move(sk));
        return 0;
    }
    std::thread([self, ip, port] {
        for (int i = 0; i < 40 && self->running; ++i) {
            int fd = tcp_connect(ip, port, 500);
            if (fd >= 0) {
                std::lock_guard<std::mutex> lk(self->dist_mu);
                RtpSink sk;
                sk.fd = fd;
                sk.tcp = true;
                sk.ip = ip;
                sk.port = port;
                self->sinks.push_back(std::move(sk));
                std::cerr << "级联 TCP 已连到 " << ip << ":" << port << std::endl;
                return;
            }
            std::this_thread::sleep_for(std::chrono::milliseconds(250));
        }
        std::cerr << "级联 TCP 连接失败 " << ip << ":" << port << std::endl;
    }).detach();
    return 0;
}

void Session::forward_rtp(const uint8_t* pkt, size_t n) {
    if (n < 12) return;
    std::vector<uint8_t> copy(pkt, pkt + n);
    if (!ssrc.empty()) {
        unsigned long v = strtoul(ssrc.c_str(), nullptr, 10);
        copy[8] = (uint8_t)(v >> 24);
        copy[9] = (uint8_t)(v >> 16);
        copy[10] = (uint8_t)(v >> 8);
        copy[11] = (uint8_t)v;
    }
    std::lock_guard<std::mutex> lk(dist_mu);
    if (sinks.empty()) return;
    for (auto it = sinks.begin(); it != sinks.end();) {
        bool ok = true;
        if (it->tcp) {
            uint8_t hdr[2] = {(uint8_t)(copy.size() >> 8), (uint8_t)copy.size()};
            ok = send_all_fd(it->fd, hdr, 2) && send_all_fd(it->fd, copy.data(), copy.size());
        } else if (it->udp) {
            sockaddr_in addr{};
            addr.sin_family = AF_INET;
            addr.sin_port = htons((uint16_t)it->port);
            if (inet_pton(AF_INET, it->ip.c_str(), &addr.sin_addr) != 1) ok = false;
            else if (sendto(it->fd, copy.data(), copy.size(), MSG_NOSIGNAL, (sockaddr*)&addr, sizeof(addr)) < 0) ok = false;
        }
        if (!ok) {
            ::close(it->fd);
            it = sinks.erase(it);
        } else {
            ++it;
        }
    }
}

bool Hub::relay(const std::string& id, const std::string& transport, const std::string& mode, const std::string& ip, int port, const std::string& direction, int& listen_port) {
    auto s = get(id);
    if (!s) return false;
    listen_port = 0;
    if (direction == "recv") return s->start_recv_tcp(s, ip, port);
    int bind_port = (transport == "TCP" && mode == "active") ? grab_media_port() : 0;
    int rc = s->start_send_relay(s, transport, mode, ip, port, bind_port);
    if (rc < 0) return false;
    listen_port = rc;
    return true;
}

std::string Hub::public_ip() const {
    if (!advertise_ip.empty()) return advertise_ip;
    return host_ip();
}
