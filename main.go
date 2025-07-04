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
}
