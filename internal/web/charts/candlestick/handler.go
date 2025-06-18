package candlestick

import (
	"fmt"
	"net/http"
	"time"

	"github.com/JamesTiberiusKirk/fishstox/internal/components"
	"github.com/JamesTiberiusKirk/fishstox/internal/db"
	"github.com/JamesTiberiusKirk/fishstox/internal/prices"
	"github.com/JamesTiberiusKirk/fishstox/internal/slogctx"
)

func NewHandler(db *db.Client) http.Handler {
	return &handler{
		db: db,
	}
}

type handler struct {
	db *db.Client
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.get(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func (h *handler) get(w http.ResponseWriter, r *http.Request) {
	tickerQuery := r.PathValue("tickerQuery")
	if tickerQuery == "" {
		w.WriteHeader(http.StatusNotFound)
		components.NotFound(r, "Ticker not found").Render(r.Context(), w)
		return
	}

	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()

	rawPrices, err := h.db.GetStockPricesByTimeFrame(tickerQuery, from, to)
	if err != nil {
		slogctx.Ctx(r.Context()).Error("Error getting prices", "ticker", tickerQuery, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		components.ServerError(r, err.Error()).Render(r.Context(), w)
		return
	}

	interval := 3600000 // 1 hour in milliseconds
	candles, err := prices.CalculateCandlestick(rawPrices, interval)
	if err != nil {
		slogctx.Ctx(r.Context()).Error("Error processing prices", "ticker", tickerQuery, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		components.ServerError(r, err.Error()).Render(r.Context(), w)
		return
	}

	fmt.Println(len(candles))

	pageData := pageProps{
		tickerQuery: tickerQuery,
		candles:     candles,
	}

	w.WriteHeader(http.StatusOK)
	page(r, pageData).Render(r.Context(), w)
}
