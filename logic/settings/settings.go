package settings

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Frosin/shoplist-telegram-bot/consts"
	"github.com/Frosin/shoplist-telegram-bot/helpers"
	"github.com/Frosin/shoplist-telegram-bot/logic"
	"github.com/Frosin/shoplist-telegram-bot/session"
	"github.com/dchest/uniuri"
	"github.com/spf13/viper"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	backBtnText          = "Назад"
	versionText          = "Версия бота: \n"
	IDUserText           = "ID текущего пользователя"
	NoComunityText       = "Вы не состоите в группе, для вступления в группу, введите ID участника"
	InYourGroupText      = "В вашей группе"
	YouCanText           = "вы можете создавать общие списки покупок"
	LeaveSuccessText     = "<Вы вышли из группы>"
	AlreadyInGroupText   = "<Вы уже состоите в группе>"
	InvalidUserIDText    = "<ID участника должен состоять из цифр>"
	UserNotFound         = "<Участник не найден. Проверьте корректность ID>"
	JoinGroupSuccessText = "<Вы вступили в группу>"

	leaveComunityText = "Покинуть группу"
	LeaveCommand      = "leave"
)

var (
	ErrBadComunityUsersCount = errors.New("comunity users not found")
)

type settings struct {
	sessionItem *session.SessionItem
}

func New() *settings {
	return &settings{}
}

func (s *settings) SetSession(sessionItem *session.SessionItem) {
	s.sessionItem = sessionItem
}

func (s *settings) getStartPage(withMessage string) (logic.Output, error) {
	//debug
	log.Printf("curItem: sAPI=%v, userID=%v, communityID=%v", s.sessionItem.SListAPI, *s.sessionItem.User.Id, *s.sessionItem.User.ComunityId)
	//
	comunityUsers, err := s.sessionItem.SListAPI.GetUsersByComunityID(*s.sessionItem.User.ComunityId)
	if err != nil {
		return logic.Output{}, err
	}
	comunityUsersCount := len(comunityUsers)

	buttonsRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(backBtnText, consts.FirstPageStart), //back Btn
	}

	version := viper.GetString("SHOPLIST-BOT_SERVICE_VERSION")
	message := fmt.Sprintf("%s%s: \"%v\".", versionText+version, IDUserText, *s.sessionItem.User.TelegramId)
	switch {
	case comunityUsersCount > 1:
		users := []string{}
		for _, v := range comunityUsers {
			userName := "no username"
			if *v.TelegramUsername != "" {
				userName = *v.TelegramUsername
			}
			user := fmt.Sprintf("ID=%v(%s)", *v.TelegramId, userName)
			users = append(users, user)
		}
		message += fmt.Sprintf(
			"\n%s: %s, %s",
			InYourGroupText,
			strings.Join(users, ", "),
			YouCanText,
		)
		leaveParam := helpers.GetParam(consts.SettingsWord, LeaveCommand)
		leaveBtn := tgbotapi.NewInlineKeyboardButtonData(leaveComunityText, leaveParam) // leaveComunity btn
		buttonsRow = append(buttonsRow, leaveBtn)

	case comunityUsersCount == 1:
		message += "\n" + NoComunityText
	default:
		return logic.Output{}, ErrBadComunityUsersCount
	}

	message = withMessage + "\n" + message

	keyboard := &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{buttonsRow},
	}
	return logic.Output{
		Message:  message,
		Keyboard: keyboard,
	}, nil
}

func (s *settings) GetCallbackOutput(command string) (logic.Output, error) {
	switch command {
	case LeaveCommand:
		// leave comunity handler
		newComunityID := uniuri.New()
		err := s.sessionItem.SListAPI.UpdateUser(
			*s.sessionItem.User.Id,
			&newComunityID,
			nil,
		)
		if err != nil {
			return logic.Output{}, err
		}
		// update in session
		s.sessionItem.User.ComunityId = &newComunityID
		return s.getStartPage(LeaveSuccessText)
	}
	return s.getStartPage("")
}

func (s *settings) GetMessageOutput(curData string, msg string) (logic.Output, error) {
	comunityUsers, err := s.sessionItem.SListAPI.GetUsersByComunityID(*s.sessionItem.User.ComunityId)
	if err != nil {
		return logic.Output{}, err
	}
	comunityUsersCount := len(comunityUsers)
	if comunityUsersCount > 1 {
		return s.getStartPage(AlreadyInGroupText)
	}

	intUserID, err := strconv.Atoi(msg)
	if err != nil {
		return s.getStartPage(InvalidUserIDText)
	}

	groupOwner, err := s.sessionItem.SListAPI.GetUserByTelegramID(intUserID)
	switch {
	case err == consts.ErrNotFound:
		return s.getStartPage(UserNotFound)
	case err != nil:
		return logic.Output{}, err
	}

	// update user comunityID
	err = s.sessionItem.SListAPI.UpdateUser(
		*s.sessionItem.User.Id,
		groupOwner.ComunityId,
		nil,
	)
	if err != nil {
		return logic.Output{}, err
	}
	// update in session
	s.sessionItem.User.ComunityId = groupOwner.ComunityId
	return s.getStartPage(JoinGroupSuccessText)
}
