#include "Common/sms.h"

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


static uint32_t mpeg_crc(const uint8_t* data, size_t n) {
    uint32_t crc = 0xFFFFFFFF;
    for (size_t i = 0; i < n; ++i) {
        crc ^= (uint32_t)data[i] << 24;
        for (int b = 0; b < 8; ++b) crc = (crc & 0x80000000u) ? (crc << 1) ^ 0x04C11DB7u : (crc << 1);
    }
    return crc;
}

static void ts_packet(std::vector<uint8_t>& out, uint16_t pid, uint8_t& cc, const uint8_t* payload, size_t n, bool pusi, const uint8_t* adapt, size_t adapt_n) {
    uint8_t pkt[188];
    memset(pkt, 0xFF, sizeof(pkt));
    pkt[0] = 0x47;
    pkt[1] = (pusi ? 0x40 : 0x00) | ((pid >> 8) & 0x1F);
    pkt[2] = (uint8_t)pid;
    bool has_adapt = adapt != nullptr;
    bool has_payload = n > 0;
    pkt[3] = (uint8_t)((has_adapt && has_payload ? 0x30 : has_adapt ? 0x20 : 0x10) | (cc & 0x0F));
    cc = (uint8_t)((cc + 1) & 0x0F);
    size_t off = 4;
    if (has_adapt) {
        pkt[off++] = (uint8_t)adapt_n;
        if (adapt_n && off + adapt_n <= 188) memcpy(pkt + off, adapt, adapt_n);
        off += adapt_n;
    }
    if (has_payload && off < 188) {
        size_t room = 188 - off;
        if (n > room) n = room;
        memcpy(pkt + off, payload, n);
    }
    out.insert(out.end(), pkt, pkt + 188);
}

static void ts_psi(std::vector<uint8_t>& out, uint16_t pid, uint8_t& cc, const std::vector<uint8_t>& section) {
    uint8_t payload[184];
    memset(payload, 0xFF, sizeof(payload));
    payload[0] = 0;
    size_t n = std::min(section.size(), (size_t)183);
    memcpy(payload + 1, section.data(), n);
    ts_packet(out, pid, cc, payload, 184, true, nullptr, 0);
}

static std::vector<uint8_t> psi_pat() {
    uint8_t s[16] = {};
    s[0] = 0x00;
    s[1] = 0xB0;
    s[2] = 0x0D;
    s[3] = 0x00;
    s[4] = 0x01;
    s[5] = 0xC1;
    s[8] = 0x00;
    s[9] = 0x01;
    s[10] = 0xF0;
    s[11] = 0x00;
    uint32_t crc = mpeg_crc(s, 12);
    s[12] = crc >> 24;
    s[13] = crc >> 16;
    s[14] = crc >> 8;
    s[15] = crc;
    return std::vector<uint8_t>(s, s + 16);
}

static std::vector<uint8_t> psi_pmt(bool h265) {
    uint8_t s[21] = {};
    s[0] = 0x02;
    s[1] = 0xB0;
    s[2] = 0x12;
    s[3] = 0x00;
    s[4] = 0x01;
    s[5] = 0xC1;
    s[8] = 0xE1;
    s[9] = 0x00;
    s[12] = h265 ? 0x24 : 0x1B;
    s[13] = 0xE1;
    s[14] = 0x00;
    uint32_t crc = mpeg_crc(s, 17);
    s[17] = crc >> 24;
    s[18] = crc >> 16;
    s[19] = crc >> 8;
    s[20] = crc;
    return std::vector<uint8_t>(s, s + 21);
}

static void write_pts(uint8_t* p, uint32_t pts) {
    p[0] = (uint8_t)(0x21 | (((pts >> 30) & 0x07) << 1));
    p[1] = (uint8_t)((pts >> 22) & 0xFF);
    p[2] = (uint8_t)((((pts >> 15) & 0x7F) << 1) | 1);
    p[3] = (uint8_t)((pts >> 7) & 0xFF);
    p[4] = (uint8_t)(((pts & 0x7F) << 1) | 1);
}

static void write_pcr(uint8_t* p, uint32_t pts90) {
    p[0] = (uint8_t)(pts90 >> 25);
    p[1] = (uint8_t)(pts90 >> 17);
    p[2] = (uint8_t)(pts90 >> 9);
    p[3] = (uint8_t)(pts90 >> 1);
    p[4] = (uint8_t)(((pts90 & 1) << 7) | 0x7E);
    p[5] = 0;
}

static void pes_to_ts(std::vector<uint8_t>& out, uint8_t& cc, const uint8_t* pes, size_t n, uint32_t pts) {
    size_t off = 0;
    bool first = true;
    while (off < n) {
        if (first) {
            size_t payload = std::min((size_t)176, n - off);
            size_t adapt_len = 183 - payload;
            std::vector<uint8_t> adapt(adapt_len);
            adapt[0] = 0x10;
            write_pcr(adapt.data() + 1, pts);
            for (size_t i = 7; i < adapt.size(); ++i) adapt[i] = 0xFF;
            ts_packet(out, 0x100, cc, pes + off, payload, true, adapt.data(), adapt.size());
            off += payload;
            first = false;
        } else if (n - off >= 184) {
            ts_packet(out, 0x100, cc, pes + off, 184, false, nullptr, 0);
            off += 184;
        } else {
            size_t payload = n - off;
            size_t adapt_len = 183 - payload;
            if (adapt_len == 0) {
                uint8_t dummy = 0;
                ts_packet(out, 0x100, cc, pes + off, payload, false, &dummy, 0);
            } else {
                std::vector<uint8_t> adapt(adapt_len, 0xFF);
                adapt[0] = 0x00;
                ts_packet(out, 0x100, cc, pes + off, payload, false, adapt.data(), adapt.size());
            }
            off += payload;
        }
    }
}
static void hls_close(Session& s, int64_t now) {
    if (s.hls_cur.empty()) return;
    HlsSeg seg;
    seg.seq = s.hls_seq++;
    double dur = s.hls_begin ? (now - s.hls_begin) / 1000.0 : 1;
    if (dur < 0.5) dur = 0.5;
    if (dur > 4) dur = 4;
    seg.dur = dur;
    seg.ts.swap(s.hls_cur);
    s.hls.push_back(std::move(seg));
    while (s.hls.size() > 5) s.hls.pop_front();
    s.hls_begin = now;
}

static void hls_push(Session& s, const std::vector<NAL>& nals, uint32_t pts) {
    bool key = false;
    bool h265 = false;
    for (auto& n : nals) {
        if (n.key) key = true;
        if (n.h265) h265 = true;
    }
    std::lock_guard<std::mutex> lk(s.dist_mu);
    if (!s.hls_armed) {
        if (!key) return;
        s.hls_armed = true;
        s.hls_begin = std::chrono::duration_cast<std::chrono::milliseconds>(
                          std::chrono::steady_clock::now().time_since_epoch())
                          .count();
    }
    int64_t now = std::chrono::duration_cast<std::chrono::milliseconds>(
                      std::chrono::steady_clock::now().time_since_epoch())
                      .count();
    if (key && !s.hls_cur.empty() && s.hls_begin && now - s.hls_begin >= 1000) hls_close(s, now);
    if (s.hls_cur.empty()) {
        ts_psi(s.hls_cur, 0, s.cc_pat, psi_pat());
        ts_psi(s.hls_cur, 0x1000, s.cc_pmt, psi_pmt(h265));
        if (!s.hls_begin) s.hls_begin = now;
    }
    std::vector<uint8_t> pes;
    pes.reserve(256);
    const uint8_t pes_hdr[] = {0, 0, 1, 0xE0, 0, 0, 0x80, 0x80, 5};
    pes.insert(pes.end(), pes_hdr, pes_hdr + sizeof(pes_hdr));
    uint8_t ptsb[5];
    write_pts(ptsb, pts);
    pes.insert(pes.end(), ptsb, ptsb + 5);
    for (auto& n : nals) pes.insert(pes.end(), n.data.begin(), n.data.end());
    pes_to_ts(s.hls_cur, s.cc_vid, pes.data(), pes.size(), pts);
}

std::string Session::hls_m3u8() {
    std::lock_guard<std::mutex> lk(dist_mu);
    std::deque<HlsSeg> view = hls;
    if (!hls_cur.empty()) {
        HlsSeg edge;
        edge.seq = hls_seq;
        edge.dur = 1;
        edge.ts = hls_cur;
        view.push_back(std::move(edge));
    }
    if (view.empty()) {
        return "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:2\n#EXT-X-MEDIA-SEQUENCE:0\n";
    }
    double maxd = 1;
    for (auto& g : view) maxd = std::max(maxd, g.dur);
    std::ostringstream o;
    o << "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:" << (int)(maxd + 0.999)
      << "\n#EXT-X-MEDIA-SEQUENCE:" << view.front().seq << "\n";
    for (auto& g : view) {
        o << "#EXTINF:" << g.dur << ",\nseg" << g.seq << ".ts\n";
    }
    return o.str();
}

bool Session::hls_seg(int seq, std::vector<uint8_t>& out) {
    std::lock_guard<std::mutex> lk(dist_mu);
    for (auto& g : hls) {
        if (g.seq == seq) {
            out = g.ts;
            return !out.empty();
        }
    }
    if (seq == hls_seq && !hls_cur.empty()) {
        out = hls_cur;
        return true;
    }
    return false;
}

void hls_on_video(Session& s, const std::vector<NAL>& nals, uint32_t ts) {
    hls_push(s, nals, ts);
}
