package post

import (
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/plate/query"
	"github.com/rail44/plate/tables"
	"github.com/rail44/plate/types"
)

// Author joins with user table (belongs_to relationship)
func Author(opts ...types.Option[tables.User]) types.QueryOption[tables.Post] {
	return func(s *types.State, q *ast.Query) {
		sl := q.Query.(*ast.Select)

		baseAlias := s.CurrentAlias()

		s.WithRelationship("author", func(alias string) {
			// Create and apply JOIN
			sl.From.Source = query.Join(query.JoinConfig{
				Source:      sl.From.Source,
				BaseTable:   baseAlias,
				TargetTable: "user",
				TargetAlias: alias,
				BaseKey:     "user_id",
				TargetKey:   "id",
				JoinType:    ast.InnerJoin,
			})

			// Apply options
			for _, opt := range opts {
				opt.Apply(s, q)
			}
		})
	}
}

// Column accessors for type-safe column references
func ID() types.Column[tables.Post] {
	return types.Column[tables.Post]{Name: "id"}
}

func UserID() types.Column[tables.Post] {
	return types.Column[tables.Post]{Name: "user_id"}
}

func Title() types.Column[tables.Post] {
	return types.Column[tables.Post]{Name: "title"}
}

func Content() types.Column[tables.Post] {
	return types.Column[tables.Post]{Name: "content"}
}


func Limit(count int) types.QueryOption[tables.Post] {
	return func(s *types.State, q *ast.Query) {
		q.Limit = query.Limit(count)
	}
}

// And creates an AND condition that groups multiple conditions
func And(opts ...types.ExprOption[tables.Post]) types.ExprOption[tables.Post] {
	return query.And(opts...)
}

// Or creates an OR condition from multiple conditions
func Or(opts ...types.ExprOption[tables.Post]) types.ExprOption[tables.Post] {
	return query.Or(opts...)
}

func OrderBy(column types.Column[tables.Post], dir ast.Direction) types.QueryOption[tables.Post] {
	return func(s *types.State, q *ast.Query) {
		if q.OrderBy == nil {
			q.OrderBy = &ast.OrderBy{}
		}
		q.OrderBy.Items = append(q.OrderBy.Items, query.OrderByItem(s.CurrentAlias(), column.Name, dir))
	}
}

// WithInnerJoin changes the JOIN type to INNER JOIN
func WithInnerJoin() types.QueryOption[tables.Post] {
	return func(s *types.State, q *ast.Query) {
		// Find the last JOIN and change its type
		if sl, ok := q.Query.(*ast.Select); ok {
			if join := query.FindLastJoin(sl.From.Source); join != nil {
				join.Op = ast.InnerJoin
			}
		}
	}
}

// Tags joins with tag table through post_tag junction table (many-to-many relationship)
func Tags(opts ...types.Option[tables.Tag]) types.QueryOption[tables.Post] {
	return func(s *types.State, q *ast.Query) {
		sl := q.Query.(*ast.Select)

		baseAlias := s.CurrentAlias()
		junctionAlias := s.RegisterJunction("post_tag")

		s.WithRelationship("tags", func(targetAlias string) {
			// Create and apply JOIN
			sl.From.Source = query.JoinThrough(query.JoinThroughConfig{
				Source:        sl.From.Source,
				BaseTable:     baseAlias,
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

			// Apply options
			for _, opt := range opts {
				opt.Apply(s, q)
			}
		})
	}
}
