package main

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"

	"github.com/rail44/plate/profile"
	"github.com/rail44/plate/query"
	"github.com/rail44/plate/tables"
	"github.com/rail44/plate/user"
)

func main() {
	// User -> Profile JOIN
	sql1, params1 := query.Select[tables.User](
		user.JoinProfile(nil),
	)
	fmt.Printf("User->Profile JOIN only: %s (params: %v)\n", sql1, params1)

	sql2, params2 := query.Select[tables.User](
		user.JoinProfile(profile.Bio(ast.OpEqual, "Engineer")),
	)
	fmt.Printf("User->Profile JOIN with condition: %s (params: %v)\n", sql2, params2)

	// Profile -> User JOIN (新機能!)
	sql3, params3 := query.Select[tables.Profile](
		profile.JoinUser(nil),
	)
	fmt.Printf("Profile->User JOIN only: %s (params: %v)\n", sql3, params3)

	sql4, params4 := query.Select[tables.Profile](
		profile.JoinUser(user.Name(ast.OpEqual, "John")),
		profile.Where(profile.Bio(ast.OpEqual, "Engineer")),
	)
	fmt.Printf("Profile->User JOIN with conditions: %s (params: %v)\n", sql4, params4)

	// Whereを省略した新しいスタイル
	sql5, params5 := query.Select[tables.User](
		user.Name(ast.OpEqual, "John"),
		user.Limit(10),
	)
	fmt.Printf("SELECT with LIMIT: %s (params: %v)\n", sql5, params5)

	sql6, params6 := query.Select[tables.User](
		user.JoinProfile(profile.Bio(ast.OpEqual, "Engineer")),
		user.Name(ast.OpEqual, "John"),
		user.Email(ast.OpLike, "%@example.com"),
		user.Limit(5),
	)
	fmt.Printf("JOIN with multiple conditions: %s (params: %v)\n", sql6, params6)
	
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
}
