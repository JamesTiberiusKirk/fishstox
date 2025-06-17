package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/JamesTiberiusKirk/fishstox/internal/cacher"
	"github.com/JamesTiberiusKirk/fishstox/internal/config"
	"github.com/JamesTiberiusKirk/fishstox/internal/db"
)

func main() {
	config := config.GetConfig()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, err := db.InitClient(logger,
		config.DbUser, config.DbPass, config.DbHost, config.DbName,
		true, time.Now)
	if err != nil {
		panic("error connecting to db " + err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := cacher.NewCacher(logger, db)
	c.Scrape(ctx)
}
