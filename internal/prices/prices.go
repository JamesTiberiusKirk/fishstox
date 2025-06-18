package prices

import (
	"fmt"
	"time"

	"github.com/JamesTiberiusKirk/fishstox/internal/models"
)

// ConcatAndAverage processes the raw prices and averages them into specified intervals.
func ConcatAndAverage(prices []models.StockPrice, numPrices int, from, to time.Time) ([]models.StockPrice, error) {
	// Sanity check for numPrices
	if numPrices <= 0 {
		return nil, fmt.Errorf("numPrices must be greater than 0")
	}

	// Calculate the total duration in milliseconds
	duration := to.Sub(from).Milliseconds()

	// Calculate the time interval between each price in milliseconds
	interval := duration / int64(numPrices)

	var processedPrices []models.StockPrice
	var currentIntervalStart int64
	var priceSum, priceCount int

	// Iterate through the raw prices and group them by intervals
	for _, sp := range prices {
		ts := int64(sp.Timestamp)

		// If the timestamp is within the current interval, accumulate the price
		if ts >= currentIntervalStart && ts < currentIntervalStart+interval {
			priceSum += sp.Value
			priceCount++
		} else {
			// If the interval is complete, compute the average for the interval
			if priceCount > 0 {
				averagePrice := priceSum / priceCount
				processedPrices = append(processedPrices, models.StockPrice{
					Ticker:    sp.Ticker,
					Timestamp: currentIntervalStart + interval/2, // Midpoint of the interval
					Value:     averagePrice,
				})
			}

			// Reset for the new interval
			currentIntervalStart = ts - ts%interval
			priceSum = sp.Value
			priceCount = 1
		}

		// Stop if we've picked the required number of prices
		if len(processedPrices) == numPrices {
			break
		}
	}

	// If there are remaining prices after looping through all rows, compute the last average
	if priceCount > 0 && len(processedPrices) < numPrices {
		averagePrice := priceSum / priceCount
		processedPrices = append(processedPrices, models.StockPrice{
			Ticker:    "Last",
			Timestamp: currentIntervalStart + interval/2, // Midpoint of the last interval
			Value:     averagePrice,
		})
	}

	return processedPrices, nil
}
