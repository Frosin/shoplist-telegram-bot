package cache

import (
	"time"

	"github.com/Frosin/shoplist-api-client-go/client"
)

//Cache contains day->shoppingID->items
type Cache struct {
	days map[time.Time]map[int]*[]client.ShoppingItem
}

func New(token string) {

}
