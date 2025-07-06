package user

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/plate/query"
	"github.com/rail44/plate/tables"
	"github.com/rail44/plate/types"
)

func ID(op ast.BinaryOp, value string) types.ExprOption[tables.User] {
	return func(s *types.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.ColumnExpr(s.CurrentAlias(), "id", op, fmt.Sprintf("p%d", i))
	}
}

func Name(op ast.BinaryOp, value string) types.ExprOption[tables.User] {
	return func(s *types.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.ColumnExpr(s.CurrentAlias(), "name", op, fmt.Sprintf("p%d", i))
	}
}

func Email(op ast.BinaryOp, value string) types.ExprOption[tables.User] {
	return func(s *types.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.ColumnExpr(s.CurrentAlias(), "email", op, fmt.Sprintf("p%d", i))
	}
}

func Limit(count int) types.QueryOption[tables.User] {
	return func(s *types.State, q *ast.Query) {
		q.Limit = query.Limit(count)
	}
}

func OrderBy(column string, dir ast.Direction) types.QueryOption[tables.User] {
	return func(s *types.State, q *ast.Query) {
		if q.OrderBy == nil {
			q.OrderBy = &ast.OrderBy{
				Items: []*ast.OrderByItem{},
			}
		}
		q.OrderBy.Items = append(q.OrderBy.Items,
			query.OrderByItem(s.CurrentAlias(), column, dir))
	}
}

// OrderBy column names
const (
	OrderByID   = "id"
	OrderByName = "name"
)

// Or creates an OR condition that can be used at the top level
func Or(opts ...types.ExprOption[tables.User]) types.ExprOption[tables.User] {
	return func(s *types.State, expr *ast.Expr) {
		if len(opts) == 0 {
			return
		}

		var left ast.Expr
		opts[0](s, &left)

		for i := 1; i < len(opts); i++ {
			var right ast.Expr
			opts[i](s, &right)
			left = query.OrExpr(left, right)
		}

		*expr = left
	}
}

func Paren(inner types.ExprOption[tables.User]) types.ExprOption[tables.User] {
	return func(s *types.State, expr *ast.Expr) {
		var innerExpr ast.Expr
		inner(s, &innerExpr)
		*expr = query.ParenExpr(innerExpr)
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
