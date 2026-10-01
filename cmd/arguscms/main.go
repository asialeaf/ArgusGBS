package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/argus/argusgbs/internal/config"
	"github.com/argus/argusgbs/internal/httpapi"
	"github.com/argus/argusgbs/internal/media"
	"github.com/argus/argusgbs/internal/sip"
	"github.com/argus/argusgbs/internal/store"
)

func main() {
	cfgPath := flag.String("config", "configs/arguscms.ini", "配置文件")
	flag.Parse()
	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	applyEnv(cfg)
	db, err := store.Open(cfg.DBFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	mc := media.New(cfg.SMSHost, cfg.SMSHTTPPort, cfg.SMSAPISecret)
	sipSrv := sip.New(cfg, db, mc)
	if err := sipSrv.Start(); err != nil {
		log.Fatal(err)
	}
	api := &httpapi.API{
		Cfg: cfg, DB: db, SIP: sipSrv, Media: mc,
		WWW: findWWW(cfg.WWWDir), Started: time.Now(),
	}
	addr := ":" + itoa(cfg.HTTPPort)
	log.Printf("ArgusCMS HTTP %s  前端 %s", addr, api.WWW)
	log.Printf("默认账号 admin / admin")
	if cfg.HTTPSPort > 0 && cfg.HTTPSCert != "" && cfg.HTTPSKey != "" {
		go func() {
			haddr := ":" + itoa(cfg.HTTPSPort)
			log.Printf("ArgusCMS HTTPS %s", haddr)
			if err := http.ListenAndServeTLS(haddr, cfg.HTTPSCert, cfg.HTTPSKey, api.Handler()); err != nil {
				log.Printf("HTTPS 退出: %v", err)
			}
		}()
	}
	if err := http.ListenAndServe(addr, api.Handler()); err != nil {
		log.Fatal(err)
	}
}

func applyEnv(cfg *config.Config) {
	if v := os.Getenv("ARGUS_SIP_HOST"); v != "" {
		cfg.SIPHost = v
	}
	if v := os.Getenv("ARGUS_SMS_HOST"); v != "" {
		cfg.SMSHost = v
	}
	if v := os.Getenv("ARGUS_ADVERTISE_IP"); v != "" {
		cfg.SMSPublicHost = v
		if cfg.SIPHost == "" {
			cfg.SIPHost = v
		}
	}
}

func findWWW(p string) string {
	cands := []string{}
	if p != "" {
		cands = append(cands, p)
	}
	for _, c := range cands {
		st, err := os.Stat(filepath.Join(c, "index.html"))
		if err == nil && !st.IsDir() {
			abs, err := filepath.Abs(c)
			if err == nil {
				return abs
			}
			return c
		}
	}
	return ""
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [16]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
