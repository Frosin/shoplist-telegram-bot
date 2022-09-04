package shoplist

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Frosin/shoplist-telegram-bot/consts"
	"github.com/Frosin/shoplist-telegram-bot/internal/shoplist/ent"
	"github.com/Frosin/shoplist-telegram-bot/internal/shoplist/ent/item"
	"github.com/Frosin/shoplist-telegram-bot/internal/shoplist/ent/predicate"
	"github.com/Frosin/shoplist-telegram-bot/internal/shoplist/ent/shop"
	"github.com/Frosin/shoplist-telegram-bot/internal/shoplist/ent/shopping"
	"github.com/Frosin/shoplist-telegram-bot/internal/shoplist/ent/user"
	"github.com/dchest/uniuri"
	"github.com/labstack/gommon/log"

	entSql "entgo.io/ent/dialect/sql"
)

const (
	dateLayout = "2006-01-02"
	timeLayout = "15:04:05"

	shoppingTypeDefault     = 0
	shoppingTypeCheckList   = 1
	shoppingTypeCurrentList = 2
)

type Shoplist struct {
	token string
	ent   *ent.Client
}

//NewShoplistAPI returns new Shoplist api instance
func NewShoplistAPI(e *ent.Client, token string) *Shoplist {
	return &Shoplist{
		token: token,
		ent:   e,
	}
}

func (s *Shoplist) GetUserByTelegramID(telegramID int) (*ent.User, error) {

	log.Info("METHOD GetUserByTelegramID")

	ctx, cancel := context.WithTimeout(context.Background(), consts.ReadTimeout)
	defer cancel()

	user, err := s.ent.User.
		Query().
		Where(user.TelegramIDEQ(int64(telegramID))).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetUserByTelegramID error: %w", err)
	}

	log.Info("GetUserByTelegramID", user)

	return user, nil
}

func (s *Shoplist) GetUsersByComunityID(comunityID string) ([]*ent.User, error) {
	log.Info("METHOD GetUsersByComunityID")

	ctx, cancel := context.WithTimeout(context.Background(), consts.ReadTimeout)
	defer cancel()

	users, err := s.ent.User.
		Query().
		Where(user.ComunityIDEQ(comunityID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetUsersByComunityID error: %w", err)
	}

	log.Info("GetUsersByComunityID", users)

	return users, nil
}

func (s *Shoplist) CreateUser(userID int, chatID int64, username string) (*ent.User, error) {
	log.Info("METHOD CreateUser")

	ctx, cancel := context.WithTimeout(context.Background(), consts.ReadTimeout)
	defer cancel()

	token := uniuri.New()
	comunityID := uniuri.New()
	intChatID := int64(chatID)

	user, err := s.ent.User.
		Create().
		SetChatID(int64(intChatID)).
		SetComunityID(comunityID).
		SetTelegramID(int64(userID)).
		SetTelegramUsername(username).
		SetToken(token).Save(ctx)

	if err != nil {
		return nil, fmt.Errorf("CreateUser error: %w", err)
	}

	log.Info("CreateUser", user)

	return user, nil
}

func (s *Shoplist) UpdateUser(userID int, comunityID, userName *string) error {

	log.Info("METHOD UpdateUser")

	ctx, cancel := context.WithTimeout(context.Background(), consts.ReadTimeout)
	defer cancel()

	if comunityID != nil {
		_, err := s.ent.User.
			UpdateOneID(userID).
			SetComunityID(*comunityID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("UpdateUser error: %w", err)
		}
	}

	if userName != nil {
		_, err := s.ent.User.
			UpdateOneID(userID).
			SetTelegramUsername(*userName).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("UpdateUser error: %w", err)
		}
	}

	return nil
}

func (s *Shoplist) UserInit(telegramUserID int, chatID int64, userName string) (*ent.User, error) {

	log.Info("METHOD UserInit")
	user, err := s.GetUserByTelegramID(telegramUserID)

	switch {
	case ent.IsNotFound(err):
		user, err = s.CreateUser(telegramUserID, chatID, userName)
		if err != nil {
			return nil, fmt.Errorf("UserInit error: %w", err)
		}
	case err != nil:
		return nil, fmt.Errorf("UserInit error: %w", err)
	}

	return user, nil
}

func (s *Shoplist) getCommunityUsers() (int, []int, error) {
	log.Info("METHOD getCommunityUsers")

	ctx, cancel := context.WithTimeout(context.Background(), consts.ReadTimeout)
	defer cancel()
	// get user by token
	usr, err := s.ent.User.
		Query().
		Where(user.TokenEQ(s.token)).
		Only(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("getCommunityUsers getUser error: %w", err)
	}

	log.Infof("userID=%v, userTelegramID=%v, userToken=%v, userComunityID=%v", usr.ID, usr.TelegramID, usr.Token, usr.ComunityID)

	comunityUsers, err := s.ent.User.
		Query().
		Where(user.ComunityIDEQ(usr.ComunityID)).
		All(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("getCommunityUsers get comunityUsers error: %w", err)
	}
	comUserIDs := []int{}
	for _, v := range comunityUsers {
		comUserIDs = append(comUserIDs, v.ID)
	}

	return usr.ID, comUserIDs, nil
}

//GetShoppingDays returns days with shoppings by date params
func (s *Shoplist) GetShoppingDays(time time.Time) ([]int, error) {
	log.Info("METHOD GetShoppingDays")

	ctx, cancel := context.WithTimeout(context.Background(), consts.ReadTimeout)
	defer cancel()

	strMonth := strconv.Itoa(int(time.Month()))
	if time.Month() < 10 {
		strMonth = "0" + strMonth
	}
	queryParam := fmt.Sprintf("%v-%s%%", time.Year(), strMonth)

	_, comUserIDs, err := s.getCommunityUsers()
	if err != nil {
		return nil, fmt.Errorf("GetShoppingDays: %w", err)
	}

	monthShoppings, err := s.ent.Shopping.
		Query().
		Where(
			shopping.HasUserWith(
				user.IDIn(comUserIDs...),
			),
			shopping.TypeEQ(shoppingTypeDefault),
			predicate.Shopping(func(s *entSql.Selector) {
				s.Where(entSql.Like(s.C(shopping.FieldDate), queryParam))
			})).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("GetShoppingDays get monthShoppings error: %w", err)
	}

	log.Info("GetShoppingDays", monthShoppings)

	var result []int
	for _, v := range monthShoppings {
		result = append(result, v.Date.Day())
	}

	return result, nil
}

func (s *Shoplist) GetShoppingsByDay(sDay time.Time) ([]*ent.Shopping, error) {
	log.Info("METHOD GetShoppingsByDay")

	ctx, cancel := context.WithTimeout(context.Background(), consts.ReadTimeout)
	defer cancel()

	_, comUserIDs, err := s.getCommunityUsers()
	if err != nil {
		return nil, fmt.Errorf("GetShoppingsByDay: %w", err)
	}

	strMonth := strconv.Itoa(int(sDay.Month()))
	if sDay.Month() < 10 {
		strMonth = "0" + strMonth
	}

	strDay := strconv.Itoa(int(sDay.Day()))
	if sDay.Day() < 10 {
		strDay = "0" + strDay
	}

	queryParam := fmt.Sprintf("%v-%s-%s%%", sDay.Year(), strMonth, strDay)

	shoppings, err := s.ent.Shopping.
		Query().
		WithShop().
		WithUser().
		Where(
			shopping.HasUserWith(
				user.IDIn(comUserIDs...),
			),
			shopping.TypeEQ(shoppingTypeDefault),
			predicate.Shopping(func(s *entSql.Selector) {
				s.Where(entSql.Like(s.C(shopping.FieldDate), queryParam))
			})).
		All(ctx)

	if err != nil {
		return nil, fmt.Errorf("GetShoppingsByDay get shoppings error: %w", err)
	}

	log.Info("GetShoppingsByDay", shoppings)

	return shoppings, nil
}

func (s *Shoplist) GetShoppingItems(shoppingID int) ([]*ent.Item, error) {
	log.Info("METHOD GetShoppingItems")

	ctx, cancel := context.WithTimeout(context.Background(), consts.ReadTimeout)
	defer cancel()

	goods, err := s.ent.Item.
		Query().
		WithShopping().
		Where(item.HasShoppingWith(shopping.IDEQ(int(shoppingID)))).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetShoppingItems error: %w", err)
	}

	log.Info("GetShoppingItems", goods)

	return goods, nil
}

func (s *Shoplist) GetShopping(ID int) (*ent.Shopping, error) {
	log.Info("METHOD GetShopping")

	ctx, cancel := context.WithTimeout(context.Background(), consts.ReadTimeout)
	defer cancel()

	_, comUserIDs, err := s.getCommunityUsers()
	if err != nil {
		return nil, fmt.Errorf("GetShopping: %w", err)
	}

	shopping, err := s.ent.Shopping.
		Query().
		WithShop().
		WithUser().
		Where(
			shopping.IDEQ(ID),
			shopping.HasUserWith(
				user.IDIn(comUserIDs...),
			)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetShopping: %w", err)
	}

	log.Info("GetShopping", shopping)

	return shopping, nil
}

func (s *Shoplist) GetSpecialShopping(sType consts.ShoppingType) (int, error) {
	log.Info("METHOD GetSpecialShopping")

	ctx, cancel := context.WithTimeout(context.Background(), consts.ReadTimeout)
	defer cancel()

	ownerID, _, err := s.getCommunityUsers()
	if err != nil {
		return 0, fmt.Errorf("GetSpecialShopping: %w", err)
	}

	shopping, err := s.ent.Shopping.
		Query().
		WithShop().
		WithUser().
		Where(
			shopping.TypeEQ(int(sType)),
			shopping.HasUserWith(
				user.IDEQ(ownerID),
			),
		).
		Only(ctx)

	if err != nil {
		return 0, fmt.Errorf("GetSpecialShopping: %w", err)
	}

	log.Info("GetSpecialShopping", shopping)

	return shopping.ID, nil
}

func (s *Shoplist) AddItem(shoppingID int, itemName string) error {
	log.Info("METHOD AddItem")

	ctx, cancel := context.WithTimeout(context.Background(), consts.WriteTimeout)
	defer cancel()

	shopping, err := s.ent.Shopping.
		Query().
		Where(shopping.IDEQ(shoppingID)).
		Only(ctx)

	if err != nil {
		return fmt.Errorf("AddItem getShopping: %w", err)
	}

	err = WithTx(ctx, s.ent, func(tx *ent.Tx) error {
		_, err := tx.Item.
			Create().
			SetProductName(itemName).
			SetShopping(shopping).
			Save(ctx)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("AddItem withTx: %w", err)
	}

	return nil
}

func (s *Shoplist) RemoveItems(items []int) error {
	log.Info("METHOD RemoveItems")

	ctx, cancel := context.WithTimeout(context.Background(), consts.WriteTimeout)
	defer cancel()

	_, err := s.ent.Item.Delete().Where(item.IDIn(items...)).Exec(ctx)
	if err != nil {
		return fmt.Errorf("RemoveItems: %w", err)
	}

	return nil
}

func (s *Shoplist) AddShoppingWithType(day time.Time, shopName string, shoppingType consts.ShoppingType) (int, error) {
	log.Info("METHOD AddShoppingWithType")

	ctx, cancel := context.WithTimeout(context.Background(), consts.WriteTimeout)
	defer cancel()

	shp, err := s.ent.Shop.
		Query().
		Where(shop.NameEQ(shopName)).
		First(ctx)

	if err != nil {
		switch {
		case ent.IsNotFound(err):
			shp, err = s.ent.Shop.
				Create().
				SetName(shopName).
				Save(ctx)
			if err != nil {
				return 0, fmt.Errorf("AddShoppingWithType create: %w", err)
			}
		default:
			return 0, fmt.Errorf("AddShoppingWithType get shop: %w", err)
		}
	}

	log.Info("AddShoppingWithType", shp)

	ownerID, _, err := s.getCommunityUsers()
	if err != nil {
		return 0, fmt.Errorf("AddShoppingWithType: %w", err)
	}

	var newShopping *ent.Shopping
	err = WithTx(ctx, s.ent, func(tx *ent.Tx) error {
		newShopping, err = tx.Shopping.
			Create().
			SetShop(shp).
			SetDate(day).
			SetUserID(ownerID).
			SetType(int(shoppingType)).
			Save(ctx)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("AddShoppingWithType withTx: %w", err)
	}

	log.Info("AddShoppingWithType", newShopping)

	return newShopping.ID, nil
}

func (s *Shoplist) AddShopping(day time.Time, shopName string) error {
	log.Info("METHOD AddShopping")

	_, err := s.AddShoppingWithType(day, shopName, consts.ShoppingTypeDefault)
	return err
}

func WithTx(ctx context.Context, client *ent.Client, fn func(tx *ent.Tx) error) error {
	tx, err := client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()
	if err := fn(tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("rolling back transaction: %w", rerr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}
