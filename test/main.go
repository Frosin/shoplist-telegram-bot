package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"entgo.io/ent/dialect"
	"github.com/Frosin/shoplist-telegram-bot/internal/shoplist/ent"
	"github.com/Frosin/shoplist-telegram-bot/internal/shoplist/ent/shopping"
	_ "github.com/mattn/go-sqlite3"
)

func main() {

	client, err := ent.Open(dialect.SQLite, fmt.Sprintf("file:%s?_fk=1", "./db/shoplist.db"))
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	log.Println("maxint=", math.MaxInt32)

	//2147483647
	var id int = 8589934593 //8589934733

	shops, err := client.Shopping.
		Query().
		WithShop().
		Where(shopping.IDEQ(id)).
		All(ctx)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("shops=", shops)
}
