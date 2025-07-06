package post

import (
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/plate/query"
	"github.com/rail44/plate/tables"
	"github.com/rail44/plate/types"
)

// Author joins with user table (belongs_to relationship)
func Author(opts ...types.Option[tables.User]) types.QueryOption[tables.Post] {
	return query.DirectJoin[tables.Post, tables.User](
		"author",
		"user",
		query.KeyPair{From: "user_id", To: "id"},
		ast.InnerJoin,
		opts...,
	)
}

// Column accessors for type-safe column references
func ID() types.Column[tables.Post, string] {
	return types.Column[tables.Post, string]{Name: "id"}
}

func UserID() types.Column[tables.Post, string] {
	return types.Column[tables.Post, string]{Name: "user_id"}
}

func Title() types.Column[tables.Post, string] {
	return types.Column[tables.Post, string]{Name: "title"}
}

func Content() types.Column[tables.Post, string] {
	return types.Column[tables.Post, string]{Name: "content"}
}

func Limit(count int) types.QueryOption[tables.Post] {
	return query.Limit[tables.Post](count)
}

// And creates an AND condition that groups multiple conditions
func And(opts ...types.ExprOption[tables.Post]) types.ExprOption[tables.Post] {
	return query.And(opts...)
}

// Or creates an OR condition from multiple conditions
func Or(opts ...types.ExprOption[tables.Post]) types.ExprOption[tables.Post] {
	return query.Or(opts...)
}

// Not creates a logical NOT condition that wraps any ExprOption
func Not(opt types.ExprOption[tables.Post]) types.ExprOption[tables.Post] {
	return query.Not(opt)
}

func OrderBy[V any](column types.Column[tables.Post, V], dir ast.Direction) types.QueryOption[tables.Post] {
	return query.OrderBy(column, dir)
}

// WithInnerJoin changes the JOIN type to INNER JOIN
func WithInnerJoin() types.QueryOption[tables.Post] {
	return query.WithInnerJoinOption[tables.Post]()
}

// Tags joins with tag table through post_tag junction table (many-to-many relationship)
func Tags(opts ...types.Option[tables.Tag]) types.QueryOption[tables.Post] {
	return query.JunctionJoin[tables.Post, tables.Tag](
		"tags",
		"post_tag",
		query.KeyPair{From: "id", To: "post_id"},
		"tag",
		query.KeyPair{From: "tag_id", To: "id"},
		ast.LeftOuterJoin,
		opts...,
	)
}
