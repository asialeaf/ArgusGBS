#include "Common/sms.h"
#include <iostream>
#include <fstream>
#include <string>
#include <cstdlib>

static std::string ini_get(const std::string& path, const std::string& section, const std::string& key, const std::string& def) {
    std::ifstream in(path);
    if (!in) return def;
    std::string cur, line;
    while (std::getline(in, line)) {
        if (!line.empty() && line.back() == '\r') line.pop_back();
        if (line.empty() || line[0] == ';' || line[0] == '#') continue;
        if (line.front() == '[' && line.back() == ']') {
            cur = line.substr(1, line.size() - 2);
            continue;
        }
        auto eq = line.find('=');
        if (eq == std::string::npos) continue;
        if (cur == section && line.substr(0, eq) == key) {
            auto v = line.substr(eq + 1);
            auto c = v.find(';');
            if (c != std::string::npos) v = v.substr(0, c);
            while (!v.empty() && v.back() == ' ') v.pop_back();
            return v;
        }
    }
    return def;
}

int main(int argc, char** argv) {
    std::string cfg = "configs/argussms.ini";
    for (int i = 1; i < argc; i++) {
        if (std::string(argv[i]) == "-config" && i + 1 < argc) cfg = argv[++i];
    }
    Hub hub;
    hub.http_port = std::atoi(ini_get(cfg, "http", "port", "10001").c_str());
    hub.rtsp_port = std::atoi(ini_get(cfg, "rtsp", "port", "554").c_str());
    hub.rtmp_port = std::atoi(ini_get(cfg, "rtmp", "port", "1935").c_str());
    hub.secret = ini_get(cfg, "http", "api_secret", "argus-sms");
    hub.advertise_ip = ini_get(cfg, "rtp", "advertise_ip", "");
    std::string range = ini_get(cfg, "rtp", "udp_port_range", "30000,30249");
    auto comma = range.find(',');
    if (comma != std::string::npos) {
        hub.udp_min = std::atoi(range.c_str());
        hub.udp_max = std::atoi(range.c_str() + comma + 1);
    }
    if (const char* env = std::getenv("ARGUS_ADVERTISE_IP")) {
        if (env[0] != '\0') hub.advertise_ip = env;
    }
    std::string wr = ini_get(cfg, "webrtc", "udp_port_range", "30250,30500");
    auto wcomma = wr.find(',');
    if (wcomma != std::string::npos) {
        hub.rtc_lo = std::atoi(wr.c_str());
        hub.rtc_hi = std::atoi(wr.c_str() + wcomma + 1);
    }
    std::cerr << "ArgusSMS 收国标 RTP/PS，分发 HTTP-FLV / HLS / RTSP / RTMP / WebRTC" << std::endl;
    hub.start();
    return 0;
}
