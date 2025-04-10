package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

type User struct {
	State       State
	City        string
	Region      string
	PricingFrom int
	PricingTo   int
}

type State uint

const (
	StateDefault State = iota
	StateAskCity
	StateAskRegion
	StateAskPricing
	StateConfirm
)

var ctx = context.Background()

var users = make(map[int64]*User)

var lock = sync.RWMutex{}

func Start(ctx *th.Context, update telego.Update) error {
	chatID := update.Message.Chat.ID

	_, _ = ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), fmt.Sprintf("Привет! \nЯ - бот для мониторинга сайта 'Krisha.kz'. \n\nДля начала мониторинга введи команду /monitoring!")))
	return nil
}

// Fuck it, use buttons

// func Monitor(ctx *th.Context, update telego.Update) error {
// 	chatID := update.Message.Chat.ID

// 	lock.RLock()
// 	user, exists := users[chatID]
// 	lock.RUnlock()

// 	if !exists {
// 		user = &User{}
// 		lock.Lock()
// 		users[chatID] = user
// 		lock.Unlock()
// 	}

// 	_, _ = ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), "Введите город для мониторинга\n\n Доступны:\n Астана"))

// 	user.State = StateAskCity

// 	switch user.State {
// 	case StateAskCity:
// 		lock.RLock()
// 		user, exists := users[chatID]
// 		lock.RUnlock()

// 		if !exists {
// 			user = &User{}
// 		}tyHandler(ctx, update, *user)
// 	default:
// 		Monitor(ctx, update)
// 	}

// 	return nil
// }

// func CityHandler(ctx *th.Context, update telego.Update, user User) {
// 	chatID := update.Message.Chat.ID

// 	city := update.Message.Text

// 	if strings.ToLower(city) != "астана" {
// 		ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), "Неверный город! Выберите из списка доступных."))
// 		return
// 	}

// 	user.City = city

// 	user.State = StateAskRegion // next state (example)

// 	ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), "Введите район города (например: Есильский район)"))
// }

func Monitor(ctx *th.Context, update telego.Update) error {
	return nil
}
