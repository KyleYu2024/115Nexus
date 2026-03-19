package utils

import (
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"115nexus/internal/models"
)

// 日志缓存
type LogBuffer struct {
	Lines       []string
	Limit       int
	mu          sync.RWMutex
	subscribers []chan string
}

func (l *LogBuffer) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	line := string(p)
	l.Lines = append(l.Lines, line)
	if len(l.Lines) > l.Limit {
		l.Lines = l.Lines[len(l.Lines)-l.Limit:]
	}

	for _, sub := range l.subscribers {
		select {
		case sub <- line:
		default:
		}
	}
	return len(p), nil
}

func (l *LogBuffer) GetLogs() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var sb strings.Builder
	for _, line := range l.Lines {
		sb.WriteString(line)
	}
	return sb.String()
}

func (l *LogBuffer) Subscribe() chan string {
	l.mu.Lock()
	defer l.mu.Unlock()
	ch := make(chan string, 100)
	l.subscribers = append(l.subscribers, ch)
	return ch
}

func (l *LogBuffer) Unsubscribe(ch chan string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i, sub := range l.subscribers {
		if sub == ch {
			l.subscribers = append(l.subscribers[:i], l.subscribers[i+1:]...)
			close(ch)
			break
		}
	}
}

var (
	GlobalLogBuffer = &LogBuffer{Limit: 500}
	
	httpClient    *http.Client
	proxyClient   *http.Client
	currentProxy  string
	clientMutex   sync.Mutex
)

// 获取直连客户端 (复用)
func GetDirectClient() *http.Client {
	clientMutex.Lock()
	defer clientMutex.Unlock()
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				IdleConnTimeout:     90 * time.Second,
				MaxIdleConnsPerHost: 20,
			},
		}
	}
	return httpClient
}

// 获取代理客户端 (支持动态配置更新)
func GetProxyClient(cfg models.BotConfig) *http.Client {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	if proxyClient != nil && currentProxy == cfg.ProxyUrl {
		return proxyClient
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		MaxIdleConnsPerHost: 20,
	}

	if cfg.ProxyUrl != "" {
		if u, err := url.Parse(cfg.ProxyUrl); err == nil {
			transport.Proxy = http.ProxyURL(u)
		}
	}

	proxyClient = &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
	currentProxy = cfg.ProxyUrl
	return proxyClient
}

func Extract115Link(text string) string {
	re := regexp.MustCompile(`(https?://(?:[a-zA-Z0-9-]+\.)?(?:115\.com|anxia\.com|115cdn\.com|115\.cn)/s/[a-zA-Z0-9]+(?:\?password=[a-zA-Z0-9]+)?)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func ParseSizeToMB(s string) float64 {
	s = strings.ToUpper(strings.ReplaceAll(s, " ", ""))
	re := regexp.MustCompile(`([\d\.]+)([KMGTP]?B?)`)
	m := re.FindStringSubmatch(s)
	if len(m) < 3 { return 0 }
	v, _ := strconv.ParseFloat(m[1], 64)
	u := m[2]
	switch {
	case strings.Contains(u, "T"): v *= 1024 * 1024 * 1024 * 1024
	case strings.Contains(u, "G"): v *= 1024 * 1024 * 1024
	case strings.Contains(u, "M"): v *= 1024 * 1024
	case strings.Contains(u, "K"): v *= 1024
	}
	return v / (1024 * 1024)
}

func EstimateEpisodeCount(title string) int {
	re := regexp.MustCompile(`(?i)(?:全|共|第)?\s*(\d+)\s*(?:集|Episodes|Ep|E)`)
	if m := re.FindStringSubmatch(title); len(m) > 1 {
		v, _ := strconv.Atoi(m[1])
		return v
	}
	return 1
}

func TgEscape(s string) string {
	r := strings.NewReplacer("_","\\_","*","\\*","[","\\[","]","\\]","(","\\(",")","\\)","~","\\~","`","\\`",">","\\>","#","\\#","+","\\+","-","\\-","=","\\=","|","\\|","{","\\{","}","\\}",".","\\.","!","\\!")
	return r.Replace(s)
}

func UniqueTags(tags []string) []string {
	m := make(map[string]bool)
	var res []string
	for _, t := range tags {
		t = strings.ToUpper(strings.TrimSpace(t))
		if t != "" && !m[t] {
			m[t] = true
			res = append(res, t)
		}
	}
	return res
}

func IsExcluded(name string, rules string) bool {
	if rules == "" { return false }
	for _, r := range strings.Split(rules, "\n") {
		r = strings.TrimSpace(r)
		if r == "" { continue }
		if !strings.HasPrefix(r, "(?i)") { r = "(?i)" + r }
		if m, _ := regexp.MatchString(r, name); m { return true }
	}
	return false
}
