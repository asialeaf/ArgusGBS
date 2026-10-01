#include "Common/sms.h"
#include "Common/net.h"
#include "Rtsp/rtsp.h"
#include "Rtmp/rtmp.h"
#include <thread>

void Hub::start_extra() {
    std::thread(listen_loop, rtsp_port, "RTSP", rtsp_client, this).detach();
    std::thread(listen_loop, rtmp_port, "RTMP", rtmp_client, this).detach();
}
