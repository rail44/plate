package user

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/hoge/common"
)

func Select(opts ...common.SelectOption) string {
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

func JoinProfile(whereOpt common.ExprOption) common.SelectOption {
	return func(s *common.State, sl *ast.Select) {
		// profileテーブルを追加
		s.Tables["profile"] = struct{}{}
		
		// JOIN構造を作成
		join := &ast.Join{
			Op: ast.InnerJoin,
			Left: sl.From.Source,
			Right: &ast.TableName{
				Table: &ast.Ident{Name: "profile"},
			},
			Cond: &ast.On{
				Expr: &ast.BinaryExpr{
					Left: &ast.Path{
						Idents: []*ast.Ident{
							{Name: "profile"},
							{Name: "user_id"},
						},
					},
					Op: ast.OpEqual,
					Right: &ast.Path{
						Idents: []*ast.Ident{
							{Name: "user"},
							{Name: "id"},
						},
					},
				},
			},
		}
		
		// FromのSourceを置き換え
		sl.From.Source = join
		
		// WHERE句にprofileの条件を追加
		if whereOpt != nil {
			if sl.Where == nil {
				sl.Where = &ast.Where{}
			}
			whereOpt(s, &sl.Where.Expr)
		}
	}
}



func ID(op ast.BinaryOp, value string) common.ExprOption {
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

func Name(op ast.BinaryOp, value string) common.ExprOption {
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

func Email(op ast.BinaryOp, value string) common.ExprOption {
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
