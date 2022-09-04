package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"entgo.io/ent/dialect"
	"github.com/Frosin/shoplist-telegram-bot/internal/shoplist/ent"
	_ "github.com/mattn/go-sqlite3"
)

func main() {

	client, err := ent.Open(dialect.SQLite, fmt.Sprintf("file:%s?_fk=1", "./db/shoplist.db"))
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	shops, err := client.Shopping.
		Query().
		All(ctx)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("shops=", shops)
}
