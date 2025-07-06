package tag

import (
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/plate/query"
	"github.com/rail44/plate/tables"
	"github.com/rail44/plate/types"
)

// Column accessors for type-safe column references
func ID() types.Column[tables.Tag, string] {
	return types.Column[tables.Tag, string]{Name: "id"}
}

func Name() types.Column[tables.Tag, string] {
	return types.Column[tables.Tag, string]{Name: "name"}
}

func Limit(count int) types.QueryOption[tables.Tag] {
	return query.Limit[tables.Tag](count)
}

func OrderBy[V any](column types.Column[tables.Tag, V], dir ast.Direction) types.QueryOption[tables.Tag] {
	return query.OrderBy(column, dir)
}

// WithInnerJoin changes the JOIN type to INNER JOIN
func WithInnerJoin() types.QueryOption[tables.Tag] {
	return query.WithInnerJoinOption[tables.Tag]()
}

// And creates an AND condition that groups multiple conditions
func And(opts ...types.ExprOption[tables.Tag]) types.ExprOption[tables.Tag] {
	return query.And(opts...)
}

// Or creates an OR condition from multiple conditions
func Or(opts ...types.ExprOption[tables.Tag]) types.ExprOption[tables.Tag] {
	return query.Or(opts...)
}

// Not creates a logical NOT condition that wraps any ExprOption
func Not(opt types.ExprOption[tables.Tag]) types.ExprOption[tables.Tag] {
	return query.Not(opt)
}

// Posts joins with post table through post_tag junction table (many-to-many relationship)
func Posts(opts ...types.Option[tables.Post]) types.QueryOption[tables.Tag] {
	return query.JunctionJoin[tables.Tag, tables.Post](
		"posts",
		"post_tag",
		query.KeyPair{From: "id", To: "tag_id"},
		"post",
		query.KeyPair{From: "post_id", To: "id"},
		ast.LeftOuterJoin,
		opts...,
	)
}
