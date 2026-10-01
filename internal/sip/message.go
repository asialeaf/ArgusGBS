package sip

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type Message struct {
	Method     string
	URI        string
	StatusCode int
	Reason     string
	Response   bool
	Header     map[string][]string
	Body       []byte
	addr       net.Addr
}

func (m *Message) Get(k string) string {
	vs := m.Header[strings.ToLower(k)]
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}

func (m *Message) Add(k, v string) {
	if m.Header == nil {
		m.Header = map[string][]string{}
	}
	m.Header[strings.ToLower(k)] = append(m.Header[strings.ToLower(k)], v)
}

func (m *Message) Set(k, v string) {
	if m.Header == nil {
		m.Header = map[string][]string{}
	}
	m.Header[strings.ToLower(k)] = []string{v}
}

func (m *Message) CallID() string { return m.Get("Call-ID") }
func (m *Message) CSeq() string   { return m.Get("CSeq") }

func Parse(b []byte) (*Message, error) {
	b = bytes.TrimSpace(b)
	head, body, _ := bytes.Cut(b, []byte("\r\n\r\n"))
	if !bytes.Contains(b, []byte("\r\n\r\n")) {
		head, body, _ = bytes.Cut(b, []byte("\n\n"))
	}
	lines := bytes.Split(head, []byte("\n"))
	if len(lines) == 0 {
		return nil, fmt.Errorf("空报文")
	}
	m := &Message{Header: map[string][]string{}}
	start := strings.TrimSpace(string(lines[0]))
	parts := strings.Split(start, " ")
	if strings.HasPrefix(start, "SIP/2.0") {
		m.Response = true
		if len(parts) >= 2 {
			m.StatusCode, _ = strconv.Atoi(parts[1])
		}
		if len(parts) >= 3 {
			m.Reason = strings.Join(parts[2:], " ")
		}
	} else {
		if len(parts) >= 1 {
			m.Method = parts[0]
		}
		if len(parts) >= 2 {
			m.URI = parts[1]
		}
	}
	for _, ln := range lines[1:] {
		s := strings.TrimRight(string(ln), "\r")
		if s == "" {
			continue
		}
		k, v, ok := strings.Cut(s, ":")
		if !ok {
			continue
		}
		m.Add(strings.TrimSpace(k), strings.TrimSpace(v))
	}
	n := 0
	if cl := m.Get("Content-Length"); cl != "" {
		n, _ = strconv.Atoi(strings.TrimSpace(cl))
	}
	if n > 0 && len(body) >= n {
		m.Body = body[:n]
	} else {
		m.Body = body
	}
	return m, nil
}

func (m *Message) Bytes() []byte {
	var buf bytes.Buffer
	if m.Response {
		reason := m.Reason
		if reason == "" {
			reason = reasonText(m.StatusCode)
		}
		fmt.Fprintf(&buf, "SIP/2.0 %d %s\r\n", m.StatusCode, reason)
	} else {
		fmt.Fprintf(&buf, "%s %s SIP/2.0\r\n", m.Method, m.URI)
	}
	if m.Body != nil {
		m.Set("Content-Length", strconv.Itoa(len(m.Body)))
	} else if m.Get("Content-Length") == "" {
		m.Set("Content-Length", "0")
	}
	order := []string{"via", "from", "to", "call-id", "cseq", "contact", "expires", "date", "user-agent", "max-forwards", "www-authenticate", "authorization", "content-type", "subject", "content-length"}
	seen := map[string]bool{}
	for _, k := range order {
		for _, v := range m.Header[k] {
			fmt.Fprintf(&buf, "%s: %s\r\n", headerName(k), v)
		}
		if len(m.Header[k]) > 0 {
			seen[k] = true
		}
	}
	for k, vs := range m.Header {
		if seen[k] {
			continue
		}
		for _, v := range vs {
			fmt.Fprintf(&buf, "%s: %s\r\n", headerName(k), v)
		}
	}
	buf.WriteString("\r\n")
	buf.Write(m.Body)
	return buf.Bytes()
}

func headerName(k string) string {
	switch k {
	case "call-id":
		return "Call-ID"
	case "cseq":
		return "CSeq"
	case "www-authenticate":
		return "WWW-Authenticate"
	case "content-type":
		return "Content-Type"
	case "content-length":
		return "Content-Length"
	case "max-forwards":
		return "Max-Forwards"
	case "user-agent":
		return "User-Agent"
	default:
		if len(k) == 0 {
			return k
		}
		return strings.ToUpper(k[:1]) + k[1:]
	}
}

func reasonText(code int) string {
	switch code {
	case 200:
		return "OK"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 408:
		return "Request Timeout"
	case 480:
		return "Temporarily Unavailable"
	case 486:
		return "Busy Here"
	case 487:
		return "Request Terminated"
	case 500:
		return "Server Internal Error"
	default:
		return "OK"
	}
}

func ReadStream(r *bufio.Reader) (*Message, error) {
	var head bytes.Buffer
	for {
		line, err := r.ReadSlice('\n')
		if err != nil {
			return nil, err
		}
		head.Write(line)
		if bytes.Equal(line, []byte("\r\n")) || bytes.Equal(line, []byte("\n")) {
			break
		}
	}
	m, err := Parse(append(head.Bytes(), []byte("\r\n")...))
	if err != nil {
		return nil, err
	}
	n := 0
	if cl := m.Get("Content-Length"); cl != "" {
		n, _ = strconv.Atoi(strings.TrimSpace(cl))
	}
	if n > 0 {
		body := make([]byte, n)
		if _, err := io.ReadFull(r, body); err != nil {
			return nil, err
		}
		m.Body = body
	}
	return m, nil
}

type digestChallenge struct {
	Realm  string
	Nonce  string
	Opaque string
}

func (c digestChallenge) Header() string {
	return fmt.Sprintf(`Digest realm="%s",nonce="%s",algorithm=MD5`, c.Realm, c.Nonce)
}

type digestAuth struct {
	Username string
	Realm    string
	Nonce    string
	URI      string
	Response string
	QOP      string
	NC       string
	CNonce   string
	Algorithm string
}

func parseDigest(h string) digestAuth {
	h = strings.TrimSpace(h)
	h = strings.TrimPrefix(h, "Digest")
	h = strings.TrimSpace(h)
	var a digestAuth
	for _, part := range splitAuth(h) {
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		k = strings.ToLower(strings.TrimSpace(k))
		v = strings.Trim(strings.TrimSpace(v), `"`)
		switch k {
		case "username":
			a.Username = v
		case "realm":
			a.Realm = v
		case "nonce":
			a.Nonce = v
		case "uri":
			a.URI = v
		case "response":
			a.Response = v
		case "qop":
			a.QOP = v
		case "nc":
			a.NC = v
		case "cnonce":
			a.CNonce = v
		case "algorithm":
			a.Algorithm = v
		}
	}
	return a
}

func splitAuth(s string) []string {
	var out []string
	var cur strings.Builder
	q := false
	for _, r := range s {
		switch r {
		case '"':
			q = !q
			cur.WriteRune(r)
		case ',':
			if q {
				cur.WriteRune(r)
			} else {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

func md5hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func checkDigest(method, password string, a digestAuth) bool {
	ha1 := md5hex(a.Username + ":" + a.Realm + ":" + password)
	ha2 := md5hex(method + ":" + a.URI)
	var expect string
	if a.QOP == "" {
		expect = md5hex(ha1 + ":" + a.Nonce + ":" + ha2)
	} else {
		expect = md5hex(ha1 + ":" + a.Nonce + ":" + a.NC + ":" + a.CNonce + ":" + a.QOP + ":" + ha2)
	}
	return strings.EqualFold(expect, a.Response)
}

func buildDigest(username, password, method, uri, realm, nonce, qop string) string {
	ha1 := md5hex(username + ":" + realm + ":" + password)
	ha2 := md5hex(method + ":" + uri)
	cnonce := randHex(8)
	nc := "00000001"
	resp := md5hex(ha1 + ":" + nonce + ":" + nc + ":" + cnonce + ":" + qop + ":" + ha2)
	if qop == "" {
		resp = md5hex(ha1 + ":" + nonce + ":" + ha2)
		return fmt.Sprintf(`Digest username="%s",realm="%s",nonce="%s",uri="%s",response="%s",algorithm=MD5`, username, realm, nonce, uri, resp)
	}
	return fmt.Sprintf(`Digest username="%s",realm="%s",nonce="%s",uri="%s",response="%s",algorithm=MD5,cnonce="%s",qop=%s,nc=%s`, username, realm, nonce, uri, resp, cnonce, qop, nc)
}

func randHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func userOf(uri string) string {
	uri = strings.TrimSpace(uri)
	uri = strings.TrimPrefix(uri, "<")
	uri = strings.TrimSuffix(uri, ">")
	uri = strings.TrimPrefix(uri, "sip:")
	uri = strings.TrimPrefix(uri, "sips:")
	if i := strings.IndexAny(uri, "@;>"); i >= 0 {
		return uri[:i]
	}
	return uri
}

func tagOf(h string) string {
	for _, p := range strings.Split(h, ";") {
		p = strings.TrimSpace(p)
		if strings.HasPrefix(strings.ToLower(p), "tag=") {
			return strings.TrimPrefix(p, "tag=")
		}
	}
	return ""
}

func sipNow() string {
	return time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
}

func branch() string { return "z9hG4bK" + randHex(8) }
func newTag() string { return randHex(4) }
func newCallID() string {
	return randHex(8) + "@argus"
}
