package logic

import (
	"fmt"

	"github.com/Frosin/shoplist-telegram-bot/session"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Input struct {
	CallbackData *string
	Message      *string
}

type Output struct {
	Message            string
	Keyboard           *tgbotapi.InlineKeyboardMarkup
	MessageToCommunity *string
}

type Node interface {
	GetCallbackOutput(command string) (Output, error)
	GetMessageOutput(currentData, msg string) (Output, error)
	SetSession(sessionItem *session.SessionItem)
}

type Logic struct {
	nodes map[string]Node
}

func New() *Logic {
	return &Logic{
		nodes: make(map[string]Node),
	}
}

func (l *Logic) AddNode(name string, node Node) *Logic {
	l.nodes[name] = node
	return l
}

func (l *Logic) GetOutput(
	i Input,
	sessionItem *session.SessionItem,
) (Output, error) {
	node, ok := l.nodes[sessionItem.CurrentNode]
	if !ok {
		return Output{}, fmt.Errorf("node %s not found", sessionItem.CurrentNode)
	}
	//set current session for logic
	node.SetSession(sessionItem)
	//debug
	fmt.Printf("IN LOGIC: i.message=%v, i/callbackdata=%v", i.Message, i.CallbackData)
	//
	switch {
	case i.Message != nil:
		return node.GetMessageOutput(sessionItem.CurrentData, *i.Message)
	case i.CallbackData != nil:
		return node.GetCallbackOutput(sessionItem.CurrentData)
	}
	return Output{}, fmt.Errorf("bad input (fields are nills)")
}
