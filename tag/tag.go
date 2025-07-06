package tag

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/plate/query"
	"github.com/rail44/plate/tables"
	"github.com/rail44/plate/types"
)

func ID(op ast.BinaryOp, value string) types.ExprOption[tables.Tag] {
	return func(s *types.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.ColumnExpr(s.CurrentAlias(), "id", op, fmt.Sprintf("p%d", i))
	}
}

func Name(op ast.BinaryOp, value string) types.ExprOption[tables.Tag] {
	return func(s *types.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.ColumnExpr(s.CurrentAlias(), "name", op, fmt.Sprintf("p%d", i))
	}
}

func Limit(count int) types.QueryOption[tables.Tag] {
	return func(s *types.State, q *ast.Query) {
		q.Limit = query.Limit(count)
	}
}

func OrderBy(column string, dir ast.Direction) types.QueryOption[tables.Tag] {
	return func(s *types.State, q *ast.Query) {
		if q.OrderBy == nil {
			q.OrderBy = &ast.OrderBy{}
		}
		q.OrderBy.Items = append(q.OrderBy.Items, query.OrderByItem(s.CurrentAlias(), column, dir))
	}
}

func Or(opts ...types.ExprOption[tables.Tag]) types.ExprOption[tables.Tag] {
	return func(s *types.State, expr *ast.Expr) {
		if len(opts) == 0 {
			return
		}
		if len(opts) == 1 {
			opts[0](s, expr)
			return
		}

		var left ast.Expr
		opts[0](s, &left)

		for i := 1; i < len(opts); i++ {
			var right ast.Expr
			opts[i](s, &right)
			left = query.OrExpr(left, right)
		}

		*expr = query.ParenExpr(left)
	}
}

// Posts joins with post table through post_tag junction table (many-to-many relationship)
func Posts(opts ...types.Option[tables.Post]) types.QueryOption[tables.Tag] {
	return func(s *types.State, q *ast.Query) {
		sl := q.Query.(*ast.Select)

		baseAlias := s.CurrentAlias()
		junctionAlias := s.RegisterJunction("post_tag")

		s.WithRelationship("posts", func(targetAlias string) {
			// Create and apply JOIN
			sl.From.Source = query.JoinThrough(query.JoinThroughConfig{
				Source:        sl.From.Source,
				BaseTable:     baseAlias,
				JunctionTable: "post_tag",
				JunctionAlias: junctionAlias,
				TargetTable:   "post",
				TargetAlias:   targetAlias,
				BaseToJunction: struct {
					BaseKey     string
					JunctionKey string
				}{
					BaseKey:     "id",
					JunctionKey: "tag_id",
				},
				JunctionToTarget: struct {
					JunctionKey string
					TargetKey   string
				}{
					JunctionKey: "post_id",
					TargetKey:   "id",
				},
				JoinType: ast.LeftOuterJoin,
			})

			// Apply options
			for _, opt := range opts {
				opt.Apply(s, q)
			}
		})
	}
}
