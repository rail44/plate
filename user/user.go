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
		WorkingTableAlias: tableName,
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

		*expr = &ast.BinaryExpr{
			Left: &ast.Path{
				Idents: []*ast.Ident{
					{Name: s.WorkingTableAlias},
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
					{Name: s.WorkingTableAlias},
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
					{Name: s.WorkingTableAlias},
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
		if sl.Where == nil {
			sl.Where = &ast.Where{}
		}
		opt(s, &sl.Where.Expr)
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

