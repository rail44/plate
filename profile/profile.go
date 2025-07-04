package profile

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/hoge/common"
)

func UserID(op ast.BinaryOp, value string) common.ExprOption {
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

func Bio(op ast.BinaryOp, value string) common.ExprOption {
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
