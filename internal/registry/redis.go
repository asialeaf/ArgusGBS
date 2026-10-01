package registry

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/argus/argusgbs/internal/config"
	"github.com/redis/go-redis/v9"
)

// LiveGBS 用 Redis 哈希做节点登记：sipcms:{编号}、sipsms:{编号}，
// 流用 stream:/playback:/download:。键带保活超时，进程停了就会过期。

type Node struct {
	Serial   string
	Name     string
	Host     string
	Domain   string
	WanIP    string
	Realm    string
	Port     int
	HTTPPort int
	RTSPPort int
	RTMPPort int
	Load     int
}

type Store struct {
	rdb *redis.Client
	cfg *config.Config
	ttl time.Duration
}

func Start(cfg *config.Config) (*Store, error) {
	if err := ensureServer(cfg); err != nil {
		return nil, err
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, err
	}
	ttl := time.Duration(cfg.KeepaliveTimeout) * time.Second
	if ttl <= 0 {
		ttl = 300 * time.Second
	}
	s := &Store{rdb: rdb, cfg: cfg, ttl: ttl}
	if err := s.touchCMS(context.Background()); err != nil {
		_ = rdb.Close()
		return nil, err
	}
	go s.loop()
	log.Printf("Redis %s:%d 已登记 sipcms:%s", cfg.RedisHost, cfg.RedisPort, cfg.Serial)
	return s, nil
}

func (s *Store) loop() {
	tk := time.NewTicker(10 * time.Second)
	defer tk.Stop()
	for range tk.C {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		if err := s.touchCMS(ctx); err != nil {
			log.Printf("Redis 刷新信令登记失败: %v", err)
		}
		cancel()
	}
}

func (s *Store) touchCMS(ctx context.Context) error {
	host := s.cfg.SIPHost
	if host == "" {
		host = s.cfg.RedisHost
	}
	key := "sipcms:" + s.cfg.Serial
	fields := map[string]string{
		"Serial": s.cfg.Serial, "Realm": s.cfg.Realm, "Host": host, "Domain": host,
		"WanIP": host, "Port": strconv.Itoa(s.cfg.SIPPort),
		"Name": fmt.Sprintf("%s:%d", host, s.cfg.SIPPort),
		"AckTimeout": strconv.Itoa(s.cfg.AckTimeout), "KeepaliveTimeout": strconv.Itoa(s.cfg.KeepaliveTimeout),
		"AllowStreamStartByURL": b01(s.cfg.AllowStreamStartByURL), "AllowOptions": "0",
		"ProxySMS": "1", "RegisterResponseDateTime": "1", "Load": "0",
		"MaxSessionCount": "0", "StreamTimeout": "30",
	}
	if err := s.rdb.HSet(ctx, key, fields).Err(); err != nil {
		return err
	}
	return s.rdb.Expire(ctx, key, s.ttl).Err()
}

func (s *Store) ListSMS() []Node {
	if s == nil || s.rdb == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var out []Node
	var cursor uint64
	for {
		keys, next, err := s.rdb.Scan(ctx, cursor, "sipsms:*", 100).Result()
		if err != nil {
			return out
		}
		for _, key := range keys {
			m, err := s.rdb.HGetAll(ctx, key).Result()
			if err != nil || m["Serial"] == "" || m["Host"] == "" {
				continue
			}
			out = append(out, Node{
				Serial: m["Serial"], Name: m["Name"], Host: m["Host"], Domain: m["Domain"],
				WanIP: m["WanIP"], Realm: m["Realm"],
				Port: atoi(m["Port"]), HTTPPort: atoi(m["HTTPPort"]),
				RTSPPort: atoi(m["RTSPPort"]), RTMPPort: atoi(m["RTMPPort"]),
				Load: atoi(m["Load"]),
			})
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return out
}

// Pick 选一台在线流媒体。设备指定了 SMSID 时用那一台，否则选 Load 最小的。
func (s *Store) Pick(prefer string) *Node {
	nodes := s.ListSMS()
	if len(nodes) == 0 {
		return nil
	}
	if prefer != "" {
		for i := range nodes {
			if nodes[i].Serial == prefer {
				return &nodes[i]
			}
		}
	}
	best := &nodes[0]
	for i := range nodes {
		if nodes[i].Load < best.Load {
			best = &nodes[i]
		}
	}
	return best
}

func (s *Store) Find(serial string) *Node {
	for _, n := range s.ListSMS() {
		if n.Serial == serial {
			cp := n
			return &cp
		}
	}
	return nil
}

func (s *Store) PutStream(kind, id string, fields map[string]string) {
	if s == nil || s.rdb == nil || id == "" {
		return
	}
	if kind != "playback" && kind != "download" {
		kind = "stream"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	key := kind + ":" + id
	if err := s.rdb.HSet(ctx, key, fields).Err(); err != nil {
		log.Printf("Redis 写入 %s 失败: %v", key, err)
		return
	}
	_ = s.rdb.Expire(ctx, key, s.ttl).Err()
}

func (s *Store) RefreshStreams(liveIDs map[string]bool) {
	if s == nil || s.rdb == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	for _, prefix := range []string{"stream:", "playback:", "download:"} {
		var cursor uint64
		for {
			keys, next, err := s.rdb.Scan(ctx, cursor, prefix+"*", 100).Result()
			if err != nil {
				return
			}
			for _, key := range keys {
				id := strings.TrimPrefix(key, prefix)
				if liveIDs[id] {
					_ = s.rdb.Expire(ctx, key, s.ttl).Err()
				} else {
					_ = s.rdb.Del(ctx, key).Err()
				}
			}
			cursor = next
			if cursor == 0 {
				break
			}
		}
	}
}

func (s *Store) DropStream(id string) {
	if s == nil || s.rdb == nil || id == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = s.rdb.Del(ctx, "stream:"+id, "playback:"+id, "download:"+id).Err()
}

func ensureServer(cfg *config.Config) error {
	addr := fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort)
	if conn, err := net.DialTimeout("tcp", addr, 300*time.Millisecond); err == nil {
		conn.Close()
		return nil
	}
	bin := redisBin()
	if bin == "" {
		return fmt.Errorf("未找到 redis-server，且 %s 没有在监听", addr)
	}
	dir := "data/redis"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	conf := filepath.Join(abs, "redis.conf")
	body := fmt.Sprintf("bind 0.0.0.0\nprotected-mode yes\nport %d\nrequirepass %s\ndaemonize no\ndir %s\ndbfilename dump.rdb\nappendonly no\nsave \"\"\nlogfile \"\"\n",
		cfg.RedisPort, cfg.RedisPassword, abs)
	if err := os.WriteFile(conf, []byte(body), 0o644); err != nil {
		return err
	}
	cmd := exec.Command(bin, conf)
	cmd.Dir = abs
	logf, err := os.OpenFile(filepath.Join(abs, "redis.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err == nil {
		cmd.Stdout = logf
		cmd.Stderr = logf
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", cfg.RedisPort), 200*time.Millisecond)
		if err == nil {
			conn.Close()
			log.Printf("已启动 Redis %s 端口 %d", bin, cfg.RedisPort)
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("redis-server 已拉起但端口 %d 未就绪", cfg.RedisPort)
}

func redisBin() string {
	if p, err := exec.LookPath("redis-server"); err == nil {
		return p
	}
	cands := []string{
		os.Getenv("ARGUS_REDIS_BIN"),
		"/home/asialeaf/Argus/LiveCMS-linux-3.6.9-26093013/redis/redis-server",
	}
	for _, p := range cands {
		if p == "" {
			continue
		}
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}

func b01(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
