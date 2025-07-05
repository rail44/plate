package main

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"

	"github.com/rail44/plate/profile"
	"github.com/rail44/plate/user"
)

func main() {
	sql1, params1 := user.Select(
		user.JoinProfile(nil),
	)
	fmt.Printf("JOIN only: %s (params: %v)\n", sql1, params1)

	sql2, params2 := user.Select(
		user.JoinProfile(profile.Bio(ast.OpEqual, "Engineer")),
	)
	fmt.Printf("JOIN with profile condition: %s (params: %v)\n", sql2, params2)

	sql3, params3 := user.Select(
		user.Where(user.Name(ast.OpEqual, "John")),
		user.Limit(10),
	)
	fmt.Printf("SELECT with LIMIT: %s (params: %v)\n", sql3, params3)

	sql4, params4 := user.Select(
		user.JoinProfile(profile.Bio(ast.OpEqual, "Engineer")),
		user.Where(user.Name(ast.OpEqual, "John")),
		user.Limit(5),
	)
	fmt.Printf("JOIN with WHERE and LIMIT: %s (params: %v)\n", sql4, params4)

	sql5, params5 := user.Select(
		user.OrderBy(user.OrderByName, ast.DirectionAsc),
		user.Limit(10),
	)
	fmt.Printf("SELECT with ORDER BY ASC: %s (params: %v)\n", sql5, params5)

	sql6, params6 := user.Select(
		user.Where(user.Email(ast.OpEqual, "john@example.com")),
		user.OrderBy(user.OrderByID, ast.DirectionDesc),
		user.OrderBy(user.OrderByName, ast.DirectionAsc),
		user.Limit(20),
	)
	fmt.Printf("SELECT with multiple ORDER BY: %s (params: %v)\n", sql6, params6)
}
