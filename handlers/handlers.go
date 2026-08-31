package handlers

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
	"gorm.io/gorm"

	"database"
	"helpers"
	"middleware"
	"parser"
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

type Handler struct {
	db    *gorm.DB
	lock  sync.RWMutex
	users map[int64]UserState
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		db:    db,
		users: make(map[int64]UserState),
	}
}

func (h *Handler) Start(ctx *th.Context, update telego.Update) error {
	chatID := update.Message.Chat.ID

	_, _ = ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), "Привет! \nЯ - бот для мониторинга сайта 'Krisha.kz'. \n\nДля начала мониторинга введи команду /monitor!"))
	return nil
}

func (h *Handler) Monitor(ctx *th.Context, update telego.Update) error {
	chatID := update.Message.Chat.ID

	h.lock.Lock()
	h.users[chatID] = UserState{State: StateAskCity}
	h.lock.Unlock()

	cityData := `{"city": "astana"}`
	button := tu.InlineKeyboardButton("Астана").WithCallbackData(cityData)
	rows := tu.InlineKeyboardRow(button)
	kb := tu.InlineKeyboard(rows)

	_, err := ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), "Выберите город:").WithReplyMarkup(kb))
	return err
}

func (h *Handler) Cancel(ctx *th.Context, update telego.Update) error {
	chatID := update.Message.Chat.ID

	err := middleware.StopUserMonitor(chatID)
	if err != nil {
		_, _ = ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), err.Error()))
		return err
	}

	h.lock.Lock()
	delete(h.users, chatID)
	h.lock.Unlock()

	_, err = ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), "Вы остановили получение информации! \n\nДля повторного процесса введите /monitor"))
	return err
}

func (h *Handler) sendDistricts(ctx *th.Context, chatID int64, city string) error {
	var rows []telego.InlineKeyboardButton

	for alias, name := range middleware.CityData[city] {
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

func (h *Handler) HandleCallback(ctx *th.Context, callback telego.CallbackQuery) error {
	data := callback.Data

	var payload map[string]string
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return nil
	}

	chatID := callback.Message.GetChat().ID

	if city, ok := payload["city"]; ok {
		h.lock.Lock()
		userState := h.users[chatID]
		userState.City = city
		userState.State = StateAskRegion
		h.users[chatID] = userState
		h.lock.Unlock()

		return h.sendDistricts(ctx, chatID, city)
	}

	if district, ok := payload["district"]; ok {
		h.lock.Lock()
		userState := h.users[chatID]
		userState.Region = district
		userState.State = StateAskPricing
		h.users[chatID] = userState
		h.lock.Unlock()

		_, err := ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), "Введите диапазон цен в формате 100000-180000"))
		return err
	}

	if confirm, ok := payload["confirm"]; ok {
		if confirm == "yes" {
			// Extract and delete user state quickly under lock
			h.lock.Lock()
			userState, exists := h.users[chatID]
			if exists {
				delete(h.users, chatID)
			}
			h.lock.Unlock()

			if !exists {
				_, err := ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), "❌ Сессия истекла. Введите /monitor заново."))
				return err
			}

			// Perform DB operations outside the lock
			_, err := database.GetUser(h.db, chatID)
			if err == nil {
				err = database.UpdateData(h.db, chatID, userState.Region, userState.PriceFrom, userState.PriceTo)
				if err != nil {
					_, _ = ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), "❌ Произошла ошибка при обновлении данных. Попробуйте снова."))
					return err
				}
			} else {
				userData := database.User{
					ChatID:      chatID,
					City:        userState.City,
					Region:      userState.Region,
					PricingFrom: userState.PriceFrom,
					PricingTo:   userState.PriceTo,
				}

				if err := database.SaveData(h.db, userData); err != nil {
					_, _ = ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), "❌ Произошла ошибка при сохранении данных. Попробуйте снова."))
					return err
				}
			}

			url := parser.FormURL(h.db, chatID)

			_, err = ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), "✅ Отлично! Ваши данные сохранены \n\nНачинаю мониторинг... (Сообщения с информацией о новых квартирах будут приходить раз в 1 час)\n\n Для отмены процесса введите /cancel"))
			if err != nil {
				return err
			}

			middleware.StartUserMonitor(ctx, chatID, url, h.db)
			return nil
		} else {
			// User canceled confirmation
			h.lock.Lock()
			userState := h.users[chatID]
			userState.State = StateDefault
			h.users[chatID] = userState
			h.lock.Unlock()

			_, err := ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), "Хорошо, давайте начнем заново! \n\nВведите /monitor для того что бы повторить процедуру ввода данных"))
			return err
		}
	}

	return nil
}

func (h *Handler) HandleMessage(ctx *th.Context, update telego.Update) error {
	chatID := update.Message.Chat.ID
	text := update.Message.Text

	h.lock.RLock()
	userState, ok := h.users[chatID]
	h.lock.RUnlock()

	if !ok {
		return nil
	}

	switch userState.State {
	case StateAskPricing:
		priceFrom, priceTo, err := helpers.DestructStringToNumbers(text, "-")
		if err != nil {
			_, err := ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), err.Error()))
			return err
		}

		h.lock.Lock()
		userState.PriceFrom = priceFrom
		userState.PriceTo = priceTo
		userState.State = StateConfirm
		h.users[chatID] = userState
		h.lock.Unlock()

		msg := fmt.Sprintf("Вы выбрали:\nГород: Астана\nРайон: %s\nЦена: от %d, до %d\nПодтвердить?",
			middleware.CityData[userState.City][userState.Region],
			userState.PriceFrom,
			userState.PriceTo,
		)

		confirmButton := tu.InlineKeyboardButton("✅ Подтвердить").WithCallbackData(`{"confirm": "yes"}`)
		declineButton := tu.InlineKeyboardButton("✖️ Отмена").WithCallbackData(`{"confirm": "no"}`)
		rows := tu.InlineKeyboardRow(confirmButton, declineButton)
		keyboard := tu.InlineKeyboard(rows)

		_, err = ctx.Bot().SendMessage(ctx, tu.Message(tu.ID(chatID), msg).WithReplyMarkup(keyboard))
		return err
	}

	return nil
}
