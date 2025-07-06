package types

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"
)

type State struct {
	Tables            map[string]struct{}
	Params            []any
	WorkingTableAlias string
}

// RegisterTableAlias finds an available alias based on the relationship context and registers it
func (s *State) RegisterTableAlias(tableName, relationshipName string) string {
	// Use relationship name as the base alias
	baseAlias := relationshipName
	if baseAlias == "" {
		baseAlias = tableName
	}

	// Check if the base alias is available
	if _, exists := s.Tables[baseAlias]; !exists {
		s.Tables[baseAlias] = struct{}{}
		return baseAlias
	}

	// If not, try with context prefix
	if s.WorkingTableAlias != "" && s.WorkingTableAlias != tableName {
		contextAlias := fmt.Sprintf("%s_%s", s.WorkingTableAlias, baseAlias)
		if _, exists := s.Tables[contextAlias]; !exists {
			s.Tables[contextAlias] = struct{}{}
			return contextAlias
		}
	}

	// Fall back to numbered suffix
	counter := 1
	for {
		numberedAlias := fmt.Sprintf("%s%d", baseAlias, counter)
		if _, exists := s.Tables[numberedAlias]; !exists {
			s.Tables[numberedAlias] = struct{}{}
			return numberedAlias
		}
		counter++
	}
}

type Table interface {
	TableName() string
}

// Option represents a query building option that can be applied to a query
type Option[T Table] interface {
	Apply(s *State, q *ast.Query)
}

// ExprOption represents an option that builds WHERE expressions
type ExprOption[T Table] func(*State, *ast.Expr)

// Apply implements the Option interface for ExprOption
func (opt ExprOption[T]) Apply(s *State, q *ast.Query) {
	sl := q.Query.(*ast.Select)
	if sl.Where == nil {
		sl.Where = &ast.Where{}
	}
	opt(s, &sl.Where.Expr)
}

// QueryOption represents an option that modifies the entire query
type QueryOption[T Table] func(*State, *ast.Query)

// Apply implements the Option interface for QueryOption
func (opt QueryOption[T]) Apply(s *State, q *ast.Query) {
	opt(s, q)
}

// AliasOption represents an alias specification for a JOIN operation
type AliasOption struct {
	Alias string
}

// As creates an AliasOption to specify a custom alias for a JOIN
func As(alias string) AliasOption {
	return AliasOption{Alias: alias}
}
