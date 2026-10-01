#include "Common/sms.h"
#include "Record/hls.h"
#include "Rtsp/rtsp.h"

void Session::playout(const std::vector<NAL>& nals) {
    if (nals.empty()) return;
    uint32_t ts = play_ts;
    play_ts += 3600;
    hls_on_video(*this, nals, ts);
    rtsp_on_video(*this, nals, ts);
    rtc_send(this, nals, ts);
}
