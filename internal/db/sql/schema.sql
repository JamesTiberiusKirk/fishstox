-- name: schema_up
CREATE TABLE tickers (
    ticker     VARCHAR(10)    NOT NULL,
    timestamp  BIGINT         NOT NULL,
    value      NUMERIC(10, 2) NOT NULL,

    PRIMARY KEY (ticker, timestamp)
);

<<<<<<< HEAD
CREATE INDEX idx_ticker_only ON tickers(ticker);
CREATE INDEX idx_timestamp_only ON tickers(timestamp);

-- name: schema_down
DROP TABLE IF EXISTS tickers;
=======
CREATE INDEX IF NOT EXISTS idx_users_updated_at ON users (updated_at);

CREATE TABLE IF NOT EXISTS test_table (
    id         UUID PRIMARY KEY,
    test_name   varchar(255),
);
>>>>>>> parent of aab4be2 (Removed test migration)
