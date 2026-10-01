#pragma once
#include <cstdint>
#include <vector>
struct NAL;
class Session;
class Hub;
void rtsp_on_video(Session& s, const std::vector<NAL>& nals, uint32_t ts);
void rtsp_client(Hub* hub, int fd);
