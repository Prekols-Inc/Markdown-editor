package main

import (
	"log/slog"

	"github.com/Prekols-Inc/Markdown-editor/lib/logger"
)

var Logger *slog.Logger

func init() {
	multi, err := logger.NewMultiHandler("auth")
	if err != nil {
		panic(err)
	}

	Logger = slog.New(multi)
}
