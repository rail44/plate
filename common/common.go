package common

import (
	"github.com/cloudspannerecosystem/memefish/ast"
)

type SelectOption func(*State, *ast.Select)
type ExprOption func(*State, *ast.Expr)

type State struct {
	Tables map[string]struct{}
	Params []any
}

func Where(opt ExprOption) SelectOption {
	return func(s *State, sl *ast.Select) {
		where := ast.Where{}
		opt(s, &where.Expr)

		sl.Where = &where
	}
}


func And(left, right ExprOption) ExprOption {
	return Paren(func(s *State, expr *ast.Expr) {
		b := &ast.BinaryExpr{
			Op: ast.OpAnd,
		}
		left(s, &b.Left)
		right(s, &b.Right)
		*expr = b
	})
}

func Or(left, right ExprOption) ExprOption {
	return Paren(func(s *State, expr *ast.Expr) {
		b := &ast.BinaryExpr{
			Op: ast.OpOr,
		}
		left(s, &b.Left)
		right(s, &b.Right)
		*expr = b
	})
}

func Paren(inner ExprOption) ExprOption {
	return func(s *State, expr *ast.Expr) {
		paren := ast.ParenExpr{}
		inner(s, &paren.Expr)
		*expr = &paren
	}
}

