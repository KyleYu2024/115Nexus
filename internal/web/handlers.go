package web

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"115nexus/internal/config"
	"115nexus/internal/models"
	"115nexus/internal/services"
	"115nexus/internal/utils"
)

var OnConfigSave func()

// 这里通过全局变量或函数获取版本号
var CurrentVersion = "0.3.0"

func sendJSON(w http.ResponseWriter, status int, resp models.APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	fs := GetFrontendFS()
	tmpl, err := template.ParseFS(fs, "frontend/index.html")
	if err != nil {
		slog.Error("❌ 模板解析失败", "err", err)
		http.Error(w, "Template Error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Logo":    template.HTML(AppLogoSVG),
		"Version": CurrentVersion,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}

func HandleStatic(w http.ResponseWriter, r *http.Request) {
	fs := GetFrontendFS()
	path := r.URL.Path
	if strings.HasPrefix(path, "/static/") {
		path = "frontend/" + path[len("/static/"):]
	}

	f, err := fs.Open(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	if strings.HasSuffix(path, ".css") {
		w.Header().Set("Content-Type", "text/css")
		w.Header().Set("Cache-Control", "public, max-age=31536000")
	} else if strings.HasSuffix(path, ".js") {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "public, max-age=31536000")
	}

	io.Copy(w, f)
}

func HandleLogo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	fmt.Fprint(w, AppLogoSVG)
}

func HandleManifest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{
		"name": "115Nexus",
		"short_name": "115Nexus",
		"start_url": "/",
		"display": "standalone",
		"background_color": "#ffffff",
		"theme_color": "#007aff",
		"icons": [
			{
				"src": "https://img.andp.cc/icons/upload/115Nexus.png",
				"sizes": "512x512",
				"type": "image/png"
			}
		]
	}`)
}

func HandleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		cfg, err := config.Read()
		if err != nil {
			slog.Warn("💡 配置文件读取异常（可能未初始化）", "err", err)
		}
		sendJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: cfg})
	} else if r.Method == "POST" {
		var n models.BotConfig
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			sendJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Message: "读取请求失败"})
			return
		}
		if err := json.Unmarshal(body, &n); err == nil {
			config.Save(n)
			models.UpdateConfig(n)
			if OnConfigSave != nil { OnConfigSave() }
			sendJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "配置保存成功"})
		} else {
			slog.Error("❌ 配置解析失败", "err", err, "body", string(body))
			sendJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Message: "配置格式错误: " + err.Error()})
		}
	}
}

func HandleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Accept") == "text/event-stream" {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		ch := utils.GlobalLogBuffer.Subscribe()
		defer utils.GlobalLogBuffer.Unsubscribe(ch)

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, ": connected\n\n")
		flusher.Flush()

		ctx := r.Context()
		for {
			select {
			case line := <-ch:
				data := strings.ReplaceAll(line, "\n", "\ndata: ")
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			case <-ctx.Done():
				return
			case <-time.After(30 * time.Second):
				fmt.Fprintf(w, ": heartbeat\n\n")
				flusher.Flush()
			}
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, utils.GlobalLogBuffer.GetLogs())
}

func HandleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		sendJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Message: "关键词不能为空"})
		return
	}
	
	source := r.URL.Query().Get("source")
	cfg := models.GetConfig()
	if source == "pansou" {
		items, err := services.DoPansouSearch(q, cfg)
		if err != nil {
			sendJSON(w, http.StatusOK, models.APIResponse{Success: false, Message: "Pansou 搜索失败: " + err.Error()})
			return
		}
		sendJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: map[string]interface{}{"results": items}})
		return
	}
	
	var res []models.MovieMetadata
	if cfg.TmdbApiKey != "" {
		target := fmt.Sprintf("https://api.themoviedb.org/3/search/multi?api_key=%s&language=zh-CN&query=%s", cfg.TmdbApiKey, url.QueryEscape(q))
		resp, err := utils.GetProxyClient(cfg).Get(target)
		if err == nil {
			defer resp.Body.Close()
			var wrapper models.TmdbSearchResponse
			json.NewDecoder(resp.Body).Decode(&wrapper)
			res = wrapper.Results
		}
	}
	sendJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: map[string]interface{}{"results": res}})
}

func HandleResources(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	mType := r.URL.Query().Get("type")
	if id == "" {
		sendJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Message: "ID 缺失"})
		return
	}
	
	cfg := models.GetConfig()
	var items []models.ProcessedResource
	var err error
	if cfg.HdhiveApiKey != "" {
		items, err = services.FetchHdhiveResources(id, mType, cfg)
	}
	if err != nil {
		sendJSON(w, http.StatusOK, models.APIResponse{Success: false, Message: "资源获取失败: " + err.Error()})
		return
	}
	sendJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: map[string]interface{}{"items": items}})
}

func HandlePush(w http.ResponseWriter, r *http.Request) {
	var req struct { Link string `json:"link"` }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Message: "非法请求格式"})
		return
	}
	
	cfg := models.GetConfig()
	msg := ""
	if strings.HasPrefix(req.Link, "magnet:?") {
		msg = services.PushMagnetToOffline(req.Link, cfg)
	} else {
		msg = services.PushToMedia302(req.Link, cfg)
	}
	
	sendJSON(w, http.StatusOK, models.APIResponse{Success: strings.Contains(msg, "✅"), Message: msg})
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if models.WebUser == "" { next(w, r); return }
		cookie, _ := r.Cookie("session")
		if cookie == nil || cookie.Value != getSessionToken() {
			sendJSON(w, http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Unauthorized"})
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
		sendJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "登录成功"})
	} else {
		sendJSON(w, http.StatusUnauthorized, models.APIResponse{Success: false, Message: "用户名或密码错误"})
	}
}
