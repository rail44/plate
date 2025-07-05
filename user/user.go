package user

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/plate/query"
	"github.com/rail44/plate/tables"
	"github.com/rail44/plate/types"
)

func Select(opts ...types.QueryOption[tables.User]) (string, []any) {
	return query.BuildSelect(opts)
}

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

func And(left, right types.ExprOption[tables.User]) types.ExprOption[tables.User] {
	return func(s *types.State, expr *ast.Expr) {
		var leftExpr, rightExpr ast.Expr
		left(s, &leftExpr)
		right(s, &rightExpr)
		*expr = query.BuildAndExpr(leftExpr, rightExpr)
	}
}

func Or(left, right types.ExprOption[tables.User]) types.ExprOption[tables.User] {
	return func(s *types.State, expr *ast.Expr) {
		var leftExpr, rightExpr ast.Expr
		left(s, &leftExpr)
		right(s, &rightExpr)
		*expr = query.BuildOrExpr(leftExpr, rightExpr)
	}
}

func Paren(inner types.ExprOption[tables.User]) types.ExprOption[tables.User] {
	return func(s *types.State, expr *ast.Expr) {
		var innerExpr ast.Expr
		inner(s, &innerExpr)
		*expr = query.BuildParenExpr(innerExpr)
	}
}
