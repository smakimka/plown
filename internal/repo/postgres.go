package repo

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"plown/internal/entity"
	"plown/internal/protocol"
	"plown/pkg/postgres"
	"plown/pkg/sq"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{
		pool: pool,
	}
}

func (p *Postgres) Games(ctx context.Context, filters []Filter) ([]entity.Game, error) {
	builder := sq.NewBuilder().Select("*").From("games").OrderBy("reviews_filtered_count")

	for _, filter := range filters {
		builder = filter.Apply(builder)
	}

	query, args, _ := builder.ToSql()
	rows, err := p.pool.Query(postgres.WithQueryName(ctx, "games"), query, args...)
	if err != nil {
		return nil, err
	}

	games, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.Game])
	if err != nil {
		return nil, err
	}

	return games, nil
}

func (p *Postgres) GetDevelopers(ctx context.Context) ([]entity.Developer, error) {
	query, args, _ := sq.NewBuilder().
		Select("id", "name").
		From("developers").
		OrderBy("name asc").
		ToSql()

	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows[entity.Developer](rows, pgx.RowToStructByName)
}

func (p *Postgres) GetPublishers(ctx context.Context) ([]entity.Publisher, error) {
	query, args, _ := sq.NewBuilder().
		Select("id", "name").
		From("publishers").
		OrderBy("name asc").
		ToSql()

	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return pgx.CollectRows[entity.Publisher](rows, pgx.RowToStructByName)
}

func (p *Postgres) GameCount(ctx context.Context, filters []Filter) (int, error) {
	builder := sq.NewBuilder().Select("count(*)").From("games")

	for _, filter := range filters {
		builder = filter.Apply(builder)
	}

	query, args, _ := builder.ToSql()

	row := p.pool.QueryRow(postgres.WithQueryName(ctx, "game_count"), query, args...)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (p *Postgres) TagCount(ctx context.Context, filters []Filter) ([]protocol.TagCount, error) {
	sub := sq.NewBuilder().Select("id").From("games")

	for _, filter := range filters {
		sub = filter.Apply(sub)
	}

	subSQL, subArgs, _ := sub.ToSql()

	builder := sq.NewBuilder().
		Select("tag_id", "count(game_id)").
		From("game_tags").
		Where("game_id IN ("+subSQL+")", subArgs...).
		GroupBy("tag_id").
		OrderBy("count desc")

	query, args, _ := builder.ToSql()

	rows, err := p.pool.Query(postgres.WithQueryName(ctx, "tag_count"), query, args...)
	if err != nil {
		return nil, err
	}

	counts, err := pgx.CollectRows(rows, pgx.RowToStructByPos[protocol.TagCount])
	if err != nil {
		return nil, err
	}

	return counts, nil
}

func (p *Postgres) LabelCount(ctx context.Context, filters []Filter) ([]protocol.LabelCount, error) {
	builder := sq.NewBuilder().
		Select("reviews_filtered_score_label", "count(id)").
		From("games").
		GroupBy("reviews_filtered_score_label").
		OrderBy("count desc")

	for _, filter := range filters {
		builder = filter.Apply(builder)
	}

	query, args, _ := builder.ToSql()

	rows, err := p.pool.Query(postgres.WithQueryName(ctx, "label_count"), query, args...)
	if err != nil {
		return nil, err
	}

	counts, err := pgx.CollectRows(rows, pgx.RowToStructByPos[protocol.LabelCount])
	if err != nil {
		return nil, err
	}

	return counts, nil
}

func (p *Postgres) FreeCount(ctx context.Context, filters []Filter) (int, error) {
	builder := sq.NewBuilder().
		Select("count(id)").
		From("games").
		Where("is_free")

	for _, filter := range filters {
		builder = filter.Apply(builder)
	}

	query, args, _ := builder.ToSql()

	row := p.pool.QueryRow(postgres.WithQueryName(ctx, "free_count"), query, args...)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (p *Postgres) Tags(ctx context.Context) ([]entity.Tag, error) {
	query, args, _ := sq.NewBuilder().
		Select("id", "name").
		From("tags").
		ToSql()

	rows, err := p.pool.Query(postgres.WithQueryName(ctx, "tags"), query, args...)
	if err != nil {
		return nil, err
	}

	tags, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.Tag])
	if err != nil {
		return nil, err
	}

	return tags, nil
}

func (p *Postgres) Labels(ctx context.Context) ([]string, error) {
	query, args, _ := sq.NewBuilder().
		Select("distinct reviews_filtered_score_label").
		From("games").
		ToSql()

	rows, err := p.pool.Query(postgres.WithQueryName(ctx, "labels"), query, args...)
	if err != nil {
		return nil, err
	}

	labels, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, err
	}

	return labels, nil
}
