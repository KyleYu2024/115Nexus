package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"115nexus/internal/bot"
	"115nexus/internal/config"
	"115nexus/internal/models"
	"115nexus/internal/services"
	"115nexus/internal/utils"
	"115nexus/internal/web"

	"github.com/robfig/cron/v3"
)

const AppVersion = "v0.2.2"

var globalCron *cron.Cron

func main() {
	config.InitLogging()
	slog.Info("115Nexus 启动中", "version", AppVersion)

	cfg, _ := config.Read()
	models.GlobalConfig = cfg
	models.WebUser = os.Getenv("WEB_USER")
	models.WebPassword = os.Getenv("WEB_PASSWORD")

	if cfg.TgToken != "" {
		go bot.Start(cfg.TgToken)
	}

	// 初始化 Cron
	UpdateCron()

	// 注册配置更新回调，以便在 Web 界面保存配置后自动刷新 Cron
	web.OnConfigSave = func() {
		slog.Info("🔄 检测到配置更新，正在重新调度 Cron 任务...")
		UpdateCron()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", web.HandleIndex)
	mux.HandleFunc("/api/logo.png", web.HandleLogo)
	mux.HandleFunc("/api/login", web.HandleLogin)
	mux.HandleFunc("/api/config", web.AuthMiddleware(web.HandleConfig))
	mux.HandleFunc("/api/logs", web.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, utils.GlobalLogBuffer.GetLogs())
	}))
	mux.HandleFunc("/api/hdhive/me", web.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		d, err := services.GetHdhiveMe(models.GlobalConfig)
		if err != nil { 
			json.NewEncoder(w).Encode(map[string]any{"success": false, "message": err.Error()}) 
		} else {
			json.NewEncoder(w).Encode(map[string]any{"success": true, "data": d})
		}
	}))
	mux.HandleFunc("/api/hdhive/checkin", web.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		m, err := services.DoHdhiveCheckin(models.GlobalConfig)
		if err != nil { 
			json.NewEncoder(w).Encode(map[string]any{"success": false, "message": err.Error()}) 
		} else {
			json.NewEncoder(w).Encode(map[string]any{"success": true, "message": m})
		}
	}))
	mux.HandleFunc("/api/search", web.AuthMiddleware(web.HandleSearch))
	mux.HandleFunc("/api/resources", web.AuthMiddleware(web.HandleResources))
	mux.HandleFunc("/api/push", web.AuthMiddleware(web.HandlePush))

	port := os.Getenv("PORT")
	if port == "" { port = "7833" }
	slog.Info("🌐 Web Server Ready", "port", port)
	http.ListenAndServe(":"+port, mux)
}

func UpdateCron() {
	if globalCron != nil {
		globalCron.Stop()
	}
	
	cfg := models.GlobalConfig
	if cfg.HdhiveCheckinEnabled && cfg.HdhiveCheckinCron != "" {
		globalCron = cron.New(cron.WithLocation(time.Local))
		_, err := globalCron.AddFunc(cfg.HdhiveCheckinCron, func() {
			slog.Info("⏰ Cron 定时触发签到")
			services.DoHdhiveCheckin(models.GlobalConfig)
		})
		
		if err != nil {
			slog.Error("❌ Cron 表达式语法错误", "cron", cfg.HdhiveCheckinCron, "error", err)
			return
		}
		
		globalCron.Start()
		
		// 计算并显示下次执行时间
		entry := globalCron.Entries()[0]
		slog.Info("📅 签到任务已就绪", "cron", cfg.HdhiveCheckinCron, "next_run", entry.Next.Format("2006-01-02 15:04:05"))
	} else {
		slog.Warn("⏳ 自动签到任务未启用或 Cron 表达式为空")
	}
}
