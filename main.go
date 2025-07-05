package main

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"

	"github.com/rail44/plate/profile"
	"github.com/rail44/plate/user"
)

func main() {
	// User -> Profile JOIN
	sql1, params1 := user.Select(
		user.JoinProfile(nil),
	)
	fmt.Printf("User->Profile JOIN only: %s (params: %v)\n", sql1, params1)

	sql2, params2 := user.Select(
		user.JoinProfile(profile.Bio(ast.OpEqual, "Engineer")),
	)
	fmt.Printf("User->Profile JOIN with condition: %s (params: %v)\n", sql2, params2)

	// Profile -> User JOIN (新機能!)
	sql3, params3 := profile.Select(
		profile.JoinUser(nil),
	)
	fmt.Printf("Profile->User JOIN only: %s (params: %v)\n", sql3, params3)

	sql4, params4 := profile.Select(
		profile.JoinUser(user.Name(ast.OpEqual, "John")),
		profile.Where(profile.Bio(ast.OpEqual, "Engineer")),
	)
	fmt.Printf("Profile->User JOIN with conditions: %s (params: %v)\n", sql4, params4)

	// 既存のクエリ例
	sql5, params5 := user.Select(
		user.Where(user.Name(ast.OpEqual, "John")),
		user.Limit(10),
	)
	fmt.Printf("SELECT with LIMIT: %s (params: %v)\n", sql5, params5)

	sql6, params6 := user.Select(
		user.JoinProfile(profile.Bio(ast.OpEqual, "Engineer")),
		user.Where(user.Name(ast.OpEqual, "John")),
		user.Limit(5),
	)
	fmt.Printf("JOIN with WHERE and LIMIT: %s (params: %v)\n", sql6, params6)
}
