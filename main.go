package main

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"

	"github.com/rail44/hoge/common"
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
		common.Where(
			common.And(
				user.Name(ast.OpEqual, "Alice"),
				common.Or(
					profile.UserID(ast.OpEqual, "user123"),
					profile.Bio(ast.OpLike, "%developer%"),
				),
			),
		),
	)
	fmt.Printf("JOIN with multiple conditions: %s\n", sql3)
}
