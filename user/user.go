package user

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/hoge/common"
	"github.com/rail44/hoge/profile"
	"github.com/rail44/hoge/query"
)

type ExprOption func(*common.State, *ast.Expr)
type QueryOption func(*common.State, *ast.Query)

func Select(opts ...QueryOption) (string, []any) {
	tableName := "user"

	s := &common.State{
		Tables: make(map[string]struct{}),
		Params: []any{},
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

func JoinProfile(whereOpt profile.ExprOption) QueryOption {
	return func(s *common.State, q *ast.Query) {
		sl := q.Query.(*ast.Select)
		// Find available alias for profile table
		baseTableName := "profile"
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
			Op: ast.InnerJoin,
			Right: &ast.TableName{
				Table: &ast.Ident{
					Name: tableName,
				},
			},
			Cond: &ast.On{
				Expr: &ast.BinaryExpr{
					Left: &ast.Path{
						Idents: []*ast.Ident{
							{Name: "user"},
							{Name: "id"},
						},
					},
					Op: ast.OpEqual,
					Right: &ast.Path{
						Idents: []*ast.Ident{
							{Name: tableName},
							{Name: "user_id"},
						},
					},
				},
			},
		}
		
		// Replace the FROM source with the JOIN
		sl.From.Source = join
		
		// Apply WHERE condition if provided
		if whereOpt != nil {
			// Set working table alias for profile expressions
			s.WorkingTableAlias = tableName
			
			// Initialize WHERE clause if not exists
			if sl.Where == nil {
				sl.Where = &ast.Where{}
			}
			
			// If there's already a WHERE clause expression, combine with AND
			if sl.Where.Expr != nil {
				existingExpr := sl.Where.Expr
				and := &ast.BinaryExpr{
					Op: ast.OpAnd,
					Left: existingExpr,
				}
				whereOpt(s, &and.Right)
				sl.Where.Expr = &ast.ParenExpr{
					Expr: and,
				}
			} else {
				whereOpt(s, &sl.Where.Expr)
			}
			
			// Restore previous working table alias
			s.WorkingTableAlias = previousAlias
		}
	}
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

