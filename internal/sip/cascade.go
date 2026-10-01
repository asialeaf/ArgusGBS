package sip

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/argus/argusgbs/internal/manscdp"
	"github.com/argus/argusgbs/internal/media"
	"github.com/argus/argusgbs/internal/model"

	"golang.org/x/text/encoding/simplifiedchinese"
)

func (s *Server) isCascadeFrom(from string) bool {
	list, err := s.db.EnabledCascades()
	if err != nil {
		return false
	}
	for _, c := range list {
		if c.Serial == from {
			return true
		}
	}
	return false
}

func (s *Server) cascadeBySerial(serial string) *model.Cascade {
	list, err := s.db.EnabledCascades()
	if err != nil {
		return nil
	}
	for i := range list {
		if list[i].Serial == serial {
			return &list[i]
		}
	}
	return nil
}

func (s *Server) onCascadeMessage(from string, env *manscdp.Envelope, m *Message, transport string, udp *net.UDPAddr, tcp net.Conn) {
	c := s.cascadeBySerial(from)
	if c == nil {
		return
	}
	root := env.XMLName.Local
	if root == "Query" && env.CmdType == "Catalog" {
		s.pushCatalog(c, env.SN, transport, udp, tcp, m)
		return
	}
	if env.CmdType == "Keepalive" || root == "Response" {
		s.db.SetCascadeOnline(c.ID, true, "OK")
	}
}

func (s *Server) pushCatalog(c *model.Cascade, sn, transport string, udp *net.UDPAddr, tcp net.Conn, req *Message) {
	items, err := s.db.SharedChannels(c.ID, c.ShareAllChannel)
	if err != nil {
		return
	}
	group := c.CatalogGroupSize
	if group <= 0 {
		group = 1
	}
	if len(items) == 0 {
		body, _ := manscdp.Encode("Response", c.Charset, [][2]string{
			{"CmdType", "Catalog"}, {"SN", sn}, {"DeviceID", c.LocalSerial}, {"SumNum", "0"},
		})
		s.reply(req, 200, "OK", body, transport, udp, tcp)
		return
	}
	// 目录通过 MESSAGE 上报，先对 Query 回 200，再分片推送。
	for i := 0; i < len(items); i += group {
		j := i + group
		if j > len(items) {
			j = len(items)
		}
		body := catalogXML(c.Charset, sn, orSerial(c), len(items), items[i:j])
		dst := fmt.Sprintf("sip:%s@%s:%d", c.Serial, c.Host, c.Port)
		msg := &Message{Method: "MESSAGE", URI: dst, Header: map[string][]string{}}
		msg.Set("Via", fmt.Sprintf("SIP/2.0/%s %s:%d;rport;branch=%s", c.CommandTransport, s.ip, s.cfg.SIPPort, branch()))
		msg.Set("From", fmt.Sprintf("<sip:%s@%s>;tag=%s", orSerial(c), s.cfg.Realm, newTag()))
		msg.Set("To", fmt.Sprintf("<sip:%s@%s>", c.Serial, c.Host))
		msg.Set("Call-ID", newCallID())
		msg.Set("CSeq", fmt.Sprintf("%d MESSAGE", s.cseq.Add(1)))
		msg.Set("Max-Forwards", "70")
		msg.Set("Content-Type", "Application/MANSCDP+xml")
		msg.Body = body
		s.sendRawCascade(c, msg)
	}
}

func orSerial(c *model.Cascade) string {
	if c.LocalSerial != "" {
		return c.LocalSerial
	}
	return c.Serial
}

func catalogXML(charset, sn, deviceID string, sum int, items []model.Channel) []byte {
	var b strings.Builder
	b.WriteString("<Response>\r\n")
	b.WriteString("<CmdType>Catalog</CmdType>\r\n")
	b.WriteString("<SN>" + sn + "</SN>\r\n")
	b.WriteString("<DeviceID>" + deviceID + "</DeviceID>\r\n")
	b.WriteString("<SumNum>" + strconv.Itoa(sum) + "</SumNum>\r\n")
	b.WriteString(fmt.Sprintf("<DeviceList Num=\"%d\">\r\n", len(items)))
	for _, it := range items {
		name := it.Name
		if it.CustomName != "" {
			name = it.CustomName
		}
		fmt.Fprintf(&b, "<Item><DeviceID>%s</DeviceID><Name>%s</Name><Manufacturer>%s</Manufacturer><Parental>%d</Parental><ParentID>%s</ParentID><Status>%s</Status></Item>\r\n",
			it.ID, xmlEscapeLocal(name), xmlEscapeLocal(it.Manufacturer), it.Parental, it.ParentID, orStatus(it.Status))
	}
	b.WriteString("</DeviceList>\r\n</Response>\r\n")
	return encodePayload(charset, b.String())
}

func xmlEscapeLocal(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")
	return r.Replace(s)
}

func orStatus(s string) string {
	if s == "" {
		return "ON"
	}
	return s
}

func encodePayload(charset, payload string) []byte {
	header := "<?xml version=\"1.0\" encoding=\"GB2312\"?>\r\n"
	raw := []byte(payload)
	if strings.EqualFold(charset, "UTF-8") || strings.EqualFold(charset, "UTF8") {
		header = "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\r\n"
		return append([]byte(header), raw...)
	}
	enc, err := simplifiedchinese.GBK.NewEncoder().Bytes(raw)
	if err != nil {
		return append([]byte(header), raw...)
	}
	return append([]byte(header), enc...)
}

func (s *Server) sendRawCascade(c *model.Cascade, m *Message) {
	b := m.Bytes()
	if strings.EqualFold(c.CommandTransport, "TCP") {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port), 5*time.Second)
		if err != nil {
			s.db.SetCascadeOnline(c.ID, false, err.Error())
			return
		}
		defer conn.Close()
		_, _ = conn.Write(b)
		return
	}
	if s.udp != nil {
		addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", c.Host, c.Port))
		if err == nil {
			_, _ = s.udp.WriteToUDP(b, addr)
		}
	}
}

func (s *Server) onCascadeInvite(m *Message, transport string, udp *net.UDPAddr, tcp net.Conn) {
	from := userOf(m.Get("From"))
	c := s.cascadeBySerial(from)
	if c == nil {
		s.reply(m, 404, "Not Found", nil, transport, udp, tcp)
		return
	}
	channelID := userOf(m.URI)
	deviceID, err := s.db.FindChannelOwner(channelID)
	if err != nil {
		s.reply(m, 404, "Not Found", nil, transport, udp, tcp)
		return
	}
	if s.media == nil {
		s.reply(m, 480, "Temporarily Unavailable", nil, transport, udp, tcp)
		return
	}
	ssrc := NewSSRC(false)
	opened, err := s.media.Open(media.OpenReq{StreamID: ssrc, Transport: "UDP", Mode: "passive", SSRC: ssrc})
	if err != nil {
		s.reply(m, 500, "Server Internal Error", nil, transport, udp, tcp)
		return
	}
	_, _, err = s.Invite(PlayOpt{
		DeviceID: deviceID, ChannelID: channelID, Transport: "UDP", Mode: "passive",
		SSRC: ssrc, RecvIP: opened.PublicIP, RecvPort: opened.Port,
	})
	if err != nil {
		_ = s.media.Close(ssrc)
		s.reply(m, 480, "Temporarily Unavailable", nil, transport, udp, tcp)
		return
	}
	uip, uport := parseSDPIPPort(string(m.Body))
	setup := "passive"
	if strings.Contains(string(m.Body), "a=setup:active") {
		setup = "active"
	}
	tr := "UDP"
	if strings.Contains(string(m.Body), "TCP/RTP") {
		tr = "TCP"
	}
	sdpPort := opened.Port
	sdpMode := "active"
	relayMode := "passive"
	if setup == "active" {
		sdpMode = "passive"
		relayMode = "active"
	}
	if uip != "" && uport > 0 {
		if p, err := s.media.Relay(ssrc, tr, relayMode, uip, uport, ssrc, "send"); err == nil && p > 0 {
			sdpPort = p
		}
	}
	sdp := buildSDP(orSerial(c), opened.PublicIP, sdpPort, tr, sdpMode, ssrc, "", "", false)
	sdp = strings.Replace(sdp, "a=recvonly", "a=sendonly", 1)
	resp := s.baseResp(m, 200, "OK")
	resp.Set("Content-Type", "Application/SDP")
	resp.Set("Contact", fmt.Sprintf("<sip:%s@%s:%d>", orSerial(c), s.ip, s.cfg.SIPPort))
	resp.Body = []byte(sdp)
	s.send(resp, transport, udp, tcp)
}

func (s *Server) cascadeLoop() {
	tk := time.NewTicker(20 * time.Second)
	defer tk.Stop()
	s.registerCascades()
	for {
		select {
		case <-s.stop:
			return
		case <-tk.C:
			s.registerCascades()
		}
	}
}

func (s *Server) registerCascades() {
	list, err := s.db.EnabledCascades()
	if err != nil {
		return
	}
	for i := range list {
		c := list[i]
		if err := s.registerOne(&c); err != nil {
			s.db.SetCascadeOnline(c.ID, false, err.Error())
			log.Printf("级联 %s 注册失败: %v", c.Name, err)
			continue
		}
		s.db.SetCascadeOnline(c.ID, true, "OK")
		s.cascadeKeepalive(&c)
	}
}

func (s *Server) registerOne(c *model.Cascade) error {
	local := orSerial(c)
	if local == c.Serial {
		local = s.cfg.Serial
	}
	uri := fmt.Sprintf("sip:%s@%s", c.Serial, c.Realm)
	host := c.LocalHost
	if host == "" {
		host = s.ip
	}
	port := c.LocalPort
	if port == 0 {
		port = s.cfg.SIPPort
	}
	send := func(auth string) (*Message, error) {
		m := &Message{Method: "REGISTER", URI: uri, Header: map[string][]string{}}
		m.Set("Via", fmt.Sprintf("SIP/2.0/%s %s:%d;rport;branch=%s", orEmpty(c.CommandTransport, "UDP"), host, port, branch()))
		m.Set("From", fmt.Sprintf("<sip:%s@%s>;tag=%s", local, c.Realm, newTag()))
		m.Set("To", fmt.Sprintf("<sip:%s@%s>", local, c.Realm))
		m.Set("Call-ID", newCallID())
		m.Set("CSeq", fmt.Sprintf("%d REGISTER", s.cseq.Add(1)))
		m.Set("Contact", fmt.Sprintf("<sip:%s@%s:%d>", local, host, port))
		m.Set("Expires", strconv.Itoa(orInt(c.RegisterTimeout, 3600)))
		m.Set("Max-Forwards", "70")
		m.Set("User-Agent", "ArgusGBS")
		if auth != "" {
			m.Set("Authorization", auth)
		}
		return s.transactCascade(c, m)
	}
	resp, err := send("")
	if err != nil {
		return err
	}
	if resp.StatusCode == 401 {
		da := parseWWW(resp.Get("WWW-Authenticate"))
		user := c.Username
		if user == "" {
			user = local
		}
		auth := buildDigest(user, c.Password, "REGISTER", uri, da.Realm, da.Nonce, da.QOP)
		resp, err = send(auth)
		if err != nil {
			return err
		}
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("注册返回 %d", resp.StatusCode)
	}
	return nil
}

type wwwAuth struct {
	Realm, Nonce, QOP string
}

func parseWWW(h string) wwwAuth {
	d := parseDigest(h)
	qop := d.QOP
	if qop == "" {
		qop = "auth"
	}
	if strings.Contains(strings.ToLower(h), "qop") && d.QOP == "" {
		qop = "auth"
	}
	return wwwAuth{Realm: d.Realm, Nonce: d.Nonce, QOP: qop}
}

func (s *Server) cascadeKeepalive(c *model.Cascade) {
	local := orSerial(c)
	if local == c.Serial {
		local = s.cfg.Serial
	}
	sn := s.nextSN()
	body, err := manscdp.Encode("Notify", c.Charset, [][2]string{
		{"CmdType", "Keepalive"}, {"SN", sn}, {"DeviceID", local}, {"Status", "OK"},
	})
	if err != nil {
		return
	}
	m := &Message{Method: "MESSAGE", URI: fmt.Sprintf("sip:%s@%s:%d", c.Serial, c.Host, c.Port), Header: map[string][]string{}}
	m.Set("Via", fmt.Sprintf("SIP/2.0/%s %s:%d;rport;branch=%s", orEmpty(c.CommandTransport, "UDP"), s.ip, s.cfg.SIPPort, branch()))
	m.Set("From", fmt.Sprintf("<sip:%s@%s>;tag=%s", local, s.cfg.Realm, newTag()))
	m.Set("To", fmt.Sprintf("<sip:%s@%s>", c.Serial, c.Host))
	m.Set("Call-ID", newCallID())
	m.Set("CSeq", fmt.Sprintf("%d MESSAGE", s.cseq.Add(1)))
	m.Set("Content-Type", "Application/MANSCDP+xml")
	m.Set("Max-Forwards", "70")
	m.Body = body
	_, _ = s.transactCascade(c, m)
}

func (s *Server) transactCascade(c *model.Cascade, m *Message) (*Message, error) {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	b := m.Bytes()
	if strings.EqualFold(c.CommandTransport, "TCP") {
		conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
		if err != nil {
			return nil, err
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(8 * time.Second))
		if _, err := conn.Write(b); err != nil {
			return nil, err
		}
		return ReadStream(bufio.NewReader(conn))
	}
	raddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(8 * time.Second))
	if _, err := conn.Write(b); err != nil {
		return nil, err
	}
	buf := make([]byte, 64*1024)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return Parse(buf[:n])
}

func orEmpty(a, b string) string {
	if a == "" {
		return b
	}
	return a
}

func orInt(a, b int) int {
	if a == 0 {
		return b
	}
	return a
}

func NewSSRC(playback bool) string {
	n := time.Now().UnixNano() % 1000000000
	if playback {
		return fmt.Sprintf("1%09d", n)
	}
	return fmt.Sprintf("0%09d", n)
}
