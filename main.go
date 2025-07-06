package main

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"
	"strings"

	"github.com/rail44/plate/post"
	"github.com/rail44/plate/query"
	"github.com/rail44/plate/tables"
	"github.com/rail44/plate/tag"
	"github.com/rail44/plate/user"
)

func printExample(title, sql string, params []any) {
	fmt.Printf("\n[%s]\n", title)
	fmt.Printf("SQL: %s\n", sql)
	fmt.Printf("Params: %v\n", params)
}

func printSection(title string) {
	fmt.Printf("\n%s\n%s\n", title, strings.Repeat("=", len(title)))
}

func main() {
	printSection("Basic SELECT Examples")

	// Example: Find a specific user with pagination
	// Use case: User search, profile lookup
	sql1, params1 := query.Select[tables.User](
		user.Name(ast.OpEqual, "John"),
		user.Limit(10),
	)
	printExample("Simple WHERE with LIMIT", sql1, params1)

	// Example: Find users by multiple names with additional filters
	// Use case: Bulk user lookup, team member search
	sql2, params2 := query.Select[tables.User](
		user.Or(
			user.Name(ast.OpEqual, "John"),
			user.Name(ast.OpEqual, "Jane"),
			user.Name(ast.OpEqual, "Bob"),
		),
		user.ID(ast.OpGreater, "100"),
		user.OrderBy("name", ast.DirectionAsc),
	)
	printExample("OR condition with multiple filters", sql2, params2)

	// Example: Find admin users from a specific company
	// Use case: Permission checks, admin user identification
	sql3, params3 := query.Select[tables.User](
		user.Email(ast.OpLike, "%@company.com"),
		user.Or(
			user.Name(ast.OpEqual, "Admin"),
			user.ID(ast.OpEqual, "1"),
		),
		user.Limit(1),
	)
	printExample("Complex nested conditions", sql3, params3)

	printSection("One-to-Many Relationship Examples")

	// Example: Get all users with their posts (including users without posts)
	// Use case: User dashboard, activity overview
	sql4, params4 := query.Select[tables.User](
		user.Posts(),
		user.OrderBy("name", ast.DirectionAsc),
	)
	printExample("All users with posts (LEFT JOIN)", sql4, params4)

	// Example: Find non-admin users who have important posts
	// Use case: Content moderation, important content tracking
	sql5, params5 := query.Select[tables.User](
		user.Posts(post.Content(ast.OpLike, "%important%")),
		user.Name(ast.OpNotEqual, "Admin"),
		user.OrderBy("name", ast.DirectionAsc),
	)
	printExample("Users with filtered posts", sql5, params5)

	// Example: Find company announcements with author information
	// Use case: Company news feed, official announcements
	sql6, params6 := query.Select[tables.Post](
		post.Author(user.Email(ast.OpLike, "%@company.com")),
		post.Title(ast.OpLike, "Announcement:%"),
		post.OrderBy("created_at", ast.DirectionDesc),
		post.Limit(10),
	)
	printExample("Posts with author filter (INNER JOIN)", sql6, params6)

	printSection("Many-to-Many Relationship Examples")

	// Example: Find Go tutorials
	// Use case: Tag-based content filtering, topic search
	sql7, params7 := query.Select[tables.Post](
		post.Tags(tag.Name(ast.OpEqual, "Go")),
		post.Title(ast.OpLike, "%tutorial%"),
		post.OrderBy("created_at", ast.DirectionDesc),
	)
	printExample("Posts with specific tag (through junction table)", sql7, params7)

	// Example: Find user-generated content with specific tags
	// Use case: Community content, excluding official posts
	sql8, params8 := query.Select[tables.Post](
		post.Tags(tag.Or(
			tag.Name(ast.OpEqual, "Go"),
			tag.Name(ast.OpEqual, "Tutorial"),
		)),
		post.Author(user.Name(ast.OpNotEqual, "Admin")),
		post.Limit(20),
	)
	printExample("Posts with OR tag conditions and author filter", sql8, params8)

	// Example: Find tech tags used in announcements
	// Use case: Tag analytics, content categorization
	sql9, params9 := query.Select[tables.Tag](
		tag.Posts(post.Title(ast.OpLike, "%announcement%")),
		tag.Name(ast.OpLike, "tech%"),
		tag.OrderBy("name", ast.DirectionAsc),
	)
	printExample("Tags used in specific posts", sql9, params9)

	printSection("Multi-level JOIN Examples")

	// Example: Complete user profile with all posts and tags
	// Use case: User profile page, export user data
	sql10, params10 := query.Select[tables.User](
		user.Email(ast.OpEqual, "john@example.com"),
		user.Posts(
			post.Tags(),
			post.OrderBy("created_at", ast.DirectionDesc),
		),
	)
	printExample("User → Posts → Tags (3-level JOIN)", sql10, params10)

	// Example: Find tutorial authors for a specific technology
	// Use case: Expert identification, content contributor search
	sql11, params11 := query.Select[tables.User](
		user.Posts(
			post.Tags(
				tag.Name(ast.OpEqual, "Go"),
			),
			post.Title(ast.OpLike, "%tutorial%"),
		),
		user.OrderBy("name", ast.DirectionAsc),
	)
	printExample("Users with Go tutorial posts (filtered multi-level)", sql11, params11)

	// Example: Analyze tag usage in company communications
	// Use case: Content strategy, tag effectiveness analysis
	sql12, params12 := query.Select[tables.Tag](
		tag.Name(ast.OpLike, "tech%"),
		tag.Posts(
			post.Author(
				user.Email(ast.OpLike, "%@company.com"),
			),
			post.Title(ast.OpLike, "%announcement%"),
		),
		tag.OrderBy("name", ast.DirectionAsc),
	)
	printExample("Tag → Posts → User (reverse navigation)", sql12, params12)

	sql13, params13 := query.Select[tables.Post](
		post.Or(
			post.ID(ast.OpEqual, "12345"),
			post.Title(ast.OpLike, "%example%"),
		),
		post.Or(
			post.UserID(ast.OpEqual, "6780"),
			post.UserID(ast.OpEqual, "67890"),
		),
	)
	printExample("Multiple OR conditions", sql13, params13)

	// Example: Complex boolean expression (a AND b) OR (c AND d)
	sql14, params14 := query.Select[tables.User](
		user.Or(
			user.And(
				user.Name(ast.OpEqual, "John"),
				user.Email(ast.OpLike, "%@example.com"),
			),
			user.And(
				user.Name(ast.OpEqual, "Jane"),
				user.Email(ast.OpLike, "%@company.com"),
			),
		),
	)
	printExample("Complex boolean expression: (a AND b) OR (c AND d)", sql14, params14)
}
