package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"sync"
)

type TwoLineHandler struct {
	w      io.Writer
	mu     sync.Mutex
	isTerm bool
	opts   *slog.HandlerOptions
}

func NewTwoLineHandler(w io.Writer, isTerm bool, opts *slog.HandlerOptions) *TwoLineHandler {
	return &TwoLineHandler{w: w, isTerm: isTerm, opts: opts}
}

func (h *TwoLineHandler) Enabled(_ context.Context, l slog.Level) bool {
	if h.opts != nil && h.opts.Level != nil {
		return l >= h.opts.Level.Level()
	}
	return true
}

func (h *TwoLineHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	timeStr := r.Time.Format("2006-01-02 15:04:05")

	if h.isTerm {
		var levelTag string
		switch r.Level {
		case slog.LevelDebug: levelTag = "\033[44;37m DEBUG \033[0m"
		case slog.LevelInfo:  levelTag = "\033[42;37m  INFO  \033[0m"
		case slog.LevelWarn:  levelTag = "\033[43;37m  WARN  \033[0m"
		case slog.LevelError: levelTag = "\033[41;37m ERROR \033[0m"
		default:              levelTag = fmt.Sprintf(" %s ", r.Level.String())
		}
		fmt.Fprintf(h.w, "%s \033[90m%s\033[0m\n", levelTag, timeStr)
	} else {
		fmt.Fprintf(h.w, "[%s] %s\n", r.Level.String(), timeStr)
	}

	fs := runtime.CallersFrames([]uintptr{r.PC})
	f, _ := fs.Next()
	source := fmt.Sprintf("%s:%d", filepath.Base(f.File), f.Line)

	attrs := make(map[string]any)
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})

	if h.isTerm {
		fmt.Fprintf(h.w, "\033[90m%s\033[0m - \033[1;37m%s\033[0m", source, r.Message)
		if len(attrs) > 0 {
			b, _ := json.Marshal(attrs)
			fmt.Fprintf(h.w, " - \033[36m%s\033[0m", string(b))
		}
	} else {
		fmt.Fprintf(h.w, "%s - %s", source, r.Message)
		if len(attrs) > 0 {
			b, _ := json.Marshal(attrs)
			fmt.Fprintf(h.w, " - %s", string(b))
		}
	}

	fmt.Fprint(h.w, "\n\n")
	return nil
}

func (h *TwoLineHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return h }
func (h *TwoLineHandler) WithGroup(name string) slog.Handler     { return h }

type MultiHandler struct {
	Handlers []slog.Handler
	Level    slog.Level
}

func (m *MultiHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return l >= m.Level
}
func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.Handlers { h.Handle(ctx, r) }
	return nil
}
func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return m }
func (m *MultiHandler) WithGroup(name string) slog.Handler     { return m }
