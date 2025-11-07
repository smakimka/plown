package lobby

import (
	"plown/internal/repo"

	sq "github.com/Masterminds/squirrel"
)

type filters struct {
	Tag         tagFilter
	Games       gameFilter
	Free        freeFilter
	ReviewLabel reviewLabelFilter
}

func (f *filters) All() []repo.Filter {
	return []repo.Filter{&f.Tag, &f.Games, &f.Free, &f.ReviewLabel}
}

type gameFilter struct {
	IDs []int
}

func (f *gameFilter) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	if len(f.IDs) > 0 {
		return query.Where(sq.Expr("id != ALL(?)", f.IDs))
	}

	return query
}

type tagFilter struct {
	IncludeIDs []int
	ExcludeIDs []int
}

func (f *tagFilter) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	if len(f.IncludeIDs) > 0 {
		query = query.Where(sq.Expr("tag_ids @> ?", f.IncludeIDs))
	}

	if len(f.ExcludeIDs) > 0 {
		query = query.Where(sq.Expr("not (tag_ids && ?)", f.ExcludeIDs))
	}

	return query
}

type freeFilter struct {
	isFree *bool
}

func (f *freeFilter) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	if f.isFree == nil {
		return query
	}

	if *f.isFree {
		query = query.Where("is_free")
	} else {
		query = query.Where("not is_free")
	}

	return query
}

type reviewLabelFilter struct {
	is    []string
	isNot []string
}

func (f *reviewLabelFilter) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	if len(f.is) > 0 {
		query = query.Where(sq.Expr("reviews_filtered_score_label = ANY(?)", f.is))
	}

	if len(f.isNot) > 0 {
		query = query.Where(sq.Expr("not reviews_filtered_score_label = ANY(?)", f.isNot))
	}

	return query
}
