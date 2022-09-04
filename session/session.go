package session

import (
	"log"
	"time"

	"github.com/Frosin/shoplist-telegram-bot/internal/shoplist/ent"
	"github.com/Frosin/shoplist-telegram-bot/shoplist"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	sessionLiveTime = time.Minute * 3
)

type SessionItem struct {
	SListAPI    *shoplist.Shoplist
	CurrentNode string
	CurrentData string
	LastMsgID   int
	ChatID      int64
	User        *ent.User
	removeTimer *time.Timer
	Data        interface{} //its field may be consists any value, we need
}

type SessionStorage struct {
	serviceURL string //database service URL
	startToken string
	items      map[int]*SessionItem //index by telegramUserID
	botAPI     *tgbotapi.BotAPI
	e          *ent.Client
}

func NewSessionStorage(serviceURL, startToken string, botAPI *tgbotapi.BotAPI, e *ent.Client) *SessionStorage {
	storage := SessionStorage{
		serviceURL: serviceURL,
		startToken: startToken,
		items:      map[int]*SessionItem{},
		botAPI:     botAPI,
		e:          e,
	}
	return &storage
}

func (s *SessionStorage) deferredDeletion(item *SessionItem) {
	item.removeTimer = time.AfterFunc(sessionLiveTime, func() {
		// startMessage := tgbotapi.NewMessage(item.ChatID, consts.StartText) //NewEditMessageText(item.ChatID, item.LastMsgID, consts.StartText)
		// startMessage.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		// 	[]tgbotapi.KeyboardButton{
		// 		tgbotapi.KeyboardButton{Text: consts.MenuText},
		// 	})
		// s.botAPI.Send(startMessage)
		delete(s.items, int(item.User.TelegramID))
	})
}

func (s *SessionStorage) Add(user *tgbotapi.User, chatID int64, startNode string) (*SessionItem, error) {
	// base client for create user
	client := shoplist.NewShoplistAPI(s.e, s.startToken)
	// init user
	userData, err := client.UserInit(user.ID, chatID, user.UserName)
	if err != nil {
		return nil, err
	}
	newTokenClient := shoplist.NewShoplistAPI(
		s.e,
		userData.Token,
	)

	// save session in storage
	item := SessionItem{
		SListAPI:    newTokenClient,
		CurrentNode: startNode,
		User:        userData,
		ChatID:      chatID,
	}
	s.items[user.ID] = &item

	// hide custom keyboard
	// msg := tgbotapi.NewMessage(item.ChatID, consts.AfterStartText)
	// //msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	// _, err = s.botAPI.Send(msg)
	// if err != nil {
	// 	log.Printf("sendMsg error=%v\n", err)
	// }

	// remove user session after the sessionLive interval
	s.deferredDeletion(&item)
	return &item, nil
}

func (s *SessionStorage) Get(update tgbotapi.Update, startNode string) (*SessionItem, error) {
	var (
		fromUser *tgbotapi.User
		chatID   int64
	)
	switch {
	case update.CallbackQuery != nil:
		fromUser = update.CallbackQuery.From
		chatID = update.CallbackQuery.Message.Chat.ID
	case update.Message != nil:
		fromUser = update.Message.From
		chatID = update.Message.Chat.ID
	}

	item, ok := s.items[fromUser.ID]
	if !ok {
		//debug
		log.Printf("added=%v", fromUser.ID)
		//
		return s.Add(fromUser, chatID, startNode)
	}
	// reset removeTimer
	if item.removeTimer != nil {
		finished := item.removeTimer.Stop()
		if finished {
			item.removeTimer.Reset(sessionLiveTime)
		}
	}

	//debug
	log.Printf("len=(%v)", len(s.items))
	for i, v := range s.items {
		log.Printf("%v-> sAPI=%v, userID=%v, communityID=%v", i, v.SListAPI, v.User.TelegramID, v.User.ComunityID)
	}
	//
	return item, nil
}

func (s *SessionItem) UpdateCallbackData(currentNode, currentData *string) {
	if currentNode != nil {
		s.CurrentNode = *currentNode
	}
	if currentData != nil {
		s.CurrentData = *currentData
	}
}

// AddIntDataToArray interprets session data as an int array and adds value to it
func (s *SessionItem) AddIntDataToArray(value int) {
	dataAsArray, _ := s.Data.([]int)
	dataAsArray = append(dataAsArray, value)
	s.Data = dataAsArray
}

func (s *SessionItem) SetDataArrayValue(value []int) {
	s.Data = value
}

func (s *SessionItem) DeleteValueInArray(value int) {
	dataAsArray, _ := s.Data.([]int)
	newArray := []int{}
	for _, v := range dataAsArray {
		if v == value {
			continue
		}
		newArray = append(newArray, v)
	}

	s.Data = newArray
}

// GetDataAsArray interprets session data as an int array and returns it
func (s *SessionItem) GetDataAsArray() []int {
	dataAsArray, _ := s.Data.([]int)
	return dataAsArray
}

// ClearDataArray clears data
func (s *SessionItem) ClearDataArray() {
	s.Data = []int{}
}
