CXX ?= g++
CXXFLAGS ?= -std=c++17 -O2 -Wall -pthread
SRC = src/main.cpp src/http.cpp src/session.cpp src/ps.cpp

.PHONY: all clean
all: argussms

argussms: $(SRC) src/sms.h
	$(CXX) $(CXXFLAGS) -o argussms $(SRC)

clean:
	rm -f argussms
