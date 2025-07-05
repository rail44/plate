package profile

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/plate/types"
	"github.com/rail44/plate/query"
)

type ExprOption func(*types.State, *ast.Expr)
type QueryOption func(*types.State, *ast.Query)

func Select(opts ...QueryOption) (string, []any) {
	// Convert typed options to untyped for helper
	untyped := make([]func(*types.State, *ast.Query), len(opts))
	for i, opt := range opts {
		untyped[i] = func(s *types.State, q *ast.Query) {
			opt(s, q)
		}
	}
	return query.BuildSelect("profile", untyped)
}

func UserID(op ast.BinaryOp, value string) ExprOption {
	return func(s *types.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.BuildColumnExpr(s.WorkingTableAlias, "user_id", op, fmt.Sprintf("p%d", i))
	}
}

func Bio(op ast.BinaryOp, value string) ExprOption {
	return func(s *types.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.BuildColumnExpr(s.WorkingTableAlias, "bio", op, fmt.Sprintf("p%d", i))
	}
}

func Where(opt ExprOption) QueryOption {
	return func(s *types.State, q *ast.Query) {
		sl := q.Query.(*ast.Select)
		if sl.Where == nil {
			sl.Where = &ast.Where{}
		}
		opt(s, &sl.Where.Expr)
	}
}

func And(left, right ExprOption) ExprOption {
	return func(s *types.State, expr *ast.Expr) {
		var leftExpr, rightExpr ast.Expr
		left(s, &leftExpr)
		right(s, &rightExpr)
		*expr = query.BuildAndExpr(leftExpr, rightExpr)
	}
}

func Or(left, right ExprOption) ExprOption {
	return func(s *types.State, expr *ast.Expr) {
		var leftExpr, rightExpr ast.Expr
		left(s, &leftExpr)
		right(s, &rightExpr)
		*expr = query.BuildOrExpr(leftExpr, rightExpr)
	}
}

func Paren(inner ExprOption) ExprOption {
	return func(s *types.State, expr *ast.Expr) {
		var innerExpr ast.Expr
		inner(s, &innerExpr)
		*expr = query.BuildParenExpr(innerExpr)
	}
}

func Limit(count int) QueryOption {
	return func(s *types.State, q *ast.Query) {
		q.Limit = query.BuildLimit(count)
	}
}

func OrderBy(column string, dir ast.Direction) QueryOption {
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
