package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"

	"github.com/JamesTiberiusKirk/fishstox/internal/models"
	"github.com/JamesTiberiusKirk/migrator/migrator"
)

type Client struct {
	log     *slog.Logger
	connUrl string
	db      *sql.DB
	sq      squirrel.StatementBuilderType
	now     func() time.Time
}

// InitClient initializes a new database client and pings the DB.
func InitClient(
	log *slog.Logger,
	user, pass, host, dbName string,
	disableSSL bool,
	now func() time.Time,
) (*Client, error) {
	connUrl := fmt.Sprintf("postgres://%s:%s@%s/%s", user, pass, host, dbName)
	if disableSSL {
		connUrl += "?sslmode=disable"
	}

	db, err := sql.Open("postgres", connUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping the database: %w", err)
	}

	migrate, err := migrator.NewMigratorWithSqlClient(db, "./internal/db/sql/")
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator instance: %w", err)
	}

	err = migrate.ApplySchemaUp()
	if err != nil && !errors.Is(err, migrator.ErrSchemaAlreadyInitialised) {
		return nil, fmt.Errorf("failed to apply schema up: %w", err)
	}

	err = migrate.ApplyMigration()
	if err != nil {
		return nil, fmt.Errorf("failed to apply schema up: %w", err)
	}

	return &Client{
		log:     log,
		connUrl: connUrl,
		db:      db,
		sq:      squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		now:     now,
	}, nil
}

// AddStockData adds stock data for a specific ticker into the stock_data table.
func (c *Client) AddStockData(ticker string, timestamp string, value int) error {
	// Build the insert query using squirrel
	query := c.sq.Insert("tickers").
		Columns("ticker", "timestamp", "value").
		Values(ticker, timestamp, value).
		Suffix("ON CONFLICT (ticker, timestamp) DO NOTHING") // To prevent duplicates on the same timestamp for a ticker.

	// Execute the query
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		c.log.Error("failed to build SQL query", slog.Any("ticker", ticker), slog.Any("timestamp", timestamp), slog.String("error", err.Error()))
		return fmt.Errorf("failed to build SQL query: %w", err)
	}

	// Run the query
	_, err = c.db.Exec(sqlQuery, args...)
	if err != nil {
		c.log.Error("failed to execute SQL query", slog.Any("ticker", ticker), slog.Any("timestamp", timestamp), slog.String("error", err.Error()))
		return fmt.Errorf("failed to insert stock data: %w", err)
	}

	c.log.Info("added stock data", slog.Any("ticker", ticker), slog.Any("timestamp", timestamp), slog.Int("value", value))
	return nil
}

type TimeFrame int

const (
	TimeFrameHourly TimeFrame = iota
	TimeFrameDaily
	TimeFrameWeekly
	TimeFrameMax
)

// GetStockPricesByTimeFrame returns prices for a ticker at the specified time frame, evenly spaced.
func (c *Client) GetStockPricesByTimeFrameOld(
	ticker string,
	from, to time.Time,
	numPrices int, // Number of prices to return
) ([]models.StockPrice, error) {
	// Sanity check for numPrices
	if numPrices <= 0 {
		return nil, fmt.Errorf("numPrices must be greater than 0")
	}

	// Calculate the total duration in milliseconds
	duration := to.Sub(from).Milliseconds()

	// Calculate the time interval between each price in milliseconds
	interval := duration / int64(numPrices-1)

	selectCols := []string{"ticker", "timestamp", "value"}

	// We will select prices based on the time intervals
	sb := c.sq.Select(selectCols...).From("tickers").
		Where(squirrel.Eq{"ticker": ticker}).
		Where("timestamp BETWEEN ? AND ?", from.UnixNano()/int64(time.Millisecond), to.UnixNano()/int64(time.Millisecond)).
		OrderBy("timestamp ASC") // Order by timestamp to ensure chronological order

	sqlQuery, args, err := sb.ToSql()
	if err != nil {
		c.log.Error("failed to build SQL query", slog.Any("ticker", ticker), slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	rows, err := c.db.Query(sqlQuery, args...)
	if err != nil {
		c.log.Error("failed to execute SQL query", slog.Any("ticker", ticker), slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to query stock data: %w", err)
	}
	defer rows.Close()

	var prices []models.StockPrice
	var lastTimestamp int64

	// Use this to track the number of records selected
	count := 0

	// Scan rows and pick prices based on intervals
	for rows.Next() {
		var sp models.StockPrice
		var val string
		if err := rows.Scan(&sp.Ticker, &sp.Timestamp, &val); err != nil {
			c.log.Error("failed to scan row", slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to scan stock data: %w", err)
		}

		// Parse the value from string
		value, err := strconv.Atoi(val)
		if err != nil {
			c.log.Error("failed to parse value", slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to parse stock value: %w", err)
		}
		sp.Value = value

		// Convert sp.Timestamp to int64 for comparison and assignments
		ts := int64(sp.Timestamp)

		// Pick the first row
		if count == 0 {
			prices = append(prices, sp)
			lastTimestamp = ts
			count++
			continue
		}

		// Check if the current timestamp is within the next interval
		if ts >= lastTimestamp+interval {
			prices = append(prices, sp)
			lastTimestamp = ts
			count++
		}

		// Stop if we've picked the required number of prices
		if count == numPrices {
			break
		}
	}

	// Check for row iteration errors
	if err := rows.Err(); err != nil {
		c.log.Error("row iteration error", slog.String("error", err.Error()))
		return nil, err
	}

	// Return the selected prices
	return prices, nil
}

// GetStockPricesByTimeFrameAveraged returns averaged prices for a ticker at the specified time frame.
func (c *Client) GetStockPricesByTimeFrameAveraged(
	ticker string,
	from, to time.Time,
	numPrices int, // Number of prices to return
) ([]models.StockPrice, error) {
	// Sanity check for numPrices
	if numPrices <= 0 {
		return nil, fmt.Errorf("numPrices must be greater than 0")
	}

	// Calculate the total duration in milliseconds
	duration := to.Sub(from).Milliseconds()

	// Calculate the time interval between each price in milliseconds
	interval := duration / int64(numPrices)

	selectCols := []string{"ticker", "timestamp", "value"}

	// We will select prices based on the time intervals
	sb := c.sq.Select(selectCols...).From("tickers").
		Where(squirrel.Eq{"ticker": ticker}).
		Where("timestamp BETWEEN ? AND ?", from.UnixNano()/int64(time.Millisecond), to.UnixNano()/int64(time.Millisecond)).
		OrderBy("timestamp ASC") // Order by timestamp to ensure chronological order

	sqlQuery, args, err := sb.ToSql()
	if err != nil {
		c.log.Error("failed to build SQL query", slog.Any("ticker", ticker), slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	rows, err := c.db.Query(sqlQuery, args...)
	if err != nil {
		c.log.Error("failed to execute SQL query", slog.Any("ticker", ticker), slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to query stock data: %w", err)
	}
	defer rows.Close()

	var prices []models.StockPrice
	var currentIntervalStart int64
	var priceSum, priceCount int

	// Iterate through the rows and group by interval
	for rows.Next() {
		var sp models.StockPrice
		var val string
		if err := rows.Scan(&sp.Ticker, &sp.Timestamp, &val); err != nil {
			c.log.Error("failed to scan row", slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to scan stock data: %w", err)
		}

		// Parse the value from string
		value, err := strconv.Atoi(val)
		if err != nil {
			c.log.Error("failed to parse value", slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to parse stock value: %w", err)
		}
		sp.Value = value

		// Convert sp.Timestamp to int64 for comparisons
		ts := int64(sp.Timestamp)

		// Check if the timestamp falls within the current interval
		if ts >= currentIntervalStart && ts < currentIntervalStart+interval {
			priceSum += sp.Value
			priceCount++
		} else {
			// If we have passed the interval, save the average for the previous interval
			if priceCount > 0 {
				averagePrice := priceSum / priceCount
				prices = append(prices, models.StockPrice{
					Ticker:    sp.Ticker,
					Timestamp: currentIntervalStart + interval/2, // Ensure this is int64
					Value:     averagePrice,
				})
			}

			// Reset for the new interval
			currentIntervalStart = ts - ts%interval
			priceSum = sp.Value
			priceCount = 1
		}

		// Stop if we've picked the required number of prices
		if len(prices) == numPrices {
			break
		}
	}

	// Check for row iteration errors
	if err := rows.Err(); err != nil {
		c.log.Error("row iteration error", slog.String("error", err.Error()))
		return nil, err
	}

	// If there are remaining prices after looping through all rows
	if priceCount > 0 && len(prices) < numPrices {
		averagePrice := priceSum / priceCount
		prices = append(prices, models.StockPrice{
			Ticker:    "Last",
			Timestamp: currentIntervalStart + interval/2, // Ensure this is int64
			Value:     averagePrice,
		})
	}

	// Return the selected prices
	return prices, nil
}

// GetStockPricesByTimeFrame retrieves all prices for a ticker between the given time range.
func (c *Client) GetStockPricesByTimeFrame(
	ticker string,
	from, to time.Time,
) ([]models.StockPrice, error) {
	selectCols := []string{"ticker", "timestamp", "value"}

	sb := c.sq.Select(selectCols...).From("tickers").
		Where(squirrel.Eq{"ticker": ticker}).
		Where("timestamp BETWEEN ? AND ?", from.UnixNano()/int64(time.Millisecond), to.UnixNano()/int64(time.Millisecond)).
		OrderBy("timestamp ASC") // Order by timestamp to ensure chronological order

	sqlQuery, args, err := sb.ToSql()
	if err != nil {
		c.log.Error("failed to build SQL query", slog.Any("ticker", ticker), slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	rows, err := c.db.Query(sqlQuery, args...)
	if err != nil {
		c.log.Error("failed to execute SQL query", slog.Any("ticker", ticker), slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to query stock data: %w", err)
	}
	defer rows.Close()

	var prices []models.StockPrice

	for rows.Next() {
		var sp models.StockPrice
		var val string
		if err := rows.Scan(&sp.Ticker, &sp.Timestamp, &val); err != nil {
			c.log.Error("failed to scan row", slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to scan stock data: %w", err)
		}

		// Parse the value from string
		value, err := strconv.Atoi(val)
		if err != nil {
			c.log.Error("failed to parse value", slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to parse stock value: %w", err)
		}
		sp.Value = value

		prices = append(prices, sp)
	}

	if err := rows.Err(); err != nil {
		c.log.Error("row iteration error", slog.String("error", err.Error()))
		return nil, err
	}

	return prices, nil
}
