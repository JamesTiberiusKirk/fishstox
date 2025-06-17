package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/JamesTiberiusKirk/fishstox/internal/cacher"
	"github.com/JamesTiberiusKirk/fishstox/internal/config"
	"github.com/JamesTiberiusKirk/fishstox/internal/db"
	"github.com/JamesTiberiusKirk/fishstox/internal/middleware"
	"github.com/JamesTiberiusKirk/fishstox/internal/web/index"
	"github.com/alexedwards/scs/v2"
	"github.com/rickb777/servefiles/v3"
)

var Version = "devel"

func main() {
	config.Version = Version
	config := config.GetConfig()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Initialize the session.
	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	db, err := db.InitClient(logger,
		config.DbUser, config.DbPass, config.DbHost, config.DbName,
		true, time.Now)
	if err != nil {
		panic("error connecting to db " + err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	{
		c := cacher.NewCacher(logger, db)
		go c.CacheStoxData(ctx)
	}

	{
		serverMux := http.NewServeMux()
		serverMux.Handle("/{$}", index.NewHandler(sessionManager))
		assets := servefiles.NewAssetHandler("./assets/").WithMaxAge(time.Hour)
		serverMux.Handle("/assets/", http.StripPrefix("/assets/", assets))
		loggedServer := middleware.Logger(logger, serverMux)
		sessionedServer := sessionManager.LoadAndSave(loggedServer)

		port := os.Getenv("PORT")
		if port == "" {
			port = "3030"
		}

		logger.Info("HTTP server listening", "port", port)
		if err := http.ListenAndServe(":"+port, sessionedServer); err != nil {
			logger.Error("failed to start server: ", "error", err)
			return
		}

	}
}
