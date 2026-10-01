package media

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	Base   string
	Secret string
	Host   string
	HTTP   *http.Client
}

func New(host string, port int, secret string) *Client {
	return &Client{
		Base:   fmt.Sprintf("http://%s:%d", host, port),
		Secret: secret,
		Host:   host,
		HTTP:   &http.Client{Timeout: 5 * time.Second},
	}
}

type OpenReq struct {
	StreamID  string `json:"stream_id"`
	Transport string `json:"transport"`
	Mode      string `json:"mode"`
	SSRC      string `json:"ssrc"`
	PeerIP    string `json:"peer_ip,omitempty"`
	PeerPort  int    `json:"peer_port,omitempty"`
	VideoCodec string `json:"video_codec,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Bitrate   int    `json:"bitrate,omitempty"`
	FrameRate int    `json:"framerate,omitempty"`
}

type OpenResp struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	PublicIP string `json:"public_ip"`
}

type Stats struct {
	StreamID     string `json:"stream_id"`
	Codec        string `json:"codec"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FPS          int    `json:"fps"`
	RTPCount     int64  `json:"rtp_count"`
	RTPLost      int64  `json:"rtp_lost"`
	InBytes      int64  `json:"in_bytes"`
	InBitRate    int    `json:"in_bitrate"`
	VideoFrames  int64  `json:"video_frames"`
	NumOutputs   int    `json:"num_outputs"`
	AudioCodec   string `json:"audio_codec"`
	Ready        bool   `json:"ready"`
}

func (c *Client) Open(req OpenReq) (*OpenResp, error) {
	var out OpenResp
	if err := c.post("/api/v1/rtp/open", req, &out); err != nil {
		return nil, err
	}
	if out.Port == 0 {
		return nil, fmt.Errorf("流媒体未返回收流端口")
	}
	return &out, nil
}

func (c *Client) Close(streamID string) error {
	return c.post("/api/v1/rtp/close", map[string]string{"stream_id": streamID}, nil)
}

func (c *Client) Relay(streamID, transport, mode, peerIP string, peerPort int, ssrc string) error {
	return c.post("/api/v1/rtp/relay", map[string]any{
		"stream_id": streamID, "transport": transport, "mode": mode,
		"peer_ip": peerIP, "peer_port": peerPort, "ssrc": ssrc,
	}, nil)
}

func (c *Client) Stats(streamID string) (*Stats, error) {
	req, err := http.NewRequest(http.MethodGet, c.Base+"/api/v1/rtp/stats?stream_id="+streamID, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Argus-Secret", c.Secret)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("sms %s", bytes.TrimSpace(b))
	}
	var st Stats
	if err := json.Unmarshal(b, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

func (c *Client) Healthy() bool {
	req, err := http.NewRequest(http.MethodGet, c.Base+"/api/v1/serverinfo", nil)
	if err != nil {
		return false
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

func (c *Client) post(path string, in any, out any) error {
	b, _ := json.Marshal(in)
	req, err := http.NewRequest(http.MethodPost, c.Base+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Argus-Secret", c.Secret)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("流媒体服务不可达: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		msg := stringsTrim(body)
		if msg == "" {
			msg = resp.Status
		}
		return fmt.Errorf("%s", msg)
	}
	if out != nil && len(body) > 0 {
		if err := json.Unmarshal(body, out); err != nil {
			return err
		}
	}
	return nil
}

func stringsTrim(b []byte) string {
	return string(bytes.TrimSpace(b))
}
