package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Addr     string
	DB       string
	Username string
	Password string
}

func New(ctx context.Context, cfg Config, l *slog.Logger) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", cfg.Username, cfg.Password, cfg.Addr, cfg.DB))
	if err != nil {
		return nil, err
	}

	config.ConnConfig.Tracer = &tracer{l: l}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

type traceKey string

var nameKey traceKey = "name"
var dataKey traceKey = "data"

func WithQueryName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, nameKey, name)
}

type traceVal struct {
	start time.Time
	name  string
}

type tracer struct {
	l *slog.Logger
}

func (t *tracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	name, ok := ctx.Value(nameKey).(string)
	if !ok {
		name = "unknown"
	}

	return context.WithValue(ctx, dataKey, traceVal{
		start: time.Now(),
		name:  name,
	})
}

func (t *tracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	traceData, ok := ctx.Value(dataKey).(traceVal)
	if !ok {
		return
	}
	took := time.Since(traceData.start)

	if data.Err != nil {
		t.l.LogAttrs(ctx, slog.LevelError, "query failed",
			slog.String("name", traceData.name),
			slog.Duration("took", took),
			slog.String("error", data.Err.Error()),
		)

		return
	}

	t.l.LogAttrs(ctx, slog.LevelInfo, "query done",
		slog.String("name", traceData.name),
		slog.Duration("took", took),
	)
}
