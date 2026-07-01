package budgethistcat

import (
	"context"
	"fmt"
	"log"
	"regexp"
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
	backText   = "⬅ Назад"
	emptyItems = "Нет категорий для отображения"
	headerTxt  = "Бюджет: '%s', освоение: %d%% (%dр, из них %dр картой + %dр наликом), остаток %d"
)

var (
	timeout         = time.Second * 5
	patternBudgetID = regexp.MustCompile(`bi(\d+)`)
)

type budgetHistCat struct {
	sessionItem *session.SessionItem
	storage     bugetstorage.Storage
}

func New(storage bugetstorage.Storage) *budgetHistCat {
	return &budgetHistCat{storage: storage}
}

func (b *budgetHistCat) SetSession(s *session.SessionItem) {
	b.sessionItem = s
}

func (b *budgetHistCat) GetCallbackOutput(command string) (logic.Output, error) {
	log.Println("** budgethistcat callback:", command)
	budgetID, err := parseBudgetID(command)
	if err != nil {
		return logic.Output{}, fmt.Errorf("budgethistcat: %w", err)
	}
	return b.getOutput(budgetID)
}

func (b *budgetHistCat) GetMessageOutput(curData string, msg string) (logic.Output, error) {
	budgetID, err := parseBudgetID(curData)
	if err != nil {
		return logic.Output{}, fmt.Errorf("budgethistcat: %w", err)
	}
	return b.getOutput(budgetID)
}

func parseBudgetID(data string) (int, error) {
	m := patternBudgetID.FindStringSubmatch(data)
	if len(m) != 2 {
		return 0, fmt.Errorf("cannot parse budget id from %q", data)
	}
	return strconv.Atoi(m[1])
}

func (b *budgetHistCat) getOutput(budgetID int) (logic.Output, error) {
	bugetCommunity := viper.GetString("SHOPLIST-BUDGET_COMMUNITY")
	if b.sessionItem.User.ComunityID != bugetCommunity {
		log.Println("ACCESS DENIED: ", b.sessionItem.User)
		return logic.Output{}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	backRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(backText, consts.BudgetHistoryStart),
	}
	emptyOut := logic.Output{
		Message: emptyItems,
		Keyboard: &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{backRow},
		},
	}

	budget, err := b.storage.GetBudget(ctx, budgetID)
	if err != nil {
		return logic.Output{}, fmt.Errorf("budgethistcat: %w", err)
	}

	categories, err := b.storage.GetBudgetCategories(ctx, budgetID)
	if err != nil {
		return logic.Output{}, fmt.Errorf("budgethistcat: %w", err)
	}
	if len(categories) == 0 {
		return emptyOut, nil
	}

	column := [][]tgbotapi.InlineKeyboardButton{}
	var targetSum, curSum, currentCashSum int

	for i, cat := range categories {
		curSum += cat.Current
		targetSum += cat.Target
		currentCashSum += cat.CashCurrent

		var fillPercent int64
		if cat.Target > 0 {
			fillPercent = int64(cat.Current * 100 / cat.Target)
		}
		remainder := cat.Target - cat.Current

		param := helpers.GetParam(consts.BudgetHistoryNotesWord, "ci", strconv.Itoa(cat.ID))
		btnTxt := fmt.Sprintf("%d. %s (%d%%), ост: %dр.", i+1, cat.Title, fillPercent, remainder)
		column = append(column, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(btnTxt, param),
		})
	}
	column = append(column, backRow)

	var totalPercent int64
	if targetSum > 0 {
		totalPercent = int64(curSum * 100 / targetSum)
	}
	remainder := targetSum - curSum
	curSumCard := curSum - currentCashSum

	msg := fmt.Sprintf(headerTxt, budget.Title, totalPercent, curSum, curSumCard, currentCashSum, remainder)
	return logic.Output{
		Message: msg,
		Keyboard: &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: column,
		},
	}, nil
}
