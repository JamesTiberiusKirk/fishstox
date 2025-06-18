package util

import (
	"encoding/json"

	"github.com/JamesTiberiusKirk/fishstox/internal/models"
)

func GenerateChartData(prices []models.StockPrice) string {
	var timestamps []int64
	var values []int

	for _, price := range prices {
		timestamps = append(timestamps, price.Timestamp)
		values = append(values, price.Value)
	}

	chartData := struct {
		Timestamps []int64 `json:"timestamps"`
		Values     []int   `json:"values"`
	}{
		Timestamps: timestamps,
		Values:     values,
	}

	// Convert chart data to JSON
	jsonData, _ := json.Marshal(chartData)
	return string(jsonData)
}
