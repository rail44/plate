package post_tag

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/plate/query"
	"github.com/rail44/plate/tables"
	"github.com/rail44/plate/types"
)

func PostID(op ast.BinaryOp, value string) types.ExprOption[tables.PostTag] {
	return func(s *types.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.BuildColumnExpr(s.WorkingTableAlias, "post_id", op, fmt.Sprintf("p%d", i))
	}
}

func TagID(op ast.BinaryOp, value string) types.ExprOption[tables.PostTag] {
	return func(s *types.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.BuildColumnExpr(s.WorkingTableAlias, "tag_id", op, fmt.Sprintf("p%d", i))
	}
}

// CreatedAt can be used for filtering when the tag was added to the post
func CreatedAt(op ast.BinaryOp, value string) types.ExprOption[tables.PostTag] {
	return func(s *types.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.BuildColumnExpr(s.WorkingTableAlias, "created_at", op, fmt.Sprintf("p%d", i))
	}
}

func Where(opt types.ExprOption[tables.PostTag]) types.QueryOption[tables.PostTag] {
	return func(s *types.State, q *ast.Query) {
		sl := q.Query.(*ast.Select)
		if sl.Where == nil {
			sl.Where = &ast.Where{}
		}
		opt(s, &sl.Where.Expr)
	}
}