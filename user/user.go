package user

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/hoge/common"
	"github.com/rail44/hoge/profile"
)

type ExprOption func(*common.State, *ast.Expr)
type SelectOption func(*common.State, *ast.Select)

func Select(opts ...SelectOption) string {
	tableName := "user"

	s := &common.State{
		Tables: make(map[string]struct{}),
		Params: []any{},
	}
	s.Tables[tableName] = struct{}{}

	stmt := ast.Select{
		Results: []ast.SelectItem {
			&ast.Star{},
		},
		From: &ast.From{
			Source: &ast.TableName{
				Table: &ast.Ident{
					Name: tableName,
				},
			},
		},
	}

	for _, opt := range opts {
		opt(s, &stmt)
	}

	return stmt.SQL()
}

func JoinProfile(whereOpt profile.ExprOption) SelectOption {
	return func(s *common.State, sl *ast.Select) {
	}
}



func ID(op ast.BinaryOp, value string) ExprOption {
	return func(s *common.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)

		*expr = &ast.BinaryExpr{
			Left: &ast.Path{
				Idents: []*ast.Ident{
					{Name: "user"},
					{Name: "id"},
				},
			},
			Op: op,
			Right: &ast.Param{
				Name: fmt.Sprintf("p%d", i),
			},
		}
	}
}

func Name(op ast.BinaryOp, value string) ExprOption {
	return func(s *common.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)

		*expr = &ast.BinaryExpr{
			Left: &ast.Path{
				Idents: []*ast.Ident{
					{Name: "user"},
					{Name: "name"},
				},
			},
			Op: op,
			Right: &ast.Param{
				Name: fmt.Sprintf("p%d", i),
			},
		}
	}
}

func Email(op ast.BinaryOp, value string) ExprOption {
	return func(s *common.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)

		*expr = &ast.BinaryExpr{
			Left: &ast.Path{
				Idents: []*ast.Ident{
					{Name: "user"},
					{Name: "email"},
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

