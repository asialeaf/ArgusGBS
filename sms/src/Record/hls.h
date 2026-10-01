#pragma once
#include <cstdint>
#include <vector>
struct NAL;
class Session;
void hls_on_video(Session& s, const std::vector<NAL>& nals, uint32_t ts);
