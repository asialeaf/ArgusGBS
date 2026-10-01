package metrics

import (
	"bufio"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Point struct {
	Time string  `json:"time"`
	Use  float64 `json:"use,omitempty"`
	Sent float64 `json:"sent,omitempty"`
	Recv float64 `json:"recv,omitempty"`
}

type Snap struct {
	Mem []Point
	CPU []Point
	Net []Point
}

var (
	lastCPU  uint64
	lastIdle uint64
	lastNet  uint64
	lastNetr uint64
	lastAt   time.Time
	memHist  []Point
	cpuHist  []Point
	netHist  []Point
)

func Hardware() string {
	return runtime.GOARCH + "/" + runtime.GOOS
}

func Running(start time.Time) string {
	d := time.Since(start).Round(time.Second)
	return d.String()
}

func Sample() Snap {
	now := time.Now().Format("15:04:05")
	mem := memUsed()
	cpu := cpuUsed()
	sent, recv := netMbps()
	memHist = append(trim(memHist), Point{Time: now, Use: mem})
	cpuHist = append(trim(cpuHist), Point{Time: now, Use: cpu})
	netHist = append(trim(netHist), Point{Time: now, Sent: sent, Recv: recv})
	return Snap{Mem: memHist, CPU: cpuHist, Net: netHist}
}

func trim(in []Point) []Point {
	if len(in) > 30 {
		return in[len(in)-30:]
	}
	return in
}

func memUsed() float64 {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0
	}
	defer f.Close()
	var total, avail float64
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fs := strings.Fields(sc.Text())
		if len(fs) < 2 {
			continue
		}
		n, _ := strconv.ParseFloat(fs[1], 64)
		switch fs[0] {
		case "MemTotal:":
			total = n
		case "MemAvailable:":
			avail = n
		}
	}
	if total == 0 {
		return 0
	}
	return (total - avail) / total
}

func cpuUsed() float64 {
	b, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0
	}
	line := strings.SplitN(string(b), "\n", 2)[0]
	fs := strings.Fields(line)
	if len(fs) < 5 {
		return 0
	}
	var total, idle uint64
	for i := 1; i < len(fs); i++ {
		n, _ := strconv.ParseUint(fs[i], 10, 64)
		total += n
		if i == 4 {
			idle = n
		}
	}
	var use float64
	if lastCPU > 0 && total > lastCPU {
		dt := float64(total - lastCPU)
		di := float64(idle - lastIdle)
		use = (dt - di) / dt
	}
	lastCPU, lastIdle = total, idle
	return use
}

func netMbps() (float64, float64) {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return 0, 0
	}
	defer f.Close()
	var rx, tx uint64
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if !strings.Contains(line, ":") || strings.HasPrefix(line, "Inter") || strings.HasPrefix(line, "face") {
			continue
		}
		parts := strings.Split(line, ":")
		name := strings.TrimSpace(parts[0])
		if name == "lo" {
			continue
		}
		fs := strings.Fields(parts[1])
		if len(fs) < 9 {
			continue
		}
		r, _ := strconv.ParseUint(fs[0], 10, 64)
		t, _ := strconv.ParseUint(fs[8], 10, 64)
		rx += r
		tx += t
	}
	now := time.Now()
	var sent, recv float64
	if !lastAt.IsZero() {
		sec := now.Sub(lastAt).Seconds()
		if sec > 0 {
			sent = float64(tx-lastNet) * 8 / sec / 1e6
			recv = float64(rx-lastNetr) * 8 / sec / 1e6
		}
	}
	lastAt, lastNet, lastNetr = now, tx, rx
	return sent, recv
}

func DiskGB(path string) (used, free float64) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, 0
	}
	total := float64(st.Blocks) * float64(st.Bsize) / (1 << 30)
	free = float64(st.Bavail) * float64(st.Bsize) / (1 << 30)
	used = total - free
	if used < 0 {
		used = 0
	}
	return round1(used), round1(free)
}

func round1(v float64) float64 {
	return float64(int(v*10)) / 10
}
