package checklist

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
	backSumbol   = "⬅ Меню"
	inputMsg     = "Чек-лист. Введите товар для добавления"
	removeMsg    = "Удалить"
	copyMsg      = "В текущий список"
	selectAllMsg = "Выделить все"
	emptyItems   = "Чек-лист пока что пуст. Для добавления введите название товара."
	copiedSucess = "Товары скопированы."
	noNewItems   = "Нет новых товаров для добавления. "
)

type checklist struct {
	sessionItem *session.SessionItem
}

func New() *checklist {
	return &checklist{}
}

func (c *checklist) SetSession(sessionItem *session.SessionItem) {
	c.sessionItem = sessionItem
}

func (c *checklist) GetCallbackOutput(command string) (logic.Output, error) {
	var checklistShoppingID int
	var err error

	// if first start of checklist page we will get checklist shoppingID
	if command == consts.Start {
		// delete items
		c.sessionItem.ClearDataArray()
		// get checklist shopping ID
		checklistShoppingID, err = c.sessionItem.SListAPI.GetSpecialShopping(consts.ShoppingTypeCheckList)
		switch {
		case err == consts.ErrNotFound:
			checklistShoppingID, err = c.sessionItem.SListAPI.AddShoppingWithType(
				time.Now(),
				consts.ChecklistWord,
				consts.ShoppingTypeCheckList,
			)
			if err != nil {
				return logic.Output{}, err
			}
		case err != nil:
			return logic.Output{}, err
		}

		parseObject := helpers.ParseResult{
			ShoppingID: checklistShoppingID,
		}

		return c.getOutput(&parseObject, "", nil)
	}

	// parse command
	parseResult, err := helpers.ParseCommand(command)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.ChecklistWord, err)
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

	// any of start operation buttons was pressed
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

			return c.getOutput(parseResult, "", nil)
		case consts.TypeOperationSelectAll:
			// select all
			checklistItems, err := c.sessionItem.SListAPI.GetShoppingItems(parseResult.ShoppingID)
			if err != nil {
				return logic.Output{}, fmt.Errorf("%v: %w", consts.ChecklistWord, err)
			}
			arItemIDs := []int{}
			for _, v := range checklistItems {
				arItemIDs = append(arItemIDs, v.ID)
			}
			c.sessionItem.SetDataArrayValue(arItemIDs)
		case consts.TypeOperationCopy:
			// copy items to current list
			checklistItems, err := c.sessionItem.SListAPI.GetShoppingItems(parseResult.ShoppingID)
			if err != nil {
				return logic.Output{}, fmt.Errorf("%v: %w", consts.ChecklistWord, err)
			}
			//get currentlist shoppingID
			currentlistShoppingID, err := c.sessionItem.SListAPI.GetSpecialShopping(consts.ShoppingTypeCurrentList)
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
			// get currentlist items
			currentlistItems, err := c.sessionItem.SListAPI.GetShoppingItems(currentlistShoppingID)
			if err != nil {
				return logic.Output{}, fmt.Errorf("%v: %w", consts.ChecklistWord, err)
			}
			//add checklist items to current list with check of duplicates

			alreadyExist := func(itemName string) bool {
				for _, currentlistItem := range currentlistItems {
					if itemName == currentlistItem.ProductName {
						return true
					}
				}
				return false
			}

			notSelected := func(itemID int) bool {
				selectedItems := c.sessionItem.GetDataAsArray()
				for _, id := range selectedItems {
					if itemID == id {
						return false
					}
				}
				return true
			}

			itemsToAdd := []string{}
			for _, checklistItem := range checklistItems {
				if alreadyExist(checklistItem.ProductName) || notSelected(checklistItem.ID) {
					continue
				}
				itemsToAdd = append(itemsToAdd, checklistItem.ProductName)
			}

			if len(itemsToAdd) > 0 {
				for _, item := range itemsToAdd {
					err = c.sessionItem.SListAPI.AddItem(currentlistShoppingID, item)
					if err != nil {
						return logic.Output{}, fmt.Errorf("%v: %w", consts.ChecklistWord, err)
					}
				}

				// clear
				c.sessionItem.ClearDataArray()

				return c.getOutput(parseResult, copiedSucess, nil)
			}
			// no new items to add message
			return c.getOutput(parseResult, noNewItems, nil)
		}
	}

	return c.getOutput(parseResult, "", nil)
}

func (c *checklist) GetMessageOutput(curData string, msg string) (logic.Output, error) {
	var result *helpers.ParseResult
	var err error
	// if first start of checklist page
	if curData == consts.Start {
		// get checklist shopping ID
		checklistShoppingID, err := c.sessionItem.SListAPI.GetSpecialShopping(consts.ShoppingTypeCheckList)
		if err != nil {
			return logic.Output{}, fmt.Errorf("%v: %w", consts.ChecklistWord, err)
		}

		result = &helpers.ParseResult{
			ShoppingID: checklistShoppingID,
		}
	} else {
		result, err = helpers.ParseCommand(curData)
		if err != nil {
			return logic.Output{}, fmt.Errorf("%v: %w", consts.ChecklistWord, err)
		}
	}

	err = c.sessionItem.SListAPI.AddItem(result.ShoppingID, msg)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.ChecklistWord, err)
	}

	return c.getOutput(result, "", &msg)
}

func (c *checklist) getOutput(parseObject *helpers.ParseResult, additionalMessage string, addedItems *string) (logic.Output, error) {
	shoppingIDStr := strconv.Itoa(parseObject.ShoppingID)
	selectedItems := c.sessionItem.GetDataAsArray()
	//listStr := ""

	// strArrayNums := []string{}
	// for _, v := range parseObject.Items {
	// 	strArrayNums = append(strArrayNums, strconv.Itoa(v))
	// }

	// if parseObject.Items != nil {
	// 	listStr = consts.ListItemSymbol + strings.Join(strArrayNums, ",")
	// }
	shoppingData, err := c.sessionItem.SListAPI.GetShopping(parseObject.ShoppingID)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.ChecklistWord, err)
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
		return logic.Output{}, fmt.Errorf("%v: %w", consts.ChecklistWord, err)
	}

	column := [][]tgbotapi.InlineKeyboardButton{}

	// create items list to show
	for i, data := range items {
		itemIDStr := strconv.Itoa(data.ID)
		itemName := data.ProductName
		// underlined item name
		if helpers.IsInArray(data.ID, selectedItems) {
			itemName = helpers.GetUnderlinedText(itemName)
		}

		// add itemID to itemsIDs string
		// listData := consts.ListItemSymbol + itemIDStr
		// if parseObject.Items != nil {
		// 	listData = listStr + "," + itemIDStr
		// }
		itemData := consts.ListItemSymbol + itemIDStr

		//make item buttom param with shopping id, removed item ids + id curItem to remove
		//as example: "checklist_123i45"
		param := helpers.GetParam(
			consts.ChecklistWord,
			shoppingIDStr,
			itemData,
		)
		row := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(i+1)+". "+itemName, param),
		}
		column = append(column, row)
	}

	if len(items) > 0 {
		// select all button
		selectAllButton := tgbotapi.NewInlineKeyboardButtonData(selectAllMsg,
			helpers.GetParam(
				consts.ChecklistWord,
				shoppingIDStr,
				consts.ListStartSelectAllSymbol,
			))
		controlButtons = append(controlButtons, selectAllButton)
	}

	// add remove and copy buttons to keyboard
	if len(selectedItems) > 0 {
		//remove button
		removeButton := tgbotapi.NewInlineKeyboardButtonData(removeMsg,
			helpers.GetParam(
				consts.ChecklistWord,
				shoppingIDStr,
				consts.ListStartRemoveSymbol,
			))
		// copy button
		copyButton := tgbotapi.NewInlineKeyboardButtonData(copyMsg,
			helpers.GetParam(
				consts.ChecklistWord,
				shoppingIDStr,
				consts.ListStartCopySymbol,
			))
		controlButtons = append(controlButtons, removeButton, copyButton)
	}

	//final keyboard
	column = append(column, controlButtons)
	keyboard := &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: column,
	}

	output := logic.Output{
		Message:  fmt.Sprintf("%s %s", additionalMessage, inputMsg),
		Keyboard: keyboard,
	}

	if addedItems != nil {
		msg := fmt.Sprintf(
			"Пользователь %s(%v) добавил '%s' в '%s'(%s)",
			c.sessionItem.User.TelegramUsername,
			c.sessionItem.User.TelegramID,
			*addedItems,
			shoppingData.Edges.Shop.Name,
			shoppingData.Date,
		)
		output.MessageToCommunity = &msg
	}

	return output, nil
}
