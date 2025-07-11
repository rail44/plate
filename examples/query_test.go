package main

import (
	"testing"
	"time"
	
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/plate/examples/generated/post"
	"github.com/rail44/plate/examples/generated/tag" 
	"github.com/rail44/plate/examples/generated/user"
)

func TestUserQueries(t *testing.T) {
	tests := []struct {
		name     string
		query    func() (string, []any)
		wantSQL  string
		wantArgs []any
	}{
		{
			name: "simple select",
			query: func() (string, []any) {
				return user.Select()
			},
			wantSQL:  "SELECT user.* FROM user",
			wantArgs: nil,
		},
		{
			name: "select with where",
			query: func() (string, []any) {
				return user.Select(
					user.Name().Eq("John"),
				)
			},
			wantSQL:  "SELECT user.* FROM user WHERE user.name = @p0",
			wantArgs: []any{"John"},
		},
		{
			name: "select with limit",
			query: func() (string, []any) {
				return user.Select(
					user.Limit(10),
				)
			},
			wantSQL:  "SELECT user.* FROM user LIMIT 10",
			wantArgs: nil,
		},
		{
			name: "select with order by",
			query: func() (string, []any) {
				return user.Select(
					user.OrderBy(user.CreatedAt(), ast.DirectionDesc),
				)
			},
			wantSQL:  "SELECT user.* FROM user ORDER BY user.created_at DESC",
			wantArgs: nil,
		},
		{
			name: "select with has_many subquery",
			query: func() (string, []any) {
				return user.Select(
					user.WithPosts(
						post.Title().Eq("Hello World"),
					),
				)
			},
			wantSQL:  "SELECT user.*, ARRAY(SELECT AS STRUCT * FROM post WHERE post.user_id = user.id AND post.title = @p0) AS posts FROM user",
			wantArgs: []any{"Hello World"},
		},
		{
			name: "select filtered by has_many",
			query: func() (string, []any) {
				return user.Select(
					user.WherePosts(
						post.Title().Eq("Hello World"),
					),
				)
			},
			wantSQL:  "SELECT user.* FROM user WHERE EXISTS(SELECT 1 FROM post WHERE post.user_id = user.id AND post.title = @p0)",
			wantArgs: []any{"Hello World"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args := tt.query()
			if sql != tt.wantSQL {
				t.Errorf("SQL mismatch\ngot:  %s\nwant: %s", sql, tt.wantSQL)
			}
			if len(args) != len(tt.wantArgs) {
				t.Errorf("Args length mismatch\ngot:  %d\nwant: %d", len(args), len(tt.wantArgs))
			} else {
				for i, arg := range args {
					if arg != tt.wantArgs[i] {
						t.Errorf("Arg[%d] mismatch\ngot:  %v\nwant: %v", i, arg, tt.wantArgs[i])
					}
				}
			}
		})
	}
}

func TestPostQueries(t *testing.T) {
	tests := []struct {
		name     string
		query    func() (string, []any)
		wantSQL  string
		wantArgs []any
	}{
		{
			name: "select with belongs_to subquery",
			query: func() (string, []any) {
				return post.Select(
					post.WithAuthor(
						user.Name().Eq("Alice"),
					),
				)
			},
			wantSQL:  "SELECT post.*, (SELECT AS STRUCT * FROM user WHERE user.id = post.user_id AND user.name = @p0) AS author FROM post",
			wantArgs: []any{"Alice"},
		},
		{
			name: "select with many_to_many subquery",
			query: func() (string, []any) {
				return post.Select(
					post.WithTags(
						tag.Name().Eq("golang"),
					),
				)
			},
			wantSQL:  "SELECT post.*, ARRAY(SELECT AS STRUCT tag.* FROM tag INNER JOIN post_tag ON tag.id = post_tag.tag_id WHERE post_tag.post_id = post.id AND tag.name = @p0) AS tags FROM post",
			wantArgs: []any{"golang"},
		},
		{
			name: "complex query with multiple conditions",
			query: func() (string, []any) {
				return post.Select(
					post.Title().Like("%tutorial%"),
					post.WhereAuthor(
						user.Email().Eq("author@example.com"),
					),
					post.OrderBy(post.CreatedAt(), ast.DirectionDesc),
					post.Limit(5),
				)
			},
			wantSQL:  "SELECT post.* FROM post WHERE post.title LIKE @p0 AND EXISTS(SELECT 1 FROM user WHERE user.id = post.user_id AND user.email = @p1) ORDER BY post.created_at DESC LIMIT 5",
			wantArgs: []any{"%tutorial%", "author@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args := tt.query()
			if sql != tt.wantSQL {
				t.Errorf("SQL mismatch\ngot:  %s\nwant: %s", sql, tt.wantSQL)
			}
			if len(args) != len(tt.wantArgs) {
				t.Errorf("Args length mismatch\ngot:  %d\nwant: %d", len(args), len(tt.wantArgs))
			} else {
				for i, arg := range args {
					if arg != tt.wantArgs[i] {
						t.Errorf("Arg[%d] mismatch\ngot:  %v\nwant: %v", i, arg, tt.wantArgs[i])
					}
				}
			}
		})
	}
}

func TestLogicalOperators(t *testing.T) {
	tests := []struct {
		name     string
		query    func() (string, []any)
		wantSQL  string
		wantArgs []any
	}{
		{
			name: "OR condition",
			query: func() (string, []any) {
				return user.Select(
					user.Or(
						user.Name().Eq("Alice"),
						user.Name().Eq("Bob"),
					),
				)
			},
			wantSQL:  "SELECT user.* FROM user WHERE (user.name = @p0 OR user.name = @p1)",
			wantArgs: []any{"Alice", "Bob"},
		},
		{
			name: "AND condition",
			query: func() (string, []any) {
				return user.Select(
					user.And(
						user.Name().Like("A%"),
						user.Email().Like("%@example.com"),
					),
				)
			},
			wantSQL:  "SELECT user.* FROM user WHERE (user.name LIKE @p0 AND user.email LIKE @p1)",
			wantArgs: []any{"A%", "%@example.com"},
		},
		{
			name: "NOT condition",
			query: func() (string, []any) {
				return user.Select(
					user.Not(user.Name().Eq("Admin")),
				)
			},
			wantSQL:  "SELECT user.* FROM user WHERE NOT (user.name = @p0)",
			wantArgs: []any{"Admin"},
		},
		{
			name: "NOT with OR condition",
			query: func() (string, []any) {
				return user.Select(
					user.Not(user.Or(
						user.Name().Eq("Admin"),
						user.Name().Eq("Root"),
					)),
				)
			},
			wantSQL:  "SELECT user.* FROM user WHERE NOT (user.name = @p0 OR user.name = @p1)",
			wantArgs: []any{"Admin", "Root"},
		},
		{
			name: "NOT with AND condition",
			query: func() (string, []any) {
				return user.Select(
					user.Not(user.And(
						user.Name().Eq("Alice"),
						user.Email().Like("%@admin.com"),
					)),
				)
			},
			wantSQL:  "SELECT user.* FROM user WHERE NOT (user.name = @p0 AND user.email LIKE @p1)",
			wantArgs: []any{"Alice", "%@admin.com"},
		},
		{
			name: "Complex NOT precedence",
			query: func() (string, []any) {
				return user.Select(
					user.And(
						user.Not(user.Name().Eq("Admin")),
						user.Email().Like("%@example.com"),
					),
				)
			},
			wantSQL:  "SELECT user.* FROM user WHERE (NOT (user.name = @p0) AND user.email LIKE @p1)",
			wantArgs: []any{"Admin", "%@example.com"},
		},
		{
			name: "NOT precedence with inequality",
			query: func() (string, []any) {
				testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
				return user.Select(
					user.Not(user.CreatedAt().Gt(testTime)),
				)
			},
			wantSQL:  "SELECT user.* FROM user WHERE NOT (user.created_at > @p0)",
			wantArgs: []any{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		},
		{
			name: "Multiple NOT conditions",
			query: func() (string, []any) {
				return user.Select(
					user.And(
						user.Not(user.Name().Eq("Admin")),
						user.Not(user.Name().Eq("Root")),
					),
				)
			},
			wantSQL:  "SELECT user.* FROM user WHERE (NOT (user.name = @p0) AND NOT (user.name = @p1))",
			wantArgs: []any{"Admin", "Root"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args := tt.query()
			if sql != tt.wantSQL {
				t.Errorf("SQL mismatch\ngot:  %s\nwant: %s", sql, tt.wantSQL)
			}
			if len(args) != len(tt.wantArgs) {
				t.Errorf("Args length mismatch\ngot:  %d\nwant: %d", len(args), len(tt.wantArgs))
			} else {
				for i, arg := range args {
					if arg != tt.wantArgs[i] {
						t.Errorf("Arg[%d] mismatch\ngot:  %v\nwant: %v", i, arg, tt.wantArgs[i])
					}
				}
			}
		})
	}
}