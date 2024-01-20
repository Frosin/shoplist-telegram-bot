package fund

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

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	bugetCategoryWord    = "fund"
	dateLayout           = "02.01.06"
	backText             = "⬅ Назад"
	newbugetCategoryText = "*** Создать новый ***"
	emptyItems           = "Нет фондов для отображения"

	fundText = "Фонд: %s состояние (%dр):\n"

	maxHistoryNotes = 10
)

var (
	timeout = time.Second * 5

	patternCallback = regexp.MustCompile(`i(\d+)`)
	patternNewNote  = regexp.MustCompile(`(-?)(\d+)\s+(.+)`)
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

	fundID, err := parseCurData(command)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.FundWord, err)
	}
	fund, err := c.storage.GetFund(ctx, fundID)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.FundWord, err)
	}

	return c.getOutput(fund)
}

// returns fundID
func parseCurData(data string) (int, error) {
	//parse msg to category
	m := patternCallback.FindStringSubmatch(data)
	if len(m) != 2 {
		return 0, errors.New("parsing error")
	}

	fundID, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, err
	}
	return fundID, nil
}

func (c *bugetCategory) GetMessageOutput(curData string, msg string) (logic.Output, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	fundID, err := parseCurData(curData)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.FundWord, err)
	}

	fund, err := c.storage.GetFund(ctx, fundID)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.FundWord, err)
	}

	m := patternNewNote.FindStringSubmatch(msg)
	if len(m) != 4 {
		return c.getOutput(fund)
	}
	noteTitle := m[3]
	noteSum, _ := strconv.Atoi(m[2])

	//if minus
	if m[1] != "" {
		noteSum = noteSum * -1
	}

	newCurrent := fund.Current + int64(noteSum)
	fund.Current = newCurrent
	//update category
	if err := c.storage.UpdateFund(ctx, fundID, int(newCurrent)); err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.FundWord, err)
	}
	//create new note
	note := bugetstorage.Note{
		CategoryID: fundID,
		Sum:        noteSum,
		Title:      noteTitle,
		Created:    time.Now().Unix(),
	}
	if err := c.storage.InsertNote(ctx, note); err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.FundWord, err)
	}

	return c.getOutput(fund)
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
		tgbotapi.NewInlineKeyboardButtonData(backText, consts.FundsStart),
	}

	column := [][]tgbotapi.InlineKeyboardButton{controlButtons}

	//final keyboard
	keyboard := &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: column,
	}
	notes, err := c.storage.GetCategoryNotes(ctx, category.ID)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.FundWord, err)
	}

	outTxt := []string{fmt.Sprintf(fundText, category.Title, category.Current)}
	for i, v := range notes {
		t := time.Unix(v.Created, 0).Format(dateLayout)
		plus := ""
		if v.Sum > 0 {
			plus = "+"
		}
		noteTxt := fmt.Sprintf("%d) %s -> %s%dр. - %s\n", i+1, t, plus, v.Sum, v.Title)
		outTxt = append(outTxt, noteTxt)
	}

	outLen := len(outTxt)
	if outLen > maxHistoryNotes {
		cut := []string{
			outTxt[0], "...\n",
		}
		cut = append(cut, outTxt[outLen-maxHistoryNotes:]...)

		outTxt = cut
	}

	output := logic.Output{
		Message:  strings.Join(outTxt, ""),
		Keyboard: keyboard,
	}

	return output, nil
}
