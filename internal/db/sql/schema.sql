-- name: schema_up
CREATE TABLE tickers (
    ticker     VARCHAR(10)    NOT NULL,
    timestamp  BIGINT         NOT NULL,
    value      INTEGER        NOT NULL,

    PRIMARY KEY (ticker, timestamp)
);

CREATE INDEX idx_ticker_only ON tickers(ticker);
CREATE INDEX idx_timestamp_only ON tickers(timestamp);

-- name: schema_down
DROP TABLE IF EXISTS tickers;
