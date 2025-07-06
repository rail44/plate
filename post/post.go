package post

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/plate/query"
	"github.com/rail44/plate/tables"
	"github.com/rail44/plate/types"
)

// Author joins with user table (belongs_to relationship)
func Author(opts ...types.Option[tables.User]) types.QueryOption[tables.Post] {
	return func(s *types.State, q *ast.Query) {
		sl := q.Query.(*ast.Select)
		
		// Find available alias for target table
		tableName := query.FindTableAlias(s, "user")
		s.Tables[tableName] = struct{}{}
		
		// Create and apply JOIN
		sl.From.Source = query.Join(query.JoinConfig{
			Source:      sl.From.Source,
			BaseTable:   "post",
			TargetTable: "user",
			TargetAlias: tableName,
			BaseKey:     "user_id",
			TargetKey:   "id",
			JoinType:    ast.InnerJoin,
		})
		
		// Apply options with the target table alias
		if len(opts) > 0 {
			previousAlias := s.WorkingTableAlias
			s.WorkingTableAlias = tableName
			
			for _, opt := range opts {
				opt.Apply(s, q)
			}
			
			s.WorkingTableAlias = previousAlias
		}
	}
}


func ID(op ast.BinaryOp, value string) types.ExprOption[tables.Post] {
	return func(s *types.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.ColumnExpr(s.WorkingTableAlias, "id", op, fmt.Sprintf("p%d", i))
	}
}

func UserID(op ast.BinaryOp, value string) types.ExprOption[tables.Post] {
	return func(s *types.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.ColumnExpr(s.WorkingTableAlias, "user_id", op, fmt.Sprintf("p%d", i))
	}
}

func Title(op ast.BinaryOp, value string) types.ExprOption[tables.Post] {
	return func(s *types.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.ColumnExpr(s.WorkingTableAlias, "title", op, fmt.Sprintf("p%d", i))
	}
}

func Content(op ast.BinaryOp, value string) types.ExprOption[tables.Post] {
	return func(s *types.State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = query.ColumnExpr(s.WorkingTableAlias, "content", op, fmt.Sprintf("p%d", i))
	}
}

func Limit(count int) types.QueryOption[tables.Post] {
	return func(s *types.State, q *ast.Query) {
		q.Limit = query.Limit(count)
	}
}

func Or(opts ...types.ExprOption[tables.Post]) types.ExprOption[tables.Post] {
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

func OrderBy(column string, dir ast.Direction) types.QueryOption[tables.Post] {
	return func(s *types.State, q *ast.Query) {
		if q.OrderBy == nil {
			q.OrderBy = &ast.OrderBy{}
		}
		q.OrderBy.Items = append(q.OrderBy.Items, query.OrderByItem("post", column, dir))
	}
}

// Tags joins with tag table through post_tag junction table (many-to-many relationship)
func Tags(opts ...types.Option[tables.Tag]) types.QueryOption[tables.Post] {
	return func(s *types.State, q *ast.Query) {
		sl := q.Query.(*ast.Select)
		
		// Find available aliases for junction and target tables
		junctionAlias := query.FindTableAlias(s, "post_tag")
		s.Tables[junctionAlias] = struct{}{}
		
		targetAlias := query.FindTableAlias(s, "tag")
		s.Tables[targetAlias] = struct{}{}
		
		// Create and apply JOIN
		sl.From.Source = query.JoinThrough(query.JoinThroughConfig{
			Source:        sl.From.Source,
			BaseTable:     "post",
			JunctionTable: "post_tag",
			JunctionAlias: junctionAlias,
			TargetTable:   "tag",
			TargetAlias:   targetAlias,
			BaseToJunction: struct {
				BaseKey     string
				JunctionKey string
			}{
				BaseKey:     "id",
				JunctionKey: "post_id",
			},
			JunctionToTarget: struct {
				JunctionKey string
				TargetKey   string
			}{
				JunctionKey: "tag_id",
				TargetKey:   "id",
			},
			JoinType: ast.LeftOuterJoin,
		})
		
		// Apply options with the target table alias
		if len(opts) > 0 {
			previousAlias := s.WorkingTableAlias
			s.WorkingTableAlias = targetAlias
			
			for _, opt := range opts {
				opt.Apply(s, q)
			}
			
			s.WorkingTableAlias = previousAlias
		}
	}
}