package httpapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"github.com/argus/argusgbs/internal/config"
)

func TestPlayQueryMatchesHMAC(t *testing.T) {
	a := &API{Cfg: &config.Config{SMSAPISecret: "argus-sms"}}
	q := a.playQuery("0123456789")
	if !strings.Contains(q, "e=") || !strings.Contains(q, "&s=") {
		t.Fatalf("query %s", q)
	}
	exp := q[strings.Index(q, "e=")+2 : strings.Index(q, "&s=")]
	sig := q[strings.Index(q, "&s=")+3:]
	mac := hmac.New(sha256.New, []byte("argus-sms"))
	mac.Write([]byte("0123456789." + exp))
	if hex.EncodeToString(mac.Sum(nil)) != sig {
		t.Fatalf("sig mismatch")
	}
	if time.Now().Add(13 * time.Hour).Unix() < atoi64(exp) {
		t.Fatalf("expiry too far")
	}
}

func TestPasswordPolicy(t *testing.T) {
	a := &API{Cfg: &config.Config{PwdLength: 6, PwdLevel: 2}}
	if a.passwordOK("admin") == "" {
		t.Fatal("default password accepted")
	}
	if a.passwordOK("ab1") == "" {
		t.Fatal("short password accepted")
	}
	if a.passwordOK("Admin1") != "" {
		t.Fatal("Admin1 rejected")
	}
}

func atoi64(s string) int64 {
	var n int64
	for _, c := range s {
		n = n*10 + int64(c-'0')
	}
	return n
}
