package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"plown/internal/config"
	"plown/internal/handler"
	"plown/internal/lobby"
	"plown/internal/repo"
	"plown/pkg/postgres"

	"github.com/go-chi/chi/v5"
	"github.com/lmittmann/tint"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	l := slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      slog.LevelInfo,
		AddSource:  true,
		TimeFormat: time.Kitchen,
	}))

	if err := run(ctx, l); err != nil {
		l.Error("app err:", "err", err.Error())
	}
}

func run(ctx context.Context, l *slog.Logger) error {
	cfg, err := config.Parse[config.ServerConfig]()
	if err != nil {
		return err
	}

	pool, err := postgres.New(ctx, postgres.Config{
		Addr:     cfg.PGAddr,
		DB:       cfg.PGDB,
		Username: cfg.PGUser,
		Password: cfg.PGPass,
	}, l)
	if err != nil {
		return err
	}
	defer pool.Close()

	pgRepo := repo.NewPostgres(pool)

	mgr := lobby.NewManager(l, pgRepo)

	handler := handler.New(l, pgRepo, mgr, cfg.StaticDir)
	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		r.Post("/lobby", handler.CreateLobby)
		r.Get("/lobby/{lobby_id}/ws", handler.LobbyWS)

		r.Get("/tags", handler.GetTags)
		r.Get("/labels", handler.GetLabels)
	})

	r.Get("/*", handler.Static)

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}

	go func() {
		<-ctx.Done()
		srv.Shutdown(context.Background())
	}()

	l.Info("server starting", "addr", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}
