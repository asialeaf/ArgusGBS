#pragma once
#include <cstdint>
#include <string>
#include <vector>
#include <mutex>
#include <deque>
#include <map>
#include <memory>
#include <atomic>
#include <thread>
#include <condition_variable>

struct NAL {
    bool key = false;
    bool h265 = false;
    std::vector<uint8_t> data; // Annex-B including start code
    uint32_t ts90 = 0;
};

struct AudioFrame {
    bool alaw = true;
    bool aac = false;
    int channels = 1;
    std::vector<uint8_t> asc; // 非空表示这帧带上新的 AudioSpecificConfig
    std::vector<uint8_t> data;
    uint32_t ts90 = 0;
};

struct StreamStats {
    std::string codec = "H264";
    int width = 0;
    int height = 0;
    int fps = 0;
    int64_t rtp_count = 0;
    int64_t rtp_lost = 0;
    int64_t in_bytes = 0;
    int in_bitrate = 0;
    int64_t video_frames = 0;
    int num_outputs = 0;
    std::string audio_codec;
    bool ready = false;
};

class PsDemux {
public:
    void push(const uint8_t* data, size_t len, uint32_t rtp_ts);
    std::vector<NAL> take();
    std::vector<AudioFrame> take_audio();
    void hint_video(const std::string& codec);
    std::string codec() const { return codec_; }
    std::string audio_codec() const { return audio_codec_; }
private:
    void feed();
    size_t pack_bytes() const;
    void parse_psm(const uint8_t* p, size_t n);
    void on_pes(uint8_t sid, const uint8_t* p, size_t n);
    void emit_nals(const uint8_t* es, size_t n, uint32_t pts, bool h265);
    void emit_aac(const uint8_t* es, size_t n, uint32_t pts);
    std::vector<uint8_t> buf_;
    std::vector<uint8_t> es_tail_;
    uint32_t es_tail_pts_ = 0;
    std::vector<uint8_t> aac_tail_;
    uint32_t aac_tail_pts_ = 0;
    std::vector<uint8_t> aac_asc_;
    std::vector<NAL> out_;
    std::vector<AudioFrame> audio_;
    std::map<uint8_t, std::string> sid_codec_;
    std::string codec_ = "H264";
    std::string audio_codec_;
    uint32_t ts_ = 0;
};

// 直播时间戳：把绝对 PTS 收成从 0 递增，跳变超过 300ms 时沿用上一增量。
// 与 ZLMediaKit Stamp / DeltaStamp 的直播策略一致。
struct LiveStamp {
    bool seen = false;
    int64_t last_in = 0;
    int64_t out = 0;
    int64_t last_delta = 40;
    uint32_t map(uint32_t pts90) {
        int64_t in = (int64_t)pts90 / 90;
        if (!seen) {
            seen = true;
            last_in = in;
            out = 0;
            return 0;
        }
        if (in == last_in) return (uint32_t)out;
        int64_t d = in - last_in;
        last_in = in;
        if (d <= 0 || d > 300) d = last_delta;
        else last_delta = d;
        out += d;
        if (out < 0) out = 0;
        return (uint32_t)out;
    }
};

inline uint32_t flv_tag_ts(const uint8_t* p) {
    return ((uint32_t)p[4] << 16) | ((uint32_t)p[5] << 8) | p[6] | ((uint32_t)p[7] << 24);
}

inline void flv_set_ts(uint8_t* p, uint32_t ts) {
    p[4] = (uint8_t)((ts >> 16) & 0xFF);
    p[5] = (uint8_t)((ts >> 8) & 0xFF);
    p[6] = (uint8_t)(ts & 0xFF);
    p[7] = (uint8_t)((ts >> 24) & 0xFF);
}

inline bool flv_is_seq(const uint8_t* p, size_t n) {
    if (n < 13) return false;
    if (p[0] == 9 && p[12] == 0x00) return true;
    if (p[0] == 8 && (p[11] >> 4) == 10 && p[12] == 0x00) return true;
    return false;
}

inline int64_t flv_first_media_ts(const uint8_t* p, size_t n) {
    size_t i = 0;
    while (i + 11 <= n) {
        uint32_t size = ((uint32_t)p[i + 1] << 16) | ((uint32_t)p[i + 2] << 8) | p[i + 3];
        if (i + 11 + size + 4 > n) break;
        if (!flv_is_seq(p + i, 11 + size)) return flv_tag_ts(p + i);
        i += 11 + size + 4;
    }
    return 0;
}

inline void flv_shift_tags(uint8_t* p, size_t n, int64_t base) {
    size_t i = 0;
    while (i + 11 <= n) {
        uint32_t size = ((uint32_t)p[i + 1] << 16) | ((uint32_t)p[i + 2] << 8) | p[i + 3];
        if (i + 11 + size + 4 > n) break;
        if (flv_is_seq(p + i, 11 + size)) flv_set_ts(p + i, 0);
        else {
            int64_t nts = (int64_t)flv_tag_ts(p + i) - base;
            if (nts < 0) nts = 0;
            flv_set_ts(p + i, (uint32_t)nts);
        }
        i += 11 + size + 4;
    }
}

struct Subscriber {
    int fd = -1;
    bool websocket = false;
    bool sent_header = false;
    int64_t stamp_base = 0;
};

struct RtpSink {
    int fd = -1;
    bool tcp = false;
    bool udp = false;
    std::string ip;
    int port = 0;
};

struct HlsSeg {
    int seq = 0;
    double dur = 1;
    std::vector<uint8_t> ts;
};

struct RtspOut {
    int tcp = -1;
    int udp = -1;
};

class Session {
public:
    explicit Session(std::string id);
    ~Session();
    std::string id;
    std::string transport = "UDP";
    std::string mode = "passive";
    std::string ssrc;
    int port = 0;
    int udp_fd = -1;
    int tcp_fd = -1;
    std::atomic<bool> running{true};
    std::thread worker;
    PsDemux demux;
    std::mutex mu;
    std::vector<NAL> gop;
    std::vector<uint8_t> sps, pps, vps;
    std::vector<uint8_t> aac_asc;
    int aac_channels = 1;
    LiveStamp video_stamp, audio_stamp;
    std::deque<std::vector<uint8_t>> flv_tags;
    std::vector<std::shared_ptr<Subscriber>> subs;
    StreamStats stats;
    int64_t last_rtp_ms = 0;
    uint16_t last_seq = 0;
    bool has_seq = false;
    std::map<uint16_t, std::vector<uint8_t>> rtp_buf;
    std::vector<uint8_t> asm_buf;
    uint32_t asm_stamp = 0;
    uint16_t asm_seq = 0;
    uint8_t asm_pt = 0;
    bool asm_has = false;
    bool asm_drop = false;
    bool asm_seq_ok = false;
    std::vector<uint8_t> fu;
    bool fu_open = false;
    std::map<uint16_t, std::vector<uint8_t>> ts_pes;
    std::map<uint16_t, std::string> ts_type;
    uint16_t ts_pmt = 0xFFFF;
    void add_nals(const std::vector<NAL>& nals);
    std::vector<uint8_t> flv_header_and_gop();
    void broadcast(const std::vector<uint8_t>& tag);

    std::mutex dist_mu;
    std::vector<RtpSink> sinks;
    int relay_listen = -1;
    std::atomic<bool> recv_started{false};
    std::deque<HlsSeg> hls;
    std::vector<uint8_t> hls_cur;
    int64_t hls_begin = 0;
    int hls_seq = 0;
    uint32_t hls_pts = 0;
    uint8_t cc_pat = 0, cc_pmt = 0, cc_vid = 0;
    bool hls_armed = false;
    std::vector<RtspOut> rtsp;
    uint16_t rtsp_seq = 1;
    uint32_t play_ts = 0;
    std::vector<int> rtmp;
    std::vector<std::shared_ptr<void>> rtc;
    std::vector<NAL> rtc_gop;

    bool start_recv_tcp(std::shared_ptr<Session> self, const std::string& ip, int port);
    int start_send_relay(std::shared_ptr<Session> self, const std::string& transport, const std::string& mode, const std::string& ip, int port, int listen_port);
    void forward_rtp(const uint8_t* pkt, size_t n);
    void playout(const std::vector<NAL>& nals);
    void fanout_flv(const std::vector<uint8_t>& tag);
    std::string hls_m3u8();
    bool hls_seg(int seq, std::vector<uint8_t>& out);
    void attach_rtsp(int tcp_fd, int udp_fd);
    void attach_rtmp(int fd);
    bool release_rtsp(int tcp_fd, int udp_fd);
    bool release_rtmp(int fd);
};

class Hub {
public:
    int http_port = 10001;
    int rtsp_port = 554;
    int rtmp_port = 1935;
    std::string secret = "argus-sms";
    std::string advertise_ip;
    int udp_min = 30000, udp_max = 30249;
    int rtc_lo = 30250, rtc_hi = 30500;
    void start();
    void start_extra();
    std::shared_ptr<Session> open(const std::string& id, const std::string& transport, const std::string& mode, const std::string& ssrc, const std::string& peer_ip, int peer_port);
    void close(const std::string& id);
    std::shared_ptr<Session> get(const std::string& id);
    bool relay(const std::string& id, const std::string& transport, const std::string& mode, const std::string& ip, int port, const std::string& direction, int& listen_port);
    std::string public_ip() const;
    int grab_media_port();
    int grab_rtc_port();
    bool whep(const std::string& id, const std::string& offer, bool as_json, int& http_code, std::string& ctype, std::string& body);
private:
    std::mutex mu_;
    std::map<std::string, std::shared_ptr<Session>> sessions_;
    int next_port_ = 30000;
    int next_rtc_ = 30250;
};

std::string host_ip();
void sms_tcp_loop(std::shared_ptr<Session> s);
void rtc_send(Session* s, const std::vector<NAL>& nals, uint32_t ts90);
void rtc_stop(const std::shared_ptr<void>& player);
void ws_send(int fd, const uint8_t* data, size_t n);

std::vector<uint8_t> annexb_to_flv_tag(const NAL& nal, uint32_t& dts, const std::vector<uint8_t>& sps, const std::vector<uint8_t>& pps, bool& sent_seq);
std::vector<uint8_t> flv_file_header(bool with_audio);
std::vector<uint8_t> avc_seq_header(const std::vector<uint8_t>& sps, const std::vector<uint8_t>& pps);
std::vector<uint8_t> hevc_seq_header(const std::vector<uint8_t>& vps, const std::vector<uint8_t>& sps, const std::vector<uint8_t>& pps);
std::vector<uint8_t> g711_flv_tag(const uint8_t* data, size_t len, uint32_t dts_ms, bool alaw);
std::vector<uint8_t> aac_flv_tag(const uint8_t* data, size_t len, uint32_t dts_ms, bool sequence, int channels);
std::string json_escape(const std::string& s);
