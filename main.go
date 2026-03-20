package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"115nexus/internal/bot"
	"115nexus/internal/config"
	"115nexus/internal/models"
	"115nexus/internal/services"
	"115nexus/internal/web"

	"github.com/robfig/cron/v3"
)

const AppVersion = "0.4.19"

var globalCron *cron.Cron

func main() {
	rand.Seed(time.Now().UnixNano())
	config.InitLogging()
	slog.Info("115Nexus 启动中", "version", AppVersion)

	// 捕获退出信号
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, _ := config.Read()
	models.UpdateConfig(cfg)
	models.WebUser = os.Getenv("WEB_USER")
	models.WebPassword = os.Getenv("WEB_PASSWORD")

	go models.StartCacheCleaner()

	if cfg.TgToken != "" {
		go bot.StartOrReload(cfg.TgToken)
	}

	// 启动异步推送 Worker (透传 ctx)
	go services.StartPushWorker(ctx)

	// 初始化 Cron
	UpdateCron()

	// 注册配置更新回调
	web.OnConfigSave = func() {
		slog.Info("🔄 检测到配置更新，正在重新应用配置...")
		UpdateCron()
		go bot.StartOrReload(models.GetConfig().TgToken)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", web.HandleIndex)
	mux.HandleFunc("/static/", web.HandleStatic)
	mux.HandleFunc("/api/logo.png", web.HandleLogo)
	mux.HandleFunc("/manifest.json", web.HandleManifest)
	mux.HandleFunc("/api/login", web.HandleLogin)
	mux.HandleFunc("/api/config", web.AuthMiddleware(web.HandleConfig))
	mux.HandleFunc("/api/logs", web.AuthMiddleware(web.HandleLogs))
	mux.HandleFunc("/api/hdhive/me", web.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		d, err := services.GetHdhiveMe(models.GetConfig())
		if err != nil { 
			json.NewEncoder(w).Encode(map[string]any{"success": false, "message": err.Error()}) 
		} else {
			json.NewEncoder(w).Encode(map[string]any{"success": true, "data": d})
		}
	}))
	mux.HandleFunc("/api/hdhive/checkin", web.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		m, err := services.DoHdhiveCheckin(models.GetConfig())
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
	
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// 在协程中启动服务
	go func() {
		slog.Info("🌐 Web Server Ready", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("❌ Web Server Error", "err", err)
		}
	}()

	// 等待退出信号
	<-ctx.Done()
	slog.Info("🛑 正在停止 115Nexus...")

	// 优雅关闭：给予 10s 超时时间
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if globalCron != nil {
		globalCron.Stop()
	}

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("❌ Web Server Shutdown Error", "err", err)
	}

	slog.Info("👋 115Nexus 已安全退出")
}

func UpdateCron() {
	if globalCron != nil {
		globalCron.Stop()
	}
	
	cfg := models.GetConfig()
	if cfg.HdhiveCheckinEnabled && cfg.HdhiveCheckinCron != "" {
		globalCron = cron.New(cron.WithLocation(time.Local))
		_, err := globalCron.AddFunc(cfg.HdhiveCheckinCron, func() {
			delay := time.Duration(rand.Intn(31)) * time.Minute
			slog.Info("⏰ Cron 定时触发签到", "delay", delay.String())
			
			if delay > 0 {
				time.Sleep(delay)
			}
			
			services.DoHdhiveCheckin(models.GetConfig())
		})
		
		if err != nil {
			slog.Error("❌ Cron 表达式语法错误", "cron", cfg.HdhiveCheckinCron, "error", err)
			return
		}
		
		globalCron.Start()
		entry := globalCron.Entries()[0]
		slog.Info("📅 签到任务已就绪", "cron", cfg.HdhiveCheckinCron, "next_run", entry.Next.Format("2006-01-02 15:04:05"))
	} else {
		slog.Warn("⏳ 自动签到任务未启用或 Cron 表达式为空")
	}
}
