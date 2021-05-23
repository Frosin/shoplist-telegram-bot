package helpers

import (
	"strconv"
	"strings"
	"time"

	"github.com/Frosin/shoplist-telegram-bot/consts"
)

const (
	dayCodeLayout   = "d2006-01-02"
	monthCodeLayout = "m2006-01"
)

type ParseResult struct {
	OperType       consts.OperationType
	StartOperation bool
	SelectItem     *int
	ShoppingID     int
}

func Time2DayCode(t time.Time) string {
	return t.Format(dayCodeLayout)
}

func DayCode2Time(dayCode string) (time.Time, error) {
	t, err := time.Parse(dayCodeLayout, dayCode)
	return t, err
}

func Time2MonthCode(t time.Time) string {
	return t.Format(monthCodeLayout)
}

func MonthCode2Time(monthCode string) (time.Time, error) {
	t, err := time.Parse(monthCodeLayout, monthCode)
	return t, err
}

// func GetParam(controlWord, commandWord string) string {
// 	return strings.Join([]string{controlWord, commandWord}, "_")
// }

func GetParam(controlWord string, words ...string) string {
	wordsStr := strings.Join(words, "")
	return strings.Join([]string{controlWord, wordsStr}, "_")
}

func GetUnderlinedText(text string) string {
	result := ""
	for _, s := range text {
		result = result + string(s) + "\u0332"
	}
	return result
}

func GetStrikeThroughText(text string) string {
	result := ""
	for _, s := range text {
		result = result + string(s) + "\u0336"
	}
	return result
}

// //GetRemovedNumsByCode gets "r34,56,78,87" adn returns []int{34,54,78,87}
// func getRemovedNumsByCode(code string) ([]int, error) {
// 	// if code consists "!" symbol
// 	if index := strings.Index(code, consts.ListStartRemoveSymbol); index > 0 {
// 		code = code[:len(code)-1]
// 	}

// 	result := []int{}
// 	arString := strings.Split(code[1:], ",")
// 	for i, v := range arString {
// 		if i == consts.ListOperationLimit {
// 			break
// 		}
// 		//check corrent int value
// 		value, err := strconv.Atoi(v)
// 		if err != nil {
// 			return nil, err
// 		}
// 		result = append(result, value)
// 	}
// 	return result, nil
// }

//IsInArray returns true if num exist in nums array
func IsInArray(num int, nums []int) bool {
	for _, v := range nums {
		if v == num {
			return true
		}
	}
	return false
}

//GetNodeName get nodeName from callBackData
func GetNodeName(word string) string {
	in := strings.Index(word, "_")
	return word[:in]
}

//GetOperationName get operationName from callBackData
func GetOperationName(word string) string {
	in := strings.Index(word, "_")
	return word[in+1:]
}

func ParseCommand(command string) (*ParseResult, error) {
	result := ParseResult{}

	//delete last symbol, it may be start symbol or numeric,
	//if it will be numeric we cut it in default switch construction
	idStr := command[:len(command)-1]

	// get lastSymbol
	lastSymbol := string(command[len(command)-1])

	// check for start operation commands
	result.StartOperation = true
	switch lastSymbol {
	case consts.ListStartSelectAllSymbol:
		result.OperType = consts.TypeOperationSelectAll
	case consts.ListStartRemoveSymbol:
		result.OperType = consts.TypeOperationDelete
	case consts.ListStartCopySymbol:
		result.OperType = consts.TypeOperationCopy
	case consts.ListStartAddFromCurrentList:
		result.OperType = consts.TypeOperationAddFromCurrent
	case consts.ListStartAddFromChecklist:
		result.OperType = consts.TypeOperationAddFromChecklist
	default:
		result.StartOperation = false
		itemIndex := strings.Index(command, consts.ListItemSymbol)
		if itemIndex > 0 {
			idStr = command[:itemIndex]
			numStr := command[itemIndex+1:]
			itemID, err := strconv.Atoi(numStr)
			if err != nil {
				return nil, err
			}
			result.SelectItem = &itemID
		} else {
			idStr = command
		}
	}

	shoppingID, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, err
	}
	result.ShoppingID = shoppingID

	return &result, nil
}
