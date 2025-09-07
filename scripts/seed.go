package main

import (
	"context"
	"log"

	"github.com/foozio/quran-go/internal/data"
	"github.com/foozio/quran-go/internal/db"
)

func main() {
	ctx := context.Background()
	d, err := db.Open("quran.db")
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Migrate(ctx, d); err != nil {
		log.Fatal(err)
	}
	if err := data.IngestAll(ctx, d, "id"); err != nil {
		log.Fatal(err)
	}
	log.Println("Done.")
}
