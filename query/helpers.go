package query

import (
	"fmt"

	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/hoge/common"
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

// BuildSelect creates a SELECT query with the given options
func BuildSelect(tableName string, opts []func(*common.State, *ast.Query)) (string, []any) {
	s := &common.State{
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
		opt(s, q)
	}

	return q.SQL(), s.Params
}

// JoinConfig contains configuration for building JOIN clauses
type JoinConfig struct {
	BaseTable   string
	TargetTable string
	BaseKey     string
	TargetKey   string
}

// BuildJoin creates a JOIN operation with the given configuration
func BuildJoin(config JoinConfig, whereOpt func(*common.State, *ast.Expr)) func(*common.State, *ast.Query) {
	return func(s *common.State, q *ast.Query) {
		sl := q.Query.(*ast.Select)
		
		// Find available alias for target table
		baseTableName := config.TargetTable
		tableName := baseTableName
		counter := 1
		
		// Check if table name is already used and find available alias
		for {
			if _, exists := s.Tables[tableName]; !exists {
				break
			}
			tableName = fmt.Sprintf("%s%d", baseTableName, counter)
			counter++
		}
		
		// Add table to state
		s.Tables[tableName] = struct{}{}
		
		// Save current working table alias
		previousAlias := s.WorkingTableAlias
		
		// Create JOIN structure
		join := &ast.Join{
			Left: sl.From.Source,
			Op:   ast.InnerJoin,
			Right: &ast.TableName{
				Table: &ast.Ident{
					Name: tableName,
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
							{Name: tableName},
							{Name: config.TargetKey},
						},
					},
				},
			},
		}
		
		// Replace the FROM source with the JOIN
		sl.From.Source = join
		
		// Apply WHERE condition if provided
		if whereOpt != nil {
			// Set working table alias for target table expressions
			s.WorkingTableAlias = tableName
			
			// Create expression holder
			var expr ast.Expr
			whereOpt(s, &expr)
			
			// Only add WHERE clause if expression was actually set
			if expr != nil {
				// Initialize WHERE clause if not exists
				if sl.Where == nil {
					sl.Where = &ast.Where{}
				}
				
				// If there's already a WHERE clause expression, combine with AND
				if sl.Where.Expr != nil {
					existingExpr := sl.Where.Expr
					sl.Where.Expr = &ast.ParenExpr{
						Expr: &ast.BinaryExpr{
							Op:    ast.OpAnd,
							Left:  existingExpr,
							Right: expr,
						},
					}
				} else {
					sl.Where.Expr = expr
				}
			}
			
			// Restore previous working table alias
			s.WorkingTableAlias = previousAlias
		}
	}
}