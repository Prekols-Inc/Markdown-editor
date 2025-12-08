package main

import (
	"log/slog"
	"os"

	"github.com/Prekols-Inc/Markdown-editor/lib/logger"
)

var Logger *slog.Logger

func init() {
	logdir := os.Getenv("LOG_DIR")
	if logdir == "" {
		panic("LOG_DIR env var not found")
	}

	multi, err := logger.NewMultiHandler(logdir, "auth.log")
	if err != nil {
		panic(err)
	}

	Logger = slog.New(multi)
}
