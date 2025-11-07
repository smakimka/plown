package sq

import "github.com/Masterminds/squirrel"

func NewBuilder() squirrel.StatementBuilderType {
	return squirrel.StatementBuilderType(squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar))
}
