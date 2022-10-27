package main

import (
	"fmt"
	"log"
	"strings"

	"entgo.io/ent/dialect"
	"github.com/getsentry/sentry-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/spf13/viper"

	"github.com/Frosin/shoplist-telegram-bot/bugetstorage"
	"github.com/Frosin/shoplist-telegram-bot/consts"
	"github.com/Frosin/shoplist-telegram-bot/helpers"
	"github.com/Frosin/shoplist-telegram-bot/internal/shoplist/ent"
	"github.com/Frosin/shoplist-telegram-bot/iot"
	"github.com/Frosin/shoplist-telegram-bot/logic"
	"github.com/Frosin/shoplist-telegram-bot/logic/buget"
	"github.com/Frosin/shoplist-telegram-bot/logic/bugetcategory"
	"github.com/Frosin/shoplist-telegram-bot/logic/calendar"
	"github.com/Frosin/shoplist-telegram-bot/logic/checklist"
	"github.com/Frosin/shoplist-telegram-bot/logic/currentlist"
	"github.com/Frosin/shoplist-telegram-bot/logic/dayshoppings"
	"github.com/Frosin/shoplist-telegram-bot/logic/firstpage"
	"github.com/Frosin/shoplist-telegram-bot/logic/iotlogic"
	"github.com/Frosin/shoplist-telegram-bot/logic/settings"
	"github.com/Frosin/shoplist-telegram-bot/logic/shoppingitems"
	"github.com/Frosin/shoplist-telegram-bot/session"
	_ "github.com/mattn/go-sqlite3"
)

const (
	debugMode = false
	startNode = "firstpage"
)

var (
	cfgFile string
)

func sentryInit(dsn string) {
	sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
	})
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("shoplist-bot")
	}
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("shoplist-bot")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func sendErrorMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, err error) {
	errMsg := err.Error()
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "sendErrorMsg"+errMsg)
	_, err = bot.Send(msg)
	if err != nil {
		log.Printf("send error message Err=%v\n", err)
	}
}

func updateHandler(
	update tgbotapi.Update,
	sessions *session.SessionStorage,
	appLogic *logic.Logic,
	bot *tgbotapi.BotAPI,
	startNode string,
) {

	// get session by updateData
	sessionItem, err := sessions.Get(
		update,
		startNode,
	)
	if err != nil {
		sendErrorMessage(bot, update, err)
	}

	//debug
	log.Printf("update.CallbackQuery=%v\n", update.CallbackQuery)
	if update.CallbackQuery != nil {
		log.Printf("update.CallbackQueryData=%v\n", update.CallbackQuery.Data)
	}
	log.Printf("update.Message=%v\n", update.Message)
	//

	if sessionItem != nil && sessionItem.LastMsgID != 0 && update.CallbackQuery != nil {
		currentNode := helpers.GetNodeName(update.CallbackQuery.Data)
		currentData := helpers.GetOperationName(update.CallbackQuery.Data)

		sessionItem.UpdateCallbackData(&currentNode, &currentData)

		inputData := logic.Input{
			CallbackData: &update.CallbackQuery.Data,
		}
		output, err := appLogic.GetOutput(
			inputData,
			sessionItem,
		)
		if err != nil {
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "error: "+err.Error())
			_, err = bot.Send(msg)
			if err != nil {
				log.Println("error sending error msg")
			}
		}

		if output.MessageToCommunity != nil {
			communityUsers, err := sessionItem.SListAPI.GetUsersByComunityID(sessionItem.User.ComunityID)
			if err != nil {
				sendErrorMessage(bot, update, err)
			}

			for _, user := range communityUsers {
				if user.TelegramID == sessionItem.User.TelegramID {
					continue
				}
				newMsg := tgbotapi.NewMessage(user.ChatID, *output.MessageToCommunity)
				_, err := bot.Send(newMsg)
				if err != nil {
					log.Println("error sending community msg")
				}
			}
			//bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "")) // Всплывашка с data нажатой кнопки
		}

		debugMsg := ""
		if debugMode {
			debugMsg = "[" + update.CallbackQuery.Data + "]"
		}

		editedMessage := tgbotapi.NewEditMessageText(
			update.CallbackQuery.Message.Chat.ID,
			sessionItem.LastMsgID,
			debugMsg+output.Message,
		)
		if output.Keyboard != nil {
			editedMessage.ReplyMarkup = output.Keyboard
		}
		_, err = bot.Send(editedMessage)
		if err != nil {
			log.Println("error sending msg")
		}
		return
	}

	if update.Message != nil {
		inputData := logic.Input{
			Message: &update.Message.Text,
		}
		output, err := appLogic.GetOutput(
			inputData,
			sessionItem,
		)

		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "error: "+err.Error())
			_, err = bot.Send(msg)
			if err != nil {
				log.Println("error sending error msg")
			}
		}

		// send message to community users
		if output.MessageToCommunity != nil {
			communityUsers, err := sessionItem.SListAPI.GetUsersByComunityID(sessionItem.User.ComunityID)
			if err != nil {
				sendErrorMessage(bot, update, err)
			}

			for _, user := range communityUsers {
				if user.TelegramID == sessionItem.User.TelegramID {
					continue
				}
				newMsg := tgbotapi.NewMessage(user.ChatID, *output.MessageToCommunity)
				_, err := bot.Send(newMsg)
				if err != nil {
					log.Println("error sending community msg")
				}
			}
			//bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "")) // Всплывашка с data нажатой кнопки
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, output.Message)

		if output.Keyboard != nil {
			msg.ReplyMarkup = *output.Keyboard
		}
		mess, err := bot.Send(msg)
		if err != nil {
			log.Println("error sending msg")
		}
		sessionItem.LastMsgID = mess.MessageID
	}
}

func main() {
	initConfig()
	port := viper.GetString("SHOPLIST-BOT_PORT")
	sentryDsn := viper.GetString("SHOPLIST-BOT_SENTRY_DSN")
	token := viper.GetString("SHOPLIST-BOT_TOKEN")
	webhookURL := viper.GetString("SHOPLIST-BOT_WEBHOOK_URL")
	serviceURI := viper.GetString("SHOPLIST-BOT_SERVICE_URI")
	startToken := viper.GetString("SHOPLIST-BOT_SERVICE_START_TOKEN")
	// output envs
	log.Printf("port = %s", port)
	log.Printf("sentryDsn = %s", sentryDsn)
	log.Printf("token = %s", token)
	log.Printf("webhookURL = %s", webhookURL)
	log.Printf("serviceURI = %s", serviceURI)
	log.Printf("startToken = %s", startToken)

	sentryInit(sentryDsn)

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	_, err = bot.RemoveWebhook()
	if err != nil {
		log.Fatal(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	iotStorage := iot.NewIOTStorageMap()

	srv := iot.NewServer(iotStorage, "8090")
	go srv.StartServer()

	// get ent
	e := getEnt()

	sessionStorage := session.NewSessionStorage(serviceURI, startToken, bot, e)

	bugetStorage, err := bugetstorage.NewStorage()
	if err != nil {
		log.Fatal(err)
	}

	//Create new logic with pages (nodes)
	appLogic := logic.New().
		AddNode(calendar.CalendarWord, calendar.New()).
		AddNode(firstpage.FirstpageWord, firstpage.New()).
		AddNode(dayshoppings.DayshoppingsWord, dayshoppings.New()).
		AddNode(consts.ShoppingitemsWord, shoppingitems.New()).
		AddNode(consts.SettingsWord, settings.New()).
		AddNode(consts.ChecklistWord, checklist.New()).
		AddNode(consts.CurrentlistWord, currentlist.New()).
		AddNode(consts.BugetWord, buget.New(bugetStorage)).
		AddNode(consts.BugetCategoryWord, bugetcategory.New(bugetStorage)).
		AddNode(consts.IOTWord, iotlogic.New(iotStorage))

	log.Println("start updates")
	for update := range updates {
		go updateHandler(update, sessionStorage, appLogic, bot, startNode)
	}
}

func getEnt() *ent.Client {
	dbFullFileName := viper.GetString("SHOPLIST-BOT_SHOPLISTTPATH")

	log.Println("shoplist file=", dbFullFileName)
	client, err := ent.Open(dialect.SQLite, fmt.Sprintf("file:%s?_fk=1", dbFullFileName))
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}

	return client
}
