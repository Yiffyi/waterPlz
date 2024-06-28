package waterplz

import (
	"sync"

	"github.com/yiffyi/gorad/data"
)

type Config struct {
	db          *data.JSONDatabase
	lock        *sync.RWMutex
	LogFileName string
	Debug       bool

	Watchdog    Watchdog
	WeComBotKey string
}

func LoadConfig() *Config {
	lock := sync.RWMutex{}
	db := data.NewJSONDatabase("config.json", true)
	cfg := Config{
		db:   db,
		lock: &lock,
	}

	db.Load(&cfg, true)
	return &cfg
}

func (c *Config) Save() {
	c.lock.RLock()
	defer c.lock.RUnlock()
	c.db.Save(c)
}
