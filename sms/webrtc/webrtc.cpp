#include "Common/sms.h"

#include <arpa/inet.h>
#include <netinet/in.h>
#include <sys/select.h>
#include <sys/socket.h>
#include <unistd.h>

#include <openssl/evp.h>
#include <openssl/hmac.h>
#include <openssl/ssl.h>
#include <openssl/srtp.h>
#include <openssl/x509.h>
#include <srtp2/srtp.h>

#include <openssl/sha.h>

#include <algorithm>
#include <cstdio>
#include <atomic>
#include <cstring>
#include <iostream>
#include <mutex>
#include <sstream>
#include <thread>
#include <vector>

static uint32_t crc32_ieee(const uint8_t* data, size_t n) {
    uint32_t crc = 0xFFFFFFFF;
    for (size_t i = 0; i < n; ++i) {
        crc ^= data[i];
        for (int b = 0; b < 8; ++b) crc = (crc & 1) ? (crc >> 1) ^ 0xEDB88320u : (crc >> 1);
    }
    return ~crc;
}

static std::string b64(const uint8_t* p, size_t n) {
    static const char* t = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
    std::string o;
    for (size_t i = 0; i < n; i += 3) {
        uint32_t v = (uint32_t)p[i] << 16;
        if (i + 1 < n) v |= (uint32_t)p[i + 1] << 8;
        if (i + 2 < n) v |= p[i + 2];
        o.push_back(t[(v >> 18) & 63]);
        o.push_back(t[(v >> 12) & 63]);
        o.push_back(i + 1 < n ? t[(v >> 6) & 63] : '=');
        o.push_back(i + 2 < n ? t[v & 63] : '=');
    }
    return o;
}

static std::string rand_token(int n) {
    std::string s(n, 'a');
    FILE* f = fopen("/dev/urandom", "rb");
    if (!f) return s;
    static const char* alphabet = "abcdefghijklmnopqrstuvwxyz0123456789";
    for (int i = 0; i < n; ++i) {
        unsigned char b = 0;
        if (fread(&b, 1, 1, f) != 1) break;
        s[i] = alphabet[b % 36];
    }
    fclose(f);
    return s;
}

struct DtlsCert {
    SSL_CTX* ctx = nullptr;
    std::string fingerprint;
};

static DtlsCert& dtls_cert() {
    static DtlsCert c;
    static std::once_flag once;
    std::call_once(once, [] {
        EVP_PKEY_CTX* pctx = EVP_PKEY_CTX_new_id(EVP_PKEY_EC, nullptr);
        EVP_PKEY_keygen_init(pctx);
        EVP_PKEY_CTX_set_ec_paramgen_curve_nid(pctx, NID_X9_62_prime256v1);
        EVP_PKEY* pkey = nullptr;
        EVP_PKEY_keygen(pctx, &pkey);
        EVP_PKEY_CTX_free(pctx);
        X509* x = X509_new();
        X509_set_version(x, 2);
        ASN1_INTEGER_set(X509_get_serialNumber(x), 1);
        X509_gmtime_adj(X509_getm_notBefore(x), 0);
        X509_gmtime_adj(X509_getm_notAfter(x), 60 * 60 * 24 * 365);
        X509_set_pubkey(x, pkey);
        X509_NAME* name = X509_get_subject_name(x);
        X509_NAME_add_entry_by_txt(name, "CN", MBSTRING_ASC, (const unsigned char*)"argussms", -1, -1, 0);
        X509_set_issuer_name(x, name);
        X509_sign(x, pkey, EVP_sha256());
        unsigned char* der = nullptr;
        int der_len = i2d_X509(x, &der);
        unsigned char md[32];
        SHA256(der, der_len, md);
        OPENSSL_free(der);
        std::ostringstream fp;
        for (int i = 0; i < 32; ++i) {
            if (i) fp << ":";
            fp.width(2);
            fp.fill('0');
            fp << std::hex << (int)md[i];
        }
        c.fingerprint = fp.str();
        c.ctx = SSL_CTX_new(DTLS_server_method());
        SSL_CTX_use_certificate(c.ctx, x);
        SSL_CTX_use_PrivateKey(c.ctx, pkey);
        SSL_CTX_set_tlsext_use_srtp(c.ctx, "SRTP_AES128_CM_SHA1_80");
        SSL_CTX_set_read_ahead(c.ctx, 1);
        SSL_CTX_set_min_proto_version(c.ctx, DTLS1_2_VERSION);
        SSL_CTX_set_max_proto_version(c.ctx, DTLS1_2_VERSION);
        X509_free(x);
        EVP_PKEY_free(pkey);
    });
    return c;
}

static void ensure_srtp() {
    static std::once_flag once;
    std::call_once(once, [] { srtp_init(); });
}

struct WebRtcPlayer {
    int fd = -1;
    int port = 0;
    std::string ufrag;
    std::string pwd;
    std::atomic<bool> run{true};
    std::thread th;
    std::mutex mu;
    sockaddr_in peer{};
    bool have_peer = false;
    bool we_server = true;
    SSL* ssl = nullptr;
    BIO* rbio = nullptr;
    BIO* wbio = nullptr;
    bool dtls_done = false;
    std::atomic<bool> keys{false};
    std::atomic<bool> need_key{true};
    srtp_t srtp = nullptr;
    uint8_t master_key[30]{};
    uint16_t seq = 1;
    bool srtp_err_logged = false;

    ~WebRtcPlayer() {
        run = false;
        if (th.joinable()) {
            if (th.get_id() == std::this_thread::get_id()) th.detach();
            else {
                if (fd >= 0) ::shutdown(fd, SHUT_RDWR);
                th.join();
            }
        }
        if (fd >= 0) {
            ::close(fd);
            fd = -1;
        }
        if (ssl) SSL_free(ssl);
        if (srtp) srtp_dealloc(srtp);
    }

    void ensure_ssl() {
        if (ssl) return;
        ssl = SSL_new(dtls_cert().ctx);
        rbio = BIO_new(BIO_s_mem());
        wbio = BIO_new(BIO_s_mem());
        BIO_set_mem_eof_return(rbio, -1);
        BIO_set_mem_eof_return(wbio, -1);
        SSL_set_bio(ssl, rbio, wbio);
        SSL_set_mtu(ssl, 1200);
        SSL_set_options(ssl, SSL_OP_NO_QUERY_MTU);
        if (we_server) SSL_set_accept_state(ssl);
        else SSL_set_connect_state(ssl);
    }

    void flush_dtls() {
        if (!ssl || !have_peer) return;
        int rc = SSL_do_handshake(ssl);
        uint8_t out[2000];
        while (BIO_ctrl_pending(wbio) > 0) {
            int n = BIO_read(wbio, out, sizeof(out));
            if (n <= 0) break;
            sendto(fd, out, n, MSG_NOSIGNAL, (sockaddr*)&peer, sizeof(peer));
        }
        if (rc == 1 && !keys) export_keys();
    }

    void export_keys() {
        unsigned char material[60];
        if (SSL_export_keying_material(ssl, material, 60, "EXTRACTOR-dtls_srtp", 19, nullptr, 0, 0) != 1) return;
        const uint8_t* master = SSL_is_server(ssl) ? material + 16 : material;
        const uint8_t* salt = SSL_is_server(ssl) ? material + 46 : material + 32;
        memcpy(master_key, master, 16);
        memcpy(master_key + 16, salt, 14);
        ensure_srtp();
        srtp_policy_t policy;
        memset(&policy, 0, sizeof(policy));
        srtp_crypto_policy_set_aes_cm_128_hmac_sha1_80(&policy.rtp);
        srtp_crypto_policy_set_aes_cm_128_hmac_sha1_80(&policy.rtcp);
        policy.ssrc.type = ssrc_any_outbound;
        policy.ssrc.value = 0;
        policy.key = master_key;
        policy.allow_repeat_tx = 1;
        policy.window_size = 1024;
        policy.next = nullptr;
        if (srtp_create(&srtp, &policy) != srtp_err_status_ok) {
            std::cerr << "WebRTC srtp_create 失败" << std::endl;
            return;
        }
        keys = true;
        dtls_done = true;
        need_key = true;
        std::cerr << "WebRTC DTLS 完成，开始发 SRTP" << std::endl;
    }

    void on_stun(const uint8_t* p, size_t n, const sockaddr_in& from) {
        if (n < 20) return;
        uint16_t typ = (p[0] << 8) | p[1];
        if (typ != 0x0001) return;
        std::string user;
        size_t i = 20;
        while (i + 4 <= n) {
            uint16_t at = (p[i] << 8) | p[i + 1];
            uint16_t al = (p[i + 2] << 8) | p[i + 3];
            if (i + 4 + al > n) break;
            if (at == 0x0006) user.assign((const char*)p + i + 4, al);
            i += 4 + al;
            if (al % 4) i += 4 - (al % 4);
        }
        if (!user.empty() && user.compare(0, ufrag.size(), ufrag) != 0) return;
        {
            std::lock_guard<std::mutex> lk(mu);
            peer = from;
            have_peer = true;
        }
        std::vector<uint8_t> msg(20);
        msg[0] = 0x01;
        msg[1] = 0x01;
        msg[4] = 0x21;
        msg[5] = 0x12;
        msg[6] = 0xA4;
        msg[7] = 0x42;
        memcpy(msg.data() + 8, p + 8, 12);
        uint16_t xport = ntohs(from.sin_port) ^ 0x2112;
        uint32_t xip = ntohl(from.sin_addr.s_addr) ^ 0x2112A442;
        uint8_t addr[8] = {0, 1, (uint8_t)(xport >> 8), (uint8_t)xport,
                           (uint8_t)(xip >> 24), (uint8_t)(xip >> 16), (uint8_t)(xip >> 8), (uint8_t)xip};
        msg.push_back(0x00);
        msg.push_back(0x20);
        msg.push_back(0x00);
        msg.push_back(0x08);
        msg.insert(msg.end(), addr, addr + 8);
        uint16_t with_mi = (uint16_t)(msg.size() - 20 + 24);
        msg[2] = (uint8_t)(with_mi >> 8);
        msg[3] = (uint8_t)with_mi;
        unsigned char mac[20];
        unsigned int maclen = 20;
        HMAC(EVP_sha1(), pwd.data(), (int)pwd.size(), msg.data(), msg.size(), mac, &maclen);
        msg.push_back(0x00);
        msg.push_back(0x08);
        msg.push_back(0x00);
        msg.push_back(0x14);
        msg.insert(msg.end(), mac, mac + 20);
        uint16_t with_fp = (uint16_t)(with_mi + 8);
        msg[2] = (uint8_t)(with_fp >> 8);
        msg[3] = (uint8_t)with_fp;
        uint32_t crc = crc32_ieee(msg.data(), msg.size()) ^ 0x5354554e;
        msg.push_back(0x80);
        msg.push_back(0x28);
        msg.push_back(0x00);
        msg.push_back(0x04);
        msg.push_back((uint8_t)(crc >> 24));
        msg.push_back((uint8_t)(crc >> 16));
        msg.push_back((uint8_t)(crc >> 8));
        msg.push_back((uint8_t)crc);
        sendto(fd, msg.data(), msg.size(), MSG_NOSIGNAL, (sockaddr*)&from, sizeof(from));
        if (!we_server) flush_dtls();
    }

    void on_dtls(const uint8_t* p, size_t n, const sockaddr_in& from) {
        std::lock_guard<std::mutex> lk(mu);
        peer = from;
        have_peer = true;
        ensure_ssl();
        BIO_write(rbio, p, (int)n);
        flush_dtls();
    }

    void loop() {
        while (run && fd >= 0) {
            fd_set fds;
            FD_ZERO(&fds);
            FD_SET(fd, &fds);
            timeval tv{0, 200000};
            int rc = select(fd + 1, &fds, nullptr, nullptr, &tv);
            if (!run) break;
            if (rc > 0) {
                uint8_t buf[2048];
                sockaddr_in from{};
                socklen_t fl = sizeof(from);
                ssize_t n = recvfrom(fd, buf, sizeof(buf), 0, (sockaddr*)&from, &fl);
                if (n <= 0) continue;
                if (buf[0] < 2) on_stun(buf, (size_t)n, from);
                else if (buf[0] >= 20 && buf[0] <= 63) on_dtls(buf, (size_t)n, from);
            } else if (ssl && !dtls_done && have_peer) {
                std::lock_guard<std::mutex> lk(mu);
                flush_dtls();
            }
        }
    }

    void srtp_send(const uint8_t* rtp, size_t n) {
        if (!keys || !srtp || n < 12 || fd < 0) return;
        std::lock_guard<std::mutex> lk(mu);
        if (!have_peer) return;
        std::vector<uint8_t> pkt(n + SRTP_MAX_TRAILER_LEN);
        memcpy(pkt.data(), rtp, n);
        uint16_t s = seq++;
        pkt[2] = (uint8_t)(s >> 8);
        pkt[3] = (uint8_t)s;
        int len = (int)n;
        srtp_err_status_t err = srtp_protect(srtp, pkt.data(), &len);
        if (err != srtp_err_status_ok) {
            if (!srtp_err_logged) {
                srtp_err_logged = true;
                std::cerr << "WebRTC srtp_protect 失败 " << (int)err << std::endl;
            }
            return;
        }
        sendto(fd, pkt.data(), len, MSG_NOSIGNAL, (sockaddr*)&peer, sizeof(peer));
    }
};

static const uint8_t* nal_bytes(const NAL& n, size_t& len) {
    const uint8_t* p = n.data.data();
    size_t nlen = n.data.size();
    if (nlen >= 4 && p[0] == 0 && p[1] == 0 && p[2] == 0 && p[3] == 1) {
        len = nlen - 4;
        return p + 4;
    }
    if (nlen >= 3 && p[0] == 0 && p[1] == 0 && p[2] == 1) {
        len = nlen - 3;
        return p + 3;
    }
    len = nlen;
    return p;
}

static void put_rtp(uint8_t* h, uint16_t seq, uint32_t ts, uint32_t ssrc, bool marker) {
    h[0] = 0x80;
    h[1] = (uint8_t)((marker ? 0x80 : 0) | 96);
    h[2] = (uint8_t)(seq >> 8);
    h[3] = (uint8_t)seq;
    h[4] = (uint8_t)(ts >> 24);
    h[5] = (uint8_t)(ts >> 16);
    h[6] = (uint8_t)(ts >> 8);
    h[7] = (uint8_t)ts;
    h[8] = (uint8_t)(ssrc >> 24);
    h[9] = (uint8_t)(ssrc >> 16);
    h[10] = (uint8_t)(ssrc >> 8);
    h[11] = (uint8_t)ssrc;
}

void rtc_stop(const std::shared_ptr<void>& player) {
    if (!player) return;
    auto p = std::static_pointer_cast<WebRtcPlayer>(player);
    p->run = false;
    if (p->fd >= 0) shutdown(p->fd, SHUT_RDWR);
}

static int h264_type(const NAL& n) {
    size_t len = 0;
    const uint8_t* nal = nal_bytes(n, len);
    if (!nal || len == 0 || n.h265) return -1;
    return nal[0] & 0x1F;
}

static std::vector<std::vector<uint8_t>> packetize_nals(const std::vector<NAL>& nals, uint32_t ts90, uint32_t ssrc) {
    std::vector<std::vector<uint8_t>> all;
    uint16_t seq = 1;
    for (size_t i = 0; i < nals.size(); ++i) {
        if (nals[i].h265) continue;
        size_t len = 0;
        const uint8_t* nal = nal_bytes(nals[i], len);
        if (!nal || len == 0) continue;
        bool marker = i + 1 == nals.size();
        if (len <= 1100) {
            std::vector<uint8_t> pkt(12 + len);
            put_rtp(pkt.data(), seq++, ts90, ssrc, marker);
            memcpy(pkt.data() + 12, nal, len);
            all.push_back(std::move(pkt));
        } else {
            uint8_t nalh = nal[0];
            size_t off = 1;
            while (off < len) {
                size_t chunk = std::min((size_t)1100, len - off);
                bool start = off == 1;
                bool end = off + chunk >= len;
                std::vector<uint8_t> pkt(14 + chunk);
                put_rtp(pkt.data(), seq++, ts90, ssrc, end && marker);
                pkt[12] = (uint8_t)((nalh & 0xE0) | 28);
                pkt[13] = (uint8_t)((start ? 0x80 : 0) | (end ? 0x40 : 0) | (nalh & 0x1F));
                memcpy(pkt.data() + 14, nal + off, chunk);
                all.push_back(std::move(pkt));
                off += chunk;
            }
        }
    }
    return all;
}

void rtc_send(Session* s, const std::vector<NAL>& nals, uint32_t ts90) {
    bool idr = false;
    bool has_sps = false;
    for (auto& n : nals) {
        int t = h264_type(n);
        if (t == 5) idr = true;
        if (t == 7) has_sps = true;
    }
    std::vector<NAL> framed = nals;
    if (idr && !has_sps && !s->sps.empty() && !s->pps.empty()) {
        NAL sps;
        NAL pps;
        sps.data = {0, 0, 0, 1};
        sps.data.insert(sps.data.end(), s->sps.begin(), s->sps.end());
        pps.data = {0, 0, 0, 1};
        pps.data.insert(pps.data.end(), s->pps.begin(), s->pps.end());
        framed.insert(framed.begin(), std::move(pps));
        framed.insert(framed.begin(), std::move(sps));
    }
    std::vector<std::shared_ptr<void>> viewers;
    std::vector<NAL> gop;
    {
        std::lock_guard<std::mutex> lk(s->dist_mu);
        if (idr) s->rtc_gop = framed;
        gop = s->rtc_gop;
        viewers = s->rtc;
    }
    if (viewers.empty()) return;
    uint32_t ssrc = s->ssrc.empty() ? 1u : (uint32_t)strtoul(s->ssrc.c_str(), nullptr, 10);
    auto cur = packetize_nals(framed, ts90, ssrc);
    std::vector<std::vector<uint8_t>> gop_pkts;
    if (!idr && !gop.empty()) {
        uint32_t gop_ts = ts90 > 3600 ? ts90 - 3600 : 0;
        gop_pkts = packetize_nals(gop, gop_ts, ssrc);
    }
    for (auto& v : viewers) {
        auto p = std::static_pointer_cast<WebRtcPlayer>(v);
        if (!p->keys) continue;
        if (p->need_key && !idr && !gop_pkts.empty()) {
            for (auto& pkt : gop_pkts) p->srtp_send(pkt.data(), pkt.size());
            p->need_key = false;
        }
        for (auto& pkt : cur) p->srtp_send(pkt.data(), pkt.size());
        if (idr) p->need_key = false;
    }
}

static std::string sdp_line_value(const std::string& line) {
    auto c = line.find(':');
    if (c == std::string::npos) return "";
    return line.substr(c + 1);
}

bool Hub::whep(const std::string& id, const std::string& offer, bool as_json, int& http_code, std::string& ctype, std::string& body) {
    auto s = get(id);
    if (!s) {
        http_code = 404;
        ctype = "application/json";
        body = "{\"error\":\"no stream\"}";
        return false;
    }
    if (offer.find("v=0") == std::string::npos) {
        http_code = 400;
        ctype = "application/json";
        body = "{\"error\":\"sdp\"}";
        return false;
    }
    struct MLine {
        bool video = false;
        bool audio = false;
        std::string mid;
        std::string proto = "UDP/TLS/RTP/SAVPF";
    };
    std::vector<MLine> lines;
    MLine cur;
    bool in = false;
    std::istringstream is(offer);
    std::string line;
    while (std::getline(is, line)) {
        if (!line.empty() && line.back() == '\r') line.pop_back();
        if (line.rfind("m=", 0) == 0) {
            if (in) lines.push_back(cur);
            in = true;
            cur = MLine();
            cur.video = line.find("video") != std::string::npos;
            cur.audio = line.find("audio") != std::string::npos;
            std::istringstream ms(line);
            std::string m, port, proto;
            ms >> m >> port >> proto;
            if (!proto.empty()) cur.proto = proto;
        } else if (in && line.rfind("a=mid:", 0) == 0) {
            cur.mid = sdp_line_value(line);
        }
    }
    if (in) lines.push_back(cur);
    if (lines.empty()) {
        MLine v;
        v.video = true;
        v.mid = "0";
        lines.push_back(v);
    }
    bool we_server = true;
    if (offer.find("a=setup:actpass") == std::string::npos && offer.find("a=setup:active") == std::string::npos &&
        offer.find("a=setup:passive") != std::string::npos) {
        we_server = false;
    }
    auto player = std::make_shared<WebRtcPlayer>();
    player->ufrag = rand_token(4);
    player->pwd = rand_token(24);
    player->we_server = we_server;
    for (int attempt = 0; attempt < 8 && player->fd < 0; ++attempt) {
        int port = grab_rtc_port();
        int fd = socket(AF_INET, SOCK_DGRAM, 0);
        int opt = 1;
        setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt));
        sockaddr_in addr{};
        addr.sin_family = AF_INET;
        addr.sin_port = htons((uint16_t)port);
        addr.sin_addr.s_addr = INADDR_ANY;
        if (bind(fd, (sockaddr*)&addr, sizeof(addr)) == 0) {
            player->fd = fd;
            player->port = port;
        } else {
            ::close(fd);
        }
    }
    if (player->fd < 0) {
        http_code = 500;
        ctype = "application/json";
        body = "{\"error\":\"webrtc port\"}";
        return false;
    }
    std::string ip = public_ip();
    std::string fp = dtls_cert().fingerprint;
    uint32_t ssrc = s->ssrc.empty() ? 1u : (uint32_t)strtoul(s->ssrc.c_str(), nullptr, 10);
    char pli[16] = "42e01f";
    if (s->sps.size() >= 4) snprintf(pli, sizeof(pli), "%02x%02x%02x", s->sps[1], s->sps[2], s->sps[3]);
    std::string sprop;
    if (!s->sps.empty() && !s->pps.empty()) sprop = b64(s->sps.data(), s->sps.size()) + "," + b64(s->pps.data(), s->pps.size());
    std::ostringstream sdp;
    sdp << "v=0\r\n"
        << "o=- " << ssrc << " 1 IN IP4 " << ip << "\r\n"
        << "s=ArgusSMS\r\n"
        << "t=0 0\r\n"
        << "a=ice-lite\r\n"
        << "a=group:BUNDLE";
    for (auto& m : lines) {
        if (m.video) sdp << " " << (m.mid.empty() ? "0" : m.mid);
    }
    sdp << "\r\n"
        << "a=msid-semantic: WMS *\r\n"
        << "a=ice-ufrag:" << player->ufrag << "\r\n"
        << "a=ice-pwd:" << player->pwd << "\r\n"
        << "a=fingerprint:sha-256 " << fp << "\r\n"
        << "a=setup:" << (we_server ? "passive" : "active") << "\r\n";
    for (size_t i = 0; i < lines.size(); ++i) {
        auto& m = lines[i];
        std::string mid = m.mid.empty() ? std::to_string(i) : m.mid;
        if (m.video) {
            sdp << "m=video " << player->port << " " << m.proto << " 96\r\n"
                << "c=IN IP4 " << ip << "\r\n"
                << "a=mid:" << mid << "\r\n"
                << "a=sendonly\r\n"
                << "a=rtcp-mux\r\n"
                << "a=rtcp:" << player->port << " IN IP4 " << ip << "\r\n"
                << "a=rtpmap:96 H264/90000\r\n"
                << "a=fmtp:96 level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=" << pli;
            if (!sprop.empty()) sdp << ";sprop-parameter-sets=" << sprop;
            sdp << "\r\n"
                << "a=ssrc:" << ssrc << " cname:argussms\r\n"
                << "a=candidate:1 1 udp 2130706431 " << ip << " " << player->port << " typ host\r\n";
        } else {
            sdp << "m=audio 0 " << m.proto << " 0\r\n"
                << "c=IN IP4 0.0.0.0\r\n"
                << "a=mid:" << mid << "\r\n"
                << "a=inactive\r\n";
        }
    }
    player->th = std::thread([player] { player->loop(); });
    {
        std::lock_guard<std::mutex> lk(s->dist_mu);
        s->rtc.push_back(player);
    }
    std::string answer = sdp.str();
    if (as_json) {
        http_code = 200;
        ctype = "application/json";
        std::string esc;
        for (char ch : answer) {
            if (ch == '\r') continue;
            if (ch == '\n') esc += "\\n";
            else if (ch == '"') esc += "\\\"";
            else esc.push_back(ch);
        }
        body = "{\"code\":0,\"type\":\"answer\",\"sdp\":\"" + esc + "\"}";
    } else {
        http_code = 201;
        ctype = "application/sdp";
        body = answer;
    }
    return true;
}
