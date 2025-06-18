package models

// StockPrice represents a row from stock_data.
type StockPrice struct {
	Ticker    string
	Timestamp int64
	Value     int
}

// {
//  x: date.valueOf(),
//  o: open,
//  h: high,
//  l: low,
//  c: close
// }

type Candle struct {
	Ticker    string `json:"ticker"`
	Timestamp int64  `json:"x"`
	Open      int    `json:"o"`
	Close     int    `json:"c"`
	High      int    `json:"h"`
	Low       int    `json:"l"`
}
