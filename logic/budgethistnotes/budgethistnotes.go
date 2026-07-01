package budgethistnotes

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
	dateLayout = "02.01.2006 15:04"
	catText    = "Категория: %s, освоение %d%% (%d картой + %d наликом / %d):\n"
)

var (
	timeout        = time.Second * 5
	patternCatID   = regexp.MustCompile(`ci(\d+)`)
)

type budgetHistNotes struct {
	sessionItem *session.SessionItem
	storage     bugetstorage.Storage
}

func New(storage bugetstorage.Storage) *budgetHistNotes {
	return &budgetHistNotes{storage: storage}
}

func (b *budgetHistNotes) SetSession(s *session.SessionItem) {
	b.sessionItem = s
}

func (b *budgetHistNotes) GetCallbackOutput(command string) (logic.Output, error) {
	log.Println("** budgethistnotes callback:", command)
	catID, err := parseCatID(command)
	if err != nil {
		return logic.Output{}, fmt.Errorf("budgethistnotes: %w", err)
	}
	return b.getOutput(catID)
}

func (b *budgetHistNotes) GetMessageOutput(curData string, msg string) (logic.Output, error) {
	catID, err := parseCatID(curData)
	if err != nil {
		return logic.Output{}, fmt.Errorf("budgethistnotes: %w", err)
	}
	return b.getOutput(catID)
}

func parseCatID(data string) (int, error) {
	m := patternCatID.FindStringSubmatch(data)
	if len(m) != 2 {
		return 0, fmt.Errorf("cannot parse category id from %q", data)
	}
	return strconv.Atoi(m[1])
}

func (b *budgetHistNotes) getOutput(catID int) (logic.Output, error) {
	bugetCommunity := viper.GetString("SHOPLIST-BUDGET_COMMUNITY")
	if b.sessionItem.User.ComunityID != bugetCommunity {
		log.Println("ACCESS DENIED: ", b.sessionItem.User)
		return logic.Output{}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	category, err := b.storage.GetCategory(ctx, catID)
	if err != nil {
		return logic.Output{}, fmt.Errorf("budgethistnotes: %w", err)
	}

	backParam := helpers.GetParam(consts.BudgetHistoryCatWord, "bi", strconv.Itoa(category.BugetID))
	backRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(backText, backParam),
	}

	notes, err := b.storage.GetCategoryNotes(ctx, catID)
	if err != nil {
		return logic.Output{}, fmt.Errorf("budgethistnotes: %w", err)
	}

	var fillPercent int64
	if category.Target > 0 {
		fillPercent = int64(category.Current * 100 / category.Target)
	}
	categoryByCard := category.Current - category.CashCurrent

	outTxt := fmt.Sprintf(catText, category.Title, fillPercent, categoryByCard, category.CashCurrent, category.Target)
	for i, v := range notes {
		t := time.Unix(v.Created, 0).Format(dateLayout)
		outTxt += fmt.Sprintf("%d) %s -> %dр.(%s) - %s\n", i+1, t, v.Sum, v.PaymentMethod.String(), v.Title)
	}

	return logic.Output{
		Message: outTxt,
		Keyboard: &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{backRow},
		},
	}, nil
}
