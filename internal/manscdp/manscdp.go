package manscdp

import (
	"bytes"
	"encoding/xml"
	"io"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// Envelope 同时覆盖 GB/T 28181-2016 与 GB/T 28181-2022 的 MANSCDP 报文。
// 两个年代的目录、心跳、设备控制主体一致；2022 在目录项和部分查询上扩展了字段。
type Envelope struct {
	XMLName        xml.Name   `xml:""`
	CmdType        string     `xml:"CmdType"`
	SN             string     `xml:"SN"`
	DeviceID       string     `xml:"DeviceID"`
	Status         string     `xml:"Status"`
	Result         string     `xml:"Result"`
	SumNum         int        `xml:"SumNum"`
	DeviceList     DeviceList `xml:"DeviceList"`
	RecordList     RecordList `xml:"RecordList"`
	AlarmPriority  int        `xml:"AlarmPriority"`
	AlarmMethod    int        `xml:"AlarmMethod"`
	AlarmTime      string     `xml:"AlarmTime"`
	AlarmType      int        `xml:"AlarmType"`
	Info           string     `xml:"Info"`
	Longitude      string     `xml:"Longitude"`
	Latitude       string     `xml:"Latitude"`
	PTZCmd         string     `xml:"PTZCmd"`
	StartTime      string     `xml:"StartTime"`
	EndTime        string     `xml:"EndTime"`
	SourceID       string     `xml:"SourceID"`
	TargetID       string     `xml:"TargetID"`
}

type DeviceList struct {
	Num  int    `xml:"Num,attr"`
	Item []Item `xml:"Item"`
}

type RecordList struct {
	Num  int      `xml:"Num,attr"`
	Item []Record `xml:"Item"`
}

type Item struct {
	DeviceID       string `xml:"DeviceID"`
	Name           string `xml:"Name"`
	Manufacturer   string `xml:"Manufacturer"`
	Model          string `xml:"Model"`
	Owner          string `xml:"Owner"`
	CivilCode      string `xml:"CivilCode"`
	Block          string `xml:"Block"`
	Address        string `xml:"Address"`
	Parental       int    `xml:"Parental"`
	ParentID       string `xml:"ParentID"`
	SafetyWay      int    `xml:"SafetyWay"`
	RegisterWay    int    `xml:"RegisterWay"`
	CertNum        string `xml:"CertNum"`
	Certifiable    int    `xml:"Certifiable"`
	ErrCode        int    `xml:"ErrCode"`
	EndTime        string `xml:"EndTime"`
	Secrecy        int    `xml:"Secrecy"`
	IPAddress      string `xml:"IPAddress"`
	Port           int    `xml:"Port"`
	Password       string `xml:"Password"`
	Status         string `xml:"Status"`
	Longitude      string `xml:"Longitude"`
	Latitude       string `xml:"Latitude"`
	PTZType        int    `xml:"PTZType"`
	PositionType   int    `xml:"PositionType"`
	RoomType       int    `xml:"RoomType"`
	UseType        int    `xml:"UseType"`
	SupplyLightType int   `xml:"SupplyLightType"`
	DirectionType  int    `xml:"DirectionType"`
	Resolution     string `xml:"Resolution"`
	BusinessGroupID string `xml:"BusinessGroupID"`
	DownloadSpeed  string `xml:"DownloadSpeed"`
	Firmware       string `xml:"Firmware"`
	SerialNumber   string `xml:"SerialNumber"`
	// GB/T 28181-2022 扩展
	SecurityLevelCode   string `xml:"SecurityLevelCode"`
	StreamNumber        int    `xml:"StreamNumber"`
	SVCSpaceSupportMode int    `xml:"SVCSpaceSupportMode"`
	SVCTimeSupportMode  int    `xml:"SVCTimeSupportMode"`
}

type Record struct {
	DeviceID   string `xml:"DeviceID"`
	Name       string `xml:"Name"`
	FilePath   string `xml:"FilePath"`
	Address    string `xml:"Address"`
	StartTime  string `xml:"StartTime"`
	EndTime    string `xml:"EndTime"`
	Secrecy    int    `xml:"Secrecy"`
	Type       string `xml:"Type"`
	RecorderID string `xml:"RecorderID"`
	FileSize   string `xml:"FileSize"`
	StreamNumber int  `xml:"StreamNumber"`
}

func (e *Envelope) Is2022() bool {
	switch e.CmdType {
	case "PTZPosition", "HomePositionQuery", "CruiseTrackListQuery", "CruiseTrackQuery",
		"SDCardStatus", "DeviceUpgrade", "SnapShotConfig", "OSDConfig", "TargetTrack":
		return true
	}
	for _, it := range e.DeviceList.Item {
		if it.DownloadSpeed != "" || it.SecurityLevelCode != "" || it.StreamNumber != 0 || it.SVCSpaceSupportMode != 0 || it.SVCTimeSupportMode != 0 {
			return true
		}
	}
	for _, it := range e.RecordList.Item {
		if it.StreamNumber != 0 {
			return true
		}
	}
	return false
}

func Decode(b []byte) (*Envelope, error) {
	body := b
	low := bytes.ToLower(b)
	if bytes.Contains(low, []byte("gb2312")) || bytes.Contains(low, []byte("gbk")) {
		if u, err := simplifiedchinese.GBK.NewDecoder().Bytes(b); err == nil {
			body = u
		}
	}
	s := string(body)
	if i := strings.Index(s, "?>"); i >= 0 && strings.Contains(strings.ToLower(s[:i]), "<?xml") {
		s = s[i+2:]
	}
	dec := xml.NewDecoder(strings.NewReader(strings.TrimSpace(s)))
	dec.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		if strings.Contains(strings.ToLower(charset), "gb") {
			return simplifiedchinese.GBK.NewDecoder().Reader(input), nil
		}
		return input, nil
	}
	var env Envelope
	if err := dec.Decode(&env); err != nil {
		return nil, err
	}
	return &env, nil
}

func Encode(root string, charset string, fields [][2]string) ([]byte, error) {
	var buf strings.Builder
	buf.WriteString("<")
	buf.WriteString(root)
	buf.WriteString(">\r\n")
	for _, f := range fields {
		if f[1] == "" {
			continue
		}
		buf.WriteString("<")
		buf.WriteString(f[0])
		buf.WriteString(">")
		buf.WriteString(xmlEscape(f[1]))
		buf.WriteString("</")
		buf.WriteString(f[0])
		buf.WriteString(">\r\n")
	}
	buf.WriteString("</")
	buf.WriteString(root)
	buf.WriteString(">\r\n")
	payload := []byte(buf.String())
	cs := strings.ToUpper(charset)
	if cs == "UTF-8" || cs == "UTF8" {
		return append([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\r\n"), payload...), nil
	}
	enc, err := simplifiedchinese.GBK.NewEncoder().Bytes(payload)
	if err != nil {
		return nil, err
	}
	return append([]byte("<?xml version=\"1.0\" encoding=\"GB2312\"?>\r\n"), enc...), nil
}

func xmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}

func Atoi(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

// PTZCmd 生成 GB28181 云台 8 字节指令的十六进制字符串。
// command: left/right/up/down/upleft/upright/downleft/downright/zoomin/zoomout/stop
func PTZCmd(command string, speed int) string {
	if speed <= 0 {
		speed = 80
	}
	if speed > 255 {
		speed = 255
	}
	var code byte
	switch strings.ToLower(command) {
	case "right":
		code = 0x01
	case "left":
		code = 0x02
	case "down":
		code = 0x04
	case "up":
		code = 0x08
	case "downright":
		code = 0x05
	case "downleft":
		code = 0x06
	case "upright":
		code = 0x09
	case "upleft":
		code = 0x0A
	case "zoomin":
		code = 0x10
	case "zoomout":
		code = 0x20
	default:
		code = 0x00
		speed = 0
	}
	h := byte(speed)
	v := byte(speed)
	z := byte(0)
	if code == 0x10 || code == 0x20 {
		h, v = 0, 0
		z = byte(speed >> 4)
	}
	return packPTZ(code, h, v, z)
}

// FICmd 焦点光圈。command: focusnear/focusfar/irisin/irisout/stop
func FICmd(command string, speed int) string {
	if speed <= 0 {
		speed = 80
	}
	if speed > 255 {
		speed = 255
	}
	var code byte = 0x40
	var data byte
	switch strings.ToLower(command) {
	case "focusfar", "far":
		code = 0x41
		data = byte(speed)
	case "focusnear", "near":
		code = 0x42
		data = byte(speed)
	case "irisin":
		code = 0x44
		data = byte(speed)
	case "irisout":
		code = 0x48
		data = byte(speed)
	default:
		code = 0x40
		data = 0
	}
	return packPTZ(code, 0, data, 0)
}

func PresetCmd(action string, preset int) string {
	var code byte
	switch strings.ToLower(action) {
	case "set":
		code = 0x81
	case "goto", "call":
		code = 0x82
	case "remove", "del":
		code = 0x83
	default:
		code = 0x82
	}
	if preset < 1 {
		preset = 1
	}
	if preset > 255 {
		preset = 255
	}
	return packPTZ(code, 0, byte(preset), 0)
}

func packPTZ(cmd, d1, d2, d3 byte) string {
	b := []byte{0xA5, 0x0F, 0x01, cmd, d1, d2, d3 << 4, 0}
	sum := 0
	for i := 0; i < 7; i++ {
		sum += int(b[i])
	}
	b[7] = byte(sum % 256)
	const hexdigits = "0123456789ABCDEF"
	out := make([]byte, 16)
	for i, v := range b {
		out[i*2] = hexdigits[v>>4]
		out[i*2+1] = hexdigits[v&0x0F]
	}
	return string(out)
}
