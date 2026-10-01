#include "sms.h"
#include <cstring>

static int start_code(const uint8_t* p, size_t n) {
    if (n >= 4 && p[0] == 0 && p[1] == 0 && p[2] == 0 && p[3] == 1) return 4;
    if (n >= 3 && p[0] == 0 && p[1] == 0 && p[2] == 1) return 3;
    return 0;
}

void PsDemux::push(const uint8_t* data, size_t len, uint32_t rtp_ts) {
    ts_ = rtp_ts;
    buf_.insert(buf_.end(), data, data + len);
    if (buf_.size() > 4 * 1024 * 1024) {
        buf_.erase(buf_.begin(), buf_.begin() + buf_.size() / 2);
    }
    feed();
}

std::vector<NAL> PsDemux::take() {
    std::vector<NAL> o;
    o.swap(out_);
    return o;
}

void PsDemux::feed() {
    size_t i = 0;
    while (i + 4 < buf_.size()) {
        int sc = start_code(&buf_[i], buf_.size() - i);
        if (!sc) {
            i++;
            continue;
        }
        size_t next = i + sc;
        size_t j = next;
        while (j + 3 < buf_.size()) {
            if (start_code(&buf_[j], buf_.size() - j)) break;
            j++;
        }
        if (j + 3 >= buf_.size() && !start_code(&buf_[j > 0 ? j - 1 : 0], 1)) {
            // 等更多数据，除非已经很长
            if (buf_.size() - i < 1024 * 1024) break;
            j = buf_.size();
        }
        uint8_t sid = buf_[i + sc];
        const uint8_t* payload = &buf_[i];
        size_t plen = j - i;
        if (sid >= 0xE0 && sid <= 0xEF) {
            on_pes(payload, plen, ts_);
        } else if (sid == 0xBC && plen > 10) {
            // PSM: 简单扫描 stream_type
            for (size_t k = sc + 6; k + 1 < plen; k++) {
                if (buf_[i + k] == 0x1B) codec_ = "H264";
                if (buf_[i + k] == 0x24) codec_ = "H265";
            }
        }
        i = j;
    }
    if (i > 0) buf_.erase(buf_.begin(), buf_.begin() + i);
}

void PsDemux::on_pes(const uint8_t* p, size_t n, uint32_t ts) {
    int sc = start_code(p, n);
    if (!sc || n < (size_t)sc + 9) return;
    size_t off = sc;
    // PES header
    if (off + 9 > n) return;
    uint8_t header_len = p[off + 8];
    size_t es = off + 9 + header_len;
    if (es >= n) return;
    const uint8_t* esd = p + es;
    size_t esn = n - es;
    size_t i = 0;
    while (i + 3 < esn) {
        int s = start_code(esd + i, esn - i);
        if (!s) { i++; continue; }
        size_t ns = i + s;
        size_t j = ns;
        while (j + 3 < esn) {
            if (start_code(esd + j, esn - j)) break;
            j++;
        }
        if (j + 3 >= esn) j = esn;
        if (ns < j) {
            uint8_t nalh = esd[ns];
            int nal_type = nalh & 0x1F;
            bool key = nal_type == 5 || nal_type == 7 || nal_type == 8;
            if (codec_ == "H265") {
                int t = (esd[ns] >> 1) & 0x3F;
                key = t == 19 || t == 20 || t == 21 || t == 32 || t == 33 || t == 34;
            }
            NAL nal;
            nal.key = key;
            nal.ts90 = ts;
            nal.data.assign(esd + i, esd + j);
            out_.push_back(std::move(nal));
        }
        i = j;
    }
}

static void be24(std::vector<uint8_t>& o, uint32_t v) {
    o.push_back((v >> 16) & 0xFF);
    o.push_back((v >> 8) & 0xFF);
    o.push_back(v & 0xFF);
}
static void be32(std::vector<uint8_t>& o, uint32_t v) {
    o.push_back((v >> 24) & 0xFF);
    o.push_back((v >> 16) & 0xFF);
    o.push_back((v >> 8) & 0xFF);
    o.push_back(v & 0xFF);
}

static std::vector<uint8_t> strip_start(const std::vector<uint8_t>& in) {
    size_t i = 0;
    if (in.size() >= 4 && in[0] == 0 && in[1] == 0 && in[2] == 0 && in[3] == 1) i = 4;
    else if (in.size() >= 3 && in[0] == 0 && in[1] == 0 && in[2] == 1) i = 3;
    return std::vector<uint8_t>(in.begin() + i, in.end());
}

std::vector<uint8_t> flv_file_header() {
    return {'F', 'L', 'V', 0x01, 0x01, 0, 0, 0, 0x09, 0, 0, 0, 0};
}

std::vector<uint8_t> avc_seq_header(const std::vector<uint8_t>& sps, const std::vector<uint8_t>& pps) {
    auto s = strip_start(sps);
    auto p = strip_start(pps);
    std::vector<uint8_t> payload;
    payload.push_back(0x17);
    payload.push_back(0x00);
    payload.push_back(0);
    payload.push_back(0);
    payload.push_back(0);
    payload.push_back(0x01);
    payload.push_back(s.size() > 1 ? s[1] : 0x64);
    payload.push_back(s.size() > 2 ? s[2] : 0x00);
    payload.push_back(s.size() > 3 ? s[3] : 0x1F);
    payload.push_back(0xFF);
    payload.push_back(0xE1);
    payload.push_back((s.size() >> 8) & 0xFF);
    payload.push_back(s.size() & 0xFF);
    payload.insert(payload.end(), s.begin(), s.end());
    payload.push_back(0x01);
    payload.push_back((p.size() >> 8) & 0xFF);
    payload.push_back(p.size() & 0xFF);
    payload.insert(payload.end(), p.begin(), p.end());
    std::vector<uint8_t> tag;
    tag.push_back(9);
    be24(tag, (uint32_t)payload.size());
    be24(tag, 0);
    tag.push_back(0);
    tag.push_back(0);
    tag.push_back(0);
    tag.insert(tag.end(), payload.begin(), payload.end());
    be32(tag, (uint32_t)(tag.size()));
    return tag;
}

std::vector<uint8_t> annexb_to_flv_tag(const NAL& nal, uint32_t& dts, const std::vector<uint8_t>&, const std::vector<uint8_t>&, bool&) {
    auto raw = strip_start(nal.data);
    if (raw.empty()) return {};
    int nal_type = raw[0] & 0x1F;
    if (nal_type == 7 || nal_type == 8 || nal_type == 6 || nal_type == 9) return {};
    bool key = nal_type == 5;
    std::vector<uint8_t> payload;
    payload.push_back(key ? 0x17 : 0x27);
    payload.push_back(0x01);
    payload.push_back(0);
    payload.push_back(0);
    payload.push_back(0);
    be32(payload, (uint32_t)raw.size());
    payload.insert(payload.end(), raw.begin(), raw.end());
    dts += 40;
    std::vector<uint8_t> tag;
    tag.push_back(9);
    be24(tag, (uint32_t)payload.size());
    be24(tag, dts & 0xFFFFFF);
    tag.push_back((dts >> 24) & 0xFF);
    tag.push_back(0);
    tag.push_back(0);
    tag.push_back(0);
    tag.insert(tag.end(), payload.begin(), payload.end());
    be32(tag, (uint32_t)tag.size());
    return tag;
}
