package user

import (
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/plate/query"
	"github.com/rail44/plate/tables"
	"github.com/rail44/plate/types"
)

// Column accessors for type-safe column references
func ID() types.Column[tables.User, string] {
	return types.Column[tables.User, string]{Name: "id"}
}

func Name() types.Column[tables.User, string] {
	return types.Column[tables.User, string]{Name: "name"}
}

func Email() types.Column[tables.User, string] {
	return types.Column[tables.User, string]{Name: "email"}
}


func Limit(count int) types.QueryOption[tables.User] {
	return query.LimitOption[tables.User](count)
}

func OrderBy[V any](column types.Column[tables.User, V], dir ast.Direction) types.QueryOption[tables.User] {
	return query.OrderBy(column, dir)
}

// Or creates an OR condition that can be used at the top level
func Or(opts ...types.ExprOption[tables.User]) types.ExprOption[tables.User] {
	return query.Or(opts...)
}

// And creates an AND condition that groups multiple conditions
func And(opts ...types.ExprOption[tables.User]) types.ExprOption[tables.User] {
	return query.And(opts...)
}

// Not creates a logical NOT condition that wraps any ExprOption
func Not(opt types.ExprOption[tables.User]) types.ExprOption[tables.User] {
	return query.Not(opt)
}

// WithInnerJoin changes the JOIN type to INNER JOIN
func WithInnerJoin() types.QueryOption[tables.User] {
	return query.WithInnerJoinOption[tables.User]()
}

// Posts joins with post table (has_many relationship)
func Posts(opts ...types.Option[tables.Post]) types.QueryOption[tables.User] {
	return func(s *types.State, q *ast.Query) {
		sl := q.Query.(*ast.Select)

		baseAlias := s.CurrentAlias()

		s.WithRelationship("posts", func(alias string) {
			// Create and apply JOIN
			sl.From.Source = query.Join(query.JoinConfig{
				Source:      sl.From.Source,
				BaseTable:   baseAlias,
				TargetTable: "post",
				TargetAlias: alias,
				BaseKey:     "id",
				TargetKey:   "user_id",
				JoinType:    ast.LeftOuterJoin,
			})

			// Apply options
			for _, opt := range opts {
				opt.Apply(s, q)
			}
		})
	}
}
