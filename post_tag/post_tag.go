package post_tag

import (
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/plate/query"
	"github.com/rail44/plate/tables"
	"github.com/rail44/plate/types"
)

// Column accessors for type-safe column references
func PostID() types.Column[tables.PostTag, string] {
	return types.Column[tables.PostTag, string]{Name: "post_id"}
}

func TagID() types.Column[tables.PostTag, string] {
	return types.Column[tables.PostTag, string]{Name: "tag_id"}
}

func CreatedAt() types.Column[tables.PostTag, string] {
	return types.Column[tables.PostTag, string]{Name: "created_at"}
}

func Limit(count int) types.QueryOption[tables.PostTag] {
	return query.Limit[tables.PostTag](count)
}

func OrderBy[V any](column types.Column[tables.PostTag, V], dir ast.Direction) types.QueryOption[tables.PostTag] {
	return query.OrderBy(column, dir)
}

// WithInnerJoin changes the JOIN type to INNER JOIN
func WithInnerJoin() types.QueryOption[tables.PostTag] {
	return query.WithInnerJoinOption[tables.PostTag]()
}

// And creates an AND condition that groups multiple conditions
func And(opts ...types.ExprOption[tables.PostTag]) types.ExprOption[tables.PostTag] {
	return query.And(opts...)
}

// Or creates an OR condition from multiple conditions
func Or(opts ...types.ExprOption[tables.PostTag]) types.ExprOption[tables.PostTag] {
	return query.Or(opts...)
}

// Not creates a logical NOT condition that wraps any ExprOption
func Not(opt types.ExprOption[tables.PostTag]) types.ExprOption[tables.PostTag] {
	return query.Not(opt)
}