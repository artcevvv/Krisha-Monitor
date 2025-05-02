package main

import (
	"database"
	"encoding/json"
	"fmt"
	"parser"
	"sync"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

type State uint

const (
	StateDefault State = iota
	StateAskCity
	StateAskRegion
	StateAskPricing
	StateConfirm
)

type UserState struct {
	State     State
	City      string
	Region    string
	PriceFrom int
	PriceTo   int
}

var users = make(map[int64]UserState)

var lock = sync.RWMutex{}

func Start(ctx *th.Context, update telego.Update) error {
	chatID := update.Message.Chat.ID

	_, _ = ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), "Привет! \nЯ - бот для мониторинга сайта 'Krisha.kz'. \n\nДля начала мониторинга введи команду /monitor!"))
	return nil
}

func Monitor(ctx *th.Context, update telego.Update) error {
	chatID := update.Message.Chat.ID

	lock.Lock()
	users[chatID] = UserState{State: StateAskCity}
	lock.Unlock()

	cityData := `{"city": "astana"}`
	button := tu.InlineKeyboardButton("Астана").WithCallbackData(cityData)
	rows := tu.InlineKeyboardRow(button)
	kb := tu.InlineKeyboard(rows)

	ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), "Выберите город:").WithReplyMarkup(kb))

	return nil
}

func Cancel(ctx *th.Context, update telego.Update) error {
	chatID := update.Message.Chat.ID

	err := StopUserMonitor(chatID)

	if err != nil {
		ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), err.Error()))
		return err
	}

	ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), "Вы остановили получение информации! \n\nДля повторного процесса введите /monitor"))

	return nil
}

func sendDistricts(ctx *th.Context, chatID int64, city string) error {
	rows := []telego.InlineKeyboardButton{}

	for alias, name := range cityData[city] {
		data := map[string]string{
			"district": alias,
		}
		dataJSON, _ := json.Marshal(data)

		button := tu.InlineKeyboardButton(name).WithCallbackData(string(dataJSON))
		rows = append(rows, button)
	}

	var kb [][]telego.InlineKeyboardButton

	for i := 0; i < len(rows); i += 2 {
		end := i + 2
		if end > len(rows) {
			end = len(rows)
		}

		kb = append(kb, rows[i:end])
	}
	_, err := ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), "Выберите район:").WithReplyMarkup(tu.InlineKeyboard(kb...)))
	return err
}

func HandleCallback(ctx *th.Context, callback telego.CallbackQuery) error {
	data := callback.Data

	var payload map[string]string
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return nil
	}

	chatID := callback.Message.GetChat().ID
	lock.Lock()
	userState := users[chatID]
	lock.Unlock()

	if city, ok := payload["city"]; ok {
		lock.Lock()
		userState.City = city
		userState.State = StateAskRegion
		users[chatID] = userState
		lock.Unlock()

		return sendDistricts(ctx, chatID, city)
	}

	if district, ok := payload["district"]; ok {
		lock.Lock()
		userState.Region = district
		userState.State = StateAskPricing
		users[chatID] = userState
		lock.Unlock()

		_, err := ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), "Введите диапазон цен в формате 100000-180000"))
		return err
	}

	if confirm, ok := payload["confirm"]; ok {
		if confirm == "yes" {
			lock.Lock()

			_, err := database.GetUser(db, chatID)
			if err == nil {
				err = database.UpdateData(db, chatID, userState.Region, userState.PriceFrom, userState.PriceTo)
				if err != nil {
					lock.Unlock()
					ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), "❌ Произошла ошибка при обновлении данных. Попробуйте снова."))
					return err
				}
			} else {
				data := database.User{
					ChatID:      chatID,
					City:        userState.City,
					Region:      userState.Region,
					PricingFrom: userState.PriceFrom,
					PricingTo:   userState.PriceTo,
				}

				if err := database.SaveData(db, data); err != nil {
					lock.Unlock()
					ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), "❌ Произошла ошибка при сохранении данных. Попробуйте снова."))
					return err
				}
			}

			url := parser.FormURL(db, chatID)
			delete(users, chatID)
			lock.Unlock()

			_, err = ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), "✅ Отлично! Ваши данные сохранены \n\nНачинаю мониторинг... (Сообщения с информацией о новых квартирах будут приходить раз в 1 час)\n\n Для отмены процесса введите /cancel"))

			StartUserMonitor(ctx, chatID, url, db)

			if err != nil {
				return err
			}

			return nil
		} else {
			lock.Lock()
			userState.State = StateDefault
			users[chatID] = userState
			lock.Unlock()

			_, err := ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), "Хорошо, давайте начнем заново! \n\nВведите /monitor для того что бы повторить процедуру ввода данных"))

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func HandleMessage(ctx *th.Context, update telego.Update) error {
	chatID := update.Message.Chat.ID
	text := update.Message.Text

	lock.Lock()
	userState, ok := users[chatID]
	lock.Unlock()

	if !ok {
		return nil
	}

	switch userState.State {
	case StateAskPricing:
		lock.Lock()

		priceFrom, priceTo, err := DestructStringToNumbers(text, "-")
		if err != nil {
			_, err := ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), err.Error()))
			return err
		}

		userState.PriceFrom = priceFrom
		userState.PriceTo = priceTo
		userState.State = StateConfirm

		users[chatID] = userState

		msg := fmt.Sprintf("Вы выбрали:\nГород: Астана\nРайон: %s\nЦена: от %d, до %d\nПодтвердить?",
			cityData[userState.City][userState.Region],
			userState.PriceFrom,
			userState.PriceTo,
		)
		lock.Unlock()

		confirmButton := tu.InlineKeyboardButton("✅ Подтвердить").WithCallbackData(`{"confirm": "yes"}`)
		declineButton := tu.InlineKeyboardButton("✖️ Отмена").WithCallbackData(`{"confirm": "no"}`)
		rows := tu.InlineKeyboardRow(confirmButton, declineButton)
		keyboard := tu.InlineKeyboard(rows)

		_, err = ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), msg).WithReplyMarkup(keyboard))
		return err
	}

	return nil
}
