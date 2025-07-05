package query

import (
	"fmt"

	"github.com/cloudspannerecosystem/memefish/ast"
)

// BuildLimit creates a Limit AST node
func BuildLimit(count int) *ast.Limit {
	return &ast.Limit{
		Count: &ast.IntLiteral{
			Value: fmt.Sprintf("%d", count),
		},
	}
}

// BuildOrderByItem creates an OrderByItem AST node
func BuildOrderByItem(tableAlias, column string, dir ast.Direction) *ast.OrderByItem {
	return &ast.OrderByItem{
		Expr: &ast.Path{
			Idents: []*ast.Ident{
				{Name: tableAlias},
				{Name: column},
			},
		},
		Dir: dir,
	}
}

// BuildWhereClause creates a Where AST node
func BuildWhereClause(expr ast.Expr) *ast.Where {
	return &ast.Where{
		Expr: expr,
	}
}

// BuildAndExpr creates an AND expression with parentheses
func BuildAndExpr(left, right ast.Expr) ast.Expr {
	return &ast.ParenExpr{
		Expr: &ast.BinaryExpr{
			Op:    ast.OpAnd,
			Left:  left,
			Right: right,
		},
	}
}

// BuildOrExpr creates an OR expression with parentheses
func BuildOrExpr(left, right ast.Expr) ast.Expr {
	return &ast.ParenExpr{
		Expr: &ast.BinaryExpr{
			Op:    ast.OpOr,
			Left:  left,
			Right: right,
		},
	}
}

// BuildParenExpr wraps an expression in parentheses
func BuildParenExpr(inner ast.Expr) ast.Expr {
	return &ast.ParenExpr{
		Expr: inner,
	}
}

// BuildColumnExpr creates a column reference expression with parameter
func BuildColumnExpr(tableAlias, column string, op ast.BinaryOp, paramName string) ast.Expr {
	return &ast.BinaryExpr{
		Left: &ast.Path{
			Idents: []*ast.Ident{
				{Name: tableAlias},
				{Name: column},
			},
		},
		Op: op,
		Right: &ast.Param{
			Name: paramName,
		},
	}
}