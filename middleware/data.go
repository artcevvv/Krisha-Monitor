package middleware

import (
	"context"
	"sync"
	"time"
)

var CityData = map[string]map[string]string{
	"astana": {
		"astana-almatinskij":   "Алматы р-н",
		"astana-esilskij":      "Есильский р-н",
		"astana-nura":          "Нура р-н",
		"r-n-bajkonur":         "р-н Байконур",
		"astana-saraishyk":     "Сарайшык р-н",
		"astana-saryarkinskij": "Сарыарка р-н",
	},
}

type UserMonitor struct {
	ChatID      int64
	URL         string
	KnownFlats  map[string]bool // для отслеживания уже отправленных
	CommandChan chan string     // канал для команд (start, stop и др.)
	Ticker      *time.Ticker
	CancelFunc  context.CancelFunc
}

var monitors = make(map[int64]*UserMonitor)
var monitorLock sync.Mutex
