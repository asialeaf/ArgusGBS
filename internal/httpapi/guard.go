package httpapi

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/argus/argusgbs/internal/model"
)

const (
	roleSuper = "超级管理员"
	roleAdmin = "管理员"
)

func md5hex(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func isDefaultPassword(hash string) bool {
	h := strings.ToLower(hash)
	return h == md5hex("admin") || h == md5hex("Argus@123")
}

func hasAnyRole(u *model.User, roles ...string) bool {
	if u == nil {
		return false
	}
	set := map[string]bool{}
	for _, r := range strings.Split(u.Role, ",") {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		set[r] = true
		if r == roleSuper {
			return true
		}
	}
	for _, r := range roles {
		if set[r] {
			return true
		}
	}
	return false
}

func (a *API) authRoles(next http.HandlerFunc, roles ...string) http.HandlerFunc {
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
		if len(roles) > 0 && !hasAnyRole(u, roles...) {
			http.Error(w, "没有权限", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

func (a *API) auth(next http.HandlerFunc) http.HandlerFunc {
	return a.authRoles(next)
}

func remoteHost(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func firstIP(v string) string {
	v = strings.TrimSpace(strings.Split(v, ",")[0])
	if h, _, err := net.SplitHostPort(v); err == nil {
		return h
	}
	return v
}

func clientIP(r *http.Request) string {
	host := remoteHost(r)
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return host
	}
	if v := strings.TrimSpace(r.Header.Get("X-Real-IP")); v != "" {
		return firstIP(v)
	}
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		return firstIP(v)
	}
	return host
}

func (a *API) setSessionCookie(w http.ResponseWriter, r *http.Request, token string, maxAge int) {
	c := &http.Cookie{
		Name: "gbs_token", Value: token, Path: "/", HttpOnly: true,
		SameSite: http.SameSiteLaxMode, MaxAge: maxAge,
	}
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		c.Secure = true
	}
	http.SetCookie(w, c)
}

type loginGate struct {
	mu    sync.Mutex
	fails map[string]*loginFail
}

type loginFail struct {
	n      int
	window time.Time
	until  time.Time
}

func (a *API) gate() *loginGate {
	a.loginOnce.Do(func() {
		a.fails = &loginGate{fails: map[string]*loginFail{}}
	})
	return a.fails
}

func (g *loginGate) blocked(key string, now time.Time) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	f := g.fails[key]
	return f != nil && now.Before(f.until)
}

func (g *loginGate) fail(key string, now time.Time) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if len(g.fails) > 4000 {
		for k, f := range g.fails {
			if now.After(f.until) && now.Sub(f.window) > 30*time.Minute {
				delete(g.fails, k)
			}
		}
	}
	f := g.fails[key]
	if f == nil || now.Sub(f.window) > 10*time.Minute {
		f = &loginFail{window: now}
		g.fails[key] = f
	}
	f.n++
	if f.n >= 5 {
		f.until = now.Add(15 * time.Minute)
	}
}

func (g *loginGate) reset(key string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.fails, key)
}

func passwordClasses(s string) int {
	var digit, lower, upper, special bool
	for _, r := range s {
		switch {
		case unicode.IsDigit(r):
			digit = true
		case unicode.IsLower(r):
			lower = true
		case unicode.IsUpper(r):
			upper = true
		default:
			special = true
		}
	}
	n := 0
	for _, ok := range []bool{digit, lower, upper, special} {
		if ok {
			n++
		}
	}
	return n
}

func (a *API) passwordOK(plain string) string {
	minLen := a.Cfg.PwdLength
	if minLen < 6 {
		minLen = 6
	}
	level := a.Cfg.PwdLevel
	if level < 2 {
		level = 2
	}
	if len([]rune(plain)) < minLen {
		return "密码长度不足"
	}
	if passwordClasses(plain) < level {
		return "密码种类不足，需要数字、大小写字母或符号"
	}
	if isDefaultPassword(md5hex(plain)) {
		return "不能使用初始密码"
	}
	return ""
}

func (a *API) playQuery(streamID string) string {
	secret := a.Cfg.SMSAPISecret
	if secret == "" || streamID == "" {
		return ""
	}
	exp := time.Now().Add(12 * time.Hour).Unix()
	mac := hmac.New(sha256.New, []byte(secret))
	msg := streamID + "." + itoa64(exp)
	mac.Write([]byte(msg))
	return "e=" + itoa64(exp) + "&s=" + hex.EncodeToString(mac.Sum(nil))
}

func itoa64(n int64) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

func ctEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
