package buget

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"context"

	"github.com/Frosin/shoplist-telegram-bot/bugetstorage"
	"github.com/Frosin/shoplist-telegram-bot/consts"
	"github.com/Frosin/shoplist-telegram-bot/helpers"
	"github.com/Frosin/shoplist-telegram-bot/logic"
	"github.com/Frosin/shoplist-telegram-bot/session"
	"github.com/spf13/viper"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	bugetTxt = `Бюджет: '%s', освоение: %d%%, остаток %d
	Пример добавления категории: "25000 продукты"
	Пример добавления бюджета: "!Июнь"`
	backText   = "⬅ Назад"
	emptyItems = "Нет категорий для отоброжения"
)

var (
	timeout = time.Second * 5

	patternNewCategory = regexp.MustCompile(`(\d+)\s+(.+)`)
	patternNewBudget   = regexp.MustCompile(`!(.+)`)
)

type buget struct {
	sessionItem *session.SessionItem
	storage     bugetstorage.Storage
}

func New(storage bugetstorage.Storage) *buget {
	return &buget{
		storage: storage,
	}
}

func (d *buget) SetSession(sessionItem *session.SessionItem) {
	d.sessionItem = sessionItem
}

func (c *buget) GetCallbackOutput(command string) (logic.Output, error) {
	log.Println("** message callback:", command)
	return c.getOutput()
}

func (c *buget) GetMessageOutput(curData string, msg string) (logic.Output, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	//parse msg to budget
	m := patternNewBudget.FindStringSubmatch(msg)
	if len(m) == 2 {
		//create new budget
		err := c.storage.InsertBuget(ctx, m[1])
		if err != nil {
			return logic.Output{}, fmt.Errorf("%v: %w", consts.BugetWord, err)
		}
	}

	//parse msg to category
	m = patternNewCategory.FindStringSubmatch(msg)
	if len(m) != 3 {
		return c.getOutput()
	}
	title := m[2]
	targetSum, _ := strconv.Atoi(m[1])

	//debug
	fmt.Printf("m=%#v, title=%#v, sum=%#vn\n", m, title, targetSum)
	//create keyboard and add back button to keyboard
	controlButtons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(backText, consts.FirstPageStart),
	}
	emptyOut := logic.Output{
		Message: emptyItems,
		Keyboard: &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				controlButtons,
			},
		},
	}

	lastBuget, err := c.storage.GetLastBugets(ctx, 1)
	if err != nil && err != sql.ErrNoRows {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.BugetWord, err)
	}
	if err == sql.ErrNoRows {
		// no buget, maybe db is empty
		return emptyOut, nil
	}
	newCategory := bugetstorage.Category{
		BugetID: lastBuget[0].ID,
		Title:   title,
		Current: 0,
		Target:  int64(targetSum),
	}

	err = c.storage.InsertCategory(ctx, newCategory)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.BugetWord, err)
	}
	return c.getOutput()
}

func (c *buget) getOutput() (logic.Output, error) {
	bugetCommunity := viper.GetString("SHOPLIST-BUDGET_COMMUNITY")
	if *c.sessionItem.User.ComunityId != bugetCommunity {
		return logic.Output{}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	//create keyboard and add back button to keyboard
	controlButtons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(backText, consts.FirstPageStart),
	}
	emptyOut := logic.Output{
		Message: emptyItems,
		Keyboard: &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				controlButtons,
			},
		},
	}

	lastBuget, err := c.storage.GetLastBugets(ctx, 1)
	if err != nil && err != sql.ErrNoRows {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.BugetWord, err)
	}
	if err == sql.ErrNoRows {
		// no buget, maybe db is empty
		return emptyOut, nil
	}

	categories, err := c.storage.GetBugetCategories(ctx, lastBuget[0].ID)
	if err != nil && err != sql.ErrNoRows {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.BugetWord, err)
	}
	if err == sql.ErrNoRows {
		// no buget, maybe db is empty
		return emptyOut, nil
	}

	column := [][]tgbotapi.InlineKeyboardButton{controlButtons}

	var targetSum, curSum int64
	// create items list to show
	for i, category := range categories {
		curSum += category.Current
		targetSum += category.Target

		itemIDStr := strconv.Itoa(category.ID)
		itemName := category.Title

		itemData := consts.ListItemSymbol + itemIDStr

		//make item buttom param with shopping id, removed item ids + id curItem to remove
		param := helpers.GetParam(
			consts.BugetCategoryWord,
			itemData,
		)

		var fillPercent int64
		if category.Target > 0 {
			fillPercent = int64(category.Current * 100 / category.Target)
		}
		remainder := category.Target - category.Current

		btnTxt := fmt.Sprintf("%d. %s (%d%%), ост: %dр.", i+1, itemName, fillPercent, remainder)
		//debug
		fmt.Printf("btnTxt=%s\n", btnTxt)

		row := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(btnTxt, param),
		}
		column = append(column, row)
	}

	//final keyboard
	keyboard := &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: column,
	}

	var totalPercent int64
	if targetSum > 0 {
		totalPercent = int64(curSum * 100 / targetSum)
	}
	remainder := targetSum - curSum

	outTxt := fmt.Sprintf(bugetTxt, lastBuget[0].Title, totalPercent, remainder)

	output := logic.Output{
		Message:  outTxt,
		Keyboard: keyboard,
	}

	return output, nil
}
