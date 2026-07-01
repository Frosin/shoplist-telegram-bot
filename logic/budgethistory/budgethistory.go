package budgethistory

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Frosin/shoplist-telegram-bot/bugetstorage"
	"github.com/Frosin/shoplist-telegram-bot/consts"
	"github.com/Frosin/shoplist-telegram-bot/helpers"
	"github.com/Frosin/shoplist-telegram-bot/logic"
	"github.com/Frosin/shoplist-telegram-bot/session"
	"github.com/spf13/viper"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	pageSize  = 5
	backText  = "🏠 На главную"
	prevText  = "◀"
	nextText  = "▶"
	emptyMsg  = "Нет доступных бюджетов в истории"
	headerMsg = "История бюджетов (страница %d из %d):"
)

var timeout = time.Second * 5

type budgetHistory struct {
	sessionItem *session.SessionItem
	storage     bugetstorage.Storage
}

func New(storage bugetstorage.Storage) *budgetHistory {
	return &budgetHistory{storage: storage}
}

func (b *budgetHistory) SetSession(s *session.SessionItem) {
	b.sessionItem = s
}

func (b *budgetHistory) GetCallbackOutput(command string) (logic.Output, error) {
	log.Println("** budgethistory callback:", command)
	page := 0
	if command != consts.Start && len(command) > 1 && command[0] == 'p' {
		if n, err := strconv.Atoi(command[1:]); err == nil {
			page = n
		}
	}
	return b.getOutput(page)
}

func (b *budgetHistory) GetMessageOutput(curData string, msg string) (logic.Output, error) {
	return b.getOutput(0)
}

func (b *budgetHistory) getOutput(page int) (logic.Output, error) {
	bugetCommunity := viper.GetString("SHOPLIST-BUDGET_COMMUNITY")
	if b.sessionItem.User.ComunityID != bugetCommunity {
		log.Println("ACCESS DENIED: ", b.sessionItem.User)
		return logic.Output{}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	allBudgets, err := b.storage.GetAllBudgets(ctx)
	if err != nil {
		return logic.Output{}, fmt.Errorf("budgethistory: %w", err)
	}

	backRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(backText, consts.FirstPageStart),
	}

	if len(allBudgets) == 0 {
		return logic.Output{
			Message: emptyMsg,
			Keyboard: &tgbotapi.InlineKeyboardMarkup{
				InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{backRow},
			},
		}, nil
	}

	totalPages := (len(allBudgets) + pageSize - 1) / pageSize
	if page >= totalPages {
		page = totalPages - 1
	}
	if page < 0 {
		page = 0
	}

	start := page * pageSize
	end := start + pageSize
	if end > len(allBudgets) {
		end = len(allBudgets)
	}

	column := [][]tgbotapi.InlineKeyboardButton{}
	for _, budget := range allBudgets[start:end] {
		param := helpers.GetParam(consts.BudgetHistoryCatWord, "bi", strconv.Itoa(budget.ID))
		column = append(column, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(budget.Title, param),
		})
	}

	navRow := []tgbotapi.InlineKeyboardButton{}
	if page > 0 {
		prevParam := helpers.GetParam(consts.BudgetHistoryWord, "p", strconv.Itoa(page-1))
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData(prevText, prevParam))
	}
	if page < totalPages-1 {
		nextParam := helpers.GetParam(consts.BudgetHistoryWord, "p", strconv.Itoa(page+1))
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData(nextText, nextParam))
	}
	if len(navRow) > 0 {
		column = append(column, navRow)
	}
	column = append(column, backRow)

	return logic.Output{
		Message: fmt.Sprintf(headerMsg, page+1, totalPages),
		Keyboard: &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: column,
		},
	}, nil
}
