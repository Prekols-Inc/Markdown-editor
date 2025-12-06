package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(name string) (*MultiHandler, error) {
	file, err := os.OpenFile(
		fmt.Sprintf("/var/log/markdown-editor-%s.log", name),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot open log file: %w", err)
	}

	stdoutHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	fileHandler := slog.NewTextHandler(file, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	return &MultiHandler{
		handlers: []slog.Handler{stdoutHandler, fileHandler},
	}, nil
}

func (m *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{handlers}
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return &MultiHandler{handlers}
}

var Logger *slog.Logger

func init() {
	multi, err := NewMultiHandler("auth")
	if err != nil {
		panic(err)
	}

	Logger = slog.New(multi)
}
