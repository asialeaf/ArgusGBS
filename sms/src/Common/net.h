#pragma once
#include <cstddef>
#include <string>
class Hub;
bool send_all_fd(int fd, const uint8_t* p, size_t n);
bool read_full(int fd, std::string& stash, size_t n);
void listen_loop(int port, const char* name, void (*client)(Hub*, int), Hub* hub);
