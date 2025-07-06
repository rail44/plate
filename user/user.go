package user

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/plate/query"
	"github.com/rail44/plate/tables"
	"github.com/rail44/plate/types"
)

// Column accessors for type-safe column references
func ID() types.Column[tables.User] {
	return types.Column[tables.User]{Name: "id"}
}

func Name() types.Column[tables.User] {
	return types.Column[tables.User]{Name: "name"}
}

func Email() types.Column[tables.User] {
	return types.Column[tables.User]{Name: "email"}
}

// Legacy functions for backward compatibility
func IDLegacy(op ast.BinaryOp, value string) types.ExprOption[tables.User] {
	return func(s *types.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.ColumnExpr(s.CurrentAlias(), "id", op, fmt.Sprintf("p%d", i))
	}
}


func Limit(count int) types.QueryOption[tables.User] {
	return func(s *types.State, q *ast.Query) {
		q.Limit = query.Limit(count)
	}
}

func OrderBy(column types.Column[tables.User], dir ast.Direction) types.QueryOption[tables.User] {
	return func(s *types.State, q *ast.Query) {
		if q.OrderBy == nil {
			q.OrderBy = &ast.OrderBy{
				Items: []*ast.OrderByItem{},
			}
		}
		q.OrderBy.Items = append(q.OrderBy.Items,
			query.OrderByItem(s.CurrentAlias(), column.Name, dir))
	}
}

// OrderBy column names
const (
	OrderByID   = "id"
	OrderByName = "name"
)

// Or creates an OR condition that can be used at the top level
func Or(opts ...types.ExprOption[tables.User]) types.ExprOption[tables.User] {
	return query.Or(opts...)
}

// And creates an AND condition that groups multiple conditions
func And(opts ...types.ExprOption[tables.User]) types.ExprOption[tables.User] {
	return query.And(opts...)
}

// WithInnerJoin changes the JOIN type to INNER JOIN
func WithInnerJoin() types.QueryOption[tables.User] {
	return func(s *types.State, q *ast.Query) {
		// Find the last JOIN and change its type
		if sl, ok := q.Query.(*ast.Select); ok {
			if join := query.FindLastJoin(sl.From.Source); join != nil {
				join.Op = ast.InnerJoin
			}
		}
	}
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
