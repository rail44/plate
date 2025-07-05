package types

import (
	"github.com/cloudspannerecosystem/memefish/ast"
)

type State struct {
	Tables            map[string]struct{}
	Params            []any
	WorkingTableAlias string
}

type Table interface {
	TableName() string
}


type ExprOption[T Table] func(*State, *ast.Expr)
type QueryOption[T Table] func(*State, *ast.Query)
