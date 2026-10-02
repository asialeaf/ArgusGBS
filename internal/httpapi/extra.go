package httpapi

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/argus/argusgbs/internal/model"
)

func (a *API) userList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, total, err := a.DB.ListUsers(q.Get("q"), q.Get("enable"), q.Get("sort"), q.Get("order"), atoi(q.Get("start"), 0), atoi(q.Get("limit"), 20))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{"UserCount": total, "UserList": list})
}

func (a *API) userGet(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	u, err := a.DB.GetUser(id)
	if err != nil {
		http.Error(w, "用户不存在", 404)
		return
	}
	writeJSON(w, u)
}

func (a *API) userSave(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	id, _ := strconv.ParseInt(r.Form.Get("ID"), 10, 64)
	u := &model.User{
		ID: id, Username: r.Form.Get("Username"), Role: r.Form.Get("Role"),
		ControlPriority: atoi(r.Form.Get("ControlPriority"), 0),
		PhoneNumber:     r.Form.Get("PhoneNumber"), Email: r.Form.Get("Email"),
		Description: r.Form.Get("Description"), Enable: formBool(r, "Enable", true),
		HasAllChannel: true,
	}
	me := a.currentUser(r)
	if me != nil {
		u.Creator = me.Username
		if strings.Contains(u.Role, roleSuper) && !strings.Contains(me.Role, roleSuper) {
			http.Error(w, "没有权限", http.StatusForbidden)
			return
		}
	}
	pwd := ""
	plain := ""
	if id == 0 {
		plain = "Argus@123"
		sum := md5.Sum([]byte(plain))
		pwd = hex.EncodeToString(sum[:])
	}
	nid, err := a.DB.SaveUser(u, pwd)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	writeJSON(w, map[string]any{"ID": nid, "DefaultUserPassword": plain})
}

func (a *API) userRemove(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	id, _ := strconv.ParseInt(first(r, "id"), 10, 64)
	if err := a.DB.RemoveUser(id); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	writeJSON(w, map[string]any{})
}

func (a *API) userEnable(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	id, _ := strconv.ParseInt(first(r, "id"), 10, 64)
	_ = a.DB.SetUserEnable(id, formBool(r, "enable", true))
	writeJSON(w, map[string]any{})
}

func (a *API) userResetPwd(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	me := a.currentUser(r)
	if me == nil {
		http.Error(w, "未登录或登录已过期", http.StatusUnauthorized)
		return
	}
	idStr := first(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	plain := r.Form.Get("password")
	timestamp := r.Form.Get("timestamp")
	verify := strings.ToLower(r.Form.Get("verify"))
	if plain == "" || timestamp == "" || verify == "" {
		http.Error(w, "缺少密码校验", http.StatusBadRequest)
		return
	}
	ts, err := time.ParseInLocation("20060102150405", timestamp, time.Local)
	if err != nil || time.Since(ts) > 10*time.Minute || ts.After(time.Now().Add(2*time.Minute)) {
		http.Error(w, "校验已过期，请重试", http.StatusBadRequest)
		return
	}
	expect := md5hex(idStr + plain + timestamp + strings.ToLower(me.PasswordMD5))
	if !ctEqual(expect, verify) {
		http.Error(w, "我的密码不正确", http.StatusBadRequest)
		return
	}
	target, err := a.DB.GetUser(id)
	if err != nil || target == nil {
		http.Error(w, "用户不存在", http.StatusNotFound)
		return
	}
	if strings.Contains(target.Role, roleSuper) && !hasAnyRole(me, roleSuper) {
		http.Error(w, "没有权限", http.StatusForbidden)
		return
	}
	if msg := a.passwordOK(plain); msg != "" {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	if err := a.DB.SetUserPassword(id, md5hex(plain)); err != nil {
		http.Error(w, "保存密码失败", http.StatusInternalServerError)
		return
	}
	a.DB.DeleteUserSessions(id)
	writeJSON(w, map[string]any{})
	a.audit(r, me, "重置密码", "200")
}

func (a *API) userUnlock(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	id, _ := strconv.ParseInt(first(r, "id"), 10, 64)
	_ = a.DB.UnlockUser(id)
	writeJSON(w, map[string]any{})
}

func (a *API) userAllChannel(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	id, _ := strconv.ParseInt(first(r, "id"), 10, 64)
	_ = a.DB.SetUserHasAll(id, formBool(r, "hasallchannel", true) || formBool(r, "shared", true))
	writeJSON(w, map[string]any{})
}

func (a *API) userSaveChannels(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	id, _ := strconv.ParseInt(first(r, "id"), 10, 64)
	_ = a.DB.SaveUserChannels(id, parsePairs(r))
	writeJSON(w, map[string]any{})
}

func (a *API) userRemoveChannels(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	id, _ := strconv.ParseInt(first(r, "id"), 10, 64)
	_ = a.DB.RemoveUserChannels(id, parsePairs(r))
	writeJSON(w, map[string]any{})
}

func parsePairs(r *http.Request) [][2]string {
	serials := r.Form["serial"]
	codes := r.Form["code"]
	var out [][2]string
	n := len(serials)
	if len(codes) < n {
		n = len(codes)
	}
	for i := 0; i < n; i++ {
		out = append(out, [2]string{serials[i], codes[i]})
	}
	if len(out) == 0 {
		if s, c := r.Form.Get("serial"), r.Form.Get("code"); s != "" && c != "" {
			out = append(out, [2]string{s, c})
		}
	}
	return out
}

func (a *API) cascadeList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, total, err := a.DB.ListCascades(q.Get("q"), q.Get("online"), q.Get("enable"), q.Get("sort"), q.Get("order"), atoi(q.Get("start"), 0), atoi(q.Get("limit"), 20))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{"CascadeCount": total, "CascadeList": list})
}

func (a *API) cascadeSave(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	c := &model.Cascade{
		ID: r.Form.Get("ID"), Name: r.Form.Get("Name"), Serial: r.Form.Get("Serial"), Realm: r.Form.Get("Realm"),
		Host: r.Form.Get("Host"), Port: atoi(r.Form.Get("Port"), 5060),
		LocalSerial: r.Form.Get("LocalSerial"), LocalHost: r.Form.Get("LocalHost"), LocalPort: atoi(r.Form.Get("LocalPort"), 0),
		Username: r.Form.Get("Username"), Password: r.Form.Get("Password"),
		KeepaliveMaxCount: atoi(r.Form.Get("KeepaliveMaxCount"), 3), KeepaliveInterval: atoi(r.Form.Get("KeepaliveInterval"), 60),
		RegisterInterval: atoi(r.Form.Get("RegisterInterval"), 60), RegisterTimeout: atoi(r.Form.Get("RegisterTimeout"), 3600),
		ControlPriority: atoi(r.Form.Get("ControlPriority"), 0), LoadLimit: atoi(r.Form.Get("LoadLimit"), 0),
		CatalogGroupSize: atoi(r.Form.Get("CatalogGroupSize"), 1), CommandTransport: r.Form.Get("CommandTransport"),
		Charset: r.Form.Get("Charset"), AllowControl: formBool(r, "AllowControl", true),
		ShareAllChannel: formBool(r, "ShareAllChannel", false), ShareRecord: formBool(r, "ShareRecord", true),
		StreamKeepalive: formBool(r, "StreamKeepalive", false), StreamReader: formBool(r, "StreamReader", false),
		BindLocalIP: formBool(r, "BindLocalIP", false), Enable: formBool(r, "Enable", true),
	}
	if c.CommandTransport == "" {
		c.CommandTransport = "UDP"
	}
	if c.Charset == "" {
		c.Charset = "GB2312"
	}
	if err := a.DB.SaveCascade(c); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	writeJSON(w, map[string]any{"ID": c.ID})
}

func (a *API) cascadeRemove(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	_ = a.DB.RemoveCascade(first(r, "id"))
	writeJSON(w, map[string]any{})
}

func (a *API) cascadeEnable(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	_ = a.DB.SetCascadeEnable(first(r, "id"), formBool(r, "enable", true))
	writeJSON(w, map[string]any{})
}

func (a *API) cascadeShareAll(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	_ = a.DB.SetCascadeShareAll(first(r, "id"), formBool(r, "shareallchannel", false) || formBool(r, "shared", false))
	writeJSON(w, map[string]any{})
}

func (a *API) cascadeSaveChannels(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	_ = a.DB.SaveCascadeChannels(first(r, "id"), parsePairs(r))
	writeJSON(w, map[string]any{})
}

func (a *API) cascadeRemoveChannels(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	_ = a.DB.RemoveCascadeChannels(first(r, "id"), parsePairs(r))
	writeJSON(w, map[string]any{})
}

func (a *API) alarmList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, total, err := a.DB.ListAlarms(q.Get("q"), q.Get("serial"), atoi(q.Get("start"), 0), atoi(q.Get("limit"), 20))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{
		"AlarmCount": total, "AlarmList": list, "AlarmReserveDays": 3, "AlarmPublishToRedis": false,
	})
}

func (a *API) alarmRemove(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	_ = a.DB.RemoveAlarm(first(r, "id"))
	writeJSON(w, map[string]any{})
}

func (a *API) alarmClear(w http.ResponseWriter, r *http.Request) {
	_ = a.DB.ClearAlarms()
	writeJSON(w, map[string]any{})
}

func (a *API) logList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, total, err := a.DB.ListLogs(q.Get("q"), q.Get("method"), q.Get("starttime"), q.Get("endtime"), q.Get("sort"), q.Get("order"), atoi(q.Get("start"), 0), atoi(q.Get("limit"), 20))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{"LogCount": total, "LogList": list})
}

func (a *API) logClear(w http.ResponseWriter, r *http.Request) {
	_ = a.DB.ClearLogs()
	writeJSON(w, map[string]any{})
}

func (a *API) ruleList(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		list, total, err := a.DB.ListRules(kind, q.Get("q"), atoi(q.Get("start"), 0), atoi(q.Get("limit"), 20))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		key := "Black"
		if kind == "white" {
			key = "White"
		}
		writeJSON(w, map[string]any{key + "Count": total, key + "List": list})
	}
}

func (a *API) ruleSave(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		id, _ := strconv.ParseInt(r.Form.Get("ID"), 10, 64)
		rule := &model.AccessRule{
			ID: id, Serial: r.Form.Get("Serial"), IP: r.Form.Get("IP"), UA: r.Form.Get("UA"),
			Password: r.Form.Get("Password"), Description: r.Form.Get("Description"),
		}
		if rule.Serial == "" {
			rule.Serial = r.Form.Get("serial")
		}
		nid, err := a.DB.SaveRule(kind, rule)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		writeJSON(w, map[string]any{"ID": nid})
	}
}

func (a *API) ruleRemove(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		id, _ := strconv.ParseInt(first(r, "id"), 10, 64)
		_ = a.DB.RemoveRule(kind, id)
		writeJSON(w, map[string]any{})
	}
}

func (a *API) cloudDaily(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	day := q.Get("day")
	if day == "" {
		day = q.Get("period")
	}
	list, err := a.DB.ListCloudByDay(q.Get("serial"), q.Get("code"), day)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]any{"RecordList": list, "SumNum": len(list)})
}

func (a *API) cloudRemove(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	_ = a.DB.RemoveCloud(first(r, "id"))
	writeJSON(w, map[string]any{})
}

func (a *API) channelMove(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	serial := first(r, "serial")
	code := first(r, "code")
	parent := first(r, "parent")
	id := serial + ":" + code
	name := first(r, "name")
	if name == "" {
		name = code
	}
	_ = a.DB.SaveGroup(model.GroupNode{ID: id, ParentID: parent, Name: name, Serial: serial, Code: code, CustomID: code})
	writeJSON(w, map[string]any{})
}

func (a *API) channelRemoveGroup(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	id := first(r, "serial") + ":" + first(r, "code")
	_ = a.DB.RemoveGroup(id)
	writeJSON(w, map[string]any{})
}

