package store

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/argus/argusgbs/internal/model"

	_ "modernc.org/sqlite"
)

type Store struct {
	mu sync.Mutex
	db *sql.DB
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return nil, err
	}
	db, err := sql.Open("sqlite", path+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	if err := s.seedAdmin(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT UNIQUE,
  password_md5 TEXT,
  role TEXT,
  control_priority INTEGER DEFAULT 0,
  phone TEXT DEFAULT '',
  email TEXT DEFAULT '',
  description TEXT DEFAULT '',
  creator TEXT DEFAULT '',
  locked INTEGER DEFAULT 0,
  enable INTEGER DEFAULT 1,
  has_all_channel INTEGER DEFAULT 1,
  last_login_at TEXT DEFAULT '',
  updated_at TEXT,
  created_at TEXT
);
CREATE TABLE IF NOT EXISTS sessions (
  token TEXT PRIMARY KEY,
  url_token TEXT,
  user_id INTEGER,
  remote_ip TEXT,
  expire_at INTEGER,
  created_at TEXT
);
CREATE TABLE IF NOT EXISTS devices (
  id TEXT PRIMARY KEY,
  name TEXT DEFAULT '',
  custom_name TEXT DEFAULT '',
  manufacturer TEXT DEFAULT '',
  model TEXT DEFAULT '',
  firmware TEXT DEFAULT '',
  gb_ver TEXT DEFAULT '2016',
  dev_type TEXT DEFAULT 'GB',
  channel_count INTEGER DEFAULT 0,
  recv_stream_ip TEXT DEFAULT '',
  contact_ip TEXT DEFAULT '',
  contact_port INTEGER DEFAULT 0,
  drop_channel_type TEXT DEFAULT '',
  sms_id TEXT DEFAULT '',
  sms_group_id TEXT DEFAULT '',
  catalog_interval INTEGER DEFAULT 0,
  subscribe_interval INTEGER DEFAULT 0,
  catalog_subscribe INTEGER DEFAULT 0,
  alarm_subscribe INTEGER DEFAULT 0,
  position_subscribe INTEGER DEFAULT 0,
  ptz_subscribe INTEGER DEFAULT 0,
  online INTEGER DEFAULT 0,
  password TEXT DEFAULT '',
  record_center INTEGER DEFAULT 0,
  record_indistinct INTEGER DEFAULT 0,
  civil_code_first INTEGER DEFAULT 0,
  keep_original_tree INTEGER DEFAULT 0,
  command_transport TEXT DEFAULT 'UDP',
  media_transport TEXT DEFAULT 'UDP',
  media_transport_mode TEXT DEFAULT 'passive',
  remote_ip TEXT DEFAULT '',
  remote_port INTEGER DEFAULT 0,
  longitude REAL DEFAULT 0,
  latitude REAL DEFAULT 0,
  last_register_at TEXT DEFAULT '',
  last_keepalive_at TEXT DEFAULT '',
  updated_at TEXT,
  created_at TEXT,
  catalog_progress TEXT DEFAULT '',
  ua TEXT DEFAULT '',
  charset TEXT DEFAULT 'GB2312'
);
CREATE TABLE IF NOT EXISTS channels (
  device_id TEXT,
  id TEXT,
  name TEXT DEFAULT '',
  custom_name TEXT DEFAULT '',
  manufacturer TEXT DEFAULT '',
  custom_manufacturer TEXT DEFAULT '',
  model TEXT DEFAULT '',
  custom_model TEXT DEFAULT '',
  owner TEXT DEFAULT '',
  civil_code TEXT DEFAULT '',
  custom_civil_code TEXT DEFAULT '',
  address TEXT DEFAULT '',
  custom_address TEXT DEFAULT '',
  firmware TEXT DEFAULT '',
  custom_firmware TEXT DEFAULT '',
  serial_number TEXT DEFAULT '',
  custom_serial_number TEXT DEFAULT '',
  ip_address TEXT DEFAULT '',
  custom_ip TEXT DEFAULT '',
  port INTEGER DEFAULT 0,
  custom_port INTEGER DEFAULT 0,
  parental INTEGER DEFAULT 0,
  parent_id TEXT DEFAULT '',
  custom_parent_id TEXT DEFAULT '',
  secrecy INTEGER DEFAULT 0,
  register_way INTEGER DEFAULT 1,
  status TEXT DEFAULT 'OFF',
  custom_status TEXT DEFAULT '',
  longitude REAL DEFAULT 0,
  latitude REAL DEFAULT 0,
  custom_longitude REAL DEFAULT 0,
  custom_latitude REAL DEFAULT 0,
  altitude REAL DEFAULT 0,
  speed REAL DEFAULT 0,
  direction REAL DEFAULT 0,
  ptz_type INTEGER DEFAULT 0,
  custom_ptz_type INTEGER DEFAULT 0,
  battery_level TEXT DEFAULT '',
  signal_level TEXT DEFAULT '',
  download_speed TEXT DEFAULT '',
  block TEXT DEFAULT '',
  custom_block TEXT DEFAULT '',
  sub_count INTEGER DEFAULT 0,
  channel_no INTEGER DEFAULT 0,
  ondemand INTEGER DEFAULT 1,
  audio_enable INTEGER DEFAULT 0,
  cloud_record INTEGER DEFAULT 0,
  shared INTEGER DEFAULT 1,
  custom INTEGER DEFAULT 0,
  custom_id TEXT DEFAULT '',
  stream_id TEXT DEFAULT '',
  snap_url TEXT DEFAULT '',
  updated_at TEXT,
  created_at TEXT,
  PRIMARY KEY (device_id, id)
);
CREATE TABLE IF NOT EXISTS cascades (
  id TEXT PRIMARY KEY,
  name TEXT,
  serial TEXT,
  realm TEXT,
  host TEXT,
  port INTEGER,
  local_serial TEXT,
  local_host TEXT,
  local_port INTEGER,
  username TEXT,
  password TEXT,
  keepalive_max INTEGER DEFAULT 3,
  keepalive_interval INTEGER DEFAULT 60,
  register_interval INTEGER DEFAULT 60,
  register_timeout INTEGER DEFAULT 3600,
  control_priority INTEGER DEFAULT 0,
  load_limit INTEGER DEFAULT 0,
  catalog_group_size INTEGER DEFAULT 1,
  command_transport TEXT DEFAULT 'UDP',
  charset TEXT DEFAULT 'GB2312',
  online INTEGER DEFAULT 0,
  status TEXT DEFAULT '',
  load INTEGER DEFAULT 0,
  allow_control INTEGER DEFAULT 1,
  share_all INTEGER DEFAULT 0,
  share_record INTEGER DEFAULT 1,
  stream_keepalive INTEGER DEFAULT 0,
  stream_reader INTEGER DEFAULT 0,
  bind_local_ip INTEGER DEFAULT 0,
  enable INTEGER DEFAULT 1,
  updated_at TEXT,
  created_at TEXT
);
CREATE TABLE IF NOT EXISTS cascade_channels (
  cascade_id TEXT,
  device_id TEXT,
  channel_id TEXT,
  PRIMARY KEY (cascade_id, device_id, channel_id)
);
CREATE TABLE IF NOT EXISTS user_channels (
  user_id INTEGER,
  device_id TEXT,
  channel_id TEXT,
  PRIMARY KEY (user_id, device_id, channel_id)
);
CREATE TABLE IF NOT EXISTS alarms (
  id TEXT PRIMARY KEY,
  device_id TEXT,
  channel_id TEXT,
  priority INTEGER,
  alarm_time TEXT,
  method INTEGER,
  alarm_type INTEGER,
  event_type TEXT,
  ext_info TEXT,
  record_path TEXT,
  snap_path TEXT,
  created_at TEXT
);
CREATE TABLE IF NOT EXISTS op_logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT,
  method TEXT,
  request_uri TEXT,
  remote_addr TEXT,
  status TEXT,
  duration REAL,
  username TEXT,
  start_at TEXT,
  ext_info TEXT,
  description TEXT
);
CREATE TABLE IF NOT EXISTS streams (
  stream_id TEXT PRIMARY KEY,
  sms_id TEXT,
  device_id TEXT,
  channel_id TEXT,
  channel_name TEXT,
  transport TEXT,
  start_at TEXT,
  call_id TEXT,
  ssrc TEXT,
  playback INTEGER DEFAULT 0,
  start_time TEXT,
  end_time TEXT,
  audio INTEGER DEFAULT 0,
  ondemand INTEGER DEFAULT 1,
  cloud_record INTEGER DEFAULT 0,
  ptz_type INTEGER DEFAULT 0,
  stopped INTEGER DEFAULT 0,
  flv TEXT, ws_flv TEXT, hls TEXT, rtmp TEXT, rtsp TEXT, webrtc TEXT, whep TEXT, cdn TEXT
);
CREATE TABLE IF NOT EXISTS access_rules (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  kind TEXT,
  serial TEXT DEFAULT '',
  ip TEXT DEFAULT '',
  ua TEXT DEFAULT '',
  password TEXT DEFAULT '',
  description TEXT DEFAULT '',
  created_at TEXT
);
CREATE TABLE IF NOT EXISTS cloud_records (
  id TEXT PRIMARY KEY,
  device_id TEXT,
  channel_id TEXT,
  name TEXT,
  start_time TEXT,
  end_time TEXT,
  file_path TEXT,
  file_size INTEGER,
  important INTEGER DEFAULT 0,
  duration INTEGER
);
CREATE TABLE IF NOT EXISTS groups (
  id TEXT PRIMARY KEY,
  parent_id TEXT DEFAULT '',
  name TEXT,
  custom_id TEXT DEFAULT '',
  serial TEXT DEFAULT '',
  code TEXT DEFAULT ''
);
CREATE TABLE IF NOT EXISTS status_logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  device_id TEXT,
  status TEXT,
  created_at TEXT
);
CREATE TABLE IF NOT EXISTS presets (
  device_id TEXT,
  channel_id TEXT,
  preset_id INTEGER,
  name TEXT,
  PRIMARY KEY (device_id, channel_id, preset_id)
);
`
	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) seedAdmin() error {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(1) FROM users`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	sum := md5.Sum([]byte("admin"))
	now := model.FormatTime(time.Now())
	_, err := s.db.Exec(`INSERT INTO users(username,password_md5,role,enable,has_all_channel,creator,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?)`, "admin", hex.EncodeToString(sum[:]), "超级管理员", 1, 1, "system", now, now)
	return err
}

func bint(b bool) int {
	if b {
		return 1
	}
	return 0
}

func now() string { return model.FormatTime(time.Now()) }

func (s *Store) GetUserByName(name string) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	row := s.db.QueryRow(`SELECT id,username,password_md5,role,control_priority,phone,email,description,creator,locked,enable,has_all_channel,last_login_at,updated_at,created_at FROM users WHERE username=?`, name)
	return scanUser(row)
}

func (s *Store) GetUser(id int64) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	row := s.db.QueryRow(`SELECT id,username,password_md5,role,control_priority,phone,email,description,creator,locked,enable,has_all_channel,last_login_at,updated_at,created_at FROM users WHERE id=?`, id)
	return scanUser(row)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanUser(row scanner) (*model.User, error) {
	u := &model.User{}
	var locked, enable, all int
	err := row.Scan(&u.ID, &u.Username, &u.PasswordMD5, &u.Role, &u.ControlPriority, &u.PhoneNumber, &u.Email, &u.Description, &u.Creator, &locked, &enable, &all, &u.LastLoginAt, &u.UpdatedAt, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	u.Lock, u.Enable, u.HasAllChannel = locked == 1, enable == 1, all == 1
	return u, nil
}

func (s *Store) ListUsers(q string, enable string, sort, order string, start, limit int) ([]model.User, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	where := "WHERE 1=1"
	args := []any{}
	if q != "" {
		where += " AND (username LIKE ? OR role LIKE ? OR description LIKE ?)"
		like := "%" + q + "%"
		args = append(args, like, like, like)
	}
	if enable == "true" || enable == "1" {
		where += " AND enable=1"
	} else if enable == "false" || enable == "0" {
		where += " AND enable=0"
	}
	var total int
	if err := s.db.QueryRow(`SELECT COUNT(1) FROM users `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	col := safeSort(sort, map[string]string{"ID": "id", "Username": "username", "CreatedAt": "created_at", "UpdatedAt": "updated_at"}, "id")
	ord := "ASC"
	if strings.EqualFold(order, "desc") {
		ord = "DESC"
	}
	if limit <= 0 {
		limit = 50
	}
	qargs := append(append([]any{}, args...), limit, start)
	rows, err := s.db.Query(`SELECT id,username,password_md5,role,control_priority,phone,email,description,creator,locked,enable,has_all_channel,last_login_at,updated_at,created_at FROM users `+where+` ORDER BY `+col+` `+ord+` LIMIT ? OFFSET ?`, qargs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []model.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *u)
	}
	if list == nil {
		list = []model.User{}
	}
	return list, total, nil
}

func (s *Store) SaveUser(u *model.User, passwordMD5 string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := now()
	if u.ID == 0 {
		res, err := s.db.Exec(`INSERT INTO users(username,password_md5,role,control_priority,phone,email,description,creator,enable,has_all_channel,created_at,updated_at)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, u.Username, passwordMD5, u.Role, u.ControlPriority, u.PhoneNumber, u.Email, u.Description, u.Creator, bint(u.Enable), bint(u.HasAllChannel), t, t)
		if err != nil {
			return 0, err
		}
		return res.LastInsertId()
	}
	_, err := s.db.Exec(`UPDATE users SET username=?, role=?, control_priority=?, phone=?, email=?, description=?, enable=?, updated_at=? WHERE id=?`,
		u.Username, u.Role, u.ControlPriority, u.PhoneNumber, u.Email, u.Description, bint(u.Enable), t, u.ID)
	return u.ID, err
}

func (s *Store) SetUserPassword(id int64, md5hex string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE users SET password_md5=?, updated_at=? WHERE id=?`, md5hex, now(), id)
	return err
}

func (s *Store) SetUserEnable(id int64, enable bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE users SET enable=?, updated_at=? WHERE id=?`, bint(enable), now(), id)
	return err
}

func (s *Store) SetUserHasAll(id int64, all bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE users SET has_all_channel=? WHERE id=?`, bint(all), id)
	return err
}

func (s *Store) UnlockUser(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE users SET locked=0 WHERE id=?`, id)
	return err
}

func (s *Store) TouchLogin(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE users SET last_login_at=?, locked=0 WHERE id=?`, now(), id)
	return err
}

func (s *Store) RemoveUser(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM users WHERE id=? AND username<>'admin'`, id)
	if err == nil {
		_, _ = s.db.Exec(`DELETE FROM user_channels WHERE user_id=?`, id)
		_, _ = s.db.Exec(`DELETE FROM sessions WHERE user_id=?`, id)
	}
	return err
}

func (s *Store) CreateSession(token, urlToken string, userID int64, ip string, expire time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`INSERT INTO sessions(token,url_token,user_id,remote_ip,expire_at,created_at) VALUES(?,?,?,?,?,?)`,
		token, urlToken, userID, ip, expire.Unix(), now())
	return err
}

func (s *Store) SessionUser(token string) (*model.User, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var uid int64
	var exp int64
	var ip string
	err := s.db.QueryRow(`SELECT user_id, expire_at, remote_ip FROM sessions WHERE token=? OR url_token=?`, token, token).Scan(&uid, &exp, &ip)
	if err != nil {
		return nil, "", err
	}
	if time.Now().Unix() > exp {
		_, _ = s.db.Exec(`DELETE FROM sessions WHERE token=? OR url_token=?`, token, token)
		return nil, "", sql.ErrNoRows
	}
	row := s.db.QueryRow(`SELECT id,username,password_md5,role,control_priority,phone,email,description,creator,locked,enable,has_all_channel,last_login_at,updated_at,created_at FROM users WHERE id=?`, uid)
	u, err := scanUser(row)
	return u, ip, err
}

func (s *Store) DeleteSession(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, _ = s.db.Exec(`DELETE FROM sessions WHERE token=? OR url_token=?`, token, token)
}

func safeSort(sort string, allow map[string]string, def string) string {
	if c, ok := allow[sort]; ok {
		return c
	}
	return def
}

func (s *Store) UpsertDeviceFromRegister(d *model.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := now()
	var exists int
	_ = s.db.QueryRow(`SELECT COUNT(1) FROM devices WHERE id=?`, d.ID).Scan(&exists)
	if exists == 0 {
		_, err := s.db.Exec(`INSERT INTO devices(id,name,manufacturer,model,firmware,gb_ver,dev_type,command_transport,media_transport,media_transport_mode,online,password,remote_ip,remote_port,contact_ip,contact_port,last_register_at,last_keepalive_at,ua,charset,created_at,updated_at)
			VALUES(?,?,?,?,?,?,?,?,?,?,1,?,?,?,?,?,?,?,?,?,?,?)`,
			d.ID, d.Name, d.Manufacturer, d.Model, d.Firmware, d.GBVer, or(d.Type, "GB"), or(d.CommandTransport, "UDP"), or(d.MediaTransport, "UDP"), or(d.MediaTransportMode, "passive"),
			d.Password, d.RemoteIP, d.RemotePort, d.ContactIP, d.RemotePort, t, t, d.UA, or(d.Charset, "GB2312"), t, t)
		return err
	}
	_, err := s.db.Exec(`UPDATE devices SET online=1, remote_ip=?, remote_port=?, contact_ip=?, contact_port=?, command_transport=?, last_register_at=?, last_keepalive_at=?, ua=?, updated_at=?,
		manufacturer=CASE WHEN ?<>'' THEN ? ELSE manufacturer END,
		gb_ver=CASE WHEN ?='2022' OR gb_ver='2022' THEN '2022' ELSE '' END,
		name=CASE WHEN name='' THEN ? ELSE name END
		WHERE id=?`,
		d.RemoteIP, d.RemotePort, d.ContactIP, d.RemotePort, or(d.CommandTransport, "UDP"), t, t, d.UA, t,
		d.Manufacturer, d.Manufacturer, d.GBVer, d.Name, d.ID)
	if err == nil {
		_, _ = s.db.Exec(`INSERT INTO status_logs(device_id,status,created_at) VALUES(?,?,?)`, d.ID, "ON", t)
	}
	return err
}

func or(a, b string) string {
	if a == "" {
		return b
	}
	return a
}

func (s *Store) TouchKeepalive(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := now()
	_, _ = s.db.Exec(`UPDATE devices SET online=1, last_keepalive_at=?, updated_at=? WHERE id=?`, t, t, id)
}

func (s *Store) SetDeviceOnline(id string, online bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st := "OFF"
	if online {
		st = "ON"
	}
	t := now()
	_, _ = s.db.Exec(`UPDATE devices SET online=?, updated_at=? WHERE id=?`, bint(online), t, id)
	_, _ = s.db.Exec(`INSERT INTO status_logs(device_id,status,created_at) VALUES(?,?,?)`, id, st, t)
	if !online {
		_, _ = s.db.Exec(`UPDATE channels SET status='OFF', updated_at=? WHERE device_id=? AND custom_status=''`, t, id)
	}
}

func (s *Store) MarkOfflineBefore(deadline time.Time) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	rows, err := s.db.Query(`SELECT id FROM devices WHERE online=1 AND last_keepalive_at<>'' AND last_keepalive_at<?`, model.FormatTime(deadline))
	if err != nil {
		return nil
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	t := now()
	for _, id := range ids {
		_, _ = s.db.Exec(`UPDATE devices SET online=0, updated_at=? WHERE id=?`, t, id)
		_, _ = s.db.Exec(`INSERT INTO status_logs(device_id,status,created_at) VALUES(?,?,?)`, id, "OFF", t)
		_, _ = s.db.Exec(`UPDATE channels SET status='OFF' WHERE device_id=? AND custom_status=''`, id)
	}
	return ids
}

func (s *Store) GetDevice(id string) (*model.Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.getDevice(id)
}

func (s *Store) getDevice(id string) (*model.Device, error) {
	row := s.db.QueryRow(`SELECT id,name,custom_name,manufacturer,model,firmware,gb_ver,dev_type,channel_count,recv_stream_ip,contact_ip,drop_channel_type,sms_id,sms_group_id,catalog_interval,subscribe_interval,catalog_subscribe,alarm_subscribe,position_subscribe,ptz_subscribe,online,password,record_center,record_indistinct,civil_code_first,keep_original_tree,command_transport,media_transport,media_transport_mode,remote_ip,remote_port,longitude,latitude,last_register_at,last_keepalive_at,updated_at,created_at,catalog_progress,ua,charset FROM devices WHERE id=?`, id)
	return scanDevice(row)
}

func scanDevice(row scanner) (*model.Device, error) {
	d := &model.Device{}
	var cat, alarm, pos, ptz, online, rc, ri, cf, ko int
	err := row.Scan(&d.ID, &d.Name, &d.CustomName, &d.Manufacturer, &d.Model, &d.Firmware, &d.GBVer, &d.Type, &d.ChannelCount, &d.RecvStreamIP, &d.ContactIP, &d.DropChannelType, &d.SMSID, &d.SMSGroupID, &d.CatalogInterval, &d.SubscribeInterval, &cat, &alarm, &pos, &ptz, &online, &d.Password, &rc, &ri, &cf, &ko, &d.CommandTransport, &d.MediaTransport, &d.MediaTransportMode, &d.RemoteIP, &d.RemotePort, &d.Longitude, &d.Latitude, &d.LastRegisterAt, &d.LastKeepaliveAt, &d.UpdatedAt, &d.CreatedAt, &d.CatalogProgress, &d.UA, &d.Charset)
	if err != nil {
		return nil, err
	}
	d.CatalogSubscribe, d.AlarmSubscribe, d.PositionSubscribe, d.PTZSubscribe = cat == 1, alarm == 1, pos == 1, ptz == 1
	d.Online, d.RecordCenter, d.RecordIndistinct, d.CivilCodeFirst, d.KeepOriginalTree = online == 1, rc == 1, ri == 1, cf == 1, ko == 1
	return d, nil
}

func (s *Store) ListDevices(q, deviceType, online, sort, order string, start, limit int) ([]model.Device, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	where := "WHERE 1=1"
	args := []any{}
	if q != "" {
		where += " AND (id LIKE ? OR name LIKE ? OR custom_name LIKE ? OR manufacturer LIKE ? OR remote_ip LIKE ?)"
		like := "%" + q + "%"
		args = append(args, like, like, like, like, like)
	}
	if online == "true" || online == "1" {
		where += " AND online=1"
	} else if online == "false" || online == "0" {
		where += " AND online=0"
	}
	if deviceType != "" && deviceType != "all" && deviceType != "device" && deviceType != "decode" {
		parts := strings.Split(deviceType, ",")
		var ors []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			ors = append(ors, "substr(id,11,3)=?")
			args = append(args, p)
		}
		if len(ors) > 0 {
			where += " AND (" + strings.Join(ors, " OR ") + ")"
		}
	} else if deviceType == "decode" {
		where += " AND dev_type='Decode'"
	} else if deviceType == "device" {
		where += " AND dev_type<>'Decode'"
	}
	var total int
	if err := s.db.QueryRow(`SELECT COUNT(1) FROM devices `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	col := safeSort(sort, map[string]string{
		"ID": "id", "ChannelCount": "channel_count", "LastKeepaliveAt": "last_keepalive_at",
		"LastRegisterAt": "last_register_at", "UpdatedAt": "updated_at", "CreatedAt": "created_at",
	}, "id")
	ord := "ASC"
	if strings.EqualFold(order, "desc") {
		ord = "DESC"
	}
	if limit <= 0 {
		limit = 50
	}
	qargs := append(append([]any{}, args...), limit, start)
	rows, err := s.db.Query(`SELECT id,name,custom_name,manufacturer,model,firmware,gb_ver,dev_type,channel_count,recv_stream_ip,contact_ip,drop_channel_type,sms_id,sms_group_id,catalog_interval,subscribe_interval,catalog_subscribe,alarm_subscribe,position_subscribe,ptz_subscribe,online,password,record_center,record_indistinct,civil_code_first,keep_original_tree,command_transport,media_transport,media_transport_mode,remote_ip,remote_port,longitude,latitude,last_register_at,last_keepalive_at,updated_at,created_at,catalog_progress,ua,charset FROM devices `+where+` ORDER BY `+col+` `+ord+` LIMIT ? OFFSET ?`, qargs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := []model.Device{}
	for rows.Next() {
		d, err := scanDevice(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, *d)
	}
	return list, total, nil
}

func (s *Store) UpdateDevice(d *model.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE devices SET custom_name=?, media_transport=?, media_transport_mode=?, password=?, sms_id=?, sms_group_id=?, recv_stream_ip=?, drop_channel_type=?,
		catalog_subscribe=?, alarm_subscribe=?, position_subscribe=?, ptz_subscribe=?, record_center=?, record_indistinct=?, civil_code_first=?, keep_original_tree=?, name=?, updated_at=? WHERE id=?`,
		d.CustomName, d.MediaTransport, d.MediaTransportMode, d.Password, d.SMSID, d.SMSGroupID, d.RecvStreamIP, d.DropChannelType,
		bint(d.CatalogSubscribe), bint(d.AlarmSubscribe), bint(d.PositionSubscribe), bint(d.PTZSubscribe), bint(d.RecordCenter), bint(d.RecordIndistinct), bint(d.CivilCodeFirst), bint(d.KeepOriginalTree), d.Name, now(), d.ID)
	return err
}

func (s *Store) SetDeviceName(id, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE devices SET custom_name=?, updated_at=? WHERE id=?`, name, now(), id)
	return err
}

func (s *Store) SetMediaTransport(id, transport, mode string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE devices SET media_transport=?, media_transport_mode=?, updated_at=? WHERE id=?`, transport, mode, now(), id)
	return err
}

func (s *Store) SetDeviceSMS(id, sms, group string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE devices SET sms_id=?, sms_group_id=? WHERE id=?`, sms, group, id)
	return err
}

func (s *Store) RemoveDevice(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.db.Exec(`DELETE FROM devices WHERE id=? AND online=0`, id); err != nil {
		return err
	}
	_, _ = s.db.Exec(`DELETE FROM channels WHERE device_id=?`, id)
	return nil
}

func (s *Store) SetCatalogProgress(id, progress string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, _ = s.db.Exec(`UPDATE devices SET catalog_progress=? WHERE id=?`, progress, id)
}

func (s *Store) SetGBVer(id, ver string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, _ = s.db.Exec(`UPDATE devices SET gb_ver=? WHERE id=?`, ver, id)
}

func (s *Store) ReplaceCatalog(deviceID string, items []model.Channel, ondemand, cloud, shared bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	t := now()
	for _, c := range items {
		if c.ID == "" {
			continue
		}
		c.DeviceID = deviceID
		_, err = tx.Exec(`INSERT INTO channels(device_id,id,name,manufacturer,model,owner,civil_code,address,firmware,serial_number,ip_address,port,parental,parent_id,secrecy,register_way,status,longitude,latitude,altitude,speed,direction,ptz_type,battery_level,signal_level,download_speed,block,sub_count,channel_no,ondemand,audio_enable,cloud_record,shared,created_at,updated_at)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
			ON CONFLICT(device_id,id) DO UPDATE SET
			name=excluded.name, manufacturer=excluded.manufacturer, model=excluded.model, owner=excluded.owner, civil_code=excluded.civil_code,
			address=excluded.address, firmware=excluded.firmware, serial_number=excluded.serial_number, ip_address=excluded.ip_address, port=excluded.port,
			parental=excluded.parental, parent_id=excluded.parent_id, secrecy=excluded.secrecy, status=excluded.status, longitude=excluded.longitude,
			latitude=excluded.latitude, ptz_type=excluded.ptz_type, sub_count=excluded.sub_count, download_speed=excluded.download_speed, block=excluded.block, updated_at=excluded.updated_at`,
			deviceID, c.ID, c.Name, c.Manufacturer, c.Model, c.Owner, c.CivilCode, c.Address, c.Firmware, c.SerialNumber, c.IPAddress, c.Port, c.Parental, c.ParentID, c.Secrecy, orInt(c.RegisterWay, 1), or(c.Status, "ON"), c.Longitude, c.Latitude, c.Altitude, c.Speed, c.Direction, c.PTZType, c.BatteryLevel, c.SignalLevel, c.DownloadSpeed, c.Block, c.SubCount, c.Channel, bint(ondemand), 0, bint(cloud), bint(shared), t, t)
		if err != nil {
			return err
		}
	}
	var cnt int
	_ = tx.QueryRow(`SELECT COUNT(1) FROM channels WHERE device_id=?`, deviceID).Scan(&cnt)
	_, _ = tx.Exec(`UPDATE devices SET channel_count=?, catalog_progress='', updated_at=? WHERE id=?`, cnt, t, deviceID)
	return tx.Commit()
}

func orInt(a, b int) int {
	if a == 0 {
		return b
	}
	return a
}

func (s *Store) ListChannels(deviceID, q, dirSerial, sort, order string, start, limit int) ([]model.Channel, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	where := "WHERE c.device_id=?"
	args := []any{deviceID}
	if dirSerial != "" {
		where += " AND (c.parent_id=? OR c.parent_id LIKE ?)"
		args = append(args, dirSerial, "%/"+dirSerial)
	}
	if q != "" {
		where += " AND (c.id LIKE ? OR c.name LIKE ? OR c.custom_name LIKE ?)"
		like := "%" + q + "%"
		args = append(args, like, like, like)
	}
	var total int
	if err := s.db.QueryRow(`SELECT COUNT(1) FROM channels c `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 50
	}
	col := safeSort(sort, map[string]string{"ID": "c.id", "Name": "c.name", "Status": "c.status"}, "c.id")
	ord := "ASC"
	if strings.EqualFold(order, "desc") {
		ord = "DESC"
	}
	qargs := append(append([]any{}, args...), limit, start)
	rows, err := s.db.Query(`SELECT c.device_id,c.id,c.name,c.custom_name,c.manufacturer,c.custom_manufacturer,c.model,c.custom_model,c.owner,c.civil_code,c.custom_civil_code,c.address,c.custom_address,c.firmware,c.custom_firmware,c.serial_number,c.custom_serial_number,c.ip_address,c.custom_ip,c.port,c.custom_port,c.parental,c.parent_id,c.custom_parent_id,c.secrecy,c.register_way,c.status,c.custom_status,c.longitude,c.latitude,c.custom_longitude,c.custom_latitude,c.altitude,c.speed,c.direction,c.ptz_type,c.custom_ptz_type,c.battery_level,c.signal_level,c.download_speed,c.block,c.custom_block,c.sub_count,c.channel_no,c.ondemand,c.audio_enable,c.cloud_record,c.shared,c.custom,c.custom_id,c.stream_id,c.snap_url,d.name,d.custom_name,d.dev_type,d.online
		FROM channels c JOIN devices d ON d.id=c.device_id `+where+` ORDER BY `+col+` `+ord+` LIMIT ? OFFSET ?`, qargs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return scanChannels(rows, total)
}

func scanChannels(rows *sql.Rows, total int) ([]model.Channel, int, error) {
	list := []model.Channel{}
	for rows.Next() {
		c := model.Channel{}
		var ond, audio, cloud, shared, custom, devOnline int
		if err := rows.Scan(&c.DeviceID, &c.ID, &c.Name, &c.CustomName, &c.Manufacturer, &c.CustomManufacturer, &c.Model, &c.CustomModel, &c.Owner, &c.CivilCode, &c.CustomCivilCode, &c.Address, &c.CustomAddress, &c.Firmware, &c.CustomFirmware, &c.SerialNumber, &c.CustomSerialNumber, &c.IPAddress, &c.CustomIPAddress, &c.Port, &c.CustomPort, &c.Parental, &c.ParentID, &c.CustomParentID, &c.Secrecy, &c.RegisterWay, &c.Status, &c.CustomStatus, &c.Longitude, &c.Latitude, &c.CustomLongitude, &c.CustomLatitude, &c.Altitude, &c.Speed, &c.Direction, &c.PTZType, &c.CustomPTZType, &c.BatteryLevel, &c.SignalLevel, &c.DownloadSpeed, &c.Block, &c.CustomBlock, &c.SubCount, &c.Channel, &ond, &audio, &cloud, &shared, &custom, &c.CustomID, &c.StreamID, &c.SnapURL, &c.DeviceName, &c.DeviceCustomName, &c.DeviceType, &devOnline); err != nil {
			return nil, 0, err
		}
		c.Ondemand, c.AudioEnable, c.CloudRecord, c.Shared, c.Custom = ond == 1, audio == 1, cloud == 1, shared == 1, custom == 1
		c.DeviceOnline = devOnline == 1
		if c.SnapURL == "" {
			c.SnapURL = "/images/default_snap.png"
		}
		list = append(list, c)
	}
	return list, total, nil
}

func (s *Store) GetChannel(deviceID, id string) (*model.Channel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rows, err := s.db.Query(`SELECT c.device_id,c.id,c.name,c.custom_name,c.manufacturer,c.custom_manufacturer,c.model,c.custom_model,c.owner,c.civil_code,c.custom_civil_code,c.address,c.custom_address,c.firmware,c.custom_firmware,c.serial_number,c.custom_serial_number,c.ip_address,c.custom_ip,c.port,c.custom_port,c.parental,c.parent_id,c.custom_parent_id,c.secrecy,c.register_way,c.status,c.custom_status,c.longitude,c.latitude,c.custom_longitude,c.custom_latitude,c.altitude,c.speed,c.direction,c.ptz_type,c.custom_ptz_type,c.battery_level,c.signal_level,c.download_speed,c.block,c.custom_block,c.sub_count,c.channel_no,c.ondemand,c.audio_enable,c.cloud_record,c.shared,c.custom,c.custom_id,c.stream_id,c.snap_url,d.name,d.custom_name,d.dev_type,d.online
		FROM channels c JOIN devices d ON d.id=c.device_id WHERE c.device_id=? AND c.id=?`, deviceID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list, _, err := scanChannels(rows, 1)
	if err != nil || len(list) == 0 {
		if err == nil {
			err = sql.ErrNoRows
		}
		return nil, err
	}
	return &list[0], nil
}

func (s *Store) SetChannelField(deviceID, id, column string, value any) error {
	allow := map[string]bool{
		"custom_name": true, "custom_id": true, "audio_enable": true, "cloud_record": true, "ondemand": true,
		"custom_ptz_type": true, "shared": true, "custom_longitude": true, "custom_latitude": true, "stream_id": true,
		"custom_manufacturer": true, "custom_status": true, "name": true,
	}
	if !allow[column] {
		return fmt.Errorf("字段不允许修改")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE channels SET `+column+`=?, updated_at=? WHERE device_id=? AND id=?`, value, now(), deviceID, id)
	return err
}

func (s *Store) SetPresetName(deviceID, channelID string, preset int, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`INSERT INTO presets(device_id,channel_id,preset_id,name) VALUES(?,?,?,?)
		ON CONFLICT(device_id,channel_id,preset_id) DO UPDATE SET name=excluded.name`, deviceID, channelID, preset, name)
	return err
}

func (s *Store) OnlineStats() (devTotal, devOnline, chTotal, chOnline int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.db.QueryRow(`SELECT COUNT(1), COALESCE(SUM(online),0) FROM devices`).Scan(&devTotal, &devOnline)
	_ = s.db.QueryRow(`SELECT COUNT(1), COALESCE(SUM(CASE WHEN status='ON' THEN 1 ELSE 0 END),0) FROM channels`).Scan(&chTotal, &chOnline)
	return
}

func (s *Store) ChannelTree(serial, pcode string) ([]model.TreeNode, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if serial == "" {
		rows, err := s.db.Query(`SELECT id,name,custom_name,online,channel_count FROM devices ORDER BY id`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var list []model.TreeNode
		for rows.Next() {
			var id, name, custom string
			var online, cnt int
			if err := rows.Scan(&id, &name, &custom, &online, &cnt); err != nil {
				return nil, err
			}
			st := "OFF"
			if online == 1 {
				st = "ON"
			}
			show := name
			if custom != "" {
				show = custom
			}
			if show == "" {
				show = id
			}
			list = append(list, model.TreeNode{
				ID: id, Name: show, CustomName: custom, Serial: id, Status: st,
				Parental: 1, SubCount: cnt, SubCountDevice: cnt, Manufacturer: "",
			})
		}
		if list == nil {
			list = []model.TreeNode{}
		}
		return list, nil
	}
	parent := pcode
	q := `SELECT id,name,custom_name,custom,custom_id,status,ptz_type,longitude,latitude,parental,manufacturer,sub_count FROM channels WHERE device_id=?`
	args := []any{serial}
	if parent == "" {
		q += ` AND (parent_id='' OR parent_id=? OR parent_id LIKE ?)`
		args = append(args, serial, serial+"/%")
		// 顶层：父节点是设备本身，或父节点不在本设备通道集合中。先取全部再过滤，避免漏目录。
		q = `SELECT id,name,custom_name,custom,custom_id,status,ptz_type,longitude,latitude,parental,manufacturer,sub_count,parent_id FROM channels WHERE device_id=?`
		args = []any{serial}
	} else {
		q += ` AND (parent_id=? OR parent_id LIKE ?)`
		args = append(args, parent, "%/"+parent)
		q = `SELECT id,name,custom_name,custom,custom_id,status,ptz_type,longitude,latitude,parental,manufacturer,sub_count,parent_id FROM channels WHERE device_id=? AND (parent_id=? OR parent_id LIKE ?)`
	}
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type row struct {
		n      model.TreeNode
		parent string
	}
	var all []row
	ids := map[string]bool{}
	for rows.Next() {
		var n model.TreeNode
		var custom int
		var parentID string
		if err := rows.Scan(&n.Code, &n.Name, &n.CustomName, &custom, &n.CustomID, &n.Status, &n.PTZType, &n.Longitude, &n.Latitude, &n.Parental, &n.Manufacturer, &n.SubCount, &parentID); err != nil {
			return nil, err
		}
		n.Custom = custom == 1
		n.Serial = serial
		n.ID = serial + ":" + n.Code
		if n.CustomName != "" {
			n.Name = n.CustomName
		}
		if n.Status == "" {
			n.Status = "OFF"
		}
		all = append(all, row{n, parentID})
		ids[n.Code] = true
	}
	var list []model.TreeNode
	for _, r := range all {
		if parent == "" {
			p := lastID(r.parent)
			if p == "" || p == serial || !ids[p] {
				list = append(list, r.n)
			}
			continue
		}
		if lastID(r.parent) == parent {
			list = append(list, r.n)
		}
	}
	if list == nil {
		list = []model.TreeNode{}
	}
	return list, nil
}

func lastID(parent string) string {
	parent = strings.TrimSpace(parent)
	if parent == "" {
		return ""
	}
	if i := strings.LastIndex(parent, "/"); i >= 0 {
		return parent[i+1:]
	}
	return parent
}

func (s *Store) GroupTree(serial, pcode string) ([]model.TreeNode, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var n int
	_ = s.db.QueryRow(`SELECT COUNT(1) FROM groups`).Scan(&n)
	if n == 0 {
		s.mu.Unlock()
		list, err := s.ChannelTree(serial, pcode)
		s.mu.Lock()
		return list, err
	}
	parent := ""
	if serial != "" || pcode != "" {
		parent = serial
		if pcode != "" {
			parent = serial + ":" + pcode
		}
	}
	rows, err := s.db.Query(`SELECT id,parent_id,name,custom_id,serial,code FROM groups WHERE parent_id=? ORDER BY name`, parent)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []model.TreeNode{}
	for rows.Next() {
		var id, pid, name, cid, ser, code string
		if err := rows.Scan(&id, &pid, &name, &cid, &ser, &code); err != nil {
			return nil, err
		}
		list = append(list, model.TreeNode{
			ID: id, Name: name, Custom: true, CustomID: cid, CustomName: name,
			Serial: ser, Code: code, Status: "ON", Parental: 1,
		})
	}
	return list, nil
}

func (s *Store) SaveGroup(g model.GroupNode) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`INSERT INTO groups(id,parent_id,name,custom_id,serial,code) VALUES(?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET parent_id=excluded.parent_id, name=excluded.name, custom_id=excluded.custom_id, serial=excluded.serial, code=excluded.code`,
		g.ID, g.ParentID, g.Name, g.CustomID, g.Serial, g.Code)
	return err
}

func (s *Store) RemoveGroup(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM groups WHERE id=? OR parent_id=?`, id, id)
	return err
}

func (s *Store) SaveCascade(c *model.Cascade) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if c.ID == "" {
		c.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	t := now()
	_, err := s.db.Exec(`INSERT INTO cascades(id,name,serial,realm,host,port,local_serial,local_host,local_port,username,password,keepalive_max,keepalive_interval,register_interval,register_timeout,control_priority,load_limit,catalog_group_size,command_transport,charset,allow_control,share_all,share_record,stream_keepalive,stream_reader,bind_local_ip,enable,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(id) DO UPDATE SET name=excluded.name, serial=excluded.serial, realm=excluded.realm, host=excluded.host, port=excluded.port, local_serial=excluded.local_serial, local_host=excluded.local_host, local_port=excluded.local_port, username=excluded.username, password=excluded.password, keepalive_max=excluded.keepalive_max, keepalive_interval=excluded.keepalive_interval, register_interval=excluded.register_interval, register_timeout=excluded.register_timeout, control_priority=excluded.control_priority, load_limit=excluded.load_limit, catalog_group_size=excluded.catalog_group_size, command_transport=excluded.command_transport, charset=excluded.charset, allow_control=excluded.allow_control, share_all=excluded.share_all, share_record=excluded.share_record, stream_keepalive=excluded.stream_keepalive, stream_reader=excluded.stream_reader, bind_local_ip=excluded.bind_local_ip, enable=excluded.enable, updated_at=excluded.updated_at`,
		c.ID, c.Name, c.Serial, c.Realm, c.Host, c.Port, c.LocalSerial, c.LocalHost, c.LocalPort, c.Username, c.Password, orInt(c.KeepaliveMaxCount, 3), orInt(c.KeepaliveInterval, 60), orInt(c.RegisterInterval, 60), orInt(c.RegisterTimeout, 3600), c.ControlPriority, c.LoadLimit, orInt(c.CatalogGroupSize, 1), or(c.CommandTransport, "UDP"), or(c.Charset, "GB2312"), bint(c.AllowControl), bint(c.ShareAllChannel), bint(c.ShareRecord), bint(c.StreamKeepalive), bint(c.StreamReader), bint(c.BindLocalIP), bint(c.Enable), t, t)
	return err
}

func (s *Store) ListCascades(q, online, enable, sort, order string, start, limit int) ([]model.Cascade, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	where := "WHERE 1=1"
	args := []any{}
	if q != "" {
		where += " AND (name LIKE ? OR serial LIKE ? OR host LIKE ?)"
		like := "%" + q + "%"
		args = append(args, like, like, like)
	}
	if online == "true" {
		where += " AND online=1"
	} else if online == "false" {
		where += " AND online=0"
	}
	if enable == "true" {
		where += " AND enable=1"
	} else if enable == "false" {
		where += " AND enable=0"
	}
	var total int
	_ = s.db.QueryRow(`SELECT COUNT(1) FROM cascades `+where, args...).Scan(&total)
	if limit <= 0 {
		limit = 50
	}
	qargs := append(append([]any{}, args...), limit, start)
	rows, err := s.db.Query(`SELECT id,name,serial,realm,host,port,local_serial,local_host,local_port,username,password,keepalive_max,keepalive_interval,register_interval,register_timeout,control_priority,load_limit,catalog_group_size,command_transport,charset,online,status,load,allow_control,share_all,share_record,stream_keepalive,stream_reader,bind_local_ip,enable FROM cascades `+where+` ORDER BY created_at DESC LIMIT ? OFFSET ?`, qargs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := []model.Cascade{}
	for rows.Next() {
		c := model.Cascade{}
		var online, allow, share, rec, sk, sr, bind, en int
		if err := rows.Scan(&c.ID, &c.Name, &c.Serial, &c.Realm, &c.Host, &c.Port, &c.LocalSerial, &c.LocalHost, &c.LocalPort, &c.Username, &c.Password, &c.KeepaliveMaxCount, &c.KeepaliveInterval, &c.RegisterInterval, &c.RegisterTimeout, &c.ControlPriority, &c.LoadLimit, &c.CatalogGroupSize, &c.CommandTransport, &c.Charset, &online, &c.Status, &c.Load, &allow, &share, &rec, &sk, &sr, &bind, &en); err != nil {
			return nil, 0, err
		}
		c.Online, c.AllowControl, c.ShareAllChannel, c.ShareRecord = online == 1, allow == 1, share == 1, rec == 1
		c.StreamKeepalive, c.StreamReader, c.BindLocalIP, c.Enable = sk == 1, sr == 1, bind == 1, en == 1
		list = append(list, c)
	}
	return list, total, nil
}

func (s *Store) EnabledCascades() ([]model.Cascade, error) {
	list, _, err := s.ListCascades("", "", "true", "", "", 0, 1000)
	return list, err
}

func (s *Store) SetCascadeEnable(id string, enable bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE cascades SET enable=? WHERE id=?`, bint(enable), id)
	return err
}

func (s *Store) SetCascadeOnline(id string, online bool, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, _ = s.db.Exec(`UPDATE cascades SET online=?, status=? WHERE id=?`, bint(online), status, id)
}

func (s *Store) RemoveCascade(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM cascades WHERE id=?`, id)
	_, _ = s.db.Exec(`DELETE FROM cascade_channels WHERE cascade_id=?`, id)
	return err
}

func (s *Store) SaveCascadeChannels(id string, pairs [][2]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range pairs {
		_, _ = s.db.Exec(`INSERT OR IGNORE INTO cascade_channels(cascade_id,device_id,channel_id) VALUES(?,?,?)`, id, p[0], p[1])
	}
	return nil
}

func (s *Store) RemoveCascadeChannels(id string, pairs [][2]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range pairs {
		_, _ = s.db.Exec(`DELETE FROM cascade_channels WHERE cascade_id=? AND device_id=? AND channel_id=?`, id, p[0], p[1])
	}
	return nil
}

func (s *Store) SetCascadeShareAll(id string, all bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`UPDATE cascades SET share_all=? WHERE id=?`, bint(all), id)
	return err
}

func (s *Store) SharedChannels(cascadeID string, shareAll bool) ([]model.Channel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	q := `SELECT c.device_id,c.id,c.name,c.custom_name,c.status,c.parental,c.parent_id,c.manufacturer FROM channels c WHERE c.shared=1`
	args := []any{}
	if !shareAll {
		q += ` AND EXISTS (SELECT 1 FROM cascade_channels cc WHERE cc.cascade_id=? AND cc.device_id=c.device_id AND cc.channel_id=c.id)`
		args = append(args, cascadeID)
	}
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []model.Channel
	for rows.Next() {
		var c model.Channel
		if err := rows.Scan(&c.DeviceID, &c.ID, &c.Name, &c.CustomName, &c.Status, &c.Parental, &c.ParentID, &c.Manufacturer); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func (s *Store) AddAlarm(a *model.Alarm) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if a.ID == "" {
		a.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if a.CreatedAt == "" {
		a.CreatedAt = now()
	}
	_, err := s.db.Exec(`INSERT INTO alarms(id,device_id,channel_id,priority,alarm_time,method,alarm_type,event_type,ext_info,record_path,snap_path,created_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`,
		a.ID, a.DeviceID, a.ChannelID, a.AlarmPriority, a.Time, a.AlarmMethod, a.AlarmType, a.AlarmEventType, a.ExtInfo, a.RecordPath, a.SnapPath, a.CreatedAt)
	return err
}

func (s *Store) ListAlarms(q, deviceID string, start, limit int) ([]model.Alarm, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	where := "WHERE 1=1"
	args := []any{}
	if deviceID != "" {
		where += " AND device_id=?"
		args = append(args, deviceID)
	}
	if q != "" {
		where += " AND (device_id LIKE ? OR channel_id LIKE ?)"
		like := "%" + q + "%"
		args = append(args, like, like)
	}
	var total int
	_ = s.db.QueryRow(`SELECT COUNT(1) FROM alarms `+where, args...).Scan(&total)
	if limit <= 0 {
		limit = 50
	}
	qargs := append(append([]any{}, args...), limit, start)
	rows, err := s.db.Query(`SELECT id,device_id,channel_id,priority,alarm_time,method,alarm_type,event_type,ext_info,record_path,snap_path,created_at FROM alarms `+where+` ORDER BY created_at DESC LIMIT ? OFFSET ?`, qargs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := []model.Alarm{}
	for rows.Next() {
		var a model.Alarm
		if err := rows.Scan(&a.ID, &a.DeviceID, &a.ChannelID, &a.AlarmPriority, &a.Time, &a.AlarmMethod, &a.AlarmType, &a.AlarmEventType, &a.ExtInfo, &a.RecordPath, &a.SnapPath, &a.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, a)
	}
	return list, total, nil
}

func (s *Store) ClearAlarms() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM alarms`)
	return err
}

func (s *Store) RemoveAlarm(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM alarms WHERE id=?`, id)
	return err
}

func (s *Store) AddLog(l *model.OpLog) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, _ = s.db.Exec(`INSERT INTO op_logs(name,method,request_uri,remote_addr,status,duration,username,start_at,ext_info,description) VALUES(?,?,?,?,?,?,?,?,?,?)`,
		l.Name, l.Method, l.RequestURI, l.RemoteAddr, l.Status, l.Duration, l.Username, l.StartAt, l.ExtInfo, l.Description)
}

func (s *Store) ListLogs(q, method, startAt, endAt, sort, order string, start, limit int) ([]model.OpLog, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	where := "WHERE 1=1"
	args := []any{}
	if q != "" {
		where += " AND (name LIKE ? OR request_uri LIKE ? OR username LIKE ?)"
		like := "%" + q + "%"
		args = append(args, like, like, like)
	}
	if method != "" {
		where += " AND method=?"
		args = append(args, method)
	}
	if startAt != "" {
		where += " AND start_at>=?"
		args = append(args, startAt)
	}
	if endAt != "" {
		where += " AND start_at<=?"
		args = append(args, endAt+" 23:59:59")
	}
	var total int
	_ = s.db.QueryRow(`SELECT COUNT(1) FROM op_logs `+where, args...).Scan(&total)
	if limit <= 0 {
		limit = 50
	}
	col := safeSort(sort, map[string]string{"StartAt": "start_at", "Duration": "duration"}, "id")
	ord := "DESC"
	if strings.EqualFold(order, "asc") {
		ord = "ASC"
	}
	qargs := append(append([]any{}, args...), limit, start)
	rows, err := s.db.Query(`SELECT id,name,method,request_uri,remote_addr,status,duration,username,start_at,ext_info,description FROM op_logs `+where+` ORDER BY `+col+` `+ord+` LIMIT ? OFFSET ?`, qargs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := []model.OpLog{}
	for rows.Next() {
		var l model.OpLog
		if err := rows.Scan(&l.ID, &l.Name, &l.Method, &l.RequestURI, &l.RemoteAddr, &l.Status, &l.Duration, &l.Username, &l.StartAt, &l.ExtInfo, &l.Description); err != nil {
			return nil, 0, err
		}
		list = append(list, l)
	}
	return list, total, nil
}

func (s *Store) ClearLogs() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM op_logs`)
	return err
}

func (s *Store) SaveStream(st *model.Stream) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`INSERT INTO streams(stream_id,sms_id,device_id,channel_id,channel_name,transport,start_at,call_id,ssrc,playback,start_time,end_time,audio,ondemand,cloud_record,ptz_type,stopped,flv,ws_flv,hls,rtmp,rtsp,webrtc,whep,cdn)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,0,?,?,?,?,?,?,?,?)
		ON CONFLICT(stream_id) DO UPDATE SET stopped=0, call_id=excluded.call_id, start_at=excluded.start_at, flv=excluded.flv, ws_flv=excluded.ws_flv, hls=excluded.hls, rtmp=excluded.rtmp, rtsp=excluded.rtsp, webrtc=excluded.webrtc, whep=excluded.whep`,
		st.StreamID, st.SMSID, st.DeviceID, st.ChannelID, st.ChannelName, st.Transport, st.StartAt, st.CallID, st.SSRC, bint(st.Playback), st.StartTime, st.EndTime, bint(st.AudioEnable), bint(st.Ondemand), bint(st.CloudRecord), st.ChannelPTZType,
		st.FLV, st.WSFLV, st.HLS, st.RTMP, st.RTSP, st.WEBRTC, st.WHEP, st.CDN)
	if err == nil {
		_, _ = s.db.Exec(`UPDATE channels SET stream_id=? WHERE device_id=? AND id=?`, st.StreamID, st.DeviceID, st.ChannelID)
	}
	return err
}

func (s *Store) GetStream(id string) (*model.Stream, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	row := s.db.QueryRow(`SELECT stream_id,sms_id,device_id,channel_id,channel_name,transport,start_at,call_id,ssrc,playback,start_time,end_time,audio,ondemand,cloud_record,ptz_type,stopped,flv,ws_flv,hls,rtmp,rtsp,webrtc,whep,cdn FROM streams WHERE stream_id=?`, id)
	return scanStream(row)
}

func (s *Store) FindLiveStream(deviceID, channelID string) (*model.Stream, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	row := s.db.QueryRow(`SELECT stream_id,sms_id,device_id,channel_id,channel_name,transport,start_at,call_id,ssrc,playback,start_time,end_time,audio,ondemand,cloud_record,ptz_type,stopped,flv,ws_flv,hls,rtmp,rtsp,webrtc,whep,cdn FROM streams WHERE device_id=? AND channel_id=? AND playback=0 AND stopped=0`, deviceID, channelID)
	return scanStream(row)
}

func scanStream(row scanner) (*model.Stream, error) {
	st := &model.Stream{}
	var pb, audio, ond, cloud, stopped int
	err := row.Scan(&st.StreamID, &st.SMSID, &st.DeviceID, &st.ChannelID, &st.ChannelName, &st.Transport, &st.StartAt, &st.CallID, &st.SSRC, &pb, &st.StartTime, &st.EndTime, &audio, &ond, &cloud, &st.ChannelPTZType, &stopped, &st.FLV, &st.WSFLV, &st.HLS, &st.RTMP, &st.RTSP, &st.WEBRTC, &st.WHEP, &st.CDN)
	if err != nil {
		return nil, err
	}
	st.Playback, st.AudioEnable, st.Ondemand, st.CloudRecord, st.Stopped = pb == 1, audio == 1, ond == 1, cloud == 1, stopped == 1
	return st, nil
}

func (s *Store) ListStreams(playback bool) ([]model.Stream, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rows, err := s.db.Query(`SELECT stream_id,sms_id,device_id,channel_id,channel_name,transport,start_at,call_id,ssrc,playback,start_time,end_time,audio,ondemand,cloud_record,ptz_type,stopped,flv,ws_flv,hls,rtmp,rtsp,webrtc,whep,cdn FROM streams WHERE playback=? AND stopped=0`, bint(playback))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []model.Stream{}
	for rows.Next() {
		st, err := scanStream(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *st)
	}
	return list, nil
}

func (s *Store) StopStream(id string) (*model.Stream, error) {
	st, err := s.GetStream(id)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err = s.db.Exec(`UPDATE streams SET stopped=1 WHERE stream_id=?`, id)
	_, _ = s.db.Exec(`UPDATE channels SET stream_id='' WHERE device_id=? AND id=? AND stream_id=?`, st.DeviceID, st.ChannelID, id)
	return st, err
}

func (s *Store) SaveRule(kind string, r *model.AccessRule) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r.ID == 0 {
		res, err := s.db.Exec(`INSERT INTO access_rules(kind,serial,ip,ua,password,description,created_at) VALUES(?,?,?,?,?,?,?)`,
			kind, r.Serial, r.IP, r.UA, r.Password, r.Description, now())
		if err != nil {
			return 0, err
		}
		return res.LastInsertId()
	}
	_, err := s.db.Exec(`UPDATE access_rules SET serial=?, ip=?, ua=?, password=?, description=? WHERE id=? AND kind=?`,
		r.Serial, r.IP, r.UA, r.Password, r.Description, r.ID, kind)
	return r.ID, err
}

func (s *Store) ListRules(kind, q string, start, limit int) ([]model.AccessRule, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	where := "WHERE kind=?"
	args := []any{kind}
	if q != "" {
		where += " AND (serial LIKE ? OR ip LIKE ? OR ua LIKE ?)"
		like := "%" + q + "%"
		args = append(args, like, like, like)
	}
	var total int
	_ = s.db.QueryRow(`SELECT COUNT(1) FROM access_rules `+where, args...).Scan(&total)
	if limit <= 0 {
		limit = 50
	}
	qargs := append(append([]any{}, args...), limit, start)
	rows, err := s.db.Query(`SELECT id,serial,ip,ua,password,description,created_at FROM access_rules `+where+` ORDER BY id DESC LIMIT ? OFFSET ?`, qargs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := []model.AccessRule{}
	for rows.Next() {
		var r model.AccessRule
		if err := rows.Scan(&r.ID, &r.Serial, &r.IP, &r.UA, &r.Password, &r.Description, &r.CreatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, r)
	}
	return list, total, nil
}

func (s *Store) RemoveRule(kind string, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM access_rules WHERE kind=? AND id=?`, kind, id)
	return err
}

func (s *Store) IsBlack(serial, ip, ua string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	var n int
	_ = s.db.QueryRow(`SELECT COUNT(1) FROM access_rules WHERE kind='black' AND ((serial<>'' AND serial=?) OR (ip<>'' AND ip=?) OR (ua<>'' AND ? LIKE '%'||ua||'%'))`, serial, ip, ua).Scan(&n)
	return n > 0
}

func (s *Store) DevicePassword(serial string, global string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	var pwd string
	_ = s.db.QueryRow(`SELECT password FROM access_rules WHERE kind='white' AND serial=? AND password<>'' LIMIT 1`, serial).Scan(&pwd)
	if pwd != "" {
		return pwd
	}
	_ = s.db.QueryRow(`SELECT password FROM devices WHERE id=?`, serial).Scan(&pwd)
	if pwd != "" {
		return pwd
	}
	return global
}

func (s *Store) AddCloudRecord(r *model.CloudRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r.ID == "" {
		r.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	_, err := s.db.Exec(`INSERT INTO cloud_records(id,device_id,channel_id,name,start_time,end_time,file_path,file_size,important,duration) VALUES(?,?,?,?,?,?,?,?,?,?)`,
		r.ID, r.DeviceID, r.ChannelID, r.Name, r.StartTime, r.EndTime, r.FilePath, r.FileSize, bint(r.Important), r.Duration)
	return err
}

func (s *Store) ListCloudByDay(deviceID, channelID, day string) ([]model.CloudRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rows, err := s.db.Query(`SELECT id,device_id,channel_id,name,start_time,end_time,file_path,file_size,important,duration FROM cloud_records WHERE device_id=? AND channel_id=? AND start_time LIKE ? ORDER BY start_time`, deviceID, channelID, day+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []model.CloudRecord{}
	for rows.Next() {
		var r model.CloudRecord
		var imp int
		if err := rows.Scan(&r.ID, &r.DeviceID, &r.ChannelID, &r.Name, &r.StartTime, &r.EndTime, &r.FilePath, &r.FileSize, &imp, &r.Duration); err != nil {
			return nil, err
		}
		r.Important = imp == 1
		list = append(list, r)
	}
	return list, nil
}

func (s *Store) RemoveCloud(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM cloud_records WHERE id=?`, id)
	return err
}

func (s *Store) StatusLogs(deviceID string, limit int) ([]map[string]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(`SELECT status, created_at FROM status_logs WHERE device_id=? ORDER BY id DESC LIMIT ?`, deviceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []map[string]string
	for rows.Next() {
		var st, at string
		if rows.Scan(&st, &at) == nil {
			list = append(list, map[string]string{"Status": st, "CreatedAt": at, "DeviceID": deviceID})
		}
	}
	if list == nil {
		list = []map[string]string{}
	}
	return list, nil
}

func (s *Store) FindChannelOwner(channelID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var deviceID string
	err := s.db.QueryRow(`SELECT device_id FROM channels WHERE id=? LIMIT 1`, channelID).Scan(&deviceID)
	return deviceID, err
}

func (s *Store) SaveUserChannels(uid int64, pairs [][2]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range pairs {
		_, _ = s.db.Exec(`INSERT OR IGNORE INTO user_channels(user_id,device_id,channel_id) VALUES(?,?,?)`, uid, p[0], p[1])
	}
	return nil
}

func (s *Store) RemoveUserChannels(uid int64, pairs [][2]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range pairs {
		_, _ = s.db.Exec(`DELETE FROM user_channels WHERE user_id=? AND device_id=? AND channel_id=?`, uid, p[0], p[1])
	}
	return nil
}
