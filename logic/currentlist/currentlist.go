package currentlist

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Frosin/shoplist-telegram-bot/consts"
	"github.com/Frosin/shoplist-telegram-bot/helpers"
	"github.com/Frosin/shoplist-telegram-bot/logic"
	"github.com/Frosin/shoplist-telegram-bot/session"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	backSumbol = "⬅ Меню"
	inputMsg   = "Текущий список. Введите товар для добавления"
	removeMsg  = "Удалить"
	emptyItems = "Текущий список пока что пуст. Для добавления введите название товара."
)

type currentlist struct {
	sessionItem *session.SessionItem
}

func New() *currentlist {
	return &currentlist{}
}

func (c *currentlist) SetSession(sessionItem *session.SessionItem) {
	c.sessionItem = sessionItem
}

func (c *currentlist) GetCallbackOutput(command string) (logic.Output, error) {
	var currentlistShoppingID int
	var err error

	// if first start of current page
	if command == consts.Start {
		// get currentlist shopping ID
		currentlistShoppingID, err = c.sessionItem.SListAPI.GetSpecialShopping(consts.ShoppingTypeCurrentList)
		switch {
		case err == consts.ErrNotFound:
			currentlistShoppingID, err = c.sessionItem.SListAPI.AddShoppingWithType(
				time.Now(),
				consts.CurrentlistWord,
				consts.ShoppingTypeCurrentList,
			)
			if err != nil {
				return logic.Output{}, err
			}
		case err != nil:
			return logic.Output{}, err
		}

		parseObject := helpers.ParseResult{
			ShoppingID: currentlistShoppingID,
		}

		return c.getOutput(&parseObject, nil)
	}

	// parse command
	parseResult, err := helpers.ParseCommand(command)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.CurrentlistWord, err)
	}

	// item button is pressed add itemID to session data storage
	if parseResult.SelectItem != nil {
		selectedItems := c.sessionItem.GetDataAsArray()
		if helpers.IsInArray(*parseResult.SelectItem, selectedItems) {
			// unselect item
			c.sessionItem.DeleteValueInArray(*parseResult.SelectItem)
		} else {
			c.sessionItem.AddIntDataToArray(*parseResult.SelectItem)
		}

	}

	if parseResult.StartOperation {
		switch parseResult.OperType {
		case consts.TypeOperationDelete:
			//remove items and get output
			selectItems := c.sessionItem.GetDataAsArray()
			err = c.sessionItem.SListAPI.RemoveItems(selectItems)
			if err != nil {
				return logic.Output{}, err
			}
			// delete items
			c.sessionItem.ClearDataArray()

			return c.getOutput(parseResult, nil)
		}
	}

	return c.getOutput(parseResult, nil)
}

func (c *currentlist) GetMessageOutput(curData string, msg string) (logic.Output, error) {
	var result *helpers.ParseResult
	var err error
	// if first start of currentlist page
	if curData == consts.Start {
		// get current shopping ID
		currentlistShoppingID, err := c.sessionItem.SListAPI.GetSpecialShopping(consts.ShoppingTypeCurrentList)
		if err != nil {
			return logic.Output{}, fmt.Errorf("%v: %w", consts.CurrentlistWord, err)
		}

		result = &helpers.ParseResult{
			ShoppingID: currentlistShoppingID,
		}
	} else {
		result, err = helpers.ParseCommand(curData)
		if err != nil {
			return logic.Output{}, fmt.Errorf("%v: %w", consts.CurrentlistWord, err)
		}
	}

	err = c.sessionItem.SListAPI.AddItem(result.ShoppingID, msg)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.CurrentlistWord, err)
	}

	return c.getOutput(result, &msg)
}

func (c *currentlist) getOutput(parseObject *helpers.ParseResult, addedItemName *string) (logic.Output, error) {
	shoppingIDStr := strconv.Itoa(parseObject.ShoppingID)
	selectedItems := c.sessionItem.GetDataAsArray()

	shoppingData, err := c.sessionItem.SListAPI.GetShopping(parseObject.ShoppingID)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.CurrentlistWord, err)
	}

	//create keyboard and add back button to keyboard
	controlButtons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(backSumbol, consts.FirstPageStart),
	}

	items, err := c.sessionItem.SListAPI.GetShoppingItems(parseObject.ShoppingID)
	if err != nil {
		//if get empty items list
		// heroku's go not understand errors.Is
		if err == consts.ErrNotFound {
			return logic.Output{
				Message: emptyItems,
				Keyboard: &tgbotapi.InlineKeyboardMarkup{
					InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
						controlButtons,
					},
				},
			}, nil
		}
		return logic.Output{}, fmt.Errorf("%v: %w", consts.CurrentlistWord, err)
	}

	column := [][]tgbotapi.InlineKeyboardButton{}

	// create items list to show
	for i, data := range items {
		itemIDStr := strconv.Itoa(data.ID)
		itemName := data.ProductName
		// strikethrough item name
		if helpers.IsInArray(data.ID, selectedItems) {
			itemName = helpers.GetStrikeThroughText(itemName)
		}

		itemData := consts.ListItemSymbol + itemIDStr

		//make item buttom param with shopping id, removed item ids + id curItem to remove
		param := helpers.GetParam(
			consts.CurrentlistWord,
			shoppingIDStr,
			itemData,
		)
		row := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(i+1)+". "+itemName, param),
		}
		column = append(column, row)
	}

	// add remove and copy buttons to keyboard
	if len(selectedItems) > 0 {
		//remove button
		removeButton := tgbotapi.NewInlineKeyboardButtonData(removeMsg,
			helpers.GetParam(
				consts.CurrentlistWord,
				shoppingIDStr,
				consts.ListStartRemoveSymbol,
			))
		controlButtons = append(controlButtons, removeButton)
	}

	//final keyboard
	column = append(column, controlButtons)
	keyboard := &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: column,
	}

	output := logic.Output{
		Message:  inputMsg,
		Keyboard: keyboard,
	}

	if addedItemName != nil {
		msg := fmt.Sprintf(
			"Пользователь %s(%v) добавил '%s' в '%s'(%s)",
			c.sessionItem.User.TelegramUsername,
			c.sessionItem.User.TelegramID,
			*addedItemName,
			shoppingData.Edges.Shop.Name,
			shoppingData.Date,
		)
		output.MessageToCommunity = &msg
	}

	return output, nil
}
