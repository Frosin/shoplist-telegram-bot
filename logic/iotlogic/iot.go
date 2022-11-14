package iotlogic

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Frosin/shoplist-telegram-bot/consts"
	"github.com/Frosin/shoplist-telegram-bot/logic"
	"github.com/Frosin/shoplist-telegram-bot/session"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"

	"github.com/Frosin/shoplist-telegram-bot/iot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	backText   = "⬅ Назад"
	timeLayout = "15:04"
)

var (
	timeout = time.Second * 5
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

func newErrorOut(msg string, controlButtons []tgbotapi.InlineKeyboardButton) logic.Output {
	return logic.Output{
		Message: msg,
		Keyboard: &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				controlButtons,
			},
		},
	}
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

	dayValues, err := c.storage.GetDayValues(time.Now())
	if err != nil {
		return newErrorOut(err.Error(), controlButtons), nil
	}

	if len(dayValues) == 0 {
		return newErrorOut("no new values", controlButtons), nil
	}

	msg := getMessage(dayValues)

	name, err := c.generateGraph("t", dayValues["t"])
	if err != nil {
		return newErrorOut(err.Error(), controlButtons), nil
	}

	f, err := os.Open(name)
	if err != nil {
		return newErrorOut(err.Error(), controlButtons), nil
	}
	defer func() {
		f.Close()
		os.Remove(name)
	}()

	content, err := ioutil.ReadAll(f)

	if err != nil {
		return newErrorOut(err.Error(), controlButtons), nil
	}

	bytes := tgbotapi.FileBytes{Name: name, Bytes: content}
	img := tgbotapi.NewPhotoUpload(c.sessionItem.ChatID, bytes)

	out := logic.Output{
		Message: msg,
		Keyboard: &tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
				controlButtons,
			},
		},
		Image: &img,
	}

	return out, nil
}

func (c *iotLogic) generateGraph(param string, dayValues []iot.StorageValue) (string, error) {
	p := plot.New()

	p.Title.Text = "t"
	p.X.Label.Text = "time"
	p.Y.Label.Text = "value"

	xyValues := plotter.XYs{}
	for _, value := range dayValues {
		xyValues = append(xyValues, plotter.XY{
			X: timeToFloat(value.Time),
			Y: value.Value,
		})
	}

	err := plotutil.AddLinePoints(p,
		param, xyValues,
	)
	if err != nil {
		return "", err
	}

	name := uuid.New().String() + ".png"

	// Save the plot to a PNG file.
	if err := p.Save(4*vg.Inch, 3*vg.Inch, name); err != nil {
		return "", err
	}

	return name, nil
}

func getMessage(dayValues map[string][]iot.StorageValue) string {
	bldr := strings.Builder{}
	for param, pValues := range dayValues {
		min, max, cur := getNums(pValues)

		paramData := fmt.Sprintf(
			"%s: min=%.1f(%s), max=%.1f(%s), cur=%.1f(%s)\n",
			param, min.Value, min.Time.Format(timeLayout),
			max.Value, max.Time.Format(timeLayout),
			cur.Value, cur.Time.Format(timeLayout),
		)
		bldr.WriteString(paramData)
	}
	return bldr.String()
}

func getNums(dayValues []iot.StorageValue) (min, max, cur iot.StorageValue) {
	if len(dayValues) == 0 {
		return
	}

	min = dayValues[0]
	max = dayValues[0]
	cur = dayValues[0]

	for _, v := range dayValues {
		if v.Value > max.Value {
			max = v
		}
		if v.Value < min.Value {
			min = v
		}
		cur = v
	}
	return
}

func timeToFloat(t time.Time) float64 {
	beforeDot := float64(t.Hour())
	afterDot := float64((t.Minute()*100)/60) / 100
	return beforeDot + afterDot
}
