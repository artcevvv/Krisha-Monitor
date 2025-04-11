package main

import (
	"context"
	"database"
	"os"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func main() {
	godotenv.Load()
	ctx := context.Background()
	token := os.Getenv("TELEGRAM_BOT_TOKEN")

	database.InitDB()

	bot, err := telego.NewBot(token, telego.WithDefaultDebugLogger())
	if err != nil {
		panic(err)
	}

	updates, _ := bot.UpdatesViaLongPolling(ctx, nil)

	bh, _ := th.NewBotHandler(bot, updates)

	defer func() { _ = bh.Stop() }()

	bh.Handle(Start, th.CommandEqual("start"))
	bh.Handle(Monitor, th.CommandEqual("monitor"))
	bh.Handle(HandleCallback)
	bh.Handle(HandleMessage)

	_ = bh.Start()
}
