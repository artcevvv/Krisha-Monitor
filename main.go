package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"

	"database"
	"handlers"
	"logger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	ctx := context.Background()
	token := os.Getenv("TELEGRAM_BOT_TOKEN")

	file, err := os.OpenFile("logs/bot_logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	defer file.Close()

	customLogger := &logger.FileLogger{
		File:        file,
		DebugMode:   true,
		PrintErrors: true,
		Replacer:    strings.NewReplacer("old", "new"),
	}

	customLoggerOption := telego.WithLogger(customLogger)

	bot, err := telego.NewBot(token, customLoggerOption)
	if err != nil {
		log.Fatalf("Failed to initialize bot: %v", err)
	}

	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	h := handlers.NewHandler(db)

	updates, err := bot.UpdatesViaLongPolling(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to start long polling: %v", err)
	}

	bh, err := th.NewBotHandler(bot, updates)
	if err != nil {
		log.Fatalf("Failed to initialize bot handler: %v", err)
	}

	defer func() { _ = bh.Stop() }()

	bh.Handle(h.Start, th.CommandEqual("start"))
	bh.Handle(h.Monitor, th.CommandEqual("monitor"))
	bh.Handle(h.Cancel, th.CommandEqual("cancel"))
	bh.HandleCallbackQuery(h.HandleCallback, th.AnyCallbackQueryWithMessage())
	bh.Handle(h.HandleMessage)

	_ = bh.Start()
}
