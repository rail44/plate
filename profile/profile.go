package profile

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/plate/query"
	"github.com/rail44/plate/tables"
	"github.com/rail44/plate/types"
)

func Select(opts ...types.QueryOption[tables.Profile]) (string, []any) {
	return query.BuildSelect(opts)
}

func JoinUser(whereOpt types.ExprOption[tables.User]) types.QueryOption[tables.Profile] {
	return types.QueryOption[tables.Profile](query.BuildJoin(query.JoinConfig{
		BaseTable:   "profile",
		TargetTable: "user",
		BaseKey:     "user_id",
		TargetKey:   "id",
	}, func(s *types.State, expr *ast.Expr) {
		if whereOpt != nil {
			whereOpt(s, expr)
		}
	}))
}

func UserID(op ast.BinaryOp, value string) types.ExprOption[tables.Profile] {
	return func(s *types.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.BuildColumnExpr(s.WorkingTableAlias, "user_id", op, fmt.Sprintf("p%d", i))
	}
}

func Bio(op ast.BinaryOp, value string) types.ExprOption[tables.Profile] {
	return func(s *types.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.BuildColumnExpr(s.WorkingTableAlias, "bio", op, fmt.Sprintf("p%d", i))
	}
}

func Where(opt types.ExprOption[tables.Profile]) types.QueryOption[tables.Profile] {
	return func(s *types.State, q *ast.Query) {
		sl := q.Query.(*ast.Select)
		if sl.Where == nil {
			sl.Where = &ast.Where{}
		}
		opt(s, &sl.Where.Expr)
	}
}

func And(left, right types.ExprOption[tables.Profile]) types.ExprOption[tables.Profile] {
	return func(s *types.State, expr *ast.Expr) {
		var leftExpr, rightExpr ast.Expr
		left(s, &leftExpr)
		right(s, &rightExpr)
		*expr = query.BuildAndExpr(leftExpr, rightExpr)
	}
}

func Or(left, right types.ExprOption[tables.Profile]) types.ExprOption[tables.Profile] {
	return func(s *types.State, expr *ast.Expr) {
		var leftExpr, rightExpr ast.Expr
		left(s, &leftExpr)
		right(s, &rightExpr)
		*expr = query.BuildOrExpr(leftExpr, rightExpr)
	}
}

func Paren(inner types.ExprOption[tables.Profile]) types.ExprOption[tables.Profile] {
	return func(s *types.State, expr *ast.Expr) {
		var innerExpr ast.Expr
		inner(s, &innerExpr)
		*expr = query.BuildParenExpr(innerExpr)
	}
}

func Limit(count int) types.QueryOption[tables.Profile] {
	return func(s *types.State, q *ast.Query) {
		q.Limit = query.BuildLimit(count)
	}
}

func OrderBy(column string, dir ast.Direction) types.QueryOption[tables.Profile] {
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
	OrderByUserID = "user_id"
	OrderByBio    = "bio"
)
