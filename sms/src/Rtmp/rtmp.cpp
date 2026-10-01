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


bool Session::release_rtmp(int fd) {
    std::lock_guard<std::mutex> lk(dist_mu);
    for (auto it = rtmp.begin(); it != rtmp.end(); ++it) {
        if (*it == fd) {
            rtmp.erase(it);
            return true;
        }
    }
    return false;
}

static void rtmp_chunk(int fd, int csid, uint8_t type, uint32_t ts, uint32_t stream_id, const uint8_t* data, size_t len) {
    const size_t chunk = 4096;
    size_t off = 0;
    bool first = true;
    do {
        uint8_t hdr[16];
        size_t h = 0;
        bool ext = ts >= 0xFFFFFF;
        if (first) {
            hdr[h++] = (uint8_t)csid;
            uint32_t ts24 = ext ? 0xFFFFFF : ts;
            hdr[h++] = (uint8_t)(ts24 >> 16);
            hdr[h++] = (uint8_t)(ts24 >> 8);
            hdr[h++] = (uint8_t)ts24;
            hdr[h++] = (uint8_t)(len >> 16);
            hdr[h++] = (uint8_t)(len >> 8);
            hdr[h++] = (uint8_t)len;
            hdr[h++] = type;
            hdr[h++] = (uint8_t)stream_id;
            hdr[h++] = (uint8_t)(stream_id >> 8);
            hdr[h++] = (uint8_t)(stream_id >> 16);
            hdr[h++] = (uint8_t)(stream_id >> 24);
            if (ext) {
                hdr[h++] = (uint8_t)(ts >> 24);
                hdr[h++] = (uint8_t)(ts >> 16);
                hdr[h++] = (uint8_t)(ts >> 8);
                hdr[h++] = (uint8_t)ts;
            }
            first = false;
        } else {
            hdr[h++] = (uint8_t)(0xC0 | csid);
            if (ext) {
                hdr[h++] = (uint8_t)(ts >> 24);
                hdr[h++] = (uint8_t)(ts >> 16);
                hdr[h++] = (uint8_t)(ts >> 8);
                hdr[h++] = (uint8_t)ts;
            }
        }
        size_t n = std::min(chunk, len - off);
        if (!send_all_fd(fd, hdr, h)) return;
        if (n && !send_all_fd(fd, data + off, n)) return;
        off += n;
        if (len == 0) break;
    } while (off < len);
}

static void rtmp_write_tag(int fd, const std::vector<uint8_t>& tag) {
    if (tag.size() < 16) return;
    uint8_t typ = tag[0];
    if (typ != 8 && typ != 9) return;
    uint32_t ds = ((uint32_t)tag[1] << 16) | ((uint32_t)tag[2] << 8) | tag[3];
    uint32_t ts = ((uint32_t)tag[4] << 16) | ((uint32_t)tag[5] << 8) | tag[6] | ((uint32_t)tag[7] << 24);
    if (tag.size() < 11u + ds) return;
    rtmp_chunk(fd, 6, typ, ts, 1, tag.data() + 11, ds);
}

void Session::fanout_flv(const std::vector<uint8_t>& tag) {
    std::lock_guard<std::mutex> lk(dist_mu);
    for (auto it = rtmp.begin(); it != rtmp.end();) {
        rtmp_write_tag(*it, tag);
        ++it;
    }
}

void Session::attach_rtmp(int fd) {
    std::vector<uint8_t> seq;
    std::vector<std::vector<uint8_t>> tags;
    {
        std::lock_guard<std::mutex> lk(mu);
        if (!vps.empty() && !sps.empty() && !pps.empty()) seq = hevc_seq_header(vps, sps, pps);
        else if (!sps.empty() && !pps.empty()) seq = avc_seq_header(sps, pps);
        tags.assign(flv_tags.begin(), flv_tags.end());
    }
    {
        std::lock_guard<std::mutex> lk(dist_mu);
        rtmp.push_back(fd);
    }
    if (!seq.empty()) rtmp_write_tag(fd, seq);
    for (auto& t : tags) rtmp_write_tag(fd, t);
}
static void amf_str(std::vector<uint8_t>& o, const std::string& s) {
    o.push_back(0x02);
    o.push_back((uint8_t)(s.size() >> 8));
    o.push_back((uint8_t)s.size());
    o.insert(o.end(), s.begin(), s.end());
}

static void amf_num(std::vector<uint8_t>& o, double d) {
    o.push_back(0x00);
    uint64_t u;
    memcpy(&u, &d, 8);
    for (int i = 7; i >= 0; --i) o.push_back((uint8_t)(u >> (i * 8)));
}

static void amf_key(std::vector<uint8_t>& o, const std::string& s) {
    o.push_back((uint8_t)(s.size() >> 8));
    o.push_back((uint8_t)s.size());
    o.insert(o.end(), s.begin(), s.end());
}

static void amf_obj_end(std::vector<uint8_t>& o) {
    o.push_back(0);
    o.push_back(0);
    o.push_back(0x09);
}

static void skip_amf(const uint8_t* p, size_t n, size_t& i) {
    if (i >= n) return;
    uint8_t t = p[i++];
    if (t == 0x00) i += 8;
    else if (t == 0x01) i += 1;
    else if (t == 0x02) {
        if (i + 2 > n) return;
        int l = (p[i] << 8) | p[i + 1];
        i += 2 + l;
    } else if (t == 0x05 || t == 0x06) {
    } else if (t == 0x03 || t == 0x08) {
        if (t == 0x08) i += 4;
        while (i + 3 <= n) {
            int l = (p[i] << 8) | p[i + 1];
            i += 2;
            if (l == 0) {
                i++;
                break;
            }
            i += l;
            skip_amf(p, n, i);
        }
    } else i = n;
}

static std::string amf_cmd(const uint8_t* p, size_t n, double& trans, std::string& arg) {
    size_t i = 0;
    if (n > 0 && p[0] == 0x00) i = 1;
    if (i >= n || p[i] != 0x02) return "";
    i++;
    if (i + 2 > n) return "";
    int l = (p[i] << 8) | p[i + 1];
    i += 2;
    if (i + l > n) return "";
    std::string cmd((const char*)p + i, l);
    i += l;
    if (i < n && p[i] == 0x00 && i + 9 <= n) {
        uint64_t u = 0;
        for (int k = 0; k < 8; ++k) u = (u << 8) | p[i + 1 + k];
        memcpy(&trans, &u, 8);
        i += 9;
    }
    skip_amf(p, n, i);
    if (i < n && p[i] == 0x02) {
        i++;
        if (i + 2 <= n) {
            int al = (p[i] << 8) | p[i + 1];
            i += 2;
            if (i + al <= n) arg.assign((const char*)p + i, al);
        }
    }
    return cmd;
}

struct RtmpCs {
    uint32_t ts = 0;
    uint32_t len = 0;
    uint32_t stream = 0;
    uint8_t type = 0;
    std::vector<uint8_t> body;
    bool have = false;
};

void rtmp_client(Hub* hub, int fd) {
    std::string stash;
    if (!read_full(fd, stash, 1 + 1536)) {
        close(fd);
        return;
    }
    std::string c1 = stash.substr(1, 1536);
    stash.erase(0, 1 + 1536);
    std::string s0s1s2;
    s0s1s2.push_back(3);
    s0s1s2.append(1536, '\0');
    s0s1s2 += c1;
    send_all_fd(fd, (const uint8_t*)s0s1s2.data(), s0s1s2.size());
    if (!read_full(fd, stash, 1536)) {
        close(fd);
        return;
    }
    stash.erase(0, 1536);
    auto be32 = [](uint32_t v) {
        uint8_t b[4] = {(uint8_t)(v >> 24), (uint8_t)(v >> 16), (uint8_t)(v >> 8), (uint8_t)v};
        return std::vector<uint8_t>(b, b + 4);
    };
    auto win = be32(2500000);
    rtmp_chunk(fd, 2, 5, 0, 0, win.data(), 4);
    std::vector<uint8_t> bw = win;
    bw.push_back(2);
    rtmp_chunk(fd, 2, 6, 0, 0, bw.data(), 5);
    auto cs = be32(4096);
    rtmp_chunk(fd, 2, 1, 0, 0, cs.data(), 4);

    uint32_t in_chunk = 128;
    std::map<int, RtmpCs> channels;
    std::shared_ptr<Session> playing;
    while (hub) {
        if (stash.size() < 1 && !read_full(fd, stash, 1)) break;
        uint8_t b0 = (uint8_t)stash[0];
        int fmt = b0 >> 6;
        int csid = b0 & 0x3F;
        size_t basic = 1;
        if (csid == 0) basic = 2;
        if (csid == 1) basic = 3;
        size_t mh = basic + (fmt == 0 ? 11 : fmt == 1 ? 7 : fmt == 2 ? 3 : 0);
        if (!read_full(fd, stash, mh)) break;
        const uint8_t* h = (const uint8_t*)stash.data();
        if (csid == 0) csid = h[1] + 64;
        if (csid == 1) csid = h[2] + 64 + (h[1] << 8);
        auto& m = channels[csid];
        const uint8_t* mhptr = h + basic;
        if (fmt == 0) {
            uint32_t ts = (mhptr[0] << 16) | (mhptr[1] << 8) | mhptr[2];
            m.len = (mhptr[3] << 16) | (mhptr[4] << 8) | mhptr[5];
            m.type = mhptr[6];
            m.stream = mhptr[7] | (mhptr[8] << 8) | (mhptr[9] << 16) | (mhptr[10] << 24);
            m.ts = ts;
            m.body.clear();
            m.have = true;
        } else if (fmt == 1) {
            uint32_t delta = (mhptr[0] << 16) | (mhptr[1] << 8) | mhptr[2];
            m.len = (mhptr[3] << 16) | (mhptr[4] << 8) | mhptr[5];
            m.type = mhptr[6];
            m.ts += delta;
            m.body.clear();
            m.have = true;
        } else if (fmt == 2) {
            uint32_t delta = (mhptr[0] << 16) | (mhptr[1] << 8) | mhptr[2];
            m.ts += delta;
            m.body.clear();
        }
        bool ts_ext = (fmt == 0 || fmt == 1) && mhptr[0] == 0xFF && mhptr[1] == 0xFF && mhptr[2] == 0xFF;
        stash.erase(0, mh);
        if (ts_ext) {
            if (!read_full(fd, stash, 4)) break;
            m.ts = ((uint8_t)stash[0] << 24) | ((uint8_t)stash[1] << 16) | ((uint8_t)stash[2] << 8) | (uint8_t)stash[3];
            stash.erase(0, 4);
        }
        size_t need = std::min((size_t)in_chunk, (size_t)m.len - m.body.size());
        if (!read_full(fd, stash, need)) break;
        m.body.insert(m.body.end(), stash.begin(), stash.begin() + need);
        stash.erase(0, need);
        if (m.body.size() < m.len) continue;
        if (m.type == 1 && m.body.size() >= 4) {
            in_chunk = ((uint32_t)m.body[0] << 24) | ((uint32_t)m.body[1] << 16) | ((uint32_t)m.body[2] << 8) | m.body[3];
            if (in_chunk == 0) in_chunk = 128;
        } else if (m.type == 20 || m.type == 17) {
            double trans = 0;
            std::string arg;
            std::string cmd = amf_cmd(m.body.data(), m.body.size(), trans, arg);
            if (cmd == "connect") {
                std::vector<uint8_t> body;
                amf_str(body, "_result");
                amf_num(body, trans);
                body.push_back(0x03);
                amf_key(body, "fmsVer");
                amf_str(body, "FMS/3,0,1,123");
                amf_key(body, "capabilities");
                amf_num(body, 31);
                amf_obj_end(body);
                body.push_back(0x03);
                amf_key(body, "level");
                amf_str(body, "status");
                amf_key(body, "code");
                amf_str(body, "NetConnection.Connect.Success");
                amf_key(body, "description");
                amf_str(body, "ok");
                amf_obj_end(body);
                rtmp_chunk(fd, 3, 20, 0, 0, body.data(), body.size());
            } else if (cmd == "createStream") {
                std::vector<uint8_t> body;
                amf_str(body, "_result");
                amf_num(body, trans);
                body.push_back(0x05);
                amf_num(body, 1);
                rtmp_chunk(fd, 3, 20, 0, 0, body.data(), body.size());
            } else if (cmd == "play") {
                std::string name = arg;
                auto slash = name.rfind('/');
                if (slash != std::string::npos) name = name.substr(slash + 1);
                playing = hub->get(name);
                if (!playing) playing = hub->get(arg);
                uint8_t begin[6] = {0, 0, 0, 0, 0, 1};
                rtmp_chunk(fd, 2, 4, 0, 0, begin, 6);
                auto status = [](const char* code) {
                    std::vector<uint8_t> body;
                    amf_str(body, "onStatus");
                    amf_num(body, 0);
                    body.push_back(0x05);
                    body.push_back(0x03);
                    amf_key(body, "level");
                    amf_str(body, "status");
                    amf_key(body, "code");
                    amf_str(body, code);
                    amf_key(body, "description");
                    amf_str(body, code);
                    amf_obj_end(body);
                    return body;
                };
                auto reset = status("NetStream.Play.Reset");
                auto start = status("NetStream.Play.Start");
                rtmp_chunk(fd, 4, 20, 0, 1, reset.data(), reset.size());
                rtmp_chunk(fd, 4, 20, 0, 1, start.data(), start.size());
                if (playing) playing->attach_rtmp(fd);
            } else if (!cmd.empty() && cmd != "play") {
                std::vector<uint8_t> body;
                amf_str(body, "_result");
                amf_num(body, trans);
                body.push_back(0x05);
                rtmp_chunk(fd, 3, 20, 0, 0, body.data(), body.size());
            }
        }
        m.body.clear();
    }
    if (!playing || playing->release_rtmp(fd)) close(fd);
}
