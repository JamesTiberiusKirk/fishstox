package cacher

import (
	"context"
	"log/slog"
	"time"

	"github.com/JamesTiberiusKirk/fishstox/internal/db"
	"github.com/JamesTiberiusKirk/fishstox/internal/stox"
)

type Cacher struct {
	log *slog.Logger
	db  *db.Client
}

func NewCacher(log *slog.Logger, db *db.Client) *Cacher {
	return &Cacher{
		log: log,
		db:  db,
	}
}

func (c *Cacher) processPrices(data stox.PriceData) {
	for ticker, timeseries := range data.Prices {
		for ts, price := range timeseries {
			err := c.db.AddStockData(ticker, ts, price)
			if err != nil {
				c.log.Error("Error getting price data from stox", "error", err)
				continue
			}
		}
	}
}

func (c *Cacher) Scrape(ctx context.Context) {
	intervals := []stox.PriceInterval{stox.PriceIntervalMax, stox.PriceIntervalWeek,
		stox.PriceIntervalDay, stox.PriceIntervalHour}

	for {
		c.log.Info("Scraping interval")
		for _, i := range intervals {
			c.log.Info("Scraping", "interval", string(i))
			data, err := stox.GetPriceData(i)
			if err != nil {
				c.log.Error("Error getting price data from stox", "error", err)
				continue
			}

			c.processPrices(data)
		}

		c.log.Info("Done scraping interval")

		time.Sleep(10 * time.Minute)
	}
}

func (c *Cacher) CacheStoxData(ctx context.Context) {
	for {
		c.log.Info("Caching stox price data on hour interval")
		data, err := stox.GetPriceData(stox.PriceIntervalHour)
		if err != nil {
			c.log.Error("Error getting price data from stox", "error", err)
			continue
		}

		go c.processPrices(data)

		c.log.Info("Done caching stox price data")

		time.Sleep(1 * time.Minute)
	}
}
