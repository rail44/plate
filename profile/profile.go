package profile

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/hoge/common"
)

type ExprOption func(*common.State, *ast.Expr)
type SelectOption func(*common.State, *ast.Select)

func UserID(op ast.BinaryOp, value string) ExprOption {
	return func(s *common.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)

		*expr = &ast.BinaryExpr{
			Left: &ast.Path{
				Idents: []*ast.Ident{
					{Name: "profile"},
					{Name: "user_id"},
				},
			},
			Op: op,
			Right: &ast.Param{
				Name: fmt.Sprintf("p%d", i),
			},
		}
	}
}

func Bio(op ast.BinaryOp, value string) ExprOption {
	return func(s *common.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)

		*expr = &ast.BinaryExpr{
			Left: &ast.Path{
				Idents: []*ast.Ident{
					{Name: "profile"},
					{Name: "bio"},
				},
			},
			Op: op,
			Right: &ast.Param{
				Name: fmt.Sprintf("p%d", i),
			},
		}
	}
}

func Where(opt ExprOption) SelectOption {
	return func(s *common.State, sl *ast.Select) {
		where := ast.Where{}
		opt(s, &where.Expr)

		sl.Where = &where
	}
}


func And(left, right ExprOption) ExprOption {
	return Paren(func(s *common.State, expr *ast.Expr) {
		b := &ast.BinaryExpr{
			Op: ast.OpAnd,
		}
		left(s, &b.Left)
		right(s, &b.Right)
		*expr = b
	})
}

func Or(left, right ExprOption) ExprOption {
	return Paren(func(s *common.State, expr *ast.Expr) {
		b := &ast.BinaryExpr{
			Op: ast.OpOr,
		}
		left(s, &b.Left)
		right(s, &b.Right)
		*expr = b
	})
}

func Paren(inner ExprOption) ExprOption {
	return func(s *common.State, expr *ast.Expr) {
		paren := ast.ParenExpr{}
		inner(s, &paren.Expr)
		*expr = &paren
	}
}

