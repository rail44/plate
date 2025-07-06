package query

import (
	"fmt"

	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/plate/types"
)

// Limit creates a Limit AST node
func Limit(count int) *ast.Limit {
	return &ast.Limit{
		Count: &ast.IntLiteral{
			Value: fmt.Sprintf("%d", count),
		},
	}
}

// OrderByItem creates an OrderByItem AST node
func OrderByItem(tableAlias, column string, dir ast.Direction) *ast.OrderByItem {
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

// WhereClause creates a Where AST node
func WhereClause(expr ast.Expr) *ast.Where {
	return &ast.Where{
		Expr: expr,
	}
}

// AndExpr creates an AND expression with parentheses
func AndExpr(left, right ast.Expr) ast.Expr {
	return &ast.ParenExpr{
		Expr: &ast.BinaryExpr{
			Op:    ast.OpAnd,
			Left:  left,
			Right: right,
		},
	}
}

// OrExpr creates an OR expression with parentheses
func OrExpr(left, right ast.Expr) ast.Expr {
	return &ast.ParenExpr{
		Expr: &ast.BinaryExpr{
			Op:    ast.OpOr,
			Left:  left,
			Right: right,
		},
	}
}

// ParenExpr wraps an expression in parentheses
func ParenExpr(inner ast.Expr) ast.Expr {
	return &ast.ParenExpr{
		Expr: inner,
	}
}

// ColumnExpr creates a column reference expression with parameter
func ColumnExpr(tableAlias, column string, op ast.BinaryOp, paramName string) ast.Expr {
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

// Select is a generic select function for any table
func Select[T types.Table](opts ...types.Option[T]) (string, []any) {
	var t T
	tableName := t.TableName()

	s := &types.State{
		Tables:            make(map[string]struct{}),
		Params:            []any{},
		WorkingTableAlias: tableName,
	}
	s.Tables[tableName] = struct{}{}

	stmt := ast.Select{
		Results: []ast.SelectItem{
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

	q := &ast.Query{
		Query: &stmt,
	}

	for _, opt := range opts {
		opt.Apply(s, q)
	}

	return q.SQL(), s.Params
}

// JoinConfig contains configuration for building JOIN clauses
type JoinConfig struct {
	Source      ast.TableExpr
	BaseTable   string
	TargetTable string
	TargetAlias string
	BaseKey     string
	TargetKey   string
	JoinType    ast.JoinOp // JOIN type (INNER, LEFT OUTER, etc.)
}

// JoinThroughConfig contains configuration for building many-to-many JOIN through a junction table
type JoinThroughConfig struct {
	Source         ast.TableExpr
	BaseTable      string
	JunctionTable  string
	JunctionAlias  string
	TargetTable    string
	TargetAlias    string
	BaseToJunction struct {
		BaseKey     string
		JunctionKey string
	}
	JunctionToTarget struct {
		JunctionKey string
		TargetKey   string
	}
	JoinType ast.JoinOp
}

// JoinThrough creates a many-to-many JOIN through a junction table
func JoinThrough(config JoinThroughConfig) *ast.Join {
	// First create junction table join
	junctionJoin := Join(JoinConfig{
		Source:      config.Source,
		BaseTable:   config.BaseTable,
		TargetTable: config.JunctionTable,
		TargetAlias: config.JunctionAlias,
		BaseKey:     config.BaseToJunction.BaseKey,
		TargetKey:   config.BaseToJunction.JunctionKey,
		JoinType:    config.JoinType,
	})

	// Then create target table join using junction join as source
	return Join(JoinConfig{
		Source:      junctionJoin,
		BaseTable:   config.JunctionAlias,
		TargetTable: config.TargetTable,
		TargetAlias: config.TargetAlias,
		BaseKey:     config.JunctionToTarget.JunctionKey,
		TargetKey:   config.JunctionToTarget.TargetKey,
		JoinType:    config.JoinType,
	})
}

// Join creates a JOIN AST node
func Join(config JoinConfig) *ast.Join {
	return &ast.Join{
		Left: config.Source,
		Op:   config.JoinType,
		Right: &ast.TableName{
			Table: &ast.Ident{
				Name: config.TargetTable,
			},
			As: &ast.AsAlias{
				Alias: &ast.Ident{
					Name: config.TargetAlias,
				},
			},
		},
		Cond: &ast.On{
			Expr: &ast.BinaryExpr{
				Left: &ast.Path{
					Idents: []*ast.Ident{
						{Name: config.BaseTable},
						{Name: config.BaseKey},
					},
				},
				Op: ast.OpEqual,
				Right: &ast.Path{
					Idents: []*ast.Ident{
						{Name: config.TargetAlias},
						{Name: config.TargetKey},
					},
				},
			},
		},
	}
}
