package model

import "time"

const TimeLayout = "2006-01-02 15:04:05"

func FormatTime(t time.Time) string {
	if t.IsZero() {
		return "0001-01-01 00:00:00"
	}
	return t.In(time.Local).Format(TimeLayout)
}

func ParseTime(s string) time.Time {
	if s == "" || stringsHasPrefixZero(s) {
		return time.Time{}
	}
	t, err := time.ParseInLocation(TimeLayout, s, time.Local)
	if err != nil {
		t, err = time.Parse(time.RFC3339, s)
		if err != nil {
			return time.Time{}
		}
	}
	return t
}

func stringsHasPrefixZero(s string) bool {
	return len(s) >= 4 && s[:4] == "0001"
}

type User struct {
	ID              int64  `json:"ID"`
	Username        string `json:"Username"`
	PasswordMD5     string `json:"-"`
	Role            string `json:"Role"`
	ControlPriority int    `json:"ControlPriority"`
	PhoneNumber     string `json:"PhoneNumber"`
	Email           string `json:"Email"`
	Description     string `json:"Description"`
	Creator         string `json:"Creator"`
	Lock            bool   `json:"Lock"`
	Enable          bool   `json:"Enable"`
	HasAllChannel   bool   `json:"HasAllChannel"`
	LastLoginAt     string `json:"LastLoginAt"`
	UpdatedAt       string `json:"UpdatedAt"`
	CreatedAt       string `json:"CreatedAt"`
}

type Device struct {
	ID                   string  `json:"ID"`
	Name                 string  `json:"Name"`
	CustomName           string  `json:"CustomName"`
	Manufacturer         string  `json:"Manufacturer"`
	Model                string  `json:"Model"`
	Firmware             string  `json:"Firmware"`
	GBVer                string  `json:"GBVer"`
	Type                 string  `json:"Type"`
	ChannelCount         int     `json:"ChannelCount"`
	RecvStreamIP         string  `json:"RecvStreamIP"`
	ContactIP            string  `json:"ContactIP"`
	DropChannelType      string  `json:"DropChannelType"`
	SMSID                string  `json:"SMSID"`
	SMSGroupID           string  `json:"SMSGroupID"`
	CatalogInterval      int     `json:"CatalogInterval"`
	SubscribeInterval    int     `json:"SubscribeInterval"`
	CatalogSubscribe     bool    `json:"CatalogSubscribe"`
	AlarmSubscribe       bool    `json:"AlarmSubscribe"`
	PositionSubscribe    bool    `json:"PositionSubscribe"`
	PTZSubscribe         bool    `json:"PTZSubscribe"`
	Online               bool    `json:"Online"`
	Password             string  `json:"Password"`
	RecordCenter         bool    `json:"RecordCenter"`
	RecordIndistinct     bool    `json:"RecordIndistinct"`
	CivilCodeFirst       bool    `json:"CivilCodeFirst"`
	KeepOriginalTree     bool    `json:"KeepOriginalTree"`
	CommandTransport     string  `json:"CommandTransport"`
	MediaTransport       string  `json:"MediaTransport"`
	MediaTransportMode   string  `json:"MediaTransportMode"`
	RemoteIP             string  `json:"RemoteIP"`
	RemotePort           int     `json:"RemotePort"`
	Longitude            float64 `json:"Longitude"`
	Latitude             float64 `json:"Latitude"`
	LastRegisterAt       string  `json:"LastRegisterAt"`
	LastKeepaliveAt      string  `json:"LastKeepaliveAt"`
	UpdatedAt            string  `json:"UpdatedAt"`
	CreatedAt            string  `json:"CreatedAt"`
	CatalogProgress      string  `json:"CatalogProgress,omitempty"`
	ChannelOverLoad      bool    `json:"ChannelOverLoad"`
	UA                   string  `json:"-"`
	Charset              string  `json:"-"`
}

type Channel struct {
	ID                 string  `json:"ID"`
	DeviceID           string  `json:"DeviceID"`
	DeviceName         string  `json:"DeviceName"`
	DeviceCustomName   string  `json:"DeviceCustomName"`
	DeviceType         string  `json:"DeviceType"`
	DeviceOnline       bool    `json:"DeviceOnline"`
	Channel            int     `json:"Channel"`
	Name               string  `json:"Name"`
	CustomName         string  `json:"CustomName"`
	Block              string  `json:"Block"`
	CustomBlock        string  `json:"CustomBlock"`
	Custom             bool    `json:"Custom"`
	CustomID           string  `json:"CustomID"`
	SubCount           int     `json:"SubCount"`
	SnapURL            string  `json:"SnapURL"`
	Manufacturer       string  `json:"Manufacturer"`
	CustomManufacturer string  `json:"CustomManufacturer"`
	Model              string  `json:"Model"`
	CustomModel        string  `json:"CustomModel"`
	Owner              string  `json:"Owner"`
	CivilCode          string  `json:"CivilCode"`
	CustomCivilCode    string  `json:"CustomCivilCode"`
	Address            string  `json:"Address"`
	CustomAddress      string  `json:"CustomAddress"`
	Firmware           string  `json:"Firmware"`
	CustomFirmware     string  `json:"CustomFirmware"`
	SerialNumber       string  `json:"SerialNumber"`
	CustomSerialNumber string  `json:"CustomSerialNumber"`
	IPAddress          string  `json:"IPAddress"`
	CustomIPAddress    string  `json:"CustomIPAddress"`
	Port               int     `json:"Port"`
	CustomPort         int     `json:"CustomPort"`
	Parental           int     `json:"Parental"`
	ParentID           string  `json:"ParentID"`
	CustomParentID     string  `json:"CustomParentID"`
	Secrecy            int     `json:"Secrecy"`
	RegisterWay        int     `json:"RegisterWay"`
	Status             string  `json:"Status"`
	CustomStatus       string  `json:"CustomStatus"`
	Longitude          float64 `json:"Longitude"`
	Latitude           float64 `json:"Latitude"`
	CustomLongitude    float64 `json:"CustomLongitude"`
	CustomLatitude     float64 `json:"CustomLatitude"`
	Altitude           float64 `json:"Altitude"`
	Speed              float64 `json:"Speed"`
	Direction          float64 `json:"Direction"`
	PTZType            int     `json:"PTZType"`
	CustomPTZType      int     `json:"CustomPTZType"`
	BatteryLevel       string  `json:"BatteryLevel"`
	SignalLevel        string  `json:"SignalLevel"`
	DownloadSpeed      string  `json:"DownloadSpeed"`
	Ondemand           bool    `json:"Ondemand"`
	AudioEnable        bool    `json:"AudioEnable"`
	CloudRecord        bool    `json:"CloudRecord"`
	Shared             bool    `json:"Shared"`
	StreamID           string  `json:"StreamID"`
	NumOutputs         int     `json:"NumOutputs"`
}

type TreeNode struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Custom        bool    `json:"custom"`
	CustomID      string  `json:"customID"`
	CustomName    string  `json:"customName"`
	Serial        string  `json:"serial"`
	Code          string  `json:"code"`
	Status        string  `json:"status"`
	PTZType       int     `json:"ptzType"`
	Longitude     float64 `json:"longitude"`
	Latitude      float64 `json:"latitude"`
	Parental      int     `json:"parental"`
	Manufacturer  string  `json:"manufacturer"`
	SubCount      int     `json:"subCount"`
	SubCountDevice int    `json:"subCountDevice"`
	OnlineSubCount int    `json:"onlineSubCount"`
}

type Cascade struct {
	ID                string `json:"ID"`
	Name              string `json:"Name"`
	Serial            string `json:"Serial"`
	Realm             string `json:"Realm"`
	Host              string `json:"Host"`
	Port              int    `json:"Port"`
	LocalSerial       string `json:"LocalSerial"`
	LocalHost         string `json:"LocalHost"`
	LocalPort         int    `json:"LocalPort"`
	Username          string `json:"Username"`
	Password          string `json:"Password"`
	KeepaliveMaxCount int    `json:"KeepaliveMaxCount"`
	KeepaliveInterval int    `json:"KeepaliveInterval"`
	RegisterInterval  int    `json:"RegisterInterval"`
	RegisterTimeout   int    `json:"RegisterTimeout"`
	ControlPriority   int    `json:"ControlPriority"`
	LoadLimit         int    `json:"LoadLimit"`
	CatalogGroupSize  int    `json:"CatalogGroupSize"`
	CommandTransport  string `json:"CommandTransport"`
	Charset           string `json:"Charset"`
	Online            bool   `json:"Online"`
	Status            string `json:"Status"`
	Load              int    `json:"Load"`
	AllowControl      bool   `json:"AllowControl"`
	ShareAllChannel   bool   `json:"ShareAllChannel"`
	ShareRecord       bool   `json:"ShareRecord"`
	StreamKeepalive   bool   `json:"StreamKeepalive"`
	StreamReader      bool   `json:"StreamReader"`
	BindLocalIP       bool   `json:"BindLocalIP"`
	Enable            bool   `json:"Enable"`
}

type Alarm struct {
	ID             string `json:"ID"`
	DeviceID       string `json:"DeviceID"`
	ChannelID      string `json:"ChannelID"`
	AlarmPriority  int    `json:"AlarmPriority"`
	Time           string `json:"Time"`
	AlarmMethod    int    `json:"AlarmMethod"`
	AlarmType      int    `json:"AlarmType"`
	AlarmEventType string `json:"AlarmEventType"`
	ExtInfo        string `json:"ExtInfo"`
	RecordPath     string `json:"RecordPath"`
	SnapPath       string `json:"SnapPath"`
	CreatedAt      string `json:"CreatedAt"`
}

type OpLog struct {
	ID          int64   `json:"ID"`
	Name        string  `json:"Name"`
	Method      string  `json:"Method"`
	RequestURI  string  `json:"RequestURI"`
	RemoteAddr  string  `json:"RemoteAddr"`
	Status      string  `json:"Status"`
	Duration    float64 `json:"Duration"`
	Username    string  `json:"Username"`
	StartAt     string  `json:"StartAt"`
	ExtInfo     string  `json:"ExtInfo"`
	Description string  `json:"Description"`
}

type Stream struct {
	StreamID              string  `json:"StreamID"`
	SMSID                 string  `json:"SMSID"`
	DeviceID              string  `json:"DeviceID"`
	ChannelID             string  `json:"ChannelID"`
	ChannelName           string  `json:"ChannelName"`
	WEBRTC                string  `json:"WEBRTC"`
	WHEP                  string  `json:"WHEP"`
	FLV                   string  `json:"FLV"`
	WSFLV                 string  `json:"WS_FLV"`
	RTMP                  string  `json:"RTMP"`
	HLS                   string  `json:"HLS"`
	RTSP                  string  `json:"RTSP"`
	CDN                   string  `json:"CDN"`
	SnapURL               string  `json:"SnapURL"`
	Transport             string  `json:"Transport"`
	StartAt               string  `json:"StartAt"`
	RecordStartAt         string  `json:"RecordStartAt"`
	Duration              int     `json:"Duration"`
	SourceVideoCodecName  string  `json:"SourceVideoCodecName"`
	SourceVideoWidth      int     `json:"SourceVideoWidth"`
	SourceVideoHeight     int     `json:"SourceVideoHeight"`
	SourceVideoFrameRate  int     `json:"SourceVideoFrameRate"`
	SourceAudioCodecName  string  `json:"SourceAudioCodecName"`
	SourceAudioSampleRate int     `json:"SourceAudioSampleRate"`
	RTPCount              int64   `json:"RTPCount"`
	RTPLostCount          int64   `json:"RTPLostCount"`
	RTPLostRate           float64 `json:"RTPLostRate"`
	VideoFrameCount       int64   `json:"VideoFrameCount"`
	AudioEnable           bool    `json:"AudioEnable"`
	Ondemand              bool    `json:"Ondemand"`
	CloudRecord           bool    `json:"CloudRecord"`
	InBytes               int64   `json:"InBytes"`
	InBitRate             int     `json:"InBitRate"`
	OutBytes              int64   `json:"OutBytes"`
	NumOutputs            int     `json:"NumOutputs"`
	CascadeSize           int     `json:"CascadeSize"`
	DecodeSize            int     `json:"DecodeSize"`
	RelaySize             int     `json:"RelaySize"`
	ChannelPTZType        int     `json:"ChannelPTZType"`
	ChannelOSD            string  `json:"ChannelOSD"`
	CallID                string  `json:"-"`
	SSRC                  string  `json:"-"`
	Playback              bool    `json:"-"`
	StartTime             string  `json:"-"`
	EndTime               string  `json:"-"`
	Stopped               bool    `json:"-"`
}

type RecordItem struct {
	DeviceID       string `json:"DeviceID"`
	Name           string `json:"Name"`
	FilePath       string `json:"FilePath"`
	FileSize       string `json:"FileSize"`
	Address        string `json:"Address"`
	StartTime      string `json:"StartTime"`
	EndTime        string `json:"EndTime"`
	Secrecy        string `json:"Secrecy"`
	Type           string `json:"Type"`
	RecorderID     string `json:"RecorderID"`
	RecordLocation string `json:"RecordLocation"`
	StreamNumber   int    `json:"StreamNumber"`
}

type AccessRule struct {
	ID          int64  `json:"ID"`
	Serial      string `json:"Serial"`
	IP          string `json:"IP"`
	UA          string `json:"UA"`
	Password    string `json:"Password"`
	Description string `json:"Description"`
	CreatedAt   string `json:"CreatedAt"`
}

type CloudRecord struct {
	ID        string `json:"ID"`
	DeviceID  string `json:"DeviceID"`
	ChannelID string `json:"ChannelID"`
	Name      string `json:"Name"`
	StartTime string `json:"StartTime"`
	EndTime   string `json:"EndTime"`
	FilePath  string `json:"FilePath"`
	FileSize  int64  `json:"FileSize"`
	Important bool   `json:"Important"`
	Duration  int    `json:"Duration"`
}

type GroupNode struct {
	ID       string `json:"-"`
	ParentID string `json:"-"`
	Name     string `json:"-"`
	CustomID string `json:"-"`
	Serial   string `json:"-"`
	Code     string `json:"-"`
}
