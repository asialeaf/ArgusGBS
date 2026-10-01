#include "Common/sms.h"
#include <cstring>

// MPEG-PS 按包长解析，做法与 ZLMediaKit 的 PSDecoder 一致：
// 包头、PSM、PES 都用长度字段切包，只在 PES 载荷内部再切 Annex-B。

static int start_at(const uint8_t* p, size_t n) {
    for (size_t i = 0; i + 3 < n; i++) {
        if (p[i] == 0 && p[i + 1] == 0 && p[i + 2] == 1) return (int)i;
    }
    return -1;
}

static uint32_t read_pts(const uint8_t* p) {
    uint32_t v = ((uint32_t)(p[0] & 0x0E) << 29)
        | ((uint32_t)p[1] << 22)
        | ((uint32_t)(p[2] & 0xFE) << 14)
        | ((uint32_t)p[3] << 7)
        | ((uint32_t)(p[4] & 0xFE) >> 1);
    return v;
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

std::vector<AudioFrame> PsDemux::take_audio() {
    std::vector<AudioFrame> o;
    o.swap(audio_);
    return o;
}

void PsDemux::hint_video(const std::string& codec) {
    if (codec == "H264" || codec == "H265") codec_ = codec;
}

size_t PsDemux::pack_bytes() const {
    if (buf_.size() < 14) return 0;
    if ((buf_[4] & 0xC0) == 0x40) {
        size_t stuff = buf_[13] & 0x07;
        if (buf_.size() < 14 + stuff) return 0;
        return 14 + stuff;
    }
    return buf_.size() >= 12 ? 12 : 0;
}

void PsDemux::parse_psm(const uint8_t* p, size_t n) {
    if (n < 6) return;
    uint16_t info_len = (p[2] << 8) | p[3];
    size_t es_off = 4u + info_len;
    if (es_off + 2 > n) return;
    uint16_t es_len = (p[es_off] << 8) | p[es_off + 1];
    size_t i = es_off + 2;
    size_t end = i + es_len;
    if (end > n) end = n;
    while (i + 4 <= end) {
        uint8_t st = p[i];
        uint8_t esid = p[i + 1];
        uint16_t ilen = (p[i + 2] << 8) | p[i + 3];
        std::string c;
        if (st == 0x1B) c = "H264";
        else if (st == 0x24) c = "H265";
        else if (st == 0x0F || st == 0x11) c = "AAC";
        else if (st == 0x90) c = "G711A";
        else if (st == 0x91) c = "G711U";
        if (!c.empty()) sid_codec_[esid] = c;
        if (c == "H264" || c == "H265") codec_ = c;
        else if (!c.empty() && audio_codec_.empty()) audio_codec_ = c;
        if (i + 4 + ilen > end) break;
        i += 4 + ilen;
    }
}

void PsDemux::emit_nals(const uint8_t* es, size_t n, uint32_t pts, bool h265) {
    size_t i = 0;
    while (i + 3 < n) {
        int s = start_at(es + i, n - i);
        if (s < 0) break;
        size_t sc = i + s;
        size_t hdr = sc + 3;
        if (hdr >= n) break;
        int nxt = start_at(es + hdr, n - hdr);
        size_t end = nxt < 0 ? n : hdr + nxt;
        if (nxt >= 0 && end > 0 && es[end - 1] == 0) end--;
        size_t begin = sc;
        if (begin > 0 && es[begin - 1] == 0) begin--;
        uint8_t nalh = es[hdr];
        bool key = false;
        if (h265) {
            int t = (nalh >> 1) & 0x3F;
            key = t == 19 || t == 20 || t == 21 || t == 32 || t == 33 || t == 34;
        } else {
            int t = nalh & 0x1F;
            key = t == 5 || t == 7 || t == 8;
        }
        NAL one;
        one.key = key;
        one.h265 = h265;
        one.ts90 = pts;
        one.data.assign(es + begin, es + end);
        out_.push_back(std::move(one));
        i = end;
    }
}

void PsDemux::on_pes(uint8_t sid, const uint8_t* p, size_t n) {
    if (n < 3 || (p[0] & 0xC0) != 0x80) return;
    uint8_t flags = p[1];
    uint8_t hlen = p[2];
    if (3u + hlen > n) return;
    uint32_t pts = ts_;
    if ((flags & 0x80) && hlen >= 5) pts = read_pts(p + 3);
    auto it = sid_codec_.find(sid);
    std::string c = it == sid_codec_.end() ? "" : it->second;
    bool g711 = c == "G711A" || c == "G711U" || (c.empty() && sid >= 0xC0 && sid <= 0xDF);
    if (g711) {
        bool alaw = c != "G711U";
        audio_codec_ = alaw ? "G711A" : "G711U";
        AudioFrame a;
        a.alaw = alaw;
        a.ts90 = pts;
        a.data.assign(p + 3 + hlen, p + n);
        if (!a.data.empty()) audio_.push_back(std::move(a));
        return;
    }
    if (c == "AAC" || sid == 0xBD) {
        if (audio_codec_.empty()) audio_codec_ = c.empty() ? "private" : c;
        if (c == "H264" || c == "H265") emit_nals(p + 3 + hlen, n - 3 - hlen, pts, c == "H265");
        return;
    }
    bool h265 = c == "H265" || (c.empty() && codec_ == "H265");
    if (c == "H264") codec_ = "H264";
    if (c == "H265") codec_ = "H265";
    if (c.empty() && sid >= 0xE0 && sid <= 0xEF) h265 = codec_ == "H265";
    if (sid >= 0xC0 && sid <= 0xDF && c.empty()) return;
    emit_nals(p + 3 + hlen, n - 3 - hlen, pts, h265);
}

void PsDemux::feed() {
    for (;;) {
        if (buf_.size() < 4) return;
        int at = start_at(buf_.data(), buf_.size());
        if (at < 0) {
            if (buf_.size() > 3) buf_.erase(buf_.begin(), buf_.end() - 3);
            return;
        }
        if (at > 0) buf_.erase(buf_.begin(), buf_.begin() + at);
        if (buf_.size() < 4) return;
        uint8_t sid = buf_[3];
        if (sid == 0xBA) {
            size_t n = pack_bytes();
            if (n == 0) return;
            buf_.erase(buf_.begin(), buf_.begin() + n);
            continue;
        }
        if (sid == 0xB9) {
            buf_.erase(buf_.begin(), buf_.begin() + 4);
            continue;
        }
        bool sized = sid == 0xBB || sid == 0xBC || sid == 0xBD || sid == 0xBE || sid == 0xBF
            || (sid >= 0xC0 && sid <= 0xEF);
        if (!sized) {
            buf_.erase(buf_.begin(), buf_.begin() + 3);
            continue;
        }
        if (buf_.size() < 6) return;
        uint16_t plen = (buf_[4] << 8) | buf_[5];
        if (plen == 0) {
            buf_.erase(buf_.begin(), buf_.begin() + 4);
            continue;
        }
        if (buf_.size() < 6u + plen) return;
        if (sid == 0xBC) parse_psm(&buf_[6], plen);
        else if ((sid >= 0xC0 && sid <= 0xEF) || sid == 0xBD) on_pes(sid, &buf_[6], plen);
        buf_.erase(buf_.begin(), buf_.begin() + 6 + plen);
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
    return {'F', 'L', 'V', 0x01, 0x05, 0, 0, 0, 0x09, 0, 0, 0, 0};
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

static void push_nal_array(std::vector<uint8_t>& o, uint8_t nal_type, const std::vector<uint8_t>& nal) {
    o.push_back(0x80 | nal_type);
    o.push_back(0);
    o.push_back(1);
    o.push_back((nal.size() >> 8) & 0xFF);
    o.push_back(nal.size() & 0xFF);
    o.insert(o.end(), nal.begin(), nal.end());
}

std::vector<uint8_t> hevc_seq_header(const std::vector<uint8_t>& vps, const std::vector<uint8_t>& sps, const std::vector<uint8_t>& pps) {
    auto v = strip_start(vps);
    auto s = strip_start(sps);
    auto p = strip_start(pps);
    std::vector<uint8_t> payload;
    payload.push_back(0x1C);
    payload.push_back(0x00);
    payload.push_back(0);
    payload.push_back(0);
    payload.push_back(0);
    payload.push_back(1);
    payload.push_back(s.size() > 1 ? s[1] : 0);
    payload.push_back(s.size() > 2 ? s[2] : 0);
    payload.push_back(s.size() > 3 ? s[3] : 0);
    payload.insert(payload.end(), 4, 0);
    payload.insert(payload.end(), 6, 0);
    payload.push_back(0);
    payload.push_back(0xF0);
    payload.push_back(0x00);
    payload.push_back(0xFC);
    payload.push_back(0xFD);
    payload.push_back(0xF8);
    payload.push_back(0xF8);
    payload.push_back(0);
    payload.push_back(0);
    payload.push_back(0x0F);
    payload.push_back(3);
    push_nal_array(payload, 32, v);
    push_nal_array(payload, 33, s);
    push_nal_array(payload, 34, p);
    std::vector<uint8_t> tag;
    tag.push_back(9);
    be24(tag, (uint32_t)payload.size());
    be24(tag, 0);
    tag.push_back(0);
    tag.push_back(0);
    tag.push_back(0);
    tag.insert(tag.end(), payload.begin(), payload.end());
    be32(tag, (uint32_t)tag.size());
    return tag;
}

std::vector<uint8_t> g711_flv_tag(const uint8_t* data, size_t len, uint32_t dts_ms, bool alaw) {
    if (!data || len == 0) return {};
    uint8_t flags = (uint8_t)(((alaw ? 7 : 8) << 4) | (1 << 1));
    std::vector<uint8_t> payload;
    payload.push_back(flags);
    payload.insert(payload.end(), data, data + len);
    std::vector<uint8_t> tag;
    tag.push_back(8);
    be24(tag, (uint32_t)payload.size());
    be24(tag, dts_ms & 0xFFFFFF);
    tag.push_back((dts_ms >> 24) & 0xFF);
    tag.push_back(0);
    tag.push_back(0);
    tag.push_back(0);
    tag.insert(tag.end(), payload.begin(), payload.end());
    be32(tag, (uint32_t)tag.size());
    return tag;
}

std::vector<uint8_t> annexb_to_flv_tag(const NAL& nal, uint32_t& dts, const std::vector<uint8_t>&, const std::vector<uint8_t>&, bool&) {
    auto raw = strip_start(nal.data);
    if (raw.empty()) return {};
    if (nal.h265) {
        int t = (raw[0] >> 1) & 0x3F;
        if (t == 32 || t == 33 || t == 34 || t == 35 || t == 39) return {};
        bool key = t == 19 || t == 20 || t == 21;
        std::vector<uint8_t> payload;
        payload.push_back(key ? 0x1C : 0x2C);
        payload.push_back(0x01);
        payload.push_back(0);
        payload.push_back(0);
        payload.push_back(0);
        be32(payload, (uint32_t)raw.size());
        payload.insert(payload.end(), raw.begin(), raw.end());
        if (nal.ts90) dts = nal.ts90 / 90;
        else dts += 40;
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
    if (nal.ts90) dts = nal.ts90 / 90;
    else dts += 40;
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
