package prices

import (
	"fmt"

	"github.com/JamesTiberiusKirk/fishstox/internal/models"
)

// CalculateCandlestick calculates OHLC data from stock price data based on the given interval.
func CalculateCandlestick(prices []models.StockPrice, interval int) ([]models.Candle, error) {
	var candles []models.Candle
	var currentIntervalStart int64
	var high, low, open, close int
	var intervalPrices []models.StockPrice

	// Ensure we have enough data to calculate
	if len(prices) == 0 {
		return nil, fmt.Errorf("no prices provided")
	}

	for _, price := range prices {
		// If the timestamp falls within the current interval, accumulate it
		if price.Timestamp >= currentIntervalStart && price.Timestamp < currentIntervalStart+int64(interval) {
			intervalPrices = append(intervalPrices, price)
			if price.Value > high {
				high = price.Value
			}
			if price.Value < low || low == 0 {
				low = price.Value
			}
			close = price.Value // last price in the interval will be the close
		} else {
			// If the interval has changed, finalize the previous candlestick
			if len(intervalPrices) > 0 {
				open = intervalPrices[0].Value // first price in the interval will be the open
				candles = append(candles, models.Candle{
					Ticker:    intervalPrices[0].Ticker,
					Timestamp: currentIntervalStart + int64(interval)/2, // Mid-point of the interval
					Open:      open,
					Close:     close,
					High:      high,
					Low:       low,
				})
			}

			// Start a new interval
			currentIntervalStart = price.Timestamp - price.Timestamp%int64(interval)
			high, low, open, close = price.Value, price.Value, price.Value, price.Value
			intervalPrices = []models.StockPrice{price} // Start with the current price
		}
	}

	// Finalize the last candlestick if any remaining prices
	if len(intervalPrices) > 0 {
		open = intervalPrices[0].Value // first price in the interval will be the open
		candles = append(candles, models.Candle{
			Ticker:    intervalPrices[0].Ticker,
			Timestamp: currentIntervalStart + int64(interval)/2,
			Open:      open,
			Close:     close,
			High:      high,
			Low:       low,
		})
	}

	return candles, nil
}
