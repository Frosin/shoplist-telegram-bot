package dayshoppings

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Frosin/shoplist-telegram-bot/consts"
	"github.com/Frosin/shoplist-telegram-bot/helpers"
	"github.com/Frosin/shoplist-telegram-bot/internal/shoplist/ent"
	"github.com/Frosin/shoplist-telegram-bot/logic"
	"github.com/Frosin/shoplist-telegram-bot/session"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	DayshoppingsWord  = "dayshoppings"
	dateLayout        = "02.01.2006"
	toCalendarText    = "⬅ Смотреть календарь"
	dayshoppingsText  = "Покупки в этот день. Для добавления введите название места."
	dayshoppingsEmpty = "Покупок в этот день нет. Для добавления введите название места."
)

type dayshoppings struct {
	sessionItem *session.SessionItem
}

func New() *dayshoppings {
	return &dayshoppings{}
}

func (d *dayshoppings) SetSession(sessionItem *session.SessionItem) {
	d.sessionItem = sessionItem
}

func getCalendarBtn(day time.Time) tgbotapi.InlineKeyboardButton {
	btnParam := helpers.GetParam(consts.CalendarWord, helpers.Time2MonthCode(day))
	return tgbotapi.NewInlineKeyboardButtonData(toCalendarText, btnParam)
}

// keyboard with one button - to calendar
func getToCalendarKeyboard(day time.Time) *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			[]tgbotapi.InlineKeyboardButton{
				getCalendarBtn(day),
			},
		},
	}
}

func getButtonsByData(sList []*ent.Shopping, day time.Time) *tgbotapi.InlineKeyboardMarkup {
	column := [][]tgbotapi.InlineKeyboardButton{}

	for i, sh := range sList {
		//debug
		fmt.Println("\n shopping=", sh.ID)
		fmt.Println("\n shopping=", sh)
		//
		param := helpers.GetParam(
			consts.ShoppingitemsWord,
			strconv.Itoa(sh.ID),
		)
		row := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(i+1)+". "+sh.Edges.Shop.Name, param),
		}
		column = append(column, row)
	}
	// add last button - back to calendar
	column = append(column, []tgbotapi.InlineKeyboardButton{
		getCalendarBtn(day),
	})

	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: column,
	}
}

func (d *dayshoppings) getOutput(day time.Time) (logic.Output, error) {
	sList, err := d.sessionItem.SListAPI.GetShoppingsByDay(day)
	if err != nil {
		return logic.Output{}, err
	}

	getMsg := func(template string) string {
		return strings.Join([]string{
			day.Format(dateLayout),
			template,
		}, " ")
	}

	if len(sList) == 0 {
		return logic.Output{
			Message:  getMsg(dayshoppingsEmpty),
			Keyboard: getToCalendarKeyboard(day),
		}, nil
	}

	return logic.Output{
		Message:  getMsg(dayshoppingsText),
		Keyboard: getButtonsByData(sList, day),
	}, nil
}

func (d *dayshoppings) GetCallbackOutput(command string) (logic.Output, error) {
	day, err := helpers.DayCode2Time(command)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.DayshoppingsWord, consts.ErrUnknownCommand)
	}
	return d.getOutput(day)
}

func (d *dayshoppings) GetMessageOutput(curData string, msg string) (logic.Output, error) {
	day, err := helpers.DayCode2Time(curData)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.DayshoppingsWord, consts.ErrUnknownCommand)
	}
	err = d.sessionItem.SListAPI.AddShopping(day, msg)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.DayshoppingsWord, err)
	}
	return d.getOutput(day)
}
