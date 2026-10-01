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
    std::vector<uint8_t> data; // Annex-B including start code
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
    std::string codec() const { return codec_; }
private:
    void feed();
    void on_pes(const uint8_t* p, size_t n, uint32_t ts);
    std::vector<uint8_t> buf_;
    std::vector<NAL> out_;
    std::string codec_ = "H264";
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
std::string json_escape(const std::string& s);
