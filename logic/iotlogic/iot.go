package iotlogic

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Frosin/shoplist-telegram-bot/consts"
	"github.com/Frosin/shoplist-telegram-bot/logic"
	"github.com/Frosin/shoplist-telegram-bot/session"
	"github.com/spf13/viper"

	"github.com/Frosin/shoplist-telegram-bot/iot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	iotTxt = `Бюджет: '%s', освоение: %d%%, остаток %d
	Пример добавления категории: "25000 продукты"
	Пример добавления бюджета: "!Июнь"`
	backText   = "⬅ Назад"
	emptyItems = "Нет категорий для отоброжения"
)

var (
	timeout = time.Second * 5
	limit   = 20
)

type iotLogic struct {
	sessionItem *session.SessionItem
	storage     iot.IOTStorage
}

func New(storage iot.IOTStorage) *iotLogic {
	return &iotLogic{
		storage: storage,
	}
}

func (d *iotLogic) SetSession(sessionItem *session.SessionItem) {
	d.sessionItem = sessionItem
}

func (c *iotLogic) GetCallbackOutput(command string) (logic.Output, error) {
	log.Println("** message callback:", command)
	return c.getOutput()
}

func (c *iotLogic) GetMessageOutput(curData string, msg string) (logic.Output, error) {
	return c.getOutput()
}

func (c *iotLogic) getOutput() (logic.Output, error) {
	iotCommunity := viper.GetString("SHOPLIST-BUDGET_COMMUNITY")
	if c.sessionItem.User.ComunityID != iotCommunity {
		return logic.Output{}, nil
	}

	//create keyboard and add back button to keyboard
	controlButtons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(backText, consts.FirstPageStart),
	}
	out := logic.Output{
		Message: emptyItems,
		Keyboard: &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				controlButtons,
			},
		},
	}
	return out, nil
}

func (c *iotLogic) getValuesText() string {
	bld := strings.Builder{}

	dayValues, err := c.storage.GetDayValues(time.Now())
	if err != nil {
		return err.Error()
	}

	if len(dayValues) == 0 {
		return "no new values"
	}

	for param, values := range dayValues {
		limited := values[len(values)-limit:]
		dayValues[param] = limited
	}

	prepared := map[string][]string{}

	for param, limited := range dayValues {
		for _, value := range limited {
			paramString := fmt.Sprintf("%s=%v", param, value.Value)
			prepared[value.Time.Format(iot.TimeLayout)] = append(prepared[value.Time.Format(iot.TimeLayout)], paramString)
		}
	}

	for ts, strs := range prepared {
		bld.WriteString(ts + ":" + strings.Join(strs, ",") + "\n")
	}
	return bld.String()
}