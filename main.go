package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/titagaki/pcgw-0yp/internal/config"
	"github.com/titagaki/pcgw-0yp/internal/db"
	"github.com/titagaki/pcgw-0yp/internal/handler"
	"github.com/titagaki/pcgw-0yp/internal/middleware"
	"github.com/titagaki/pcgw-0yp/internal/server"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	godotenv.Load() // .env があれば読み込む（なくてもエラーにしない）

	configPath := "config.toml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	database, err := db.Open(db.Config{
		User:   cfg.DB.User,
		Passwd: cfg.DB.Passwd,
		Host:   cfg.DB.Host,
		Port:   cfg.DB.Port,
		DBName: cfg.DB.DBName,
	})
	if err != nil {
		log.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	store := sessions.NewCookieStore([]byte(cfg.Server.SessionSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60, // 30 days
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	h := handler.New(database, cfg, log, "templates")
	h.StartCleanup()

	sessionMw := middleware.Session(store)
	authMw := middleware.Auth(database)

	router := server.NewRouter(h, sessionMw, authMw)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Info("starting server", "addr", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Error("server error", "error", err)
		os.Exit(1)
	}
}
