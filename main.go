package main

import (
	"context"
	"database"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"gorm.io/gorm"
)

var db *gorm.DB

type fileLogger struct {
	file        *os.File
	debugMode   bool
	printErrors bool
	replacer    *strings.Replacer
}

func (l *fileLogger) Debug(args ...interface{}) {
	if l.debugMode {
		fmt.Fprint(l.file, "[DEBUG] ")
		fmt.Fprintln(l.file, args...)
	}
}

func (l *fileLogger) Debugf(format string, args ...interface{}) {
	if l.debugMode {
		fmt.Fprint(l.file, "[DEBUG] ")
		if l.replacer != nil {
			format = l.replacer.Replace(format)
		}
		fmt.Fprintf(l.file, format+"\n", args...)
	}
}

func (l *fileLogger) Error(args ...interface{}) {
	if l.printErrors {
		fmt.Fprint(l.file, "[ERROR] ")
		fmt.Fprintln(l.file, args...)
	}
}

func (l *fileLogger) Errorf(format string, args ...interface{}) {
	if l.printErrors {
		fmt.Fprint(l.file, "[ERROR] ")
		if l.replacer != nil {
			format = l.replacer.Replace(format)
		}
		fmt.Fprintf(l.file, format+"\n", args...)
	}
}

func main() {
	godotenv.Load()
	ctx := context.Background()
	token := os.Getenv("TELEGRAM_BOT_TOKEN")

	file, err := os.OpenFile("logs/bot_logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	defer file.Close()

	// Create a custom logger option that writes to the file
	customLogger := &fileLogger{
		file:        file,
		debugMode:   true,
		printErrors: true,
		replacer:    strings.NewReplacer("old", "new"),
	}

	customLoggerOption := telego.WithLogger(customLogger)

	// Initialize the bot with the custom logger
	bot, err := telego.NewBot(token, customLoggerOption)

	db, _ = database.InitDB()

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
