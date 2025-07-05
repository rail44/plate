package user

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/plate/query"
	"github.com/rail44/plate/tables"
	"github.com/rail44/plate/types"
)

func JoinProfile(whereOpt types.ExprOption[tables.Profile]) types.QueryOption[tables.User] {
	return types.QueryOption[tables.User](query.BuildJoin(query.JoinConfig{
		BaseTable:   "user",
		TargetTable: "profile",
		BaseKey:     "id",
		TargetKey:   "user_id",
	}, func(s *types.State, expr *ast.Expr) {
		if whereOpt != nil {
			whereOpt(s, expr)
		}
	}))
}

func ID(op ast.BinaryOp, value string) types.ExprOption[tables.User] {
	return func(s *types.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.BuildColumnExpr(s.WorkingTableAlias, "id", op, fmt.Sprintf("p%d", i))
	}
}

func Name(op ast.BinaryOp, value string) types.ExprOption[tables.User] {
	return func(s *types.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.BuildColumnExpr(s.WorkingTableAlias, "name", op, fmt.Sprintf("p%d", i))
	}
}

func Email(op ast.BinaryOp, value string) types.ExprOption[tables.User] {
	return func(s *types.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.BuildColumnExpr(s.WorkingTableAlias, "email", op, fmt.Sprintf("p%d", i))
	}
}

func Where(opt types.ExprOption[tables.User]) types.QueryOption[tables.User] {
	return func(s *types.State, q *ast.Query) {
		sl := q.Query.(*ast.Select)
		if sl.Where == nil {
			sl.Where = &ast.Where{}
		}
		opt(s, &sl.Where.Expr)
	}
}

func Limit(count int) types.QueryOption[tables.User] {
	return func(s *types.State, q *ast.Query) {
		q.Limit = query.BuildLimit(count)
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
			query.BuildOrderByItem(s.WorkingTableAlias, column, dir))
	}
}

// OrderBy column names
const (
	OrderByID   = "id"
	OrderByName = "name"
)

// Or creates an OR condition that can be used at the top level
func Or(conditions ...types.ExprOption[tables.User]) types.QueryOption[tables.User] {
	return func(s *types.State, q *ast.Query) {
		sl := q.Query.(*ast.Select)
		
		// Build the OR expression from all conditions
		var orExpr ast.Expr
		for i, cond := range conditions {
			var expr ast.Expr
			cond(s, &expr)
			
			if i == 0 {
				orExpr = expr
			} else {
				orExpr = &ast.BinaryExpr{
					Op:    ast.OpOr,
					Left:  orExpr,
					Right: expr,
				}
			}
		}
		
		// Add to WHERE clause
		if sl.Where == nil {
			sl.Where = &ast.Where{Expr: orExpr}
		} else if sl.Where.Expr == nil {
			sl.Where.Expr = orExpr
		} else {
			// Combine with existing WHERE using AND
			sl.Where.Expr = &ast.BinaryExpr{
				Op:    ast.OpAnd,
				Left:  sl.Where.Expr,
				Right: &ast.ParenExpr{Expr: orExpr},
			}
		}
	}
}

func Paren(inner types.ExprOption[tables.User]) types.ExprOption[tables.User] {
	return func(s *types.State, expr *ast.Expr) {
		var innerExpr ast.Expr
		inner(s, &innerExpr)
		*expr = query.BuildParenExpr(innerExpr)
	}
}
