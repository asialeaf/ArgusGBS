package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/argus/argusgbs/internal/config"
	"github.com/argus/argusgbs/internal/manscdp"
	"github.com/argus/argusgbs/internal/media"
	"github.com/argus/argusgbs/internal/metrics"
	"github.com/argus/argusgbs/internal/model"
	"github.com/argus/argusgbs/internal/sip"
	"github.com/argus/argusgbs/internal/store"
)

type API struct {
	Cfg     *config.Config
	DB      *store.Store
	SIP     *sip.Server
	Media   *media.Client
	WWW     string
	Started time.Time
}

func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/login", a.login)
	mux.HandleFunc("GET /api/v1/login", a.login)
	mux.HandleFunc("GET /api/v1/logout", a.logout)
	mux.HandleFunc("GET /api/v1/userinfo", a.userinfo)
	mux.HandleFunc("GET /api/v1/getserverinfo", a.serverinfo)
	mux.HandleFunc("POST /api/v1/modifypassword", a.auth(a.modifyPassword))
	mux.HandleFunc("POST /api/v1/restart", a.auth(a.restart))
	mux.HandleFunc("GET /api/v1/getbaseconfig", a.auth(a.getBaseConfig))
	mux.HandleFunc("POST /api/v1/setbaseconfig", a.auth(a.setBaseConfig))
	mux.HandleFunc("GET /api/v1/getpwdconfig", a.auth(a.getPwdConfig))
	mux.HandleFunc("POST /api/v1/setpwdconfig", a.auth(a.setPwdConfig))
	mux.HandleFunc("GET /api/v1/sms/list", a.auth(a.smsList))
	mux.HandleFunc("GET /api/v1/sms/getbaseconfig", a.auth(a.smsGet))
	mux.HandleFunc("POST /api/v1/sms/setbaseconfig", a.auth(a.smsSet))

	mux.HandleFunc("GET /api/v1/dashboard/auth", a.auth(a.dashAuth))
	mux.HandleFunc("GET /api/v1/dashboard/store", a.auth(a.dashStore))
	mux.HandleFunc("GET /api/v1/dashboard/top", a.auth(a.dashTop))
	mux.HandleFunc("GET /api/v1/dashboard/top/net", a.auth(a.dashTop))

	mux.HandleFunc("GET /api/v1/device/list", a.auth(a.deviceList))
	mux.HandleFunc("GET /api/v1/device/info", a.auth(a.deviceInfo))
	mux.HandleFunc("GET /api/v1/device/channellist", a.auth(a.channelList))
	mux.HandleFunc("GET /api/v1/device/channelinfo", a.auth(a.channelInfo))
	mux.HandleFunc("GET /api/v1/device/channeltree", a.auth(a.channelTree))
	mux.HandleFunc("GET /api/v1/device/grouptree", a.auth(a.groupTree))
	mux.HandleFunc("GET /api/v1/device/onlinestats", a.auth(a.onlineStats))
	mux.HandleFunc("POST /api/v1/device/remove", a.auth(a.deviceRemove))
	mux.HandleFunc("POST /api/v1/device/setinfo", a.auth(a.deviceSetInfo))
	mux.HandleFunc("POST /api/v1/device/setname", a.auth(a.deviceSetName))
	mux.HandleFunc("POST /api/v1/device/setmediatransport", a.auth(a.deviceSetTransport))
	mux.HandleFunc("POST /api/v1/device/setsms", a.auth(a.deviceSetSMS))
	mux.HandleFunc("GET /api/v1/device/fetchcatalog", a.auth(a.fetchCatalog))
	mux.HandleFunc("GET /api/v1/device/fetchinfo", a.auth(a.fetchInfo))
	mux.HandleFunc("GET /api/v1/device/fetchstatus", a.auth(a.fetchStatus))
	mux.HandleFunc("GET /api/v1/device/fetchpreset", a.auth(a.fetchPreset))
	mux.HandleFunc("GET /api/v1/device/statuslog", a.auth(a.statusLog))
	mux.HandleFunc("GET /api/v1/device/streamlog", a.auth(a.streamLog))
	mux.HandleFunc("GET /api/v1/device/positionlog", a.auth(a.positionLog))
	mux.HandleFunc("POST /api/v1/device/setchannelname", a.auth(a.setChannelName))
	mux.HandleFunc("POST /api/v1/device/setchannelid", a.auth(a.setChannelCustom))
	mux.HandleFunc("POST /api/v1/device/setchannelaudio", a.auth(a.setChannelAudio))
	mux.HandleFunc("POST /api/v1/device/setchannelondemand", a.auth(a.setChannelOndemand))
	mux.HandleFunc("POST /api/v1/device/setchannelcloudrecord", a.auth(a.setChannelCloud))
	mux.HandleFunc("POST /api/v1/device/setchannelshared", a.auth(a.setChannelShared))
	mux.HandleFunc("POST /api/v1/device/setchannelptztype", a.auth(a.setChannelPTZ))
	mux.HandleFunc("POST /api/v1/device/setchannelposition", a.auth(a.setChannelPos))
	mux.HandleFunc("POST /api/v1/device/setpresetname", a.auth(a.setPresetName))

	mux.HandleFunc("POST /api/v1/stream/start", a.auth(a.streamStart))
	mux.HandleFunc("POST /api/v1/stream/stop", a.auth(a.streamStop))
	mux.HandleFunc("GET /api/v1/stream/info", a.auth(a.streamInfo))
	mux.HandleFunc("GET /api/v1/stream/list", a.auth(a.streamList))
	mux.HandleFunc("POST /api/v1/playback/start", a.auth(a.playbackStart))
	mux.HandleFunc("POST /api/v1/playback/stop", a.auth(a.streamStop))
	mux.HandleFunc("GET /api/v1/playback/streaminfo", a.auth(a.streamInfo))
	mux.HandleFunc("GET /api/v1/playback/streamlist", a.auth(a.playbackList))
	mux.HandleFunc("GET /api/v1/playback/recordlist", a.auth(a.recordList))
	mux.HandleFunc("POST /api/v1/playback/control", a.auth(a.playbackControl))

	mux.HandleFunc("POST /api/v1/control/ptz", a.auth(a.ptz))
	mux.HandleFunc("POST /api/v1/control/fi", a.auth(a.fi))
	mux.HandleFunc("POST /api/v1/control/preset", a.auth(a.preset))
	mux.HandleFunc("POST /api/v1/control/guard", a.auth(a.guard))
	mux.HandleFunc("POST /api/v1/control/record", a.auth(a.recordCtrl))
	mux.HandleFunc("POST /api/v1/control/teleboot", a.auth(a.teleboot))
	mux.HandleFunc("GET /api/v1/control/teleboot", a.auth(a.teleboot))
	mux.HandleFunc("POST /api/v1/control/iframe", a.auth(a.iframe))
	mux.HandleFunc("GET /api/v1/control/iframe", a.auth(a.iframe))
	mux.HandleFunc("POST /api/v1/control/homeposition", a.auth(a.homePosition))
	mux.HandleFunc("POST /api/v1/control/ptzprecise", a.auth(a.ptzPrecise))
	mux.HandleFunc("POST /api/v1/control/devicesubscribe", a.auth(a.subscribe))
	mux.HandleFunc("POST /api/v1/control/dragzoomin", a.auth(a.dragZoom))
	mux.HandleFunc("POST /api/v1/control/dragzoomout", a.auth(a.dragZoom))
	mux.HandleFunc("POST /api/v1/control/talk", a.auth(a.talk))
	mux.HandleFunc("POST /api/v1/control/alarmreset", a.auth(a.simpleControl("AlarmCmd", "ResetAlarm")))
	mux.HandleFunc("POST /api/v1/control/wiper", a.auth(a.simpleControl("Wiper", "")))
	mux.HandleFunc("POST /api/v1/control/filllight", a.auth(a.simpleControl("FillLight", "")))

	mux.HandleFunc("GET /api/v1/user/list", a.auth(a.userList))
	mux.HandleFunc("GET /api/v1/user/info", a.auth(a.userGet))
	mux.HandleFunc("POST /api/v1/user/save", a.auth(a.userSave))
	mux.HandleFunc("POST /api/v1/user/remove", a.auth(a.userRemove))
	mux.HandleFunc("POST /api/v1/user/setenable", a.auth(a.userEnable))
	mux.HandleFunc("POST /api/v1/user/resetpassword", a.auth(a.userResetPwd))
	mux.HandleFunc("POST /api/v1/user/unlock", a.auth(a.userUnlock))
	mux.HandleFunc("POST /api/v1/user/sethasallchannel", a.auth(a.userAllChannel))
	mux.HandleFunc("POST /api/v1/user/savechannels", a.auth(a.userSaveChannels))
	mux.HandleFunc("POST /api/v1/user/removechannels", a.auth(a.userRemoveChannels))
	mux.HandleFunc("GET /api/v1/user/channellist", a.auth(a.channelList))

	mux.HandleFunc("GET /api/v1/cascade/list", a.auth(a.cascadeList))
	mux.HandleFunc("POST /api/v1/cascade/save", a.auth(a.cascadeSave))
	mux.HandleFunc("POST /api/v1/cascade/remove", a.auth(a.cascadeRemove))
	mux.HandleFunc("POST /api/v1/cascade/setenable", a.auth(a.cascadeEnable))
	mux.HandleFunc("POST /api/v1/cascade/setshareallchannel", a.auth(a.cascadeShareAll))
	mux.HandleFunc("POST /api/v1/cascade/savechannels", a.auth(a.cascadeSaveChannels))
	mux.HandleFunc("POST /api/v1/cascade/removechannels", a.auth(a.cascadeRemoveChannels))
	mux.HandleFunc("GET /api/v1/cascade/channellist", a.auth(a.channelList))

	mux.HandleFunc("GET /api/v1/alarm/list", a.auth(a.alarmList))
	mux.HandleFunc("POST /api/v1/alarm/remove", a.auth(a.alarmRemove))
	mux.HandleFunc("POST /api/v1/alarm/clear", a.auth(a.alarmClear))
	mux.HandleFunc("GET /api/v1/log/list", a.auth(a.logList))
	mux.HandleFunc("POST /api/v1/log/clear", a.auth(a.logClear))
	mux.HandleFunc("POST /api/v1/log/remove", a.auth(a.logClear))

	mux.HandleFunc("GET /api/v1/black/list", a.auth(a.ruleList("black")))
	mux.HandleFunc("POST /api/v1/black/save", a.auth(a.ruleSave("black")))
	mux.HandleFunc("POST /api/v1/black/remove", a.auth(a.ruleRemove("black")))
	mux.HandleFunc("GET /api/v1/white/list", a.auth(a.ruleList("white")))
	mux.HandleFunc("POST /api/v1/white/save", a.auth(a.ruleSave("white")))
	mux.HandleFunc("POST /api/v1/white/remove", a.auth(a.ruleRemove("white")))

	mux.HandleFunc("GET /api/v1/cloudrecord/querydaily", a.auth(a.cloudDaily))
	mux.HandleFunc("GET /api/v1/cloudrecord/querychannels", a.auth(a.channelList))
	mux.HandleFunc("POST /api/v1/cloudrecord/remove", a.auth(a.cloudRemove))
	mux.HandleFunc("POST /api/v1/channel/move", a.auth(a.channelMove))
	mux.HandleFunc("POST /api/v1/channel/remove", a.auth(a.channelRemoveGroup))
	mux.HandleFunc("POST /api/v1/channel/virtual/remove", a.auth(a.channelRemoveGroup))

	mux.HandleFunc("POST /api/internal/sms/register", a.smsRegister)
	mux.HandleFunc("/", a.staticOrAPI)
	return mux
}

func (a *API) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !a.Cfg.APIAuth {
			next(w, r)
			return
		}
		u := a.currentUser(r)
		if u == nil {
			http.Error(w, "未登录或登录已过期", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (a *API) currentUser(r *http.Request) *model.User {
	tok := ""
	if c, err := r.Cookie("gbs_token"); err == nil {
		tok = c.Value
	}
	if tok == "" {
		tok = r.URL.Query().Get("token")
	}
	if tok == "" && r.Method == http.MethodPost {
		_ = r.ParseForm()
		tok = r.Form.Get("token")
	}
	if tok == "" {
		return nil
	}
	u, _, err := a.DB.SessionUser(tok)
	if err != nil || u == nil || !u.Enable {
		return nil
	}
	return u
}

func (a *API) login(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	if username == "" || password == "" {
		http.Error(w, "用户名或密码为空", http.StatusBadRequest)
		return
	}
	u, err := a.DB.GetUserByName(username)
	if err != nil || u.PasswordMD5 != strings.ToLower(password) {
		http.Error(w, "用户名或密码错误", http.StatusUnauthorized)
		return
	}
	if !u.Enable {
		http.Error(w, "用户已停用", http.StatusForbidden)
		return
	}
	if u.Lock {
		http.Error(w, "用户已锁定", http.StatusForbidden)
		return
	}
	token := randToken()
	urlToken := randToken()
	exp := time.Now().Add(7 * 24 * time.Hour)
	_ = a.DB.CreateSession(token, urlToken, u.ID, clientIP(r), exp)
	_ = a.DB.TouchLogin(u.ID)
	http.SetCookie(w, &http.Cookie{Name: "gbs_token", Value: token, Path: "/", HttpOnly: true, MaxAge: 7 * 24 * 3600})
	writeJSON(w, map[string]any{"CookieToken": token, "URLToken": urlToken, "TokenTimeout": 604800})
	a.audit(r, u, "登录", "200")
}

func (a *API) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie("gbs_token"); err == nil {
		a.DB.DeleteSession(c.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "gbs_token", Value: "", Path: "/", MaxAge: -1})
	writeJSON(w, map[string]any{})
}

func (a *API) userinfo(w http.ResponseWriter, r *http.Request) {
	u := a.currentUser(r)
	if u == nil {
		http.Error(w, "未登录或登录已过期", http.StatusUnauthorized)
		return
	}
	roles := strings.Split(u.Role, ",")
	writeJSON(w, map[string]any{
		"ID": u.ID, "Name": u.Username, "Roles": roles,
		"PhoneNumber": u.PhoneNumber, "Email": u.Email, "Description": u.Description,
		"HasAllChannel": u.HasAllChannel, "Cas": false, "OAuth": false,
		"RemoteIP": clientIP(r), "LoginAt": u.LastLoginAt,
	})
}

func (a *API) serverinfo(w http.ResponseWriter, r *http.Request) {
	_, _, chTotal, _ := a.DB.OnlineStats()
	writeJSON(w, map[string]any{
		"Authorization": "ArgusGBS", "Hardware": metrics.Hardware(),
		"InterfaceVersion": "v1", "APIAuth": a.Cfg.APIAuth, "LiveStreamAuth": a.Cfg.LiveStreamAuth,
		"RemainDays": 3650, "RunningTime": metrics.Running(a.Started),
		"ServerTime": model.FormatTime(time.Now()), "StartUpTime": model.FormatTime(a.Started),
		"Server": "ArgusCMS", "PreferStreamFmt": a.Cfg.PreferStreamFmt,
		"ChannelCount": chTotal, "VersionType": "旗舰版",
		"LogoText": a.Cfg.LogoText, "LogoMiniText": a.Cfg.LogoMiniText, "CopyrightText": a.Cfg.CopyrightText,
		"Captcha": a.Cfg.Captcha, "MapEnable": a.Cfg.MapEnable,
		"AllowStreamStartByURL": a.Cfg.AllowStreamStartByURL,
	})
}

func (a *API) modifyPassword(w http.ResponseWriter, r *http.Request) {
	u := a.currentUser(r)
	if u == nil {
		http.Error(w, "未登录或登录已过期", http.StatusUnauthorized)
		return
	}
	_ = r.ParseForm()
	old := r.Form.Get("oldpassword")
	if old == "" {
		old = r.Form.Get("oldPassword")
	}
	np := r.Form.Get("newpassword")
	if np == "" {
		np = r.Form.Get("newPassword")
	}
	if u.PasswordMD5 != strings.ToLower(old) {
		http.Error(w, "原密码错误", http.StatusBadRequest)
		return
	}
	if np == "" {
		http.Error(w, "新密码为空", http.StatusBadRequest)
		return
	}
	_ = a.DB.SetUserPassword(u.ID, strings.ToLower(np))
	writeJSON(w, map[string]any{})
}

func (a *API) restart(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{})
	go func() {
		time.Sleep(300 * time.Millisecond)
		os.Exit(0)
	}()
}

func (a *API) getBaseConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"Serial": a.Cfg.Serial, "Realm": a.Cfg.Realm, "Host": a.SIP.Host(), "Port": a.Cfg.SIPPort,
		"PreferStreamFmt": a.Cfg.PreferStreamFmt, "AckTimeout": a.Cfg.AckTimeout,
		"KeepaliveTimeout": a.Cfg.KeepaliveTimeout, "APIAuth": a.Cfg.APIAuth,
		"LiveStreamAuth": a.Cfg.LiveStreamAuth, "SIPLog": a.Cfg.SIPLog,
		"AllowStreamStartByURL": a.Cfg.AllowStreamStartByURL,
		"DevicePassword": a.Cfg.DevicePassword, "DropChannelType": a.Cfg.DropChannelType,
		"MediaTransport": a.Cfg.DefaultMediaTransport, "MediaTransportMode": a.Cfg.DefaultMediaTransportMode,
		"HTTPSPort": a.Cfg.HTTPSPort, "HTTPSCertFile": a.Cfg.HTTPSCert, "HTTPSKeyFile": a.Cfg.HTTPSKey,
		"MapEnable": a.Cfg.MapEnable, "MapCenter": a.Cfg.MapCenter,
		"GlobalChannelAudio": a.Cfg.GlobalChannelAudio, "GlobalChannelShared": a.Cfg.GlobalChannelShared,
		"GlobalDeviceCatalogSubscribeInterval": a.Cfg.CatalogSubscribeInterval,
		"GlobalDeviceAlarmSubscribeInterval": a.Cfg.AlarmSubscribeInterval,
		"GlobalDevicePositionSubscribeInterval": a.Cfg.PositionSubscribeInterval,
		"GlobalDevicePTZSubscribeInterval": a.Cfg.PTZSubscribeInterval,
		"GM": false,
	})
}

func (a *API) setBaseConfig(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	if v := r.Form.Get("Serial"); v != "" {
		a.Cfg.Serial = v
	}
	if v := r.Form.Get("Realm"); v != "" {
		a.Cfg.Realm = v
	}
	if v := r.Form.Get("Host"); v != "" {
		a.Cfg.SIPHost = v
	}
	if v := r.Form.Get("Port"); v != "" {
		a.Cfg.SIPPort = atoi(v, a.Cfg.SIPPort)
	}
	if v := r.Form.Get("DevicePassword"); v != "" {
		a.Cfg.DevicePassword = v
	}
	if v := r.Form.Get("PreferStreamFmt"); v != "" {
		a.Cfg.PreferStreamFmt = v
	}
	if v := r.Form.Get("MediaTransport"); v != "" {
		a.Cfg.DefaultMediaTransport = v
	}
	if v := r.Form.Get("MediaTransportMode"); v != "" {
		a.Cfg.DefaultMediaTransportMode = v
	}
	a.Cfg.APIAuth = formBool(r, "APIAuth", a.Cfg.APIAuth)
	a.Cfg.LiveStreamAuth = formBool(r, "LiveStreamAuth", a.Cfg.LiveStreamAuth)
	a.Cfg.SIPLog = formBool(r, "SIPLog", a.Cfg.SIPLog)
	if v := r.Form.Get("DropChannelType"); r.Form.Has("DropChannelType") {
		a.Cfg.DropChannelType = v
	}
	if v := r.Form.Get("AckTimeout"); v != "" {
		a.Cfg.AckTimeout = atoi(v, a.Cfg.AckTimeout)
	}
	if v := r.Form.Get("KeepaliveTimeout"); v != "" {
		a.Cfg.KeepaliveTimeout = atoi(v, a.Cfg.KeepaliveTimeout)
	}
	writeJSON(w, map[string]any{})
}

func (a *API) getPwdConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"PwdLength": a.Cfg.PwdLength, "PwdLevel": a.Cfg.PwdLevel, "Captcha": a.Cfg.Captcha,
		"LoginErrorLimit": 5, "LoginErrorLock": 0, "LoginOptTimeout": 0, "PwdExpDays": 0, "LoginLastTip": false,
	})
}

func (a *API) setPwdConfig(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	a.Cfg.Captcha = formBool(r, "Captcha", a.Cfg.Captcha)
	if v := r.Form.Get("PwdLength"); v != "" {
		a.Cfg.PwdLength = atoi(v, a.Cfg.PwdLength)
	}
	writeJSON(w, map[string]any{})
}

func (a *API) smsList(w http.ResponseWriter, r *http.Request) {
	online := a.Media != nil && a.Media.Healthy()
	writeJSON(w, []map[string]any{{
		"Serial": a.Cfg.SMSSerial, "Name": "ArgusSMS", "Host": a.Cfg.SMSHost,
		"Port": a.Cfg.SMSSIPPort, "HTTPPort": a.Cfg.SMSHTTPPort, "Online": online,
	}})
}

func (a *API) smsGet(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"Serial": a.Cfg.SMSSerial, "Realm": a.Cfg.SMSRealm, "Host": a.Cfg.SMSHost,
		"Port": a.Cfg.SMSSIPPort, "WanIP": a.Cfg.SMSPublicHost, "RTSPPort": a.Cfg.SMSRTSPPort,
		"RTMPPort": a.Cfg.SMSRTMPPort, "GroupID": "", "RecordDir": "data/record",
		"CleanOverDays": 0, "CleanFreespacePercent": 5, "CleanFreespaceSize": 5120,
		"TCPPortRange": "30000,30249", "UDPPortRange": "30000,30249", "RTCPortRange": "30250,30500",
		"GOPCache": true, "SIPLog": true, "OutHevc": true, "UseWanIPRecvStream": false,
	})
}

func (a *API) smsSet(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	if v := r.Form.Get("RTSPPort"); v != "" {
		a.Cfg.SMSRTSPPort = atoi(v, a.Cfg.SMSRTSPPort)
	}
	if v := r.Form.Get("RTMPPort"); v != "" {
		a.Cfg.SMSRTMPPort = atoi(v, a.Cfg.SMSRTMPPort)
	}
	if v := r.Form.Get("Host"); v != "" {
		a.Cfg.SMSHost = v
	}
	writeJSON(w, map[string]any{})
}

func (a *API) smsRegister(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"ok": true})
}

func (a *API) dashAuth(w http.ResponseWriter, r *http.Request) {
	dt, do, ct, co := a.DB.OnlineStats()
	writeJSON(w, map[string]any{"data": map[string]any{
		"DeviceTotal": dt, "DeviceOnline": do, "ChannelTotal": ct, "ChannelOnline": co, "ChannelCount": ct,
	}})
}

func (a *API) dashStore(w http.ResponseWriter, r *http.Request) {
	used, free := metrics.DiskGB(".")
	writeJSON(w, map[string]any{"data": []map[string]any{{"Name": "录像盘", "Used": used, "FreeSpace": free}}})
}

func (a *API) dashTop(w http.ResponseWriter, r *http.Request) {
	snap := metrics.Sample()
	streams, _ := a.DB.ListStreams(false)
	playbacks, _ := a.DB.ListStreams(true)
	writeJSON(w, map[string]any{"data": map[string]any{
		"memData": snap.Mem, "cpuData": snap.CPU, "netData": snap.Net,
		"loadData": []map[string]any{
			{"name": "直播", "load": len(streams)},
			{"name": "回放", "load": len(playbacks)},
		},
	}})
}

func (a *API) deviceList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, total, err := a.DB.ListDevices(q.Get("q"), q.Get("device_type"), q.Get("online"), q.Get("sort"), q.Get("order"), atoi(q.Get("start"), 0), atoi(q.Get("limit"), 20))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{"DeviceCount": total, "DeviceList": list, "DeviceNetwork": true, "DeviceRegion": false})
}

func (a *API) deviceInfo(w http.ResponseWriter, r *http.Request) {
	d, err := a.DB.GetDevice(r.URL.Query().Get("serial"))
	if err != nil {
		http.Error(w, "设备不存在", 404)
		return
	}
	writeJSON(w, d)
}

func (a *API) channelList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	serial := q.Get("serial")
	if serial == "" {
		serial = q.Get("device_id")
	}
	list, total, err := a.DB.ListChannels(serial, q.Get("q"), q.Get("dir_serial"), q.Get("sort"), q.Get("order"), atoi(q.Get("start"), 0), atoi(q.Get("limit"), 50))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{"ChannelCount": total, "ChannelList": list})
}

func (a *API) channelInfo(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	c, err := a.DB.GetChannel(q.Get("serial"), q.Get("code"))
	if err != nil {
		http.Error(w, "通道不存在", 404)
		return
	}
	writeJSON(w, c)
}

func (a *API) channelTree(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, err := a.DB.ChannelTree(q.Get("serial"), q.Get("pcode"))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, list)
}

func (a *API) groupTree(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, err := a.DB.GroupTree(q.Get("serial"), q.Get("pcode"))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, list)
}

func (a *API) onlineStats(w http.ResponseWriter, r *http.Request) {
	dt, do, ct, co := a.DB.OnlineStats()
	writeJSON(w, map[string]any{"DeviceTotal": dt, "DeviceOnline": do, "ChannelTotal": ct, "ChannelOnline": co})
}

func (a *API) deviceRemove(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	if err := a.DB.RemoveDevice(r.Form.Get("serial")); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	writeJSON(w, map[string]any{})
}

func (a *API) deviceSetInfo(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	id := r.Form.Get("ID")
	if id == "" {
		id = r.Form.Get("serial")
	}
	d, err := a.DB.GetDevice(id)
	if err != nil {
		http.Error(w, "设备不存在", 404)
		return
	}
	if v := r.Form.Get("CustomName"); v != "" || r.Form.Has("CustomName") {
		d.CustomName = r.Form.Get("CustomName")
	}
	if v := r.Form.Get("Name"); v != "" {
		d.Name = v
	}
	if v := r.Form.Get("MediaTransport"); v != "" {
		d.MediaTransport = v
	}
	if v := r.Form.Get("MediaTransportMode"); v != "" {
		d.MediaTransportMode = v
	}
	if r.Form.Has("Password") {
		d.Password = r.Form.Get("Password")
	}
	if v := r.Form.Get("SMSID"); r.Form.Has("SMSID") {
		d.SMSID = v
	}
	_ = a.DB.UpdateDevice(d)
	writeJSON(w, map[string]any{})
}

func (a *API) deviceSetName(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	_ = a.DB.SetDeviceName(r.Form.Get("serial"), r.Form.Get("name"))
	writeJSON(w, map[string]any{})
}

func (a *API) deviceSetTransport(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	_ = a.DB.SetMediaTransport(r.Form.Get("serial"), r.Form.Get("media_transport"), r.Form.Get("media_transport_mode"))
	writeJSON(w, map[string]any{})
}

func (a *API) deviceSetSMS(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	_ = a.DB.SetDeviceSMS(r.Form.Get("serial"), r.Form.Get("sms_id"), r.Form.Get("sms_group_id"))
	writeJSON(w, map[string]any{})
}

func (a *API) fetchCatalog(w http.ResponseWriter, r *http.Request) {
	serial := r.URL.Query().Get("serial")
	if err := a.SIP.QueryCatalog(serial); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{})
}

func (a *API) fetchInfo(w http.ResponseWriter, r *http.Request) {
	serial := r.URL.Query().Get("serial")
	env, err := a.SIP.QueryXML(serial, "DeviceInfo", serial, nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, env)
}

func (a *API) fetchStatus(w http.ResponseWriter, r *http.Request) {
	serial := r.URL.Query().Get("serial")
	code := r.URL.Query().Get("code")
	if code == "" {
		code = serial
	}
	env, err := a.SIP.QueryXML(serial, "DeviceStatus", code, nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, env)
}

func (a *API) fetchPreset(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	env, err := a.SIP.QueryXML(q.Get("serial"), "PresetQuery", q.Get("code"), nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{"PresetList": env.DeviceList.Item, "PresetItemList": env.DeviceList.Item})
}

func (a *API) statusLog(w http.ResponseWriter, r *http.Request) {
	list, err := a.DB.StatusLogs(r.URL.Query().Get("serial"), atoi(r.URL.Query().Get("limit"), 100))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{"LogCount": len(list), "LogList": list})
}

func (a *API) streamLog(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"LogCount": 0, "LogList": []any{}})
}

func (a *API) positionLog(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"LogCount": 0, "LogList": []any{}})
}

func (a *API) setChannelName(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	_ = a.DB.SetChannelField(r.Form.Get("serial"), r.Form.Get("code"), "custom_name", r.Form.Get("name"))
	writeJSON(w, map[string]any{})
}

func (a *API) setChannelCustom(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	_ = a.DB.SetChannelField(r.Form.Get("serial"), r.Form.Get("code"), "custom_id", r.Form.Get("id"))
	writeJSON(w, map[string]any{})
}

func (a *API) setChannelAudio(w http.ResponseWriter, r *http.Request) {
	a.setChannelBool(w, r, "audio_enable")
}
func (a *API) setChannelOndemand(w http.ResponseWriter, r *http.Request) {
	a.setChannelBool(w, r, "ondemand")
}
func (a *API) setChannelCloud(w http.ResponseWriter, r *http.Request) {
	a.setChannelBool(w, r, "cloud_record")
}
func (a *API) setChannelShared(w http.ResponseWriter, r *http.Request) {
	a.setChannelBool(w, r, "shared")
}

func (a *API) setChannelBool(w http.ResponseWriter, r *http.Request, col string) {
	_ = r.ParseForm()
	val := 0
	if formBool(r, "enable", false) || formBool(r, "value", false) || r.Form.Get(col) == "true" || r.Form.Get("audio") == "true" || r.Form.Get("ondemand") == "true" || r.Form.Get("cloudrecord") == "true" || r.Form.Get("shared") == "true" {
		val = 1
	}
	// 前端开关直接传布尔字段名不固定，true/false 字符串也接受
	for _, k := range []string{"enable", "value", "audio", "ondemand", "cloudrecord", "shared", "AudioEnable", "Ondemand", "CloudRecord", "Shared"} {
		if r.Form.Get(k) == "true" || r.Form.Get(k) == "1" {
			val = 1
		}
		if r.Form.Get(k) == "false" || r.Form.Get(k) == "0" {
			val = 0
		}
	}
	_ = a.DB.SetChannelField(r.Form.Get("serial"), r.Form.Get("code"), col, val)
	writeJSON(w, map[string]any{})
}

func (a *API) setChannelPTZ(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	_ = a.DB.SetChannelField(r.Form.Get("serial"), r.Form.Get("code"), "custom_ptz_type", atoi(r.Form.Get("ptz_type"), 0))
	writeJSON(w, map[string]any{})
}

func (a *API) setChannelPos(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	serial, code := r.Form.Get("serial"), r.Form.Get("code")
	lon, _ := strconv.ParseFloat(r.Form.Get("longitude"), 64)
	lat, _ := strconv.ParseFloat(r.Form.Get("latitude"), 64)
	_ = a.DB.SetChannelField(serial, code, "custom_longitude", lon)
	_ = a.DB.SetChannelField(serial, code, "custom_latitude", lat)
	writeJSON(w, map[string]any{})
}

func (a *API) setPresetName(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	_ = a.DB.SetPresetName(r.Form.Get("serial"), r.Form.Get("code"), atoi(r.Form.Get("preset"), 1), r.Form.Get("name"))
	writeJSON(w, map[string]any{})
}

func (a *API) streamStart(w http.ResponseWriter, r *http.Request) {
	a.play(w, r, false)
}

func (a *API) playbackStart(w http.ResponseWriter, r *http.Request) {
	a.play(w, r, true)
}

func (a *API) play(w http.ResponseWriter, r *http.Request, playback bool) {
	_ = r.ParseForm()
	serial := first(r, "serial")
	code := first(r, "code")
	ch, err := a.DB.GetChannel(serial, code)
	if err != nil {
		http.Error(w, "通道不存在", 404)
		return
	}
	if !playback {
		if old, err := a.DB.FindLiveStream(serial, code); err == nil && old != nil && !old.Stopped {
			a.fillStats(old)
			writeJSON(w, old)
			return
		}
	}
	dev, err := a.DB.GetDevice(serial)
	if err != nil || !dev.Online {
		http.Error(w, "设备离线", 400)
		return
	}
	transport := first(r, "transport")
	mode := first(r, "transport_mode")
	if transport == "" || transport == "config" {
		transport = dev.MediaTransport
	}
	if mode == "" || mode == "config" {
		mode = dev.MediaTransportMode
	}
	if transport == "" {
		transport = a.Cfg.DefaultMediaTransport
	}
	if mode == "" {
		mode = a.Cfg.DefaultMediaTransportMode
	}
	ssrc := sip.NewSSRC(playback)
	if a.Media == nil {
		http.Error(w, "流媒体服务尚未启动", 500)
		return
	}
	opened, err := a.Media.Open(media.OpenReq{
		StreamID: ssrc, Transport: transport, Mode: mode, SSRC: ssrc,
		Width: atoi(first(r, "width"), 0), Height: atoi(first(r, "height"), 0),
		Bitrate: atoi(first(r, "bitrate"), 0), FrameRate: atoi(first(r, "framerate"), 0),
		VideoCodec: first(r, "video_codec"),
	})
	if err != nil {
		http.Error(w, "流媒体服务尚未启动", 500)
		return
	}
	recvIP := opened.PublicIP
	if recvIP == "" {
		recvIP = opened.IP
	}
	if dev.RecvStreamIP != "" {
		recvIP = dev.RecvStreamIP
	}
	opt := sip.PlayOpt{
		DeviceID: serial, ChannelID: code, Transport: transport, Mode: mode,
		SSRC: ssrc, RecvIP: recvIP, RecvPort: opened.Port, Download: formBool(r, "download", false),
		Start: first(r, "starttime"), End: first(r, "endtime"),
	}
	callID, _, err := a.SIP.Invite(opt)
	if err != nil {
		_ = a.Media.Close(ssrc)
		http.Error(w, err.Error(), 500)
		return
	}
	st := a.buildStream(r, ssrc, dev, ch, transport, callID, playback, opt.Start, opt.End)
	_ = a.DB.SaveStream(st)
	writeJSON(w, st)
	a.audit(r, a.currentUser(r), "开始播放", "200")
}

func (a *API) buildStream(r *http.Request, ssrc string, dev *model.Device, ch *model.Channel, transport, callID string, playback bool, start, end string) *model.Stream {
	host := a.Cfg.SMSPublicHost
	if host == "" {
		host = hostOnly(r.Host)
	}
	if host == "" || host == "0.0.0.0" {
		host = a.Cfg.SMSHost
	}
	base := fmt.Sprintf("http://%s:%d", host, a.Cfg.SMSHTTPPort)
	ws := fmt.Sprintf("ws://%s:%d", host, a.Cfg.SMSHTTPPort)
	name := ch.Name
	if ch.CustomName != "" {
		name = ch.CustomName
	}
	return &model.Stream{
		StreamID: ssrc, SMSID: a.Cfg.SMSSerial, DeviceID: dev.ID, ChannelID: ch.ID, ChannelName: name,
		Transport: transport, StartAt: model.FormatTime(time.Now()), CallID: callID, SSRC: ssrc,
		Playback: playback, StartTime: start, EndTime: end, AudioEnable: ch.AudioEnable,
		Ondemand: ch.Ondemand, CloudRecord: ch.CloudRecord, ChannelPTZType: ch.PTZType,
		FLV: base + "/live/" + ssrc + ".flv", WSFLV: ws + "/live/" + ssrc + ".flv",
		HLS:    base + "/live/" + ssrc + "/index.m3u8",
		RTMP:   fmt.Sprintf("rtmp://%s:%d/live/%s", host, a.Cfg.SMSRTMPPort, ssrc),
		RTSP:   fmt.Sprintf("rtsp://%s:%d/live/%s", host, a.Cfg.SMSRTSPPort, ssrc),
		WEBRTC: base + "/webrtc/play?stream=" + ssrc,
		WHEP:   base + "/whep/" + ssrc,
		SnapURL: ch.SnapURL,
	}
}

func (a *API) fillStats(st *model.Stream) {
	if a.Media == nil {
		return
	}
	s, err := a.Media.Stats(st.StreamID)
	if err != nil {
		return
	}
	st.SourceVideoCodecName = s.Codec
	st.SourceVideoWidth = s.Width
	st.SourceVideoHeight = s.Height
	st.SourceVideoFrameRate = s.FPS
	st.RTPCount = s.RTPCount
	st.RTPLostCount = s.RTPLost
	st.InBytes = s.InBytes
	st.InBitRate = s.InBitRate
	st.VideoFrameCount = s.VideoFrames
	st.NumOutputs = s.NumOutputs
	st.SourceAudioCodecName = s.AudioCodec
	if st.StartAt != "" {
		if t := model.ParseTime(st.StartAt); !t.IsZero() {
			st.Duration = int(time.Since(t).Seconds())
		}
	}
}

func (a *API) streamStop(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	id := first(r, "streamid")
	if id == "" {
		id = first(r, "serial")
		if code := first(r, "code"); code != "" {
			if st, err := a.DB.FindLiveStream(id, code); err == nil {
				id = st.StreamID
			}
		}
	}
	st, err := a.DB.StopStream(id)
	if err == nil && st != nil {
		a.SIP.Bye(st.DeviceID, st.ChannelID, st.CallID)
		if a.Media != nil {
			_ = a.Media.Close(st.StreamID)
		}
	}
	writeJSON(w, map[string]any{})
}

func (a *API) streamInfo(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("streamid")
	st, err := a.DB.GetStream(id)
	if err != nil {
		http.Error(w, "流不存在", 404)
		return
	}
	a.fillStats(st)
	writeJSON(w, st)
}

func (a *API) streamList(w http.ResponseWriter, r *http.Request) {
	list, _ := a.DB.ListStreams(false)
	for i := range list {
		a.fillStats(&list[i])
	}
	writeJSON(w, map[string]any{"StreamCount": len(list), "StreamList": list})
}

func (a *API) playbackList(w http.ResponseWriter, r *http.Request) {
	list, _ := a.DB.ListStreams(true)
	writeJSON(w, map[string]any{"StreamCount": len(list), "StreamList": list})
}

func (a *API) recordList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	serial, code := q.Get("serial"), q.Get("code")
	start, end := q.Get("starttime"), q.Get("endtime")
	extra := [][2]string{{"StartTime", start}, {"EndTime", end}}
	if q.Get("streamnumber") != "" {
		extra = append(extra, [2]string{"StreamNumber", q.Get("streamnumber")})
	}
	env, err := a.SIP.QueryXML(serial, "RecordInfo", code, extra)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var list []model.RecordItem
	for _, it := range env.RecordList.Item {
		list = append(list, model.RecordItem{
			DeviceID: it.DeviceID, Name: it.Name, FilePath: it.FilePath, Address: it.Address,
			StartTime: it.StartTime, EndTime: it.EndTime, Type: it.Type, RecorderID: it.RecorderID,
			FileSize: it.FileSize, StreamNumber: it.StreamNumber, Secrecy: strconv.Itoa(it.Secrecy),
		})
	}
	if list == nil {
		list = []model.RecordItem{}
	}
	writeJSON(w, map[string]any{"DeviceID": code, "Name": "", "SumNum": env.SumNum, "RecordList": list})
}

func (a *API) playbackControl(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	st, err := a.DB.GetStream(first(r, "streamid"))
	if err != nil {
		http.Error(w, "流不存在", 404)
		return
	}
	cmd := strings.ToLower(first(r, "command"))
	scale := first(r, "scale")
	rng := first(r, "range")
	var body string
	switch cmd {
	case "pause":
		body = "PAUSE RTSP/1.0\r\nCSeq: 1\r\nPauseTime: now\r\n"
	case "teardown":
		a.SIP.Bye(st.DeviceID, st.ChannelID, st.CallID)
		writeJSON(w, map[string]any{})
		return
	default:
		if scale == "" {
			scale = "1.0"
		}
		if rng == "" {
			rng = "npt=now-"
		} else if !strings.Contains(rng, "=") {
			rng = "npt=" + rng
		}
		body = fmt.Sprintf("PLAY RTSP/1.0\r\nCSeq: 2\r\nScale: %s\r\nRange: %s\r\n", scale, rng)
	}
	if err := a.SIP.Info(st.DeviceID, st.ChannelID, st.CallID, body); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{})
}

func (a *API) ptz(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	cmd := manscdp.PTZCmd(first(r, "command"), atoi(first(r, "speed"), 80))
	if err := a.SIP.Control(first(r, "serial"), first(r, "code"), [][2]string{{"PTZCmd", cmd}}); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{})
}

func (a *API) fi(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	cmd := manscdp.FICmd(first(r, "command"), atoi(first(r, "speed"), 80))
	if err := a.SIP.Control(first(r, "serial"), first(r, "code"), [][2]string{{"PTZCmd", cmd}}); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{})
}

func (a *API) preset(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	cmd := manscdp.PresetCmd(first(r, "command"), atoi(first(r, "preset"), 1))
	if err := a.SIP.Control(first(r, "serial"), first(r, "code"), [][2]string{{"PTZCmd", cmd}}); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{})
}

func (a *API) guard(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	cmd := "SetGuard"
	if strings.Contains(strings.ToLower(first(r, "command")), "reset") || first(r, "command") == "0" {
		cmd = "ResetGuard"
	}
	if err := a.SIP.Control(first(r, "serial"), first(r, "code"), [][2]string{{"GuardCmd", cmd}}); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{})
}

func (a *API) recordCtrl(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	cmd := "Record"
	if strings.Contains(strings.ToLower(first(r, "command")), "stop") {
		cmd = "StopRecord"
	}
	if err := a.SIP.Control(first(r, "serial"), first(r, "code"), [][2]string{{"RecordCmd", cmd}}); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{})
}

func (a *API) teleboot(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	serial := first(r, "serial")
	if err := a.SIP.Control(serial, serial, [][2]string{{"TeleBoot", "Boot"}}); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{})
}

func (a *API) iframe(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	if err := a.SIP.Control(first(r, "serial"), first(r, "code"), [][2]string{{"IFrameCmd", "Send"}}); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{})
}

func (a *API) homePosition(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	err := a.SIP.Control(first(r, "serial"), first(r, "code"), [][2]string{
		{"HomePosition", "1"},
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{})
}

func (a *API) ptzPrecise(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	err := a.SIP.Control(first(r, "serial"), first(r, "code"), [][2]string{
		{"PTZPreciseCtrl", "1"},
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{})
}

func (a *API) subscribe(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	cmd := first(r, "command")
	if cmd == "" {
		cmd = "Catalog"
	}
	exp := atoi(first(r, "expires"), 3600)
	if err := a.SIP.Subscribe(first(r, "serial"), cmd, exp); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{})
}

func (a *API) dragZoom(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	tag := "DragZoomIn"
	if strings.Contains(r.URL.Path, "out") {
		tag = "DragZoomOut"
	}
	err := a.SIP.Control(first(r, "serial"), first(r, "code"), [][2]string{{tag, "1"}})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{})
}

func (a *API) talk(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"Audio": "websocket"})
}

func (a *API) simpleControl(field, value string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		v := value
		if v == "" {
			v = first(r, "command")
		}
		if err := a.SIP.Control(first(r, "serial"), first(r, "code"), [][2]string{{field, v}}); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, map[string]any{})
	}
}

func randToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
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

func formBool(r *http.Request, key string, def bool) bool {
	if !r.Form.Has(key) && r.URL.Query().Get(key) == "" {
		return def
	}
	v := r.Form.Get(key)
	if v == "" {
		v = r.URL.Query().Get(key)
	}
	return v == "1" || strings.EqualFold(v, "true") || v == "on"
}

func first(r *http.Request, key string) string {
	if v := r.Form.Get(key); v != "" {
		return v
	}
	return r.URL.Query().Get(key)
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func hostOnly(hostport string) string {
	h, _, err := net.SplitHostPort(hostport)
	if err != nil {
		return hostport
	}
	return h
}

func (a *API) audit(r *http.Request, u *model.User, name, status string) {
	user := ""
	if u != nil {
		user = u.Username
	}
	a.DB.AddLog(&model.OpLog{
		Name: name, Method: r.Method, RequestURI: r.URL.RequestURI(), RemoteAddr: clientIP(r),
		Status: status, Username: user, StartAt: model.FormatTime(time.Now()),
	})
}

func (a *API) staticOrAPI(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		http.Error(w, "请求服务不存在或已停止", 404)
		return
	}
	if a.WWW == "" {
		http.Error(w, "前端目录未配置", 404)
		return
	}
	rel := r.URL.Path
	switch rel {
	case "/", "":
		rel = "/index.html"
	case "/login", "/login/":
		rel = "/login.html"
	case "/play", "/playback", "/map", "/test":
		rel = rel + ".html"
	}
	path := filepath.Join(a.WWW, filepath.Clean("/"+rel))
	if !strings.HasPrefix(path, filepath.Clean(a.WWW)) {
		http.Error(w, "forbidden", 403)
		return
	}
	if st, err := os.Stat(path); err != nil || st.IsDir() {
		// 前端历史路由回退到首页
		http.ServeFile(w, r, filepath.Join(a.WWW, "index.html"))
		return
	}
	http.ServeFile(w, r, path)
}
