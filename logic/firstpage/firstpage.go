package firstpage

import (
	"runtime/debug"
	"time"

	"github.com/Frosin/shoplist-telegram-bot/consts"
	"github.com/Frosin/shoplist-telegram-bot/logic"
	"github.com/Frosin/shoplist-telegram-bot/session"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	FirstpageWord = "firstpage"

	curList       = "Текущий список"
	checkList     = "Чек-лист"
	settings      = "Настройки"
	calendar      = "Календарь"
	buget         = "Бюджет"
	funds         = "Фонды"
	benz          = "⛽ АЗС"
	budgetHistory = "История бюджетов"

	CurListCmd   = "curlist"
	CheckListCmd = "checklist"
	SettingsCmd  = "settings"
	CalendarCmd  = "calendar"
	BugetCmd     = "buget"

	commitDateLayout = "02.01.2006 15:04"
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
			{tgbotapi.NewInlineKeyboardButtonData(curList, consts.CurrentListStart)}, //TODO add correct param
			{tgbotapi.NewInlineKeyboardButtonData(checkList, consts.ChecklistStart)},
			{tgbotapi.NewInlineKeyboardButtonData(settings, consts.SettingsStart)},
			{tgbotapi.NewInlineKeyboardButtonData(calendar, consts.CalendarStart)},
			{tgbotapi.NewInlineKeyboardButtonData(buget, consts.BugetStart)},
			{tgbotapi.NewInlineKeyboardButtonData(funds, consts.FundsStart)},
			{tgbotapi.NewInlineKeyboardButtonData(benz, consts.BenzStart)},
			{tgbotapi.NewInlineKeyboardButtonData(budgetHistory, consts.BudgetHistoryStart)},
		},
	}
}

func commitDate() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	for _, s := range info.Settings {
		if s.Key == "vcs.time" {
			t, err := time.Parse(time.RFC3339, s.Value)
			if err != nil {
				return s.Value
			}
			return t.Format(commitDateLayout)
		}
	}
	return "unknown"
}

func getOutput() (logic.Output, error) {
	return logic.Output{
		Message:  commitDate(),
		Keyboard: getButtons(),
	}, nil
}
