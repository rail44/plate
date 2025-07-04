package main

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"

	"github.com/rail44/hoge/profile"
	"github.com/rail44/hoge/user"
)

func main() {
	sql1 := user.Select(
		user.JoinProfile(nil),
	)
	fmt.Printf("JOIN only: %s\n", sql1)
	
	sql2 := user.Select(
		user.JoinProfile(profile.Bio(ast.OpEqual, "Engineer")),
	)
	fmt.Printf("JOIN with profile condition: %s\n", sql2)
	
	sql3 := user.Select(
		user.Where(user.Name(ast.OpEqual, "John")),
		user.Limit(10),
	)
	fmt.Printf("SELECT with LIMIT: %s\n", sql3)
	
	sql4 := user.Select(
		user.JoinProfile(profile.Bio(ast.OpEqual, "Engineer")),
		user.Where(user.Name(ast.OpEqual, "John")),
		user.Limit(5),
	)
	fmt.Printf("JOIN with WHERE and LIMIT: %s\n", sql4)
	
	sql5 := user.Select(
		user.OrderBy(user.OrderByName, ast.DirectionAsc),
		user.Limit(10),
	)
	fmt.Printf("SELECT with ORDER BY ASC: %s\n", sql5)
	
	sql6 := user.Select(
		user.Where(user.Email(ast.OpEqual, "john@example.com")),
		user.OrderBy(user.OrderByID, ast.DirectionDesc),
		user.OrderBy(user.OrderByName, ast.DirectionAsc),
		user.Limit(20),
	)
	fmt.Printf("SELECT with multiple ORDER BY: %s\n", sql6)
}
