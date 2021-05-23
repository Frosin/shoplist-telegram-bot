package calendar

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/Frosin/shoplist-telegram-bot/consts"
	"github.com/Frosin/shoplist-telegram-bot/helpers"
	"github.com/Frosin/shoplist-telegram-bot/logic"
	"github.com/Frosin/shoplist-telegram-bot/session"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	CalendarWord = "calendar"
	backWord     = "back"
	timeLimit    = 8760 // one year

	emptyLabel      = " "
	leftLabel       = "<"
	rightLabel      = ">"
	calendarMessage = "calendar"
)

var (
	weekDays = []string{"ПН", "ВТ", "СР", "ЧТ", "ПТ", "СБ", "ВС"}
)

type calendar struct {
	sessionItem *session.SessionItem
}

func New() *calendar {
	return &calendar{}
}

func (c *calendar) SetSession(sessionItem *session.SessionItem) {
	c.sessionItem = sessionItem
}

func getFirstDayDate(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
}

//GetCalendar returns
// calendar_prev, calendar_next, calendar_<1-31>, calendar_back
//
func GetCalendar(date time.Time, shoppingDays []int) tgbotapi.InlineKeyboardMarkup {
	var numericKeyboard tgbotapi.InlineKeyboardMarkup
	var rows [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton

	date = getFirstDayDate(date)

	lShift := func(n int) int {
		if n == 0 {
			return 6
		}
		return n - 1
	}

	curMonthParam := CalendarWord + "_" + helpers.Time2MonthCode(date)

	monthBtn := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s'%v", date.Month(), date.Year()), "0"))

	days := []tgbotapi.InlineKeyboardButton{}
	for _, day := range weekDays {
		days = append(days, tgbotapi.NewInlineKeyboardButtonData(day, curMonthParam))
	}
	weekDaysBtns := tgbotapi.NewInlineKeyboardRow(days...)

	prevMonthPtr := TryGetPrevMonthDate(date)
	leftBtn := tgbotapi.NewInlineKeyboardButtonData(emptyLabel, curMonthParam)
	if prevMonthPtr != nil {
		prevMonthParam := CalendarWord + "_" + helpers.Time2MonthCode(*prevMonthPtr)
		leftBtn = tgbotapi.NewInlineKeyboardButtonData(leftLabel, prevMonthParam)
	}

	nextMonthPtr := TryGetNextMonthDate(date)
	rightBtn := tgbotapi.NewInlineKeyboardButtonData(emptyLabel, curMonthParam)
	if nextMonthPtr != nil {
		nextMonthParam := CalendarWord + "_" + helpers.Time2MonthCode(*nextMonthPtr)
		rightBtn = tgbotapi.NewInlineKeyboardButtonData(rightLabel, nextMonthParam)
	}

	navBtns := tgbotapi.NewInlineKeyboardRow(
		leftBtn,
		tgbotapi.NewInlineKeyboardButtonData(backWord, consts.FirstPageStart),
		rightBtn,
	)
	rows = append(rows, monthBtn)
	rows = append(rows, weekDaysBtns)

	curDay := date
	weekDay := 0
	for {
		if curDay.Month() != date.Month() {
			if weekDay <= 6 {
				for i := 0; i <= 6-weekDay; i++ {
					row = append(row, tgbotapi.NewInlineKeyboardButtonData(emptyLabel, curMonthParam))
				}
			}
			rowsB := tgbotapi.NewInlineKeyboardRow(row...)
			rows = append(rows, rowsB)
			break
		}
		if weekDay > 6 {
			rowsB := tgbotapi.NewInlineKeyboardRow(row...)
			rows = append(rows, rowsB)
			row = []tgbotapi.InlineKeyboardButton{}
			weekDay = 0
		}
		if lShift(int(curDay.Weekday())) == weekDay {
			label := strconv.Itoa(curDay.Day())
			param := helpers.GetParam(
				consts.DayshoppingsWord,
				helpers.Time2DayCode(curDay),
			)

			for _, v := range shoppingDays {
				if v == curDay.Day() {
					label = helpers.GetUnderlinedText(label)
					break
				}
			}

			row = append(row, tgbotapi.NewInlineKeyboardButtonData(label, param))
			curDay = curDay.AddDate(0, 0, 1)
		} else {
			// if not controll button - callBackData == curMonthParam
			row = append(row, tgbotapi.NewInlineKeyboardButtonData(emptyLabel, curMonthParam))
		}
		weekDay++
	}
	rows = append(rows, navBtns)
	numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(rows...)
	return numericKeyboard
}

func getHoursDuration(date time.Time) float64 {
	duration := time.Now().Sub(date)
	diff := duration.Hours()
	return diff
}

//TryGetPrevMonthDate shift current date to month back or nil
func TryGetPrevMonthDate(date time.Time) *time.Time {
	diff := getHoursDuration(date)
	// if it really previous month and 1 year limit
	if diff > 0 && math.Abs(diff) >= timeLimit {
		return nil
	}
	if date.Month() == time.January {
		result := time.Date(date.Year()-1, time.December, 1, 0, 0, 0, 0, date.Location())
		return &result
	}
	result := time.Date(date.Year(), date.Month()-1, 1, 0, 0, 0, 0, date.Location())
	return &result
}

//TryGetNextMonthDate shift date to the month forward
func TryGetNextMonthDate(date time.Time) *time.Time {
	diff := getHoursDuration(date)
	// if it really next month and 1 year limit
	if diff < 0 && math.Abs(diff) >= timeLimit {
		return nil
	}
	date = date.AddDate(0, 1, 0)
	return &date
}

func (c *calendar) getOutputByDate(date time.Time) (logic.Output, error) {
	days, err := c.sessionItem.SListAPI.GetShoppingDays(date)
	if err != nil {
		log.Println(err.Error())
		return logic.Output{}, fmt.Errorf("shoplist api error: %w", err)

	}
	keyboard := GetCalendar(date, days)
	return logic.Output{
		Message:  calendarMessage,
		Keyboard: &keyboard,
	}, nil
}

func (c *calendar) GetCallbackOutput(command string) (logic.Output, error) {
	switch command {
	case consts.Start:
		return c.getOutputByDate(time.Now())
	default:
		nextMonth, err := helpers.MonthCode2Time(command)
		if err != nil {
			return logic.Output{}, consts.ErrUnknownCommand
		}
		return c.getOutputByDate(nextMonth)
	}
}

func (c *calendar) GetMessageOutput(curData string, msg string) (logic.Output, error) {
	return logic.Output{
		Message: "msg",
	}, nil
}
