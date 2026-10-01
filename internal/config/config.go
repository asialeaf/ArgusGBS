package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Path string

	HTTPPort         int
	APIAuth          bool
	LiveStreamAuth   bool
	Captcha          bool
	LogoText         string
	LogoMiniText     string
	CopyrightText    string
	WWWDir           string
	PublicHost       string
	HTTPSPort        int
	HTTPSCert        string
	HTTPSKey         string

	DBFile string

	SIPHost                 string
	SIPPort                 int
	Serial                  string
	Realm                   string
	AckTimeout              int
	KeepaliveTimeout        int
	DevicePassword          string
	SIPLog                  bool
	AllowStreamStartByURL   bool
	DropChannelType         string
	DefaultMediaTransport   string
	DefaultMediaTransportMode string
	ChannelDefaultOndemand  bool
	ChannelDefaultCloudRecord bool
	Charset                 string
	StreamAuthURL           string
	PreferStreamFmt         string

	MapEnable bool
	MapCenter string

	SMSHost       string
	SMSHTTPPort   int
	SMSSerial     string
	SMSRealm      string
	SMSSIPPort    int
	SMSRTMPPort   int
	SMSRTSPPort   int
	SMSPublicHost string
	SMSAPISecret  string

	// 运行期可被基础配置接口修改的订阅间隔（秒，0 表示关闭）
	CatalogSubscribeInterval  int
	AlarmSubscribeInterval    int
	PositionSubscribeInterval int
	PTZSubscribeInterval      int
	GlobalChannelShared       int
	GlobalChannelAudio        bool
	PwdLength                 int
	PwdLevel                  int
}

func Default() *Config {
	return &Config{
		HTTPPort:                  10000,
		APIAuth:                   true,
		LogoText:                  "ArgusGBS",
		LogoMiniText:              "GBS",
		CopyrightText:             "ArgusGBS",
		DBFile:                    "data/argus.db",
		SIPPort:                   15060,
		Serial:                    "34020000002000000001",
		Realm:                     "3402000000",
		AckTimeout:                15,
		KeepaliveTimeout:          300,
		DevicePassword:            "gbs12345",
		SIPLog:                    true,
		AllowStreamStartByURL:     true,
		DropChannelType:           "134,135,136,137",
		DefaultMediaTransport:     "UDP",
		DefaultMediaTransportMode: "passive",
		ChannelDefaultOndemand:    true,
		Charset:                   "GB2312",
		PreferStreamFmt:           "FLV",
		SMSHost:                   "127.0.0.1",
		SMSHTTPPort:               10001,
		SMSSerial:                 "34020000002020000001",
		SMSRealm:                  "3402000000",
		SMSSIPPort:                15070,
		SMSRTMPPort:               1935,
		SMSRTSPPort:               554,
		SMSAPISecret:              "argus-sms",
		PwdLength:                 6,
		PwdLevel:                  2,
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()
	cfg.Path = path
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	defer f.Close()
	section := ""
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(strings.Trim(line, "[]"))
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if i := strings.Index(v, " ;"); i >= 0 {
			v = strings.TrimSpace(v[:i])
		}
		apply(cfg, section, strings.ToLower(k), v)
	}
	return cfg, sc.Err()
}

func apply(c *Config, section, key, val string) {
	switch section + "." + key {
	case "http.port":
		c.HTTPPort = atoi(val, c.HTTPPort)
	case "http.api_auth":
		c.APIAuth = val == "1" || strings.EqualFold(val, "true")
	case "http.live_stream_auth":
		c.LiveStreamAuth = val == "1" || strings.EqualFold(val, "true")
	case "http.captcha":
		c.Captcha = val == "1" || strings.EqualFold(val, "true")
	case "http.logo_text":
		c.LogoText = val
	case "http.logo_mini_text":
		c.LogoMiniText = val
	case "http.copyright_text":
		c.CopyrightText = val
	case "http.www_dir":
		c.WWWDir = val
	case "http.public_host":
		c.PublicHost = val
	case "https.port":
		c.HTTPSPort = atoi(val, 0)
	case "https.ssl_cert_file":
		c.HTTPSCert = val
	case "https.ssl_key_file":
		c.HTTPSKey = val
	case "db.file":
		c.DBFile = val
	case "sip.host":
		c.SIPHost = val
	case "sip.port":
		c.SIPPort = atoi(val, c.SIPPort)
	case "sip.serial":
		c.Serial = val
	case "sip.realm":
		c.Realm = val
	case "sip.ack_timeout":
		c.AckTimeout = atoi(val, c.AckTimeout)
	case "sip.keepalive_timeout":
		c.KeepaliveTimeout = atoi(val, c.KeepaliveTimeout)
	case "sip.device_password":
		c.DevicePassword = val
	case "sip.log":
		c.SIPLog = val == "1" || strings.EqualFold(val, "true")
	case "sip.allow_stream_start_by_url":
		c.AllowStreamStartByURL = val == "1" || strings.EqualFold(val, "true")
	case "sip.drop_channel_type":
		c.DropChannelType = val
	case "sip.device_default_media_transport":
		c.DefaultMediaTransport = strings.ToUpper(val)
	case "sip.device_default_media_transport_mode":
		c.DefaultMediaTransportMode = strings.ToLower(val)
	case "sip.channel_default_ondemand":
		c.ChannelDefaultOndemand = val == "1" || strings.EqualFold(val, "true")
	case "sip.channel_default_cloud_record":
		c.ChannelDefaultCloudRecord = val == "1" || strings.EqualFold(val, "true")
	case "sip.charset":
		c.Charset = val
	case "sip.stream_auth_url":
		c.StreamAuthURL = val
	case "sip.prefer_stream_fmt":
		c.PreferStreamFmt = val
	case "map.enable":
		c.MapEnable = val == "1" || strings.EqualFold(val, "true")
	case "map.center":
		c.MapCenter = val
	case "sms.host":
		c.SMSHost = val
	case "sms.http_port":
		c.SMSHTTPPort = atoi(val, c.SMSHTTPPort)
	case "sms.serial":
		c.SMSSerial = val
	case "sms.realm":
		c.SMSRealm = val
	case "sms.sip_port":
		c.SMSSIPPort = atoi(val, c.SMSSIPPort)
	case "sms.rtmp_port":
		c.SMSRTMPPort = atoi(val, c.SMSRTMPPort)
	case "sms.rtsp_port":
		c.SMSRTSPPort = atoi(val, c.SMSRTSPPort)
	case "sms.public_host":
		c.SMSPublicHost = val
	case "sms.api_secret":
		c.SMSAPISecret = val
	}
}

func atoi(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
