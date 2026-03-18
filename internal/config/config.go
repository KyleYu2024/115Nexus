package config

import (
	"encoding/json"
	"io/ioutil"
	"log/slog"
	"os"
	"strings"

	"115nexus/internal/logger"
	"115nexus/internal/models"
	"115nexus/internal/utils"
)

var ConfigPath = "config/bot_config.json"

func Read() (models.BotConfig, error) {
	var cfg models.BotConfig
	data, err := ioutil.ReadFile(ConfigPath)
	if err != nil { return cfg, err }
	err = json.Unmarshal(data, &cfg)
	return cfg, err
}

func Save(cfg models.BotConfig) {
	data, _ := json.MarshalIndent(cfg, "", "  ")
	_ = ioutil.WriteFile(ConfigPath, data, 0644)
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

	// 核心：使用 HandlerOptions 来强制生效 Level
	opts := &slog.HandlerOptions{Level: lvl}

	h1 := logger.NewTwoLineHandler(os.Stdout, true, opts)
	h2 := logger.NewTwoLineHandler(utils.GlobalLogBuffer, false, opts)
	
	multi := &logger.MultiHandler{Handlers: []slog.Handler{h1, h2}, Level: lvl}
	slog.SetDefault(slog.New(multi))
	
	slog.Info("🪵 日志系统初始化完成", "level", lvl.String())
}
