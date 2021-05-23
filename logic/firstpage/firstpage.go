package firstpage

import (
	"github.com/Frosin/shoplist-telegram-bot/consts"
	"github.com/Frosin/shoplist-telegram-bot/logic"
	"github.com/Frosin/shoplist-telegram-bot/session"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	FirstpageWord = "firstpage"

	curList   = "Текущий список"
	checkList = "Чек-лист"
	settings  = "Настройки"
	calendar  = "Календарь"

	CurListCmd   = "curlist"
	CheckListCmd = "checklist"
	SettingsCmd  = "settings"
	CalendarCmd  = "calendar"

	firstPageMessage = "firstPage"
)

type firstpage struct{}

func New() *firstpage {
	return &firstpage{}
}

func (f *firstpage) SetSession(session *session.SessionItem) {
	//silence is gold
}

func (f *firstpage) GetCallbackOutput(command string) (logic.Output, error) {
	switch command {
	case consts.Start:
		return getOutput()
	default:
		return logic.Output{}, consts.ErrUnknownCommand
	}
}

func (f *firstpage) GetMessageOutput(curData string, msg string) (logic.Output, error) {
	return getOutput()
}

func getButtons() *tgbotapi.InlineKeyboardMarkup {
	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(curList, consts.CurrentListStart)}, //TODO add correct param
			[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(checkList, consts.ChecklistStart)},
			[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(settings, consts.SettingsStart)},
			[]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(calendar, consts.CalendarStart)},
		},
	}
}

func getOutput() (logic.Output, error) {
	return logic.Output{
		Message:  firstPageMessage,
		Keyboard: getButtons(),
	}, nil
}
