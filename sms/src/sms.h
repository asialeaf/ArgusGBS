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
    std::vector<uint8_t> buf_;
    std::vector<NAL> out_;
    std::vector<AudioFrame> audio_;
    std::map<uint8_t, std::string> sid_codec_;
    std::string codec_ = "H264";
    std::string audio_codec_;
    uint32_t ts_ = 0;
};

struct Subscriber {
    int fd = -1;
    bool websocket = false;
    bool sent_header = false;
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
};

class Hub {
public:
    int http_port = 10001;
    int rtsp_port = 554;
    int rtmp_port = 1935;
    std::string secret = "argus-sms";
    std::string advertise_ip;
    int udp_min = 30000, udp_max = 30249;
    void start();
    std::shared_ptr<Session> open(const std::string& id, const std::string& transport, const std::string& mode, const std::string& ssrc, const std::string& peer_ip, int peer_port);
    void close(const std::string& id);
    std::shared_ptr<Session> get(const std::string& id);
private:
    std::mutex mu_;
    std::map<std::string, std::shared_ptr<Session>> sessions_;
    int next_port_ = 30000;
};

std::vector<uint8_t> annexb_to_flv_tag(const NAL& nal, uint32_t& dts, const std::vector<uint8_t>& sps, const std::vector<uint8_t>& pps, bool& sent_seq);
std::vector<uint8_t> flv_file_header();
std::vector<uint8_t> avc_seq_header(const std::vector<uint8_t>& sps, const std::vector<uint8_t>& pps);
std::vector<uint8_t> hevc_seq_header(const std::vector<uint8_t>& vps, const std::vector<uint8_t>& sps, const std::vector<uint8_t>& pps);
std::vector<uint8_t> g711_flv_tag(const uint8_t* data, size_t len, uint32_t dts_ms, bool alaw);
std::string json_escape(const std::string& s);
