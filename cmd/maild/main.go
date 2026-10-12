package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/buildwithtechx/codedock-mail/internal/api"
	"github.com/buildwithtechx/codedock-mail/internal/provisioner"
	"github.com/buildwithtechx/codedock-mail/internal/store"
)

func main() {
	ctx := context.Background()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://mail:mail@localhost:5432/mail?sslmode=disable"
	}
	st, err := store.Connect(ctx, dsn)
	if err != nil {
		slog.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		slog.Error("run migrations", "error", err)
		os.Exit(1)
	}
	var engine provisioner.Engine
	if stalwartURL := os.Getenv("STALWART_URL"); stalwartURL != "" {
		engine = provisioner.NewStalwart(stalwartURL, envOr("STALWART_USER", "admin"), os.Getenv("STALWART_SECRET"))
		slog.Info("engine configured", "url", stalwartURL)
	} else {
		slog.Info("engine not configured, running local-only")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	srv := api.New(st, engine)
	if err := srv.Start(":" + port); err != nil {
		slog.Error("start server", "error", err)
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
