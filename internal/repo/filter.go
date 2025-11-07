package repo

import "github.com/Masterminds/squirrel"

type Filter interface {
	Apply(query squirrel.SelectBuilder) squirrel.SelectBuilder
}
