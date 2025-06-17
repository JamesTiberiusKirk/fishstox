package stox

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	stoxPricesEndpoint      = "https://api.fishtank.live/v1/stocks/prices?range="
	stoxLeaderBoardEndpoint = "https://api.fishtank.live/v1/stocks/leader-board"
	stoxStocksEndpoint      = "https://api.fishtank.live/v1/stocks"
)

type PriceData struct {
	//		   ticker: timestamp:price
	Prices map[string]map[string]int `json:"prices"`
}

type PriceInterval string

const (
	PriceIntervalMax  PriceInterval = "max"
	PriceIntervalWeek PriceInterval = "week"
	PriceIntervalDay  PriceInterval = "day"
	PriceIntervalHour PriceInterval = "hour"
)

func GetPriceData(interval PriceInterval) (PriceData, error) {
	var data PriceData
	resp, err := http.Get(stoxPricesEndpoint + string(interval))
	if err != nil {
		return data, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return data, fmt.Errorf("bad status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return data, fmt.Errorf("failed to read response body: %w", err)
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return data, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return data, nil
}

type Stock struct {
	AveragePrice     int    `json:"averagePrice"`
	CurrentPrice     int    `json:"currentPrice"`
	HighestBuyOrder  int    `json:"highestBuyOrder"`
	HighestSellOrder int    `json:"highestSellOrder"`
	IpoAvailable     bool   `json:"ipoAvailable"`
	IpoPrice         int    `json:"ipoPrice"`
	IpoSharesLeft    int    `json:"ipoSharesLeft"`
	LastHour         int    `json:"lastHour"`
	LastWeek         int    `json:"lastWeek"`
	LowestBuyOrder   int    `json:"lowestBuyOrder"`
	LowestSellOrder  int    `json:"lowestSellOrder"`
	MyHoldings       int    `json:"myHoldings"`
	TickerSymbol     string `json:"tickerSymbol"`
	Today            int    `json:"today"`
	TotalInvestment  int    `json:"totalInvestment"`
	TotalShares      int    `json:"totalShares"`
}

type StocksResponse struct {
	// Orders []any   `json:"orders" // Need to be logged in for this i think
	Stocks []Stock `json:"stocks"`
}

func GetStocks() (*StocksResponse, error) {
	resp, err := http.Get(stoxStocksEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stocks: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var stocksResp StocksResponse
	if err := json.Unmarshal(body, &stocksResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &stocksResp, nil
}

type Clan struct {
	Tag    string `json:"tag"`
	Rank   int    `json:"rank"`
	Color  string `json:"color"`
	Emblem string `json:"emblem"`
}

type Profile struct {
	ID                     string         `json:"id"`
	DisplayName            string         `json:"displayName"`
	Color                  string         `json:"color"`
	Photo                  string         `json:"photo"`
	SeasonPass             bool           `json:"seasonPass"`
	SeasonPassXL           bool           `json:"seasonPassXL"`
	SeasonPassSubscription bool           `json:"seasonPassSubscription"`
	SeasonPassGift         *string        `json:"seasonPassGift"`
	XP                     int            `json:"xp"`
	Clan                   Clan           `json:"clan"`
	Joined                 int64          `json:"joined"`
	Pfps                   []string       `json:"pfps"`
	Medals                 map[string]int `json:"medals"`
	Tokens                 int            `json:"tokens"`
	Bio                    string         `json:"bio"`
	Endorsement            *string        `json:"endorsement"`
	Integrations           []string       `json:"integrations"`
}

type PortfolioValue struct {
	UserID         string  `json:"userId"`
	PortfolioValue int     `json:"portfolioValue"`
	Profile        Profile `json:"profile"`
}

type PortfolioValuesResponse struct {
	PortfolioValues []PortfolioValue `json:"portfolioValues"`
}

func GetPortfolioValues() (*PortfolioValuesResponse, error) {
	resp, err := http.Get(stoxLeaderBoardEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch portfolio values: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var pvr PortfolioValuesResponse
	if err := json.Unmarshal(body, &pvr); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &pvr, nil
}
