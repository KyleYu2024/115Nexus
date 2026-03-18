package utils

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"net/http"
	"net/url"

	"115nexus/internal/models"
)

// 日志缓存
type LogBuffer struct {
	Lines []string
	Limit int
}

func (l *LogBuffer) Write(p []byte) (n int, err error) {
	line := string(p)
	l.Lines = append(l.Lines, line)
	if len(l.Lines) > l.Limit {
		l.Lines = l.Lines[len(l.Lines)-l.Limit:]
	}
	return len(p), nil
}

func (l *LogBuffer) GetLogs() string {
	var sb strings.Builder
	for _, line := range l.Lines {
		sb.WriteString(line)
	}
	return sb.String()
}

var GlobalLogBuffer = &LogBuffer{Limit: 500}

// 网络客户端
func GetProxyClient(cfg models.BotConfig) *http.Client {
	c := &http.Client{Timeout: 30 * time.Second}
	if cfg.ProxyUrl != "" {
		if u, err := url.Parse(cfg.ProxyUrl); err == nil {
			c.Transport = &http.Transport{Proxy: http.ProxyURL(u)}
		}
	}
	return c
}

func GetDirectClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

// 提取工具
func Extract115Link(text string) string {
	re := regexp.MustCompile(`(https?://(?:[a-zA-Z0-9-]+\.)?(?:115\.com|anxia\.com|115cdn\.com|115\.cn)/s/[a-zA-Z0-9]+(?:\?password=[a-zA-Z0-9]+)?)`)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// 解析
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
