package shoppingitems

import (
	"fmt"
	"strconv"

	"github.com/Frosin/shoplist-telegram-bot/consts"
	"github.com/Frosin/shoplist-telegram-bot/helpers"
	"github.com/Frosin/shoplist-telegram-bot/logic"
	"github.com/Frosin/shoplist-telegram-bot/session"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	backSumbol          = "⬅ Вернуться к "
	inputMsg            = "Введите товар для добавления"
	addFromCurrentMsg   = "↑ из текущего"
	addFromChecklistMsg = "↑ из чек-листа"
	removeMsg           = "⊗ выбранные"
	emptyItems          = "Список товаров пока что пуст. Для добавления введите название товара."
)

type shoppingItems struct {
	sessionItem *session.SessionItem
}

func New() *shoppingItems {
	return &shoppingItems{}
}

func (s *shoppingItems) SetSession(sessionItem *session.SessionItem) {
	s.sessionItem = sessionItem
}

func (s *shoppingItems) GetCallbackOutput(command string) (logic.Output, error) {
	var err error

	shoppingID, err := strconv.Atoi(command)
	// if first show
	if err == nil {
		parseRes := &helpers.ParseResult{
			ShoppingID: shoppingID,
		}

		return s.getOutput(parseRes, nil)
	}

	// parse command
	parseResult, err := helpers.ParseCommand(command)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.ShoppingitemsWord, err)
	}

	// item button is pressed add itemID to session data storage
	if parseResult.SelectItem != nil {
		selectedItems := s.sessionItem.GetDataAsArray()
		if helpers.IsInArray(*parseResult.SelectItem, selectedItems) {
			// unselect item
			s.sessionItem.DeleteValueInArray(*parseResult.SelectItem)
		}
		s.sessionItem.AddIntDataToArray(*parseResult.SelectItem)
	}

	if parseResult.StartOperation {
		switch parseResult.OperType {
		case consts.TypeOperationDelete:
			//remove items and get output
			selectItems := s.sessionItem.GetDataAsArray()
			err = s.sessionItem.SListAPI.RemoveItems(selectItems)
			// delete items
			s.sessionItem.ClearDataArray()
			return s.getOutput(parseResult, nil)
		case consts.TypeOperationAddFromCurrent:
			items, err := s.sessionItem.SListAPI.GetShoppingItems(parseResult.ShoppingID)
			if err != nil {
				return logic.Output{}, fmt.Errorf("%v: %w", consts.ChecklistWord, err)
			}
			//get currentlist shoppingID
			currentlistShoppingID, err := s.sessionItem.SListAPI.GetSpecialShopping(consts.ShoppingTypeCurrentList)
			switch {
			case err == consts.ErrNotFound:
				// no items in checklist shopping
				return s.getOutput(parseResult, nil)
			case err != nil:
				return logic.Output{}, err
			}
			// get currentlist items
			currentlistItems, err := s.sessionItem.SListAPI.GetShoppingItems(*currentlistShoppingID)
			if err != nil {
				return logic.Output{}, fmt.Errorf("%v: %w", consts.ShoppingitemsWord, err)
			}

			//add current list items to shopping with duplicate check
			alreadyExist := func(itemName string) bool {
				for _, shoppingItem := range *items {
					if itemName == shoppingItem.ProductName {
						return true
					}
				}
				return false
			}

			currentItemsIDs := []int{}
			for _, currentlistItem := range *currentlistItems {
				if alreadyExist(currentlistItem.ProductName) {
					continue
				}

				currentItemsIDs = append(currentItemsIDs, *currentlistItem.Id)

				err = s.sessionItem.SListAPI.AddItem(parseResult.ShoppingID, currentlistItem.ProductName)
				if err != nil {
					return logic.Output{}, fmt.Errorf("%v: %w", consts.ShoppingitemsWord, err)
				}
			}

			// remove current list items after copying
			err = s.sessionItem.SListAPI.RemoveItems(currentItemsIDs)
			if err != nil {
				return logic.Output{}, fmt.Errorf("%v: %w", consts.ShoppingitemsWord, err)
			}

			// show
			return s.getOutput(parseResult, nil)
		case consts.TypeOperationAddFromChecklist:
			items, err := s.sessionItem.SListAPI.GetShoppingItems(parseResult.ShoppingID)
			if err != nil {
				return logic.Output{}, fmt.Errorf("%v: %w", consts.ChecklistWord, err)
			}
			//get checklist shoppingID
			checklistShoppingID, err := s.sessionItem.SListAPI.GetSpecialShopping(consts.ShoppingTypeCheckList)
			switch {
			case err == consts.ErrNotFound:
				// no items in checklist shopping
				return s.getOutput(parseResult, nil)
			case err != nil:
				return logic.Output{}, err
			}
			// get currentlist items
			checklistItems, err := s.sessionItem.SListAPI.GetShoppingItems(*checklistShoppingID)
			if err != nil {
				return logic.Output{}, fmt.Errorf("%v: %w", consts.ShoppingitemsWord, err)
			}

			//add current list items to shopping with duplicate check
			alreadyExist := func(itemName string) bool {
				for _, shoppingItem := range *items {
					if itemName == shoppingItem.ProductName {
						return true
					}
				}
				return false
			}

			checklistItemsIDs := []int{}
			for _, checklistItem := range *checklistItems {
				if alreadyExist(checklistItem.ProductName) {
					continue
				}

				checklistItemsIDs = append(checklistItemsIDs, *checklistItem.Id)

				err = s.sessionItem.SListAPI.AddItem(parseResult.ShoppingID, checklistItem.ProductName)
				if err != nil {
					return logic.Output{}, fmt.Errorf("%v: %w", consts.ShoppingitemsWord, err)
				}
			}

			// show
			return s.getOutput(parseResult, nil)
		}
	}

	return s.getOutput(parseResult, nil)
}

func (s *shoppingItems) GetMessageOutput(curData string, msg string) (logic.Output, error) {
	var err error
	result, err := helpers.ParseCommand(curData)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.CurrentlistWord, err)
	}
	//debug
	fmt.Println("\ncurData=", curData)
	fmt.Println("\nmsg=", msg)
	fmt.Println("\nresult=", result)
	//
	err = s.sessionItem.SListAPI.AddItem(result.ShoppingID, msg)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.CurrentlistWord, err)
	}

	return s.getOutput(result, &msg)
}

func (s *shoppingItems) getOutput(parseObject *helpers.ParseResult, addedItemName *string) (logic.Output, error) {
	shoppingIDStr := strconv.Itoa(parseObject.ShoppingID)
	selectedItems := s.sessionItem.GetDataAsArray()

	shoppingData, err := s.sessionItem.SListAPI.GetShopping(parseObject.ShoppingID)
	if err != nil {
		return logic.Output{}, fmt.Errorf("%v: %w", consts.ShoppingitemsWord, err)
	}

	backBtnParam := helpers.GetParam(
		consts.DayshoppingsWord,
		"d"+shoppingData.Date,
	)

	//create keyboard and add back button to keyboard
	controlButtons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(backSumbol, backBtnParam),
	}

	items, err := s.sessionItem.SListAPI.GetShoppingItems(parseObject.ShoppingID)
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
		return logic.Output{}, fmt.Errorf("%v: %w", consts.ShoppingitemsWord, err)
	}

	column := [][]tgbotapi.InlineKeyboardButton{}

	// create items list to show
	for i, data := range *items {
		itemIDStr := strconv.Itoa(*data.Id)
		itemName := data.ProductName
		// strikethrough item name
		if helpers.IsInArray(*data.Id, selectedItems) {
			itemName = helpers.GetStrikeThroughText(itemName)
		}

		itemData := consts.ListItemSymbol + itemIDStr

		//make item buttom param with shopping id, removed item ids + id curItem to remove
		param := helpers.GetParam(
			consts.ShoppingitemsWord,
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
		// remove button
		removeButton := tgbotapi.NewInlineKeyboardButtonData(removeMsg,
			helpers.GetParam(
				consts.ShoppingitemsWord,
				shoppingIDStr,
				consts.ListStartRemoveSymbol,
			))

		controlButtons = append(controlButtons, removeButton)
	}

	// add from current list button
	addFromCurrentButton := tgbotapi.NewInlineKeyboardButtonData(addFromCurrentMsg,
		helpers.GetParam(
			consts.ShoppingitemsWord,
			shoppingIDStr,
			consts.ListStartAddFromCurrentList,
		))
	controlButtons = append(controlButtons, addFromCurrentButton)

	// add from checklist button
	addFromChecklistButton := tgbotapi.NewInlineKeyboardButtonData(addFromChecklistMsg,
		helpers.GetParam(
			consts.ShoppingitemsWord,
			shoppingIDStr,
			consts.ListStartAddFromChecklist,
		))
	controlButtons = append(controlButtons, addFromChecklistButton)

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
			*s.sessionItem.User.TelegramUsername,
			*s.sessionItem.User.TelegramId,
			*addedItemName,
			shoppingData.Name,
			shoppingData.Date,
		)
		output.MessageToCommunity = &msg
	}

	return output, nil
}
