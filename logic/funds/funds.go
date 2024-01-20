package funds

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
	fundTxt = `Виртуальные фонды, всего: %d,
	Пример добавления фонда: "25000 фонд подарков"`
	backText   = "⬅ Назад"
	emptyItems = "Нет фондов для отображения"
)

var (
	timeout = time.Second * 5

	patternNewFund = regexp.MustCompile(`(\d+)\s+(.+)`)
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

	//parse msg to fund
	m := patternNewFund.FindStringSubmatch(msg)
	if len(m) != 3 {
		return c.getOutput()
	}
	title := m[2]
	fundSum, _ := strconv.Atoi(m[1])

	//debug
	fmt.Printf("m=%#v, title=%#v, sum=%#vn\n", m, title, fundSum)
	//create keyboard and add back button to keyboard

	newFund := bugetstorage.Category{
		BugetID: 0,
		Title:   title,
		Current: int64(fundSum),
	}

	err := c.storage.InsertFund(ctx, newFund)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.FundsWord, err)
	}
	return c.getOutput()
}

func (c *buget) getOutput() (logic.Output, error) {
	bugetCommunity := viper.GetString("SHOPLIST-BUDGET_COMMUNITY")
	if c.sessionItem.User.ComunityID != bugetCommunity {
		log.Println("ACCESS DENIED: ", c.sessionItem.User, c.sessionItem.User.ComunityID)

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

	funds, err := c.storage.GetFunds(ctx)
	if err != nil && err != sql.ErrNoRows {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.FundsWord, err)
	}
	if err == sql.ErrNoRows {
		// no buget, maybe db is empty
		return emptyOut, nil
	}

	column := [][]tgbotapi.InlineKeyboardButton{}

	var curSum int64
	// create items list to show
	for i, fund := range funds {
		curSum += fund.Current

		itemIDStr := strconv.Itoa(fund.ID)
		itemName := fund.Title

		itemData := consts.ListItemSymbol + itemIDStr

		//make item buttom param with shopping id, removed item ids + id curItem to remove
		param := helpers.GetParam(
			consts.FundWord,
			itemData,
		)

		btnTxt := fmt.Sprintf("%d. %s, ост: %dр.", i+1, itemName, fund.Current)
		//debug
		fmt.Printf("btnTxt=%s\n", btnTxt)

		row := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(btnTxt, param),
		}
		column = append(column, row)
	}
	column = append(column, controlButtons)

	//final keyboard
	keyboard := &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: column,
	}

	outTxt := fmt.Sprintf(fundTxt, curSum)

	output := logic.Output{
		Message:  outTxt,
		Keyboard: keyboard,
	}

	return output, nil
}
