package bugetcategory

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"context"

	"github.com/Frosin/shoplist-telegram-bot/bugetstorage"
	"github.com/Frosin/shoplist-telegram-bot/consts"
	"github.com/Frosin/shoplist-telegram-bot/logic"
	"github.com/Frosin/shoplist-telegram-bot/session"
	"github.com/spf13/viper"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	bugetCategoryWord    = "bugetCategory"
	dateLayout           = "02.01.2006 15:04"
	backText             = "⬅ Назад"
	newbugetCategoryText = "*** Создать новый ***"
	emptyItems           = "Нет категорий для отображения"

	catText = "Категория: %s освоение %d%% (%d картой + %d налик / %d):\n"
)

var (
	timeout = time.Second * 5

	patternCallback = regexp.MustCompile(`i(\d+)`)
	patternNewNote  = regexp.MustCompile(`(-?)(\d+)([нкНКncNC])\s+(.+)`)
)

type bugetCategory struct {
	sessionItem *session.SessionItem
	storage     bugetstorage.Storage
}

func New(storage bugetstorage.Storage) *bugetCategory {
	return &bugetCategory{
		storage: storage,
	}
}

func (d *bugetCategory) SetSession(sessionItem *session.SessionItem) {
	d.sessionItem = sessionItem
}

func (c *bugetCategory) GetCallbackOutput(command string) (logic.Output, error) {
	log.Println("** message callback:", command)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	categoryID, err := parseCurData(command)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.BugetCategoryWord, err)
	}
	category, err := c.storage.GetCategory(ctx, categoryID)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.BugetCategoryWord, err)
	}

	return c.getOutput(category)
}

// returns categoryID
func parseCurData(data string) (int, error) {
	//parse msg to category
	m := patternCallback.FindStringSubmatch(data)
	if len(m) != 2 {
		return 0, errors.New("parsing error")
	}
	categoryID, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, err
	}
	return categoryID, nil
}

func (c *bugetCategory) GetMessageOutput(curData string, msg string) (logic.Output, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	categoryID, err := parseCurData(curData)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.BugetCategoryWord, err)
	}
	category, err := c.storage.GetCategory(ctx, categoryID)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.BugetCategoryWord, err)
	}

	m := patternNewNote.FindStringSubmatch(msg)
	if len(m) != 5 {
		return logic.Output{}, fmt.Errorf("Ошибка! Пример верного написания: \"500н хлеб и молоко\" ")
	}
	noteTitle := m[4]
	noteSum, _ := strconv.Atoi(m[2])

	//if minus
	if m[1] != "" {
		noteSum = noteSum * -1
	}

	method := bugetstorage.PaymentMethodCard
	typeChar := strings.ToLower(m[3])
	switch typeChar {
	case "н":
		method = bugetstorage.PaymentMethodCash
	case "n":
		method = bugetstorage.PaymentMethodCash
	}

	newCurrent := category.Current + noteSum
	category.Current = newCurrent

	if category.Target != 0 &&
		(category.Current-category.Target) > 0 {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.BugetCategoryWord, errors.New("В категории не осталось средств!"))
	}

	note := bugetstorage.Note{
		CategoryID:    categoryID,
		Sum:           noteSum,
		Title:         noteTitle,
		PaymentMethod: method,
		Created:       time.Now().Unix(),
	}
	if err := c.storage.InsertNote(ctx, note); err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.BugetCategoryWord, err)
	}

	return c.getOutput(category)
}

func (c *bugetCategory) getOutput(category bugetstorage.Category) (logic.Output, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	bugetCommunity := viper.GetString("SHOPLIST-BUDGET_COMMUNITY")
	if c.sessionItem.User.ComunityID != bugetCommunity {
		log.Println("ACCESS DENIED: ", c.sessionItem.User, c.sessionItem.User.ComunityID)

		return logic.Output{}, nil
	}

	//create keyboard and add back button to keyboard
	controlButtons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(backText, consts.BugetStart),
	}

	column := [][]tgbotapi.InlineKeyboardButton{controlButtons}

	//final keyboard
	keyboard := &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: column,
	}
	notes, err := c.storage.GetCategoryNotes(ctx, category.ID)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.BugetCategoryWord, err)
	}
	var fillPercent int64
	if category.Target > 0 {
		fillPercent = int64(category.Current * 100 / category.Target)
	}

	categoryByCard := (category.Current - category.CashCurrent)

	outTxt := fmt.Sprintf(catText, category.Title, fillPercent, categoryByCard, category.CashCurrent, category.Target)
	for i, v := range notes {
		t := time.Unix(v.Created, 0).Format(dateLayout)
		noteTxt := fmt.Sprintf("%d) %s -> %dр.(%s) - %s\n", i+1, t, v.Sum, v.PaymentMethod.String(), v.Title)
		outTxt += noteTxt
	}

	spendInfo := checkSpend(category, time.Now())
	if spendInfo != "" {
		outTxt = outTxt + "\n" + spendInfo
	}

	output := logic.Output{
		Message:  outTxt,
		Keyboard: keyboard,
	}

	return output, nil
}

func daysIn(m time.Month, year int) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func checkSpend(category bugetstorage.Category, now time.Time) string {
	if category.Target < 10000 {
		return ""
	}
	days := daysIn(now.Month(), now.Year())

	dayBudget := category.Target / days
	curDayTargetSpent := now.Day() * dayBudget

	diff := curDayTargetSpent - category.Current

	over := diff < 0

	daysOver := int(0)
	if over && dayBudget > 0 {
		daysOver = (-1 * diff) / dayBudget
	}

	if daysOver != 0 {
		txt := fmt.Sprintf("🤬 Тормозни! Перерасход на %d дня", daysOver)
		return txt
	}

	return ""
}
