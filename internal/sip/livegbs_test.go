package sip

import (
	"net"
	"strings"
	"testing"
)

func TestDigestChallengeQOP(t *testing.T) {
	h := digestChallenge{Realm: "3402000000", Nonce: "abc"}.Header()
	if h != `Digest realm="3402000000",qop="auth",nonce="abc"` {
		t.Fatal(h)
	}
}

func TestViaRportReceived(t *testing.T) {
	addr, _ := net.ResolveUDPAddr("udp", "192.168.3.17:5060")
	got := viaWithSource("SIP/2.0/UDP 192.168.3.17:5060;rport;branch=z9hG4bK1", addr)
	want := "SIP/2.0/UDP 192.168.3.17:5060;rport=5060;received=192.168.3.17;branch=z9hG4bK1"
	if got != want {
		t.Fatal(got)
	}
}

func TestPlaySDPMatchesLiveGBS(t *testing.T) {
	sdp := buildSDP("34020000001318000001", "192.168.3.41", 30000, "UDP", "passive", "0200000001", "", "", false)
	for _, line := range []string{
		"o=34020000001318000001 0 0 IN IP4 192.168.3.41",
		"s=Play",
		"m=video 30000 RTP/AVP 96 97 98",
		"a=recvonly",
		"a=rtpmap:96 PS/90000",
		"a=rtpmap:97 MPEG4/90000",
		"a=rtpmap:98 H264/90000",
		"y=0200000001",
		"f=v/////a/6/8/1",
	} {
		if !strings.Contains(sdp, line) {
			t.Fatalf("missing %s\n%s", line, sdp)
		}
	}
	if strings.Contains(sdp, "a=setup:") {
		t.Fatal(sdp)
	}
}
