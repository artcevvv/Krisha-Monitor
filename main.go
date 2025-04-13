package main

import (
	"context"
	"database"
	"os"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	// godotenv.Load()
	ctx := context.Background()
	token := os.Getenv("TELEGRAM_BOT_TOKEN")

	db, _ = database.InitDB()

	bot, err := telego.NewBot(token, telego.WithDefaultDebugLogger())
	if err != nil {
		panic(err)
	}

	updates, _ := bot.UpdatesViaLongPolling(ctx, nil)

	bh, _ := th.NewBotHandler(bot, updates)

	defer func() { _ = bh.Stop() }()

	bh.Handle(Start, th.CommandEqual("start"))
	bh.Handle(Monitor, th.CommandEqual("monitor"))
	bh.Handle(Cancel, th.CommandEqual("cancel"))
	bh.HandleCallbackQuery(HandleCallback, th.AnyCallbackQueryWithMessage())
	bh.Handle(HandleMessage)

	_ = bh.Start()
}
