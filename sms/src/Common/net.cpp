#include "Common/net.h"
#include "Common/sms.h"
#include <arpa/inet.h>
#include <iostream>
#include <netinet/in.h>
#include <sys/socket.h>
#include <unistd.h>
#include <cerrno>
#include <thread>

bool send_all_fd(int fd, const uint8_t* p, size_t n) {
    size_t off = 0;
    while (off < n) {
        ssize_t w = send(fd, p + off, n - off, MSG_NOSIGNAL);
        if (w < 0 && (errno == EAGAIN || errno == EWOULDBLOCK)) return true;
        if (w <= 0) return false;
        off += (size_t)w;
    }
    return true;
}
bool read_full(int fd, std::string& stash, size_t n) {
    while (stash.size() < n) {
        char buf[4096];
        ssize_t r = recv(fd, buf, sizeof(buf), 0);
        if (r <= 0) return false;
        stash.append(buf, buf + r);
    }
    return true;
}
void listen_loop(int port, const char* name, void (*client)(Hub*, int), Hub* hub) {
    int fd = socket(AF_INET, SOCK_STREAM, 0);
    if (fd < 0) return;
    int opt = 1;
    setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt));
    sockaddr_in addr{};
    addr.sin_family = AF_INET;
    addr.sin_port = htons((uint16_t)port);
    addr.sin_addr.s_addr = INADDR_ANY;
    if (bind(fd, (sockaddr*)&addr, sizeof(addr)) != 0) {
        std::cerr << name << " 端口 " << port << " 绑定失败" << std::endl;
        close(fd);
        return;
    }
    listen(fd, 64);
    std::cerr << name << " :" << port << std::endl;
    while (true) {
        int c = accept(fd, nullptr, nullptr);
        if (c < 0) continue;
        std::thread(client, hub, c).detach();
    }
}
