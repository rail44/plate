package main

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"

	"github.com/rail44/plate/post"
	"github.com/rail44/plate/query"
	"github.com/rail44/plate/tables"
	"github.com/rail44/plate/tag"
	"github.com/rail44/plate/user"
)

func main() {
	// Whereを省略した新しいスタイル
	sql5, params5 := query.Select[tables.User](
		user.Name(ast.OpEqual, "John"),
		user.Limit(10),
	)
	fmt.Printf("SELECT with LIMIT: %s (params: %v)\n", sql5, params5)

	// シンプルなOR条件
	sql7, params7 := query.Select[tables.User](
		user.Or(
			user.Name(ast.OpEqual, "John"),
			user.Name(ast.OpEqual, "Jane"),
			user.Name(ast.OpEqual, "Bob"),
		),
		user.ID(ast.OpGreater, "100"),
		user.OrderBy("name", ast.DirectionAsc),
	)
	fmt.Printf("Simple OR condition: %s (params: %v)\n", sql7, params7)
	
	// 複雑な条件の組み合わせ
	sql8, params8 := query.Select[tables.User](
		user.Email(ast.OpLike, "%@company.com"),
		user.Or(
			user.Name(ast.OpEqual, "Admin"),
			user.ID(ast.OpEqual, "1"),
		),
		user.Limit(1),
	)
	fmt.Printf("Complex mixed conditions: %s (params: %v)\n", sql8, params8)

	// LEFT OUTER JOINの例
	fmt.Println("\n--- LEFT OUTER JOIN Examples ---")

	// 1対多の関係のJOIN例
	fmt.Println("\n--- 1-to-Many JOIN Examples ---")
	
	// 全ユーザーと投稿を取得（投稿がないユーザーも含む）
	sql12, params12 := query.Select[tables.User](
		user.Posts(nil),
		user.OrderBy("name", ast.DirectionAsc),
	)
	fmt.Printf("All users with their posts (including users without posts): %s (params: %v)\n", sql12, params12)
	
	// 特定の内容を含む投稿を持つユーザーを取得（投稿がないユーザーも含む）
	sql13, params13 := query.Select[tables.User](
		user.Posts(post.Content(ast.OpLike, "%important%")),
		user.Name(ast.OpNotEqual, "Admin"),
		user.OrderBy("name", ast.DirectionAsc),
	)
	fmt.Printf("Users with important posts (including users without posts): %s (params: %v)\n", sql13, params13)
	
	// 投稿から著者情報を結合して取得
	sql14, params14 := query.Select[tables.Post](
		post.Author(user.Email(ast.OpLike, "%@company.com")),
		post.Title(ast.OpLike, "Announcement:%"),
		post.OrderBy("created_at", ast.DirectionDesc),
		post.Limit(10),
	)
	fmt.Printf("Announcement posts from company authors: %s (params: %v)\n", sql14, params14)

	// 多対多の関係のJOIN例
	fmt.Println("\n--- Many-to-Many JOIN Examples ---")
	
	// 特定のタグを持つ投稿を取得
	sql15, params15 := query.Select[tables.Post](
		post.Tags(tag.Name(ast.OpEqual, "Go")),
		post.Title(ast.OpLike, "%tutorial%"),
		post.OrderBy("created_at", ast.DirectionDesc),
	)
	fmt.Printf("Posts tagged with 'Go': %s (params: %v)\n", sql15, params15)
	
	// 複数のタグ条件
	sql16, params16 := query.Select[tables.Post](
		post.Tags(tag.Or(
			tag.Name(ast.OpEqual, "Go"),
			tag.Name(ast.OpEqual, "Tutorial"),
		)),
		post.Author(user.Name(ast.OpNotEqual, "Admin")),
		post.Limit(20),
	)
	fmt.Printf("Posts tagged with 'Go' or 'Tutorial': %s (params: %v)\n", sql16, params16)
	
	// タグから関連する投稿を検索
	sql17, params17 := query.Select[tables.Tag](
		tag.Posts(post.Title(ast.OpLike, "%announcement%")),
		tag.Name(ast.OpLike, "tech%"),
		tag.OrderBy("name", ast.DirectionAsc),
	)
	fmt.Printf("Tech tags used in announcements: %s (params: %v)\n", sql17, params17)
	
	// タグがない投稿も含めて取得
	sql18, params18 := query.Select[tables.Post](
		post.Tags(nil),
		post.Author(user.Email(ast.OpLike, "%@example.com")),
		post.OrderBy("title", ast.DirectionAsc),
	)
	fmt.Printf("All posts with their tags (including posts without tags): %s (params: %v)\n", sql18, params18)
}
