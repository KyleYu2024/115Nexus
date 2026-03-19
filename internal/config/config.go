package config

import (
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"sync"

	"115nexus/internal/logger"
	"115nexus/internal/models"
	"115nexus/internal/utils"
)

var (
	ConfigPath = "config/bot_config.json"
	configMu   sync.RWMutex
)

func Read() (models.BotConfig, error) {
	configMu.RLock()
	defer configMu.RUnlock()
	
	var cfg models.BotConfig
	if _, err := os.Stat(ConfigPath); os.IsNotExist(err) {
		slog.Warn("⚠️ 配置文件不存在，返回默认值", "path", ConfigPath)
		return cfg, nil
	}

	data, err := os.ReadFile(ConfigPath)
	if err != nil {
		return cfg, err
	}
	err = json.Unmarshal(data, &cfg)
	return cfg, err
}

func Save(cfg models.BotConfig) {
	configMu.Lock()
	defer configMu.Unlock()

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		slog.Error("❌ 序列化配置失败", "err", err)
		return
	}
	
	err = os.WriteFile(ConfigPath, data, 0644)
	if err != nil {
		slog.Error("❌ 写入配置文件失败", "err", err)
	}
}

func InitLogging() {
	lvlStr := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	var lvl slog.Level
	switch lvlStr {
	case "DEBUG": lvl = slog.LevelDebug
	case "WARN":  lvl = slog.LevelWarn
	case "ERROR": lvl = slog.LevelError
	default:      lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: lvl}

	h1 := logger.NewTwoLineHandler(os.Stdout, true, opts)
	h2 := logger.NewTwoLineHandler(utils.GlobalLogBuffer, false, opts)
	
	multi := &logger.MultiHandler{Handlers: []slog.Handler{h1, h2}, Level: lvl}
	slog.SetDefault(slog.New(multi))
	
	slog.Info("🪵 日志系统初始化完成", "level", lvl.String())
}
