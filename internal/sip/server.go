package sip

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/argus/argusgbs/internal/config"
	"github.com/argus/argusgbs/internal/manscdp"
	"github.com/argus/argusgbs/internal/media"
	"github.com/argus/argusgbs/internal/model"
	"github.com/argus/argusgbs/internal/store"
)

type peer struct {
	ID        string
	Transport string
	UDP       *net.UDPAddr
	TCP       net.Conn
	FromTag   string
	To        string
	GBVer     string
	Charset   string
	UA        string
}

type waitXML struct {
	ch chan *manscdp.Envelope
}

type inviteWait struct {
	ch chan inviteAnswer
}

type inviteAnswer struct {
	code int
	sdp  string
	to   string
	err  error
}

type callDialog struct {
	deviceID string
	from     string
	to       string
	uri      string
}

type Server struct {
	cfg   *config.Config
	db    *store.Store
	media *media.Client
	ip    string

	udp *net.UDPConn
	tcp net.Listener

	mu      sync.Mutex
	peers   map[string]*peer
	nonces  map[string]time.Time
	waits   map[string]*waitXML
	invites map[string]*inviteWait
	dialogs map[string]*callDialog
	catalog map[string]*catalogBuf
	sn      atomic.Int64
	cseq    atomic.Int64

	stop chan struct{}
}

type catalogBuf struct {
	sum   int
	items []model.Channel
	seen  map[string]bool
}

func New(cfg *config.Config, db *store.Store, mediaClient *media.Client) *Server {
	ip := cfg.SIPHost
	if ip == "" {
		ip = guessIP()
	}
	return &Server{
		cfg: cfg, db: db, media: mediaClient, ip: ip,
		peers: map[string]*peer{}, nonces: map[string]time.Time{},
		waits: map[string]*waitXML{}, invites: map[string]*inviteWait{},
		dialogs: map[string]*callDialog{},
		catalog: map[string]*catalogBuf{},
		stop:    make(chan struct{}),
	}
}

func (s *Server) Host() string { return s.ip }
func (s *Server) Port() int    { return s.cfg.SIPPort }

func (s *Server) Start() error {
	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", s.cfg.SIPPort))
	if err != nil {
		return err
	}
	s.udp, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	s.tcp, err = net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.SIPPort))
	if err != nil {
		s.udp.Close()
		return err
	}
	go s.readUDP()
	go s.acceptTCP()
	go s.keepaliveLoop()
	go s.cascadeLoop()
	log.Printf("SIP 监听 %s:%d serial=%s realm=%s", s.ip, s.cfg.SIPPort, s.cfg.Serial, s.cfg.Realm)
	return nil
}

func (s *Server) readUDP() {
	buf := make([]byte, 64*1024)
	for {
		n, addr, err := s.udp.ReadFromUDP(buf)
		if err != nil {
			return
		}
		b := append([]byte(nil), buf[:n]...)
		m, err := Parse(b)
		if err != nil {
			continue
		}
		m.addr = addr
		go s.dispatch(m, "UDP", addr, nil)
	}
}

func (s *Server) acceptTCP() {
	for {
		c, err := s.tcp.Accept()
		if err != nil {
			return
		}
		go s.serveTCP(c)
	}
}

func (s *Server) serveTCP(c net.Conn) {
	defer c.Close()
	r := bufio.NewReader(c)
	for {
		_ = c.SetReadDeadline(time.Now().Add(time.Duration(s.cfg.KeepaliveTimeout+30) * time.Second))
		m, err := ReadStream(r)
		if err != nil {
			return
		}
		m.addr = c.RemoteAddr()
		s.dispatch(m, "TCP", nil, c)
	}
}

func (s *Server) dispatch(m *Message, transport string, udp *net.UDPAddr, tcp net.Conn) {
	if s.cfg.SIPLog {
		if m.Response {
			log.Printf("SIP <- %d %s %s", m.StatusCode, m.CSeq(), m.CallID())
		} else {
			log.Printf("SIP <- %s %s", m.Method, m.URI)
		}
	}
	if m.Response {
		s.onResponse(m)
		return
	}
	switch m.Method {
	case "REGISTER":
		s.onRegister(m, transport, udp, tcp)
	case "MESSAGE":
		s.onMessage(m, transport, udp, tcp)
	case "ACK":
	case "BYE":
		s.reply(m, 200, "OK", nil, transport, udp, tcp)
	case "SUBSCRIBE":
		s.onSubscribe(m, transport, udp, tcp)
	case "NOTIFY":
		s.onMessage(m, transport, udp, tcp)
	case "INVITE":
		s.onCascadeInvite(m, transport, udp, tcp)
	case "INFO":
		s.reply(m, 200, "OK", nil, transport, udp, tcp)
	default:
		s.reply(m, 200, "OK", nil, transport, udp, tcp)
	}
}

func (s *Server) onResponse(m *Message) {
	s.mu.Lock()
	w := s.invites[m.CallID()]
	s.mu.Unlock()
	if w == nil {
		return
	}
	if m.StatusCode >= 200 {
		select {
		case w.ch <- inviteAnswer{code: m.StatusCode, sdp: string(m.Body), to: m.Get("To")}:
		default:
		}
	}
}

func (s *Server) onRegister(m *Message, transport string, udp *net.UDPAddr, tcp net.Conn) {
	id := userOf(m.Get("From"))
	if id == "" {
		id = userOf(m.URI)
	}
	ip, port := splitAddr(m.addr)
	ua := m.Get("User-Agent")
	if s.db.IsBlack(id, ip, ua) {
		s.reply(m, 403, "Forbidden", nil, transport, udp, tcp)
		return
	}
	expires := 3600
	if e := m.Get("Expires"); e != "" {
		expires, _ = strconv.Atoi(e)
	}
	if contact := m.Get("Contact"); strings.Contains(strings.ToLower(contact), "expires=") {
		for _, p := range strings.Split(contact, ";") {
			p = strings.TrimSpace(p)
			if strings.HasPrefix(strings.ToLower(p), "expires=") {
				expires, _ = strconv.Atoi(strings.TrimPrefix(p, "expires="))
			}
		}
	}
	authz := m.Get("Authorization")
	if authz == "" {
		nonce := randHex(16)
		s.mu.Lock()
		s.nonces[nonce] = time.Now().Add(2 * time.Minute)
		s.mu.Unlock()
		resp := s.baseResp(m, 401, "Unauthorized")
		resp.Set("WWW-Authenticate", digestChallenge{Realm: s.cfg.Realm, Nonce: nonce}.Header())
		s.send(resp, transport, udp, tcp)
		return
	}
	da := parseDigest(authz)
	s.mu.Lock()
	exp, ok := s.nonces[da.Nonce]
	if ok && time.Now().After(exp) {
		ok = false
	}
	s.mu.Unlock()
	pwd := s.db.DevicePassword(id, s.cfg.DevicePassword)
	if !ok || !checkDigest("REGISTER", pwd, da) {
		s.reply(m, 403, "Forbidden", nil, transport, udp, tcp)
		return
	}
	if expires == 0 {
		s.mu.Lock()
		delete(s.peers, id)
		s.mu.Unlock()
		s.db.SetDeviceOnline(id, false)
		s.reply(m, 200, "OK", nil, transport, udp, tcp)
		return
	}
	gb := "2016"
	if strings.Contains(ua, "2022") || strings.Contains(strings.ToLower(ua), "gb28181-2022") {
		gb = "2022"
	}
	p := &peer{ID: id, Transport: transport, UDP: udp, TCP: tcp, FromTag: tagOf(m.Get("From")), GBVer: gb, Charset: s.cfg.Charset, UA: ua, To: m.Get("To")}
	s.mu.Lock()
	s.peers[id] = p
	s.mu.Unlock()
	_ = s.db.UpsertDeviceFromRegister(&model.Device{
		ID: id, Name: id, GBVer: gb, CommandTransport: transport, RemoteIP: ip, RemotePort: port,
		ContactIP: ip, UA: ua, Charset: s.cfg.Charset,
		MediaTransport: s.cfg.DefaultMediaTransport, MediaTransportMode: s.cfg.DefaultMediaTransportMode,
		Password: "",
	})
	resp := s.baseResp(m, 200, "OK")
	resp.Set("Expires", strconv.Itoa(expires))
	resp.Set("Date", sipNow())
	s.send(resp, transport, udp, tcp)
	go func() {
		time.Sleep(300 * time.Millisecond)
		_ = s.QueryCatalog(id)
	}()
}

func (s *Server) onMessage(m *Message, transport string, udp *net.UDPAddr, tcp net.Conn) {
	s.reply(m, 200, "OK", nil, transport, udp, tcp)
	if len(m.Body) == 0 {
		return
	}
	env, err := manscdp.Decode(m.Body)
	if err != nil {
		log.Printf("MANSCDP 解析失败: %v", err)
		return
	}
	from := userOf(m.Get("From"))
	if s.isCascadeFrom(from) {
		s.onCascadeMessage(from, env, m, transport, udp, tcp)
		return
	}
	if env.Is2022() {
		s.db.SetGBVer(from, "2022")
	}
	switch env.CmdType {
	case "Keepalive":
		s.db.TouchKeepalive(from)
		s.mu.Lock()
		if p := s.peers[from]; p != nil {
			p.UDP, p.TCP, p.Transport = udp, tcp, transport
		}
		s.mu.Unlock()
	case "Catalog":
		s.absorbCatalog(from, env)
	case "Alarm":
		_ = s.db.AddAlarm(&model.Alarm{
			DeviceID: from, ChannelID: env.DeviceID, AlarmPriority: env.AlarmPriority,
			AlarmMethod: env.AlarmMethod, AlarmType: env.AlarmType, Time: env.AlarmTime, ExtInfo: env.Info,
		})
	case "RecordInfo":
		s.deliver(from, env)
	default:
		s.deliver(from, env)
	}
}

func (s *Server) absorbCatalog(deviceID string, env *manscdp.Envelope) {
	s.mu.Lock()
	buf := s.catalog[deviceID]
	if buf == nil {
		buf = &catalogBuf{seen: map[string]bool{}}
		s.catalog[deviceID] = buf
	}
	if env.SumNum > 0 {
		buf.sum = env.SumNum
	}
	for _, it := range env.DeviceList.Item {
		if it.DeviceID == "" || buf.seen[it.DeviceID] {
			continue
		}
		buf.seen[it.DeviceID] = true
		lon, _ := strconv.ParseFloat(it.Longitude, 64)
		lat, _ := strconv.ParseFloat(it.Latitude, 64)
		st := it.Status
		if st == "" {
			st = "ON"
		}
		buf.items = append(buf.items, model.Channel{
			ID: it.DeviceID, Name: it.Name, Manufacturer: it.Manufacturer, Model: it.Model,
			Owner: it.Owner, CivilCode: it.CivilCode, Address: it.Address, Parental: it.Parental,
			ParentID: it.ParentID, Secrecy: it.Secrecy, RegisterWay: it.RegisterWay, Status: st,
			Longitude: lon, Latitude: lat, PTZType: it.PTZType, IPAddress: it.IPAddress, Port: it.Port,
			Block: it.Block, Firmware: it.Firmware, SerialNumber: it.SerialNumber, DownloadSpeed: it.DownloadSpeed,
		})
	}
	done := buf.sum > 0 && len(buf.items) >= buf.sum
	items := append([]model.Channel(nil), buf.items...)
	sum := buf.sum
	if done {
		delete(s.catalog, deviceID)
	} else if sum > 0 {
		s.db.SetCatalogProgress(deviceID, fmt.Sprintf("%d/%d", len(items), sum))
	}
	s.mu.Unlock()
	if done {
		_ = s.db.ReplaceCatalog(deviceID, items, s.cfg.ChannelDefaultOndemand, s.cfg.ChannelDefaultCloudRecord, true)
		s.deliver(deviceID, env)
	}
}

func (s *Server) onSubscribe(m *Message, transport string, udp *net.UDPAddr, tcp net.Conn) {
	resp := s.baseResp(m, 200, "OK")
	resp.Set("Expires", "90")
	s.send(resp, transport, udp, tcp)
}

func (s *Server) deliver(deviceID string, env *manscdp.Envelope) {
	key := deviceID + "|" + env.CmdType + "|" + env.SN
	s.mu.Lock()
	w := s.waits[key]
	s.mu.Unlock()
	if w == nil {
		return
	}
	select {
	case w.ch <- env:
	default:
	}
}

func (s *Server) reply(req *Message, code int, reason string, body []byte, transport string, udp *net.UDPAddr, tcp net.Conn) {
	resp := s.baseResp(req, code, reason)
	if body != nil {
		resp.Body = body
		resp.Set("Content-Type", "Application/MANSCDP+xml")
	}
	s.send(resp, transport, udp, tcp)
}

func (s *Server) baseResp(req *Message, code int, reason string) *Message {
	resp := &Message{Response: true, StatusCode: code, Reason: reason, Header: map[string][]string{}}
	resp.Set("Via", viaWithSource(req.Get("Via"), req.addr))
	from := req.Get("From")
	to := req.Get("To")
	if tagOf(to) == "" {
		to = strings.TrimRight(to, " ") + ";tag=" + newTag()
	}
	resp.Set("From", from)
	resp.Set("To", to)
	resp.Set("Call-ID", req.Get("Call-ID"))
	resp.Set("CSeq", req.Get("CSeq"))
	resp.Set("User-Agent", "ArgusGBS")
	if c := req.Get("Contact"); c != "" {
		resp.Set("Contact", c)
	}
	return resp
}

func (s *Server) send(m *Message, transport string, udp *net.UDPAddr, tcp net.Conn) {
	b := m.Bytes()
	if s.cfg.SIPLog {
		if m.Response {
			log.Printf("SIP -> %d %s", m.StatusCode, m.CSeq())
		} else {
			log.Printf("SIP -> %s %s", m.Method, m.URI)
		}
	}
	if transport == "TCP" && tcp != nil {
		_, _ = tcp.Write(b)
		return
	}
	if udp != nil && s.udp != nil {
		_, _ = s.udp.WriteToUDP(b, udp)
	}
}

func (s *Server) sendTo(id string, m *Message) error {
	s.mu.Lock()
	p := s.peers[id]
	s.mu.Unlock()
	if p == nil {
		return fmt.Errorf("设备离线")
	}
	s.send(m, p.Transport, p.UDP, p.TCP)
	return nil
}

func (s *Server) nextSN() string {
	return strconv.FormatInt(s.sn.Add(1), 10)
}

func (s *Server) request(method, deviceID, uri string, body []byte, contentType string) (*Message, error) {
	s.mu.Lock()
	p := s.peers[deviceID]
	s.mu.Unlock()
	if p == nil {
		return nil, fmt.Errorf("设备离线")
	}
	target := deviceID
	if uri != "" && !strings.HasPrefix(uri, "sip:") {
		target = uri
		uri = ""
	}
	if uri == "" {
		uri = fmt.Sprintf("sip:%s@%s", target, s.cfg.Realm)
	}
	m := &Message{Method: method, URI: uri, Header: map[string][]string{}}
	m.Set("Via", fmt.Sprintf("SIP/2.0/%s %s:%d;rport;branch=%s", p.Transport, s.ip, s.cfg.SIPPort, branch()))
	m.Set("From", fmt.Sprintf("<sip:%s@%s>;tag=%s", s.cfg.Serial, s.cfg.Realm, newTag()))
	m.Set("To", fmt.Sprintf("<sip:%s@%s>", target, s.cfg.Realm))
	m.Set("Call-ID", newCallID())
	m.Set("CSeq", fmt.Sprintf("%d %s", s.cseq.Add(1), method))
	m.Set("Max-Forwards", "70")
	m.Set("User-Agent", "ArgusGBS")
	m.Set("Contact", fmt.Sprintf("<sip:%s@%s:%d>", s.cfg.Serial, s.ip, s.cfg.SIPPort))
	if body != nil {
		m.Body = body
		m.Set("Content-Type", contentType)
	}
	if err := s.sendTo(deviceID, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Server) QueryXML(deviceID, cmd, target string, extra [][2]string) (*manscdp.Envelope, error) {
	sn := s.nextSN()
	fields := [][2]string{{"CmdType", cmd}, {"SN", sn}, {"DeviceID", target}}
	fields = append(fields, extra...)
	charset := s.cfg.Charset
	body, err := manscdp.Encode("Query", charset, fields)
	if err != nil {
		return nil, err
	}
	key := deviceID + "|" + cmd + "|" + sn
	w := &waitXML{ch: make(chan *manscdp.Envelope, 1)}
	s.mu.Lock()
	s.waits[key] = w
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.waits, key)
		s.mu.Unlock()
	}()
	if _, err := s.request("MESSAGE", deviceID, "", body, "Application/MANSCDP+xml"); err != nil {
		return nil, err
	}
	timer := time.NewTimer(time.Duration(s.cfg.AckTimeout) * time.Second)
	defer timer.Stop()
	select {
	case env := <-w.ch:
		return env, nil
	case <-timer.C:
		return nil, fmt.Errorf("设备应答超时")
	}
}

func (s *Server) Control(deviceID, channelID string, fields [][2]string) error {
	sn := s.nextSN()
	all := [][2]string{{"CmdType", "DeviceControl"}, {"SN", sn}, {"DeviceID", channelID}}
	all = append(all, fields...)
	body, err := manscdp.Encode("Control", s.cfg.Charset, all)
	if err != nil {
		return err
	}
	for _, f := range fields {
		if f[0] == "PTZCmd" {
			body = withControlPriority(body, 5)
			break
		}
	}
	target := channelID
	if target == "" {
		target = deviceID
	}
	_, err = s.request("MESSAGE", deviceID, target, body, "Application/MANSCDP+xml")
	return err
}

func (s *Server) QueryCatalog(deviceID string) error {
	s.mu.Lock()
	s.catalog[deviceID] = &catalogBuf{seen: map[string]bool{}}
	s.mu.Unlock()
	s.db.SetCatalogProgress(deviceID, "0/0")
	_, err := s.QueryXML(deviceID, "Catalog", deviceID, nil)
	return err
}

func (s *Server) Subscribe(deviceID, cmd string, expires int) error {
	sn := s.nextSN()
	body, err := manscdp.Encode("Query", s.cfg.Charset, [][2]string{
		{"CmdType", cmd}, {"SN", sn}, {"DeviceID", deviceID},
	})
	if err != nil {
		return err
	}
	s.mu.Lock()
	p := s.peers[deviceID]
	s.mu.Unlock()
	if p == nil {
		return fmt.Errorf("设备离线")
	}
	m := &Message{Method: "SUBSCRIBE", URI: fmt.Sprintf("sip:%s@%s", deviceID, s.cfg.Realm), Header: map[string][]string{}}
	m.Set("Via", fmt.Sprintf("SIP/2.0/%s %s:%d;rport;branch=%s", p.Transport, s.ip, s.cfg.SIPPort, branch()))
	m.Set("From", fmt.Sprintf("<sip:%s@%s>;tag=%s", s.cfg.Serial, s.cfg.Realm, newTag()))
	m.Set("To", fmt.Sprintf("<sip:%s@%s>", deviceID, s.cfg.Realm))
	m.Set("Call-ID", newCallID())
	m.Set("CSeq", fmt.Sprintf("%d SUBSCRIBE", s.cseq.Add(1)))
	m.Set("Event", "presence")
	m.Set("Expires", strconv.Itoa(expires))
	m.Set("Contact", fmt.Sprintf("<sip:%s@%s:%d>", s.cfg.Serial, s.ip, s.cfg.SIPPort))
	m.Set("User-Agent", "ArgusGBS")
	m.Body = body
	m.Set("Content-Type", "Application/MANSCDP+xml")
	return s.sendTo(deviceID, m)
}

type PlayOpt struct {
	DeviceID  string
	ChannelID string
	Transport string
	Mode      string
	Start     string
	End       string
	Download  bool
	Speed     int
	SSRC      string
	Subject   string
	StreamNum int
	RecvIP    string
	RecvPort  int
}

func (s *Server) Invite(opt PlayOpt) (callID string, answer string, err error) {
	s.mu.Lock()
	p := s.peers[opt.DeviceID]
	s.mu.Unlock()
	if p == nil {
		return "", "", fmt.Errorf("设备离线")
	}
	sdp := buildSDP(opt.ChannelID, opt.RecvIP, opt.RecvPort, opt.Transport, opt.Mode, opt.SSRC, opt.Start, opt.End, opt.Download)
	uri := fmt.Sprintf("sip:%s@%s", opt.ChannelID, s.cfg.Realm)
	m := &Message{Method: "INVITE", URI: uri, Header: map[string][]string{}}
	m.Set("Via", fmt.Sprintf("SIP/2.0/%s %s:%d;rport;branch=%s", p.Transport, s.ip, s.cfg.SIPPort, branch()))
	m.Set("From", fmt.Sprintf("<sip:%s@%s>;tag=%s", s.cfg.Serial, s.cfg.Realm, newTag()))
	m.Set("To", fmt.Sprintf("<sip:%s@%s>", opt.ChannelID, s.cfg.Realm))
	callID = newCallID()
	m.Set("Call-ID", callID)
	m.Set("CSeq", fmt.Sprintf("%d INVITE", s.cseq.Add(1)))
	m.Set("Contact", fmt.Sprintf("<sip:%s@%s:%d>", s.cfg.Serial, s.ip, s.cfg.SIPPort))
	m.Set("Subject", fmt.Sprintf("%s:%s,%s:0", opt.ChannelID, opt.SSRC, s.cfg.Serial))
	m.Set("Content-Type", "Application/SDP")
	m.Set("Max-Forwards", "70")
	m.Set("User-Agent", "ArgusGBS")
	m.Body = []byte(sdp)
	w := &inviteWait{ch: make(chan inviteAnswer, 1)}
	s.mu.Lock()
	s.invites[callID] = w
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.invites, callID)
		s.mu.Unlock()
	}()
	if err = s.sendTo(opt.DeviceID, m); err != nil {
		return "", "", err
	}
	timer := time.NewTimer(time.Duration(s.cfg.AckTimeout) * time.Second)
	defer timer.Stop()
	select {
	case ans := <-w.ch:
		if ans.code >= 300 {
			return callID, ans.sdp, fmt.Errorf("设备拒绝播放(%d)", ans.code)
		}
		to := ans.to
		if to == "" {
			to = m.Get("To")
		}
		s.mu.Lock()
		s.dialogs[callID] = &callDialog{deviceID: opt.DeviceID, from: m.Get("From"), to: to, uri: uri}
		s.mu.Unlock()
		s.ack(p, m, to)
		if strings.EqualFold(opt.Transport, "TCP") && strings.EqualFold(opt.Mode, "active") {
			rip, rport := parseSDPIPPort(ans.sdp)
			if rip != "" && rport > 0 && s.media != nil {
				_, _ = s.media.Relay(opt.SSRC, "TCP", "active", rip, rport, opt.SSRC, "recv")
			}
		}
		return callID, ans.sdp, nil
	case <-timer.C:
		return callID, "", fmt.Errorf("点播超时")
	}
}

func (s *Server) ack(p *peer, invite *Message, to string) {
	m := &Message{Method: "ACK", URI: invite.URI, Header: map[string][]string{}}
	m.Set("Via", fmt.Sprintf("SIP/2.0/%s %s:%d;rport;branch=%s", p.Transport, s.ip, s.cfg.SIPPort, branch()))
	m.Set("From", invite.Get("From"))
	if to == "" {
		to = invite.Get("To")
	}
	m.Set("To", to)
	m.Set("Call-ID", invite.Get("Call-ID"))
	m.Set("CSeq", strings.Replace(invite.Get("CSeq"), "INVITE", "ACK", 1))
	m.Set("Contact", fmt.Sprintf("<sip:%s@%s:%d>", s.cfg.Serial, s.ip, s.cfg.SIPPort))
	m.Set("Max-Forwards", "70")
	m.Set("User-Agent", "ArgusGBS")
	_ = s.sendTo(p.ID, m)
}

func (s *Server) Bye(deviceID, channelID, callID string) {
	s.mu.Lock()
	p := s.peers[deviceID]
	d := s.dialogs[callID]
	delete(s.dialogs, callID)
	delete(s.invites, callID)
	s.mu.Unlock()
	if p == nil {
		return
	}
	uri := fmt.Sprintf("sip:%s@%s", channelID, s.cfg.Realm)
	from := fmt.Sprintf("<sip:%s@%s>;tag=%s", s.cfg.Serial, s.cfg.Realm, newTag())
	to := fmt.Sprintf("<sip:%s@%s>", channelID, s.cfg.Realm)
	if d != nil {
		if d.uri != "" {
			uri = d.uri
		}
		if d.from != "" {
			from = d.from
		}
		if d.to != "" {
			to = d.to
		}
	}
	m := &Message{Method: "BYE", URI: uri, Header: map[string][]string{}}
	m.Set("Via", fmt.Sprintf("SIP/2.0/%s %s:%d;rport;branch=%s", p.Transport, s.ip, s.cfg.SIPPort, branch()))
	m.Set("From", from)
	m.Set("To", to)
	m.Set("Call-ID", callID)
	m.Set("CSeq", fmt.Sprintf("%d BYE", s.cseq.Add(1)))
	m.Set("Contact", fmt.Sprintf("<sip:%s@%s:%d>", s.cfg.Serial, s.ip, s.cfg.SIPPort))
	m.Set("Max-Forwards", "70")
	m.Set("User-Agent", "ArgusGBS")
	_ = s.sendTo(deviceID, m)
	go func() { _ = s.QueryCatalog(deviceID) }()
}

func (s *Server) Info(deviceID, channelID, callID, body string) error {
	s.mu.Lock()
	p := s.peers[deviceID]
	d := s.dialogs[callID]
	s.mu.Unlock()
	if p == nil {
		return fmt.Errorf("设备离线")
	}
	uri := fmt.Sprintf("sip:%s@%s", channelID, s.cfg.Realm)
	from := fmt.Sprintf("<sip:%s@%s>;tag=%s", s.cfg.Serial, s.cfg.Realm, newTag())
	to := fmt.Sprintf("<sip:%s@%s>", channelID, s.cfg.Realm)
	if d != nil {
		if d.uri != "" {
			uri = d.uri
		}
		if d.from != "" {
			from = d.from
		}
		if d.to != "" {
			to = d.to
		}
	}
	m := &Message{Method: "INFO", URI: uri, Header: map[string][]string{}}
	m.Set("Via", fmt.Sprintf("SIP/2.0/%s %s:%d;rport;branch=%s", p.Transport, s.ip, s.cfg.SIPPort, branch()))
	m.Set("From", from)
	m.Set("To", to)
	m.Set("Call-ID", callID)
	m.Set("CSeq", fmt.Sprintf("%d INFO", s.cseq.Add(1)))
	m.Set("Contact", fmt.Sprintf("<sip:%s@%s:%d>", s.cfg.Serial, s.ip, s.cfg.SIPPort))
	m.Set("Content-Type", "Application/MANSRTSP")
	m.Set("User-Agent", "ArgusGBS")
	m.Body = []byte(body)
	return s.sendTo(deviceID, m)
}

func buildSDP(serial, ip string, port int, transport, mode, ssrc, start, end string, download bool) string {
	media := "RTP/AVP"
	setup := ""
	if strings.EqualFold(transport, "TCP") {
		media = "TCP/RTP/AVP"
		if strings.EqualFold(mode, "active") {
			setup = "a=setup:active\r\n"
		} else {
			setup = "a=setup:passive\r\n"
		}
		setup += "a=connection:new\r\n"
	}
	name := "Play"
	t := "0 0"
	if start != "" && end != "" {
		name = "Playback"
		if download {
			name = "Download"
		}
		ts, _ := time.ParseInLocation("2006-01-02T15:04:05", start, time.Local)
		te, _ := time.ParseInLocation("2006-01-02T15:04:05", end, time.Local)
		t = fmt.Sprintf("%d %d", ts.Unix(), te.Unix())
	}
	payloads := "96"
	maps := "a=rtpmap:96 PS/90000\r\n"
	if !strings.EqualFold(transport, "TCP") {
		payloads = "96 97 98"
		maps += "a=rtpmap:97 MPEG4/90000\r\na=rtpmap:98 H264/90000\r\n"
	}
	return fmt.Sprintf("v=0\r\no=%s 0 0 IN IP4 %s\r\ns=%s\r\nc=IN IP4 %s\r\nt=%s\r\nm=video %d %s %s\r\na=recvonly\r\n%s%s",
		serial, ip, name, ip, t, port, media, payloads, maps, setup) + fmt.Sprintf("y=%s\r\n", ssrc)
}

func parseSDPIPPort(sdp string) (string, int) {
	ip := ""
	port := 0
	for _, ln := range strings.Split(sdp, "\n") {
		ln = strings.TrimSpace(ln)
		if strings.HasPrefix(ln, "c=") {
			fs := strings.Fields(ln)
			if len(fs) > 0 {
				ip = fs[len(fs)-1]
			}
		}
		if strings.HasPrefix(ln, "m=") {
			fs := strings.Fields(ln)
			if len(fs) >= 2 {
				port, _ = strconv.Atoi(fs[1])
			}
		}
	}
	return ip, port
}

func (s *Server) keepaliveLoop() {
	tk := time.NewTicker(15 * time.Second)
	defer tk.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-tk.C:
			dead := s.db.MarkOfflineBefore(time.Now().Add(-time.Duration(s.cfg.KeepaliveTimeout) * time.Second))
			s.mu.Lock()
			for _, id := range dead {
				delete(s.peers, id)
			}
			s.mu.Unlock()
		}
	}
}

func viaWithSource(via string, addr net.Addr) string {
	if via == "" || addr == nil {
		return via
	}
	ip, port := splitAddr(addr)
	if ip == "" {
		return via
	}
	parts := strings.Split(via, ";")
	out := make([]string, 0, len(parts)+1)
	out = append(out, parts[0])
	wrote := false
	for _, p := range parts[1:] {
		pl := strings.ToLower(strings.TrimSpace(p))
		if pl == "rport" || strings.HasPrefix(pl, "rport=") || strings.HasPrefix(pl, "received=") {
			if !wrote {
				out = append(out, fmt.Sprintf("rport=%d", port), "received="+ip)
				wrote = true
			}
			continue
		}
		out = append(out, strings.TrimSpace(p))
	}
	return strings.Join(out, ";")
}

func withControlPriority(body []byte, pri int) []byte {
	info := []byte(fmt.Sprintf("<Info>\r\n<ControlPriority>%d</ControlPriority>\r\n</Info>\r\n", pri))
	end := []byte("</Control>")
	i := bytes.LastIndex(body, end)
	if i < 0 {
		return body
	}
	out := make([]byte, 0, len(body)+len(info))
	out = append(out, body[:i]...)
	out = append(out, info...)
	out = append(out, body[i:]...)
	return out
}

func splitAddr(a net.Addr) (string, int) {
	if a == nil {
		return "", 0
	}
	return splitHostPort(a.String())
}

func splitHostPort(s string) (string, int) {
	h, p, err := net.SplitHostPort(s)
	if err != nil {
		return s, 0
	}
	n, _ := strconv.Atoi(p)
	return h, n
}

func guessIP() string {
	c, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer c.Close()
	return c.LocalAddr().(*net.UDPAddr).IP.String()
}
