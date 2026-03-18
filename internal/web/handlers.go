package web

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"115nexus/internal/config"
	"115nexus/internal/models"
	"115nexus/internal/services"
	"115nexus/internal/utils"
)

var OnConfigSave func()

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := strings.Replace(htmlPage, "{{VERSION}}", "v0.2.2", 1)
	fmt.Fprint(w, html)
}

func HandleLogo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	fmt.Fprint(w, AppLogoSVG)
}

func HandleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "GET" {
		cfg, _ := config.Read()
		json.NewEncoder(w).Encode(cfg)
	} else if r.Method == "POST" {
		var n models.BotConfig
		body, _ := ioutil.ReadAll(r.Body)
		slog.Info("📥 收到配置保存请求", "len", len(body))
		
		if err := json.Unmarshal(body, &n); err == nil {
			config.Save(n)
			models.GlobalConfig = n
			slog.Info("💾 配置已持久化", "user", n.HdhiveUser)
			if OnConfigSave != nil { OnConfigSave() }
			json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "保存成功"})
		} else {
			slog.Error("❌ 配置解析失败", "err", err, "body", string(body))
			json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "解析失败: " + err.Error()})
		}
	}
}

func HandleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	source := r.URL.Query().Get("source")
	cfg := models.GlobalConfig
	if source == "pansou" {
		items, _ := services.DoPansouSearch(q, cfg)
		json.NewEncoder(w).Encode(map[string]interface{}{"results": items})
		return
	}
	var res []models.MovieMetadata
	if cfg.TmdbApiKey != "" {
		target := fmt.Sprintf("https://api.themoviedb.org/3/search/multi?api_key=%s&language=zh-CN&query=%s", cfg.TmdbApiKey, url.QueryEscape(q))
		slog.Info("🔍 Web API TMDB Request", "url", target)
		resp, err := utils.GetProxyClient(cfg).Get(target)
		if err == nil {
			defer resp.Body.Close()
			var wrapper models.TmdbSearchResponse
			json.NewDecoder(resp.Body).Decode(&wrapper)
			res = wrapper.Results
		} else {
			slog.Error("❌ Web API TMDB Request Error", "err", err)
		}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"results": res})
}

func HandleResources(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	mType := r.URL.Query().Get("type")
	cfg := models.GlobalConfig
	var items []models.ProcessedResource
	var err error
	if cfg.HdhiveApiKey != "" {
		items, err = services.FetchHdhiveResources(id, mType, cfg)
	}
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "items": items})
}

func HandlePush(w http.ResponseWriter, r *http.Request) {
	var req struct { Link string `json:"link"` }
	json.NewDecoder(r.Body).Decode(&req)
	cfg := models.GlobalConfig
	msg := ""
	if strings.HasPrefix(req.Link, "magnet:?") {
		msg = services.PushMagnetToOffline(req.Link, cfg)
	} else {
		msg = services.PushToMedia302(req.Link, cfg)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"success": strings.Contains(msg, "✅"), "message": msg})
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if models.WebUser == "" { next(w, r); return }
		cookie, _ := r.Cookie("session")
		if cookie == nil || cookie.Value != getSessionToken() {
			http.Error(w, "Unauthorized", 401)
			return
		}
		next(w, r)
	}
}

func getSessionToken() string {
	data := []byte(models.WebUser + ":" + models.WebPassword + "@salt")
	return fmt.Sprintf("%x", md5.Sum(data))
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	var c struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&c)
	if c.Username == models.WebUser && c.Password == models.WebPassword {
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    getSessionToken(),
			Path:     "/",
			MaxAge:   31536000,
			HttpOnly: true,
		})
		json.NewEncoder(w).Encode(map[string]bool{"success": true})
	} else {
		http.Error(w, "Login Failed", 401)
	}
}
