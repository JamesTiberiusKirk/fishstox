package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"

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

// StockPrice represents a row from stock_data.
type StockPrice struct {
	Ticker    string
	Timestamp string // or time.Time if your DB column is TIMESTAMP
	Value     int
}

// GetStockPricesByTimeFrame returns prices for a ticker at the specified time frame.
func (c *Client) GetStockPricesByTimeFrame(
	ticker string,
	from, to time.Time,
	timeFrame string,
) ([]StockPrice, error) {
	var selectCols []string
	var groupBy []string

	switch timeFrame {
	case "hourly":
		selectCols = []string{"ticker", "timestamp", "value"}
		// no group by
	case "daily":
		selectCols = []string{"ticker", "date_trunc('hour', timestamp) as timestamp", "FIRST(value) as value"}
		groupBy = []string{"ticker", "date_trunc('hour', timestamp)"}
	case "weekly", "max":
		selectCols = []string{"ticker", "date_trunc('day', timestamp) as timestamp", "FIRST(value) as value"}
		groupBy = []string{"ticker", "date_trunc('day', timestamp)"}
	default:
		return nil, fmt.Errorf("invalid timeFrame: %s", timeFrame)
	}

	sb := c.sq.Select(selectCols...).From("tickers").
		Where(squirrel.Eq{"ticker": ticker}).
		Where(squirrel.GtOrEq{"timestamp": from}).
		Where(squirrel.LtOrEq{"timestamp": to})

	if len(groupBy) > 0 {
		sb = sb.GroupBy(groupBy...).OrderBy(groupBy[1] + " ASC")
	} else {
		sb = sb.OrderBy("timestamp ASC")
	}

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

	var prices []StockPrice
	for rows.Next() {
		var sp StockPrice
		if err := rows.Scan(&sp.Ticker, &sp.Timestamp, &sp.Value); err != nil {
			c.log.Error("failed to scan row", slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to scan stock data: %w", err)
		}
		prices = append(prices, sp)
	}
	if err := rows.Err(); err != nil {
		c.log.Error("row iteration error", slog.String("error", err.Error()))
		return nil, err
	}

	return prices, nil
}
