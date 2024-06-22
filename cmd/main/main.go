package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/yiffyi/waterplz"
	"github.com/yiffyi/waterplz/api"
	"github.com/yiffyi/waterplz/watchdog"
)

type Config struct {
	Watchdog watchdog.Watchdog
	WeCom    waterplz.WeComBot
}

func main() {
	if _, err := os.Stat("config.json"); err != nil {
		slog.Error("could not open config.json", "err", err)

		b, err := json.MarshalIndent(Config{}, "", "    ")
		if err != nil {
			slog.Error("could not marshal empty Config", "err", err)
		} else {
			slog.Warn("empty config.json created")
			os.WriteFile("config.json", b, 0644)
		}

		os.Exit(1)
	}

	b, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}

	c := Config{}
	err = json.Unmarshal(b, &c)
	if err != nil {
		panic(err)
	}

	http.Handle("/do", api.CreateV0Mux())

	go c.Watchdog.Start(&c.WeCom)
	http.ListenAndServe(":8080", nil)
}
