package shoplist

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Frosin/shoplist-telegram-bot/consts"
	"github.com/davecgh/go-spew/spew"
	"github.com/dchest/uniuri"

	"github.com/Frosin/shoplist-api-client-go/client"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
)

//Shoplist api struct
type Shoplist struct {
	Client *client.ClientWithResponses
}

//NewShoplistAPI returns new Shoplist api instance
func NewShoplistAPI(url, token string) (*Shoplist, error) {
	shClient := &Shoplist{}

	apiKeyProvider, apiKeyProviderErr := securityprovider.NewSecurityProviderApiKey(
		"query",
		"accesstoken",
		token,
	)
	if apiKeyProviderErr != nil {
		return nil, apiKeyProviderErr
	}

	api, apiErr := client.NewClientWithResponses(
		url,
		client.WithRequestEditorFn(apiKeyProvider.Intercept),
	)

	if apiErr != nil {
		return nil, apiErr
	}

	shClient.Client = api
	return shClient, nil
}

func badStatusError(status string) error {
	return errors.New("bad status = " + status)
}

func (s *Shoplist) GetUserByTelegramID(telegramID int) (*client.UserWithID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), consts.ReadTimeout)
	defer cancel()

	telegramUserID := client.TelegramUserId(telegramID)

	response, err := s.Client.GetUsersWithResponse(
		ctx,
		&client.GetUsersParams{
			TelegramUserId: &telegramUserID,
		},
	)
	if err != nil {
		return nil, err
	}

	switch {
	case response.StatusCode() == 404:
		return nil, consts.ErrNotFound
	case response.StatusCode() != 200:
		return nil, badStatusError(response.Status())
	}
	data := *response.JSON200.Data
	return &data[0], nil
}

func (s *Shoplist) GetUsersByComunityID(comunityID string) ([]client.UserWithID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), consts.ReadTimeout)
	defer cancel()

	comunityUserID := client.ComunityId(comunityID)
	response, err := s.Client.GetUsersWithResponse(ctx, &client.GetUsersParams{
		ComunityId: &comunityUserID,
	})
	if err != nil {
		return nil, err
	}

	switch {
	case response.StatusCode() == 404:
		return nil, consts.ErrNotFound
	case response.StatusCode() != 200:
		return nil, badStatusError(response.Status())
	}
	//debug
	fmt.Println("response dump:")
	spew.Dump(response.JSON200)
	//
	data := response.JSON200.Data
	return *data, nil
}

func (s *Shoplist) CreateUser(userID int, chatID int64, username string) (*client.UserWithID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), consts.ReadTimeout)
	defer cancel()

	token := uniuri.New()
	comunityID := uniuri.New()
	intChatID := int64(chatID)

	response, err := s.Client.
		CreateUserWithResponse(ctx, client.CreateUserJSONRequestBody(
			client.UserRequest(
				client.User{
					Token:            &token,
					TelegramUsername: &username,
					TelegramId:       &userID,
					ComunityId:       &comunityID,
					ChatId:           &intChatID,
				},
			),
		),
		)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, badStatusError(response.Status())
	}
	data := *response.JSON200.Data
	return &data[0], nil
}

func (s *Shoplist) UpdateUser(userID int, comunityID, userName *string) error {
	ctx, cancel := context.WithTimeout(context.Background(), consts.ReadTimeout)
	defer cancel()

	params := &client.UpdateUserParams{
		UserId: client.UserId(userID),
	}
	response, err := s.Client.UpdateUserWithResponse(
		ctx,
		params,
		client.UpdateUserJSONRequestBody(
			client.UserRequest(
				client.User{
					ComunityId:       comunityID,
					TelegramUsername: userName,
				},
			),
		),
	)
	if err != nil {
		return err
	}

	switch {
	case response.StatusCode() == 404:
		return consts.ErrNotFound
	case response.StatusCode() != 200:
		return badStatusError(response.Status())
	}
	return nil
}

func (s *Shoplist) UserInit(telegramUserID int, chatID int64, userName string) (*client.UserWithID, error) {
	user := &client.UserWithID{}
	user, err := s.GetUserByTelegramID(telegramUserID)

	switch {
	case err == consts.ErrNotFound:
		user, err = s.CreateUser(telegramUserID, chatID, userName)
		if err != nil {
			return nil, err
		}
	case err != nil:
		return nil, err
	}

	return user, nil
}

//GetShoppingDays returns days with shoppings by date params
func (s *Shoplist) GetShoppingDays(time time.Time) ([]int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), consts.ReadTimeout)
	defer cancel()

	response, err := s.Client.GetShoppingDaysWithResponse(
		ctx,
		client.Year(time.Year()),
		client.Month(int(time.Month())),
	)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, badStatusError(response.Status())
	}
	data := response.JSON200.Data
	if data == nil {
		//return nil, errors.New("days data is nil")
		return []int{}, nil
	}
	return *data, nil
}

func (s *Shoplist) TestPing() error {
	ctx, cancel := context.WithTimeout(context.Background(), consts.ReadTimeout)
	defer cancel()

	response, err := s.Client.LastShoppingWithResponse(ctx)
	if err != nil {
		return err
	}

	if response.JSON200.Errors == nil {
		return nil
	}
	return errors.New("bad response")
}

func (s *Shoplist) GetShoppingsByDay(sDay time.Time) ([]client.ShoppingWithId, error) {
	ctx, cancel := context.WithTimeout(context.Background(), consts.ReadTimeout)
	defer cancel()

	year := client.Year(sDay.Year())
	month := client.Month(sDay.Month())
	day := client.Day(sDay.Day())

	response, err := s.Client.GetShoppingsByDayWithResponse(ctx, year, month, day)
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, badStatusError(response.Status())
	}
	data := response.JSON200.Data

	return data, nil
}

func (s *Shoplist) GetShoppingItems(shoppingID int) (*[]client.ShoppingItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), consts.ReadTimeout)
	defer cancel()

	response, err := s.Client.GetGoodsWithResponse(ctx, client.ShoppingID(shoppingID))
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		if response.StatusCode() == 404 {
			return nil, consts.ErrNotFound
		}
		return nil, badStatusError(response.Status())
	}
	data := response.JSON200.Data

	return &data, nil
}

func (s *Shoplist) GetShopping(ID int) (*client.ShoppingWithId, error) {
	ctx, cancel := context.WithTimeout(context.Background(), consts.ReadTimeout)
	defer cancel()

	response, err := s.Client.GetShoppingWithResponse(ctx, client.ShoppingID(ID))
	if err != nil {
		return nil, err
	}
	if response.StatusCode() != 200 {
		return nil, badStatusError(response.Status())
	}
	data := response.JSON200.Data
	return data, nil
}

func (s *Shoplist) GetSpecialShopping(sType consts.ShoppingType) (*int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), consts.ReadTimeout)
	defer cancel()

	response, err := s.Client.GetSpecialShoppingWithResponse(ctx, client.ShoppingType(sType))
	if err != nil {
		return nil, err
	}
	switch {
	case response.StatusCode() == 404:
		return nil, consts.ErrNotFound
	case response.StatusCode() != 200:
		return nil, badStatusError(response.Status())
	}

	data := response.JSON200.Data.Id
	return data, nil
}

func (s *Shoplist) AddItem(shoppingID int, itemName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), consts.WriteTimeout)
	defer cancel()

	request := client.AddItemJSONRequestBody(
		client.ItemRequest{
			ShoppingItemParams: client.ShoppingItemParams{
				CategoryID:  0,
				Complete:    false,
				ListID:      shoppingID,
				ProductName: itemName,
				Quantity:    1,
			},
		},
	)
	_, err := s.Client.AddItem(ctx, request)
	return err
}

func (s *Shoplist) RemoveItems(items []int) error {
	ctx, cancel := context.WithTimeout(context.Background(), consts.WriteTimeout)
	defer cancel()

	request := client.DeleteItemsJSONRequestBody{}
	request.Ids = items
	_, err := s.Client.DeleteItems(ctx, request)
	return err
}

func (s *Shoplist) AddShoppingWithType(day time.Time, shopName string, shoppingType consts.ShoppingType) (*int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), consts.WriteTimeout)
	defer cancel()

	shoppingData := client.ShoppingParams{
		Date:    day.Format(consts.DateLayout),
		Name:    shopName,
		OwnerID: 0, // shoplist api will get userID from token
		Time:    "12:00:00",
		Type:    int(shoppingType),
	}

	request := client.AddShoppingJSONRequestBody(
		client.ShoppingRequest{
			Shopping: client.Shopping{
				ShoppingParams: shoppingData,
			},
		},
	)

	response, err := s.Client.AddShoppingWithResponse(ctx, request)
	if response.StatusCode() != 200 {
		return nil, badStatusError(response.Status())
	}
	return response.JSON200.Data.Id, err
}

func (s *Shoplist) AddShopping(day time.Time, shopName string) error {
	_, err := s.AddShoppingWithType(day, shopName, consts.ShoppingTypeDefault)
	return err
}
