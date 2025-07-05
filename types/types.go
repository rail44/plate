package types

import (
	"github.com/cloudspannerecosystem/memefish/ast"
)

type State struct {
	Tables            map[string]struct{}
	Params            []any
	WorkingTableAlias string
}

type table any

type ExprOption[T table] func(*State, *ast.Expr)
type QueryOption[T table] func(*State, *ast.Query)
