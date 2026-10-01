CXX ?= g++
CXXFLAGS ?= -std=c++17 -O2 -Wall -pthread -Isrc
LDLIBS ?= -lssl -lcrypto
SRC = server/main.cpp \
      src/Common/net.cpp src/Common/playout.cpp src/Common/server.cpp \
      src/Rtp/session.cpp src/Rtp/ps.cpp src/Rtp/relay.cpp \
      src/Http/http.cpp \
      src/Rtsp/rtsp.cpp \
      src/Rtmp/rtmp.cpp \
      src/Record/hls.cpp \
      webrtc/webrtc.cpp

.PHONY: all clean
all: argussms

argussms: $(SRC) src/Common/sms.h src/Common/net.h src/Record/hls.h src/Rtsp/rtsp.h src/Rtmp/rtmp.h
	$(CXX) $(CXXFLAGS) -o argussms $(SRC) $(LDLIBS)

clean:
	rm -f argussms
