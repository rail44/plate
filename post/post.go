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
func ID() types.Column[tables.Post, string] {
	return types.Column[tables.Post, string]{Name: "id"}
}

func UserID() types.Column[tables.Post, string] {
	return types.Column[tables.Post, string]{Name: "user_id"}
}

func Title() types.Column[tables.Post, string] {
	return types.Column[tables.Post, string]{Name: "title"}
}

func Content() types.Column[tables.Post, string] {
	return types.Column[tables.Post, string]{Name: "content"}
}


func Limit(count int) types.QueryOption[tables.Post] {
	return query.LimitOption[tables.Post](count)
}

// And creates an AND condition that groups multiple conditions
func And(opts ...types.ExprOption[tables.Post]) types.ExprOption[tables.Post] {
	return query.And(opts...)
}

// Or creates an OR condition from multiple conditions
func Or(opts ...types.ExprOption[tables.Post]) types.ExprOption[tables.Post] {
	return query.Or(opts...)
}

// Not creates a logical NOT condition that wraps any ExprOption
func Not(opt types.ExprOption[tables.Post]) types.ExprOption[tables.Post] {
	return query.Not(opt)
}

func OrderBy[V any](column types.Column[tables.Post, V], dir ast.Direction) types.QueryOption[tables.Post] {
	return query.OrderBy(column, dir)
}

// WithInnerJoin changes the JOIN type to INNER JOIN
func WithInnerJoin() types.QueryOption[tables.Post] {
	return query.WithInnerJoinOption[tables.Post]()
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
