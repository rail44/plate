package user

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/plate/common"
	"github.com/rail44/plate/profile"
	"github.com/rail44/plate/query"
)

type ExprOption func(*common.State, *ast.Expr)
type QueryOption func(*common.State, *ast.Query)

func Select(opts ...QueryOption) (string, []any) {
	// Convert typed options to untyped for helper
	untyped := make([]func(*common.State, *ast.Query), len(opts))
	for i, opt := range opts {
		opt := opt // capture loop variable
		untyped[i] = func(s *common.State, q *ast.Query) {
			opt(s, q)
		}
	}
	return query.BuildSelect("user", untyped)
}

func JoinProfile(whereOpt profile.ExprOption) QueryOption {
	return QueryOption(query.BuildJoin(query.JoinConfig{
		BaseTable:   "user",
		TargetTable: "profile",
		BaseKey:     "id",
		TargetKey:   "user_id",
	}, func(s *common.State, expr *ast.Expr) {
		if whereOpt != nil {
			whereOpt(s, expr)
		}
	}))
}

func ID(op ast.BinaryOp, value string) ExprOption {
	return func(s *common.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.BuildColumnExpr(s.WorkingTableAlias, "id", op, fmt.Sprintf("p%d", i))
	}
}

func Name(op ast.BinaryOp, value string) ExprOption {
	return func(s *common.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.BuildColumnExpr(s.WorkingTableAlias, "name", op, fmt.Sprintf("p%d", i))
	}
}

func Email(op ast.BinaryOp, value string) ExprOption {
	return func(s *common.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.BuildColumnExpr(s.WorkingTableAlias, "email", op, fmt.Sprintf("p%d", i))
	}
}

func Where(opt ExprOption) QueryOption {
	return func(s *common.State, q *ast.Query) {
		sl := q.Query.(*ast.Select)
		if sl.Where == nil {
			sl.Where = &ast.Where{}
		}
		opt(s, &sl.Where.Expr)
	}
}

func Limit(count int) QueryOption {
	return func(s *common.State, q *ast.Query) {
		q.Limit = query.BuildLimit(count)
	}
}

func OrderBy(column string, dir ast.Direction) QueryOption {
	return func(s *common.State, q *ast.Query) {
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

func And(left, right ExprOption) ExprOption {
	return func(s *common.State, expr *ast.Expr) {
		var leftExpr, rightExpr ast.Expr
		left(s, &leftExpr)
		right(s, &rightExpr)
		*expr = query.BuildAndExpr(leftExpr, rightExpr)
	}
}

func Or(left, right ExprOption) ExprOption {
	return func(s *common.State, expr *ast.Expr) {
		var leftExpr, rightExpr ast.Expr
		left(s, &leftExpr)
		right(s, &rightExpr)
		*expr = query.BuildOrExpr(leftExpr, rightExpr)
	}
}

func Paren(inner ExprOption) ExprOption {
	return func(s *common.State, expr *ast.Expr) {
		var innerExpr ast.Expr
		inner(s, &innerExpr)
		*expr = query.BuildParenExpr(innerExpr)
	}
}
