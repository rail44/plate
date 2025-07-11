package types

import (
	"fmt"
	"github.com/cloudspannerecosystem/memefish/ast"
	"strings"
)

type State struct {
	Tables           map[string]struct{}
	Params           []any
	RelationshipPath []string         // Track relationship path for alias generation
	SubqueryColumns  []SubqueryColumn // Track subqueries to add to SELECT
}

// SubqueryColumn represents a column that will be added to SELECT as a subquery
type SubqueryColumn struct {
	Alias    string
	Subquery ast.Expr
}

// registerRelationship registers a relationship and generates an alias
func (s *State) registerRelationship(relationshipName string) string {
	// Add to path
	s.RelationshipPath = append(s.RelationshipPath, relationshipName)

	// Get alias using CurrentAlias
	alias := s.CurrentAlias()

	s.Tables[alias] = struct{}{}
	return alias
}

// RegisterJunction registers a junction table alias without advancing the path
func (s *State) RegisterJunction(tableName string) string {
	// Generate junction table alias based on current alias
	currentAlias := s.CurrentAlias()
	alias := ""
	if currentAlias != "" {
		alias = currentAlias + "_" + tableName
	} else {
		alias = tableName
	}

	s.Tables[alias] = struct{}{}
	return alias
}

// CurrentAlias gets the current alias from the relationship path
func (s *State) CurrentAlias() string {
	if len(s.RelationshipPath) == 0 {
		return ""
	}
	if len(s.RelationshipPath) == 1 {
		// Root table case
		return s.RelationshipPath[0]
	}
	// For 2+ elements, join all except the first
	return strings.Join(s.RelationshipPath[1:], "_")
}

// WithRelationship executes a function within a relationship scope
// Note: This method is primarily intended for use in JOIN method implementations
func (s *State) WithRelationship(relationshipName string, fn func(alias string)) {
	// Save current path
	basePath := make([]string, len(s.RelationshipPath))
	copy(basePath, s.RelationshipPath)

	// Register relationship
	alias := s.registerRelationship(relationshipName)

	// Execute function in scope
	fn(alias)

	// Restore path
	s.RelationshipPath = basePath
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

	var expr ast.Expr
	opt(s, &expr)

	if sl.Where.Expr == nil {
		sl.Where.Expr = expr
	} else {
		// Combine with existing WHERE using AND
		sl.Where.Expr = &ast.BinaryExpr{
			Op:    ast.OpAnd,
			Left:  sl.Where.Expr,
			Right: expr,
		}
	}
}

// QueryOption represents an option that modifies the entire query
type QueryOption[T Table] func(*State, *ast.Query)

// Apply implements the Option interface for QueryOption
func (opt QueryOption[T]) Apply(s *State, q *ast.Query) {
	opt(s, q)
}

// Column represents a table column that can be used in various SQL contexts
type Column[T Table, V any] struct {
	Name string
}

// Op creates a condition using the specified operator and value
func (c Column[T, V]) Op(op ast.BinaryOp, value V) ExprOption[T] {
	return func(s *State, expr *ast.Expr) {
		i := len(s.Params)
		s.Params = append(s.Params, value)
		*expr = &ast.BinaryExpr{
			Left: &ast.Path{
				Idents: []*ast.Ident{
					{Name: s.CurrentAlias()},
					{Name: c.Name},
				},
			},
			Op: op,
			Right: &ast.Param{
				Name: fmt.Sprintf("p%d", i),
			},
		}
	}
}

// Eq creates an equality condition (=)
func (c Column[T, V]) Eq(value V) ExprOption[T] {
	return c.Op(ast.OpEqual, value)
}

// Ne creates a not equal condition (!=)
func (c Column[T, V]) Ne(value V) ExprOption[T] {
	return c.Op(ast.OpNotEqual, value)
}

// Lt creates a less than condition (<)
func (c Column[T, V]) Lt(value V) ExprOption[T] {
	return c.Op(ast.OpLess, value)
}

// Gt creates a greater than condition (>)
func (c Column[T, V]) Gt(value V) ExprOption[T] {
	return c.Op(ast.OpGreater, value)
}

// Le creates a less than or equal condition (<=)
func (c Column[T, V]) Le(value V) ExprOption[T] {
	return c.Op(ast.OpLessEqual, value)
}

// Ge creates a greater than or equal condition (>=)
func (c Column[T, V]) Ge(value V) ExprOption[T] {
	return c.Op(ast.OpGreaterEqual, value)
}

// Like creates a LIKE condition (string columns only)
func (c Column[T, string]) Like(value string) ExprOption[T] {
	return c.Op(ast.OpLike, value)
}

// NotLike creates a NOT LIKE condition (string columns only)
func (c Column[T, string]) NotLike(value string) ExprOption[T] {
	return c.Op(ast.OpNotLike, value)
}

// Numeric column specific methods

// Between creates a BETWEEN condition (comparable types only)
func (c Column[T, V]) Between(min, max V) ExprOption[T] {
	return func(s *State, expr *ast.Expr) {
		minIdx := len(s.Params)
		s.Params = append(s.Params, min)
		maxIdx := len(s.Params)
		s.Params = append(s.Params, max)

		*expr = &ast.BetweenExpr{
			Not: false,
			Left: &ast.Path{
				Idents: []*ast.Ident{
					{Name: s.CurrentAlias()},
					{Name: c.Name},
				},
			},
			RightStart: &ast.Param{Name: fmt.Sprintf("p%d", minIdx)},
			RightEnd:   &ast.Param{Name: fmt.Sprintf("p%d", maxIdx)},
		}
	}
}

// In creates an IN condition
func (c Column[T, V]) In(values ...V) ExprOption[T] {
	return func(s *State, expr *ast.Expr) {
		var exprs []ast.Expr
		for _, value := range values {
			i := len(s.Params)
			s.Params = append(s.Params, value)
			exprs = append(exprs, &ast.Param{Name: fmt.Sprintf("p%d", i)})
		}

		*expr = &ast.InExpr{
			Not: false,
			Left: &ast.Path{
				Idents: []*ast.Ident{
					{Name: s.CurrentAlias()},
					{Name: c.Name},
				},
			},
			Right: &ast.ValuesInCondition{
				Exprs: exprs,
			},
		}
	}
}

// IsNull creates an IS NULL condition
func (c Column[T, V]) IsNull() ExprOption[T] {
	return func(s *State, expr *ast.Expr) {
		*expr = &ast.IsNullExpr{
			Not: false,
			Left: &ast.Path{
				Idents: []*ast.Ident{
					{Name: s.CurrentAlias()},
					{Name: c.Name},
				},
			},
		}
	}
}

// IsNotNull creates an IS NOT NULL condition
func (c Column[T, V]) IsNotNull() ExprOption[T] {
	return func(s *State, expr *ast.Expr) {
		*expr = &ast.IsNullExpr{
			Not: true,
			Left: &ast.Path{
				Idents: []*ast.Ident{
					{Name: s.CurrentAlias()},
					{Name: c.Name},
				},
			},
		}
	}
}
