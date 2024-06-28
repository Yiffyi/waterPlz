package main

import (
	"log/slog"
	"net/http"

	"github.com/yiffyi/gorad"
	"github.com/yiffyi/gorad/notification"
	"github.com/yiffyi/waterplz"
	"github.com/yiffyi/waterplz/api"
)

func main() {
	cfg := waterplz.LoadConfig()
	level := slog.LevelInfo
	if cfg.Debug {
		level = slog.LevelDebug
	}
	logger := slog.New(gorad.NewTextFileSlogHandler(cfg.LogFileName, level))
	slog.SetDefault(logger)

	bot := notification.WeComBot{
		Key: cfg.WeComBotKey,
	}

	http.Handle("/do", api.CreateV0Mux())
	http.Handle("/v1/", api.CreateV1Mux())

	go cfg.Watchdog.Start(&bot)
	http.ListenAndServe(":8080", nil)
}
