package middleware

import (
	"context"
	"database"
	"fmt"
	"parser"
	"strings"
	"time"

	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"gorm.io/gorm"
)

func StartUserMonitor(ctx *th.Context, chatID int64, url string, db *gorm.DB) {
	monitorLock.Lock()
	defer monitorLock.Unlock()
	if _, exists := monitors[chatID]; exists {
		return
	}

	goCtx := context.Background()

	monitor := &UserMonitor{
		ChatID:      chatID,
		URL:         url,
		KnownFlats:  make(map[string]bool),
		CommandChan: make(chan string),
		Ticker:      time.NewTicker(1 * time.Hour),
	}

	monitors[chatID] = monitor

	go UserMonitorWorker(goCtx, ctx, monitor, db)
}

func UserMonitorWorker(goCtx context.Context, ctx *th.Context, monitor *UserMonitor, db *gorm.DB) {
	defer monitor.Ticker.Stop()

	err := parser.ParseKrisha(monitor.URL, db, monitor.ChatID)
	if err != nil {
		fmt.Println("Ошибка парсинга:", err)
	} else {
		flats, err := database.GetFlatsByUser(db, monitor.ChatID)
		if err != nil {
			fmt.Println("Ошибка получения квартир из БД:", err)
		} else {
			newFlats := filterNewFlats(monitor, flats)
			if len(newFlats) > 0 {
				sendFlatsSummary(ctx, monitor.ChatID, newFlats)
				saveKnownFlats(monitor, newFlats)
			}
		}
	}

	for {
		select {
		case <-goCtx.Done():
			return
		case cmd := <-monitor.CommandChan:
			if cmd == "stop" {
				return
			}

		case <-monitor.Ticker.C:
			err := parser.ParseKrisha(monitor.URL, db, monitor.ChatID)
			if err != nil {
				fmt.Println("Ошибка парсинга:", err)
				continue
			}

			flats, err := database.GetFlatsByUser(db, monitor.ChatID)
			if err != nil {
				fmt.Println("Ошибка получения квартир из БД:", err)
				continue
			}

			newFlats := filterNewFlats(monitor, flats)
			if len(newFlats) > 0 {
				sendFlatsSummary(ctx, monitor.ChatID, newFlats)
				saveKnownFlats(monitor, newFlats)
			}
		}
	}
}

func filterNewFlats(monitor *UserMonitor, flats []database.Flat) []database.Flat {
	var newFlats []database.Flat
	for _, flat := range flats {
		if !monitor.KnownFlats[flat.Link] {
			newFlats = append(newFlats, flat)
		}
	}
	return newFlats
}

func saveKnownFlats(monitor *UserMonitor, flats []database.Flat) {
	for _, flat := range flats {
		monitor.KnownFlats[flat.Link] = true
	}
}

func sendFlatsSummary(ctx *th.Context, chatID int64, flats []database.Flat) {
	const maxFlatsPerMessage = 20

	var messages []string

	for _, flat := range flats {
		message := fmt.Sprintf("🏠 %s | %s \n💸 <b>%s</b>\n🔗 %s", flat.Title, flat.Location, flat.Price, flat.Link)
		messages = append(messages, message)
	}

	for i := 0; i < len(messages); i += maxFlatsPerMessage {
		end := i + maxFlatsPerMessage
		if end > len(messages) {
			end = len(messages)
		}

		sendCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		fullMessage := strings.Join(messages[i:end], "\n\n")

		_, err := ctx.Bot().SendMessage(sendCtx, tu.Message(tu.ID(chatID), fullMessage).WithParseMode("HTML"))
		if err != nil {
			fmt.Println("Ошибка отправки сообщения:", err)
		}
	}
}

func StopUserMonitor(chatID int64) error {
	monitorLock.Lock()
	defer monitorLock.Unlock()

	if monitor, exists := monitors[chatID]; exists {
		monitor.CommandChan <- "stop"
		close(monitor.CommandChan)
		delete(monitors, chatID)
		return nil
	}
	return fmt.Errorf("вы не находитесь в процессе мониторинга")
}
