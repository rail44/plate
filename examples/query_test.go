package main

import (
	"testing"
	
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
			wantSQL:  "SELECT * FROM user",
			wantArgs: nil,
		},
		{
			name: "select with where",
			query: func() (string, []any) {
				return user.Select(
					user.Name().Eq("John"),
				)
			},
			wantSQL:  "SELECT * FROM user WHERE user.name = @p0",
			wantArgs: []any{"John"},
		},
		{
			name: "select with limit",
			query: func() (string, []any) {
				return user.Select(
					user.Limit(10),
				)
			},
			wantSQL:  "SELECT * FROM user LIMIT 10",
			wantArgs: nil,
		},
		{
			name: "select with order by",
			query: func() (string, []any) {
				return user.Select(
					user.OrderBy(user.CreatedAt(), ast.DirectionDesc),
				)
			},
			wantSQL:  "SELECT * FROM user ORDER BY user.created_at DESC",
			wantArgs: nil,
		},
		{
			name: "select with has_many join",
			query: func() (string, []any) {
				return user.Select(
					user.Posts(
						post.Title().Eq("Hello World"),
					),
				)
			},
			wantSQL:  "SELECT * FROM user LEFT OUTER JOIN post AS posts ON user.id = posts.user_id WHERE posts.title = @p0",
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
			name: "select with belongs_to join",
			query: func() (string, []any) {
				return post.Select(
					post.Author(
						user.Name().Eq("Alice"),
					),
				)
			},
			wantSQL:  "SELECT * FROM post INNER JOIN user AS author ON post.user_id = author.id WHERE author.name = @p0",
			wantArgs: []any{"Alice"},
		},
		{
			name: "select with many_to_many join",
			query: func() (string, []any) {
				return post.Select(
					post.Tags(
						tag.Name().Eq("golang"),
					),
				)
			},
			wantSQL:  "SELECT * FROM post LEFT OUTER JOIN post_tag AS post_post_tag ON post.id = post_post_tag.post_id LEFT OUTER JOIN tag AS tags ON post_post_tag.tag_id = tags.id WHERE tags.name = @p0",
			wantArgs: []any{"golang"},
		},
		{
			name: "complex query with multiple conditions",
			query: func() (string, []any) {
				return post.Select(
					post.Title().Like("%tutorial%"),
					post.Author(
						user.Email().Eq("author@example.com"),
					),
					post.OrderBy(post.CreatedAt(), ast.DirectionDesc),
					post.Limit(5),
				)
			},
			wantSQL:  "SELECT * FROM post INNER JOIN user AS author ON post.user_id = author.id WHERE post.title LIKE @p0 AND author.email = @p1 ORDER BY post.created_at DESC LIMIT 5",
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
			wantSQL:  "SELECT * FROM user WHERE (user.name = @p0 OR user.name = @p1)",
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
			wantSQL:  "SELECT * FROM user WHERE (user.name LIKE @p0 AND user.email LIKE @p1)",
			wantArgs: []any{"A%", "%@example.com"},
		},
		{
			name: "NOT condition",
			query: func() (string, []any) {
				return user.Select(
					user.Not(user.Name().Eq("Admin")),
				)
			},
			wantSQL:  "SELECT * FROM user WHERE NOT user.name = @p0",
			wantArgs: []any{"Admin"},
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