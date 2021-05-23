package consts

import (
	"errors"
	"time"
)

type ShoppingType int
type OperationType string

const (
	ReadTimeout  = 15 * time.Second
	WriteTimeout = 20 * time.Second

	StartText      = "Время сессии истекло"
	MenuText       = "Меню"
	AfterStartText = "\xE2\x9C\x8C"

	FirstPageStart   = "firstpage_start"
	CalendarStart    = "calendar_start"
	SettingsStart    = "settings_start"
	ChecklistStart   = "checklist_start"
	CurrentListStart = "currentlist_start"

	CalendarWord      = "calendar"
	DayshoppingsWord  = "dayshoppings"
	ShoppingitemsWord = "shoppingitems"
	SettingsWord      = "settings"
	ChecklistWord     = "checklist"
	CurrentlistWord   = "currentlist"

	Start = "start"

	DateLayout = "2006-01-02"

	ListItemSymbol              = "i"
	ListOperationLimit          = 3
	ListStartRemoveSymbol       = "!"
	ListStartCopySymbol         = "&"
	ListStartSelectAllSymbol    = "*"
	ListStartAddFromCurrentList = "^"
	ListStartAddFromChecklist   = "#"

	ShoppingTypeDefault     ShoppingType = 0
	ShoppingTypeCheckList   ShoppingType = 1
	ShoppingTypeCurrentList ShoppingType = 2

	TypeOperationDelete           OperationType = "delete"
	TypeOperationCopy             OperationType = "copy"
	TypeOperationSelectAll        OperationType = "selectAll"
	TypeOperationAddFromCurrent   OperationType = "addFromCurrent"
	TypeOperationAddFromChecklist OperationType = "addFromChecklist"
)

var (
	ErrUnknownCommand = errors.New("unknown command")
	ErrNotFound       = errors.New("not found")

	ErrGetUsersBadStatus = errors.New("bad status get user response")
)
