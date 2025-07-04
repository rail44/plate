package main

import (
	"fmt"
	parser "github.com/cloudspannerecosystem/memefish"
	"github.com/cloudspannerecosystem/memefish/ast"
)

type state struct {
	params []any
}

type SelectOption func(*state, *ast.Select)

func Select(opts ...SelectOption) string {
	var state state
	stmt := ast.Select{
		Results: []ast.SelectItem {
			&ast.Star{},
		},
		From: &ast.From{
			Source: &ast.TableName{
				Table: &ast.Ident{
					Name: "users",
				},
			},
		},
	}

	for _, opt := range opts {
		opt(&state, &stmt)
	}

	return stmt.SQL()
}

type ExprOption func(*state, *ast.Expr)

func Where(opts ...ExprOption) SelectOption {
	return func(state *state, sl *ast.Select) {
		where := ast.Where{}
		for _, opt := range opts {
			opt(state, &where.Expr)
		}

		sl.Where = &where
	}
}

func Name(op ast.BinaryOp, name string) ExprOption {
	return func(state *state, expr *ast.Expr) {
		i := len(state.params)
		state.params = append(state.params, name)

		*expr = &ast.BinaryExpr{
			Left: &ast.Ident{
				Name: "name",
			},
			Op: op,
			Right: &ast.Param{
				Name: fmt.Sprintf("p%d", i),
			},
		}
	}
}

func And(left, right ExprOption) ExprOption {
	return Paren(func(state *state, expr *ast.Expr) {
		b := &ast.BinaryExpr{
			Op: ast.OpAnd,
		}
		left(state, &b.Left)
		right(state, &b.Right)
		*expr = b
	})
}

func Or(left, right ExprOption) ExprOption {
	return Paren(func(state *state, expr *ast.Expr) {
		b := &ast.BinaryExpr{
			Op: ast.OpOr,
		}
		left(state, &b.Left)
		right(state, &b.Right)
		*expr = b
	})
}

func Paren(inner ExprOption) ExprOption {
	return func(state *state, expr *ast.Expr) {
		paren := ast.ParenExpr{}
		inner(state, &paren.Expr)
		*expr = &paren
	}
}

func main() {
	sql := Select(
		Where(
			Or(
				And(
					Name(ast.OpEqual, "name"),
					Name(ast.OpEqual, "name2"),
				),
				Name(ast.OpEqual, "name3"),
			),
		),
	)
	fmt.Printf("Generated SQL: %s", sql)
}
