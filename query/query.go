package query

import (
	"fmt"

	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/plate/types"
)

// And creates an AND condition that groups multiple conditions
func And[T types.Table](opts ...types.ExprOption[T]) types.ExprOption[T] {
	return func(s *types.State, expr *ast.Expr) {
		if len(opts) == 0 {
			return
		}

		var left ast.Expr
		opts[0](s, &left)

		for i := 1; i < len(opts); i++ {
			var right ast.Expr
			opts[i](s, &right)
			left = &ast.ParenExpr{
				Expr: &ast.BinaryExpr{
					Op:    ast.OpAnd,
					Left:  left,
					Right: right,
				},
			}
		}

		*expr = left
	}
}

// Or creates an OR condition from multiple conditions
func Or[T types.Table](opts ...types.ExprOption[T]) types.ExprOption[T] {
	return func(s *types.State, expr *ast.Expr) {
		if len(opts) == 0 {
			return
		}

		var left ast.Expr
		opts[0](s, &left)

		for i := 1; i < len(opts); i++ {
			var right ast.Expr
			opts[i](s, &right)
			left = &ast.ParenExpr{
				Expr: &ast.BinaryExpr{
					Op:    ast.OpOr,
					Left:  left,
					Right: right,
				},
			}
		}

		*expr = left
	}
}

// Select is a generic select function for any table
func Select[T types.Table](opts ...types.Option[T]) (string, []any) {
	var t T
	tableName := t.TableName()

	s := &types.State{
		Tables:          make(map[string]struct{}),
		Params:          []any{},
		CurrentTable:    tableName,
		SubqueryColumns: []types.SubqueryColumn{},
	}
	s.Tables[tableName] = struct{}{}

	// Start with table.* instead of just *
	stmt := ast.Select{
		Results: []ast.SelectItem{
			&ast.DotStar{
				Expr: &ast.Path{
					Idents: []*ast.Ident{
						{Name: tableName},
					},
				},
			},
		},
		From: &ast.From{
			Source: &ast.TableName{
				Table: &ast.Ident{
					Name: tableName,
				},
			},
		},
	}

	q := &ast.Query{
		Query: &stmt,
	}

	// Apply all options
	for _, opt := range opts {
		opt.Apply(s, q)
	}

	// Add subquery columns to SELECT
	if len(s.SubqueryColumns) > 0 {
		for _, col := range s.SubqueryColumns {
			stmt.Results = append(stmt.Results, &ast.Alias{
				Expr: col.Subquery,
				As: &ast.AsAlias{
					Alias: &ast.Ident{Name: col.Alias},
				},
			})
		}
	}

	return q.SQL(), s.Params
}

// KeyPair represents a relationship between two tables through their keys
type KeyPair struct {
	From string // Key from the source table
	To   string // Key in the target table
}

// OrderBy creates an ORDER BY clause for any table type
func OrderBy[T types.Table, V any](column types.Column[T, V], dir ast.Direction) types.QueryOption[T] {
	return func(s *types.State, q *ast.Query) {
		if q.OrderBy == nil {
			q.OrderBy = &ast.OrderBy{
				Items: []*ast.OrderByItem{},
			}
		}
		q.OrderBy.Items = append(q.OrderBy.Items, &ast.OrderByItem{
			Expr: &ast.Path{
				Idents: []*ast.Ident{
					{Name: s.CurrentAlias()},
					{Name: column.Name},
				},
			},
			Dir: dir,
		})
	}
}

// Not creates a logical NOT condition that wraps any ExprOption
// This allows negation of complex expressions including And() and Or() combinations
func Not[T types.Table](opt types.ExprOption[T]) types.ExprOption[T] {
	return func(s *types.State, expr *ast.Expr) {
		// Build the inner expression first
		opt(s, expr)

		// Check if the expression already has parentheses
		// ParenExpr means it's already wrapped (from And/Or)
		if _, isParenExpr := (*expr).(*ast.ParenExpr); isParenExpr {
			// Already has parentheses, just wrap with NOT
			*expr = &ast.UnaryExpr{
				Op:   ast.OpNot,
				Expr: *expr,
			}
		} else {
			// Simple expression, add parentheses for clarity
			*expr = &ast.UnaryExpr{
				Op: ast.OpNot,
				Expr: &ast.ParenExpr{
					Expr: *expr,
				},
			}
		}
	}
}

// Limit creates a LIMIT clause for any table type
func Limit[T types.Table](count int) types.QueryOption[T] {
	return func(s *types.State, q *ast.Query) {
		q.Limit = &ast.Limit{
			Count: &ast.IntLiteral{
				Value: fmt.Sprintf("%d", count),
			},
		}
	}
}

// WithSubquery adds a subquery column to the SELECT clause
// For belongs_to: (SELECT AS STRUCT t.* FROM t WHERE ...)
// For has_many/many_to_many: ARRAY(SELECT AS STRUCT t.* FROM t WHERE ...)
func WithSubquery[TBase types.Table, TTarget types.Table](
	relationshipName string,
	targetTable string,
	keys KeyPair,
	isArray bool,
	junctionTable string, // empty for direct relationships
	junctionKeys KeyPair, // empty for direct relationships
	opts ...types.Option[TTarget],
) types.QueryOption[TBase] {
	return func(s *types.State, q *ast.Query) {
		baseAlias := s.CurrentAlias()

		// Create a new state for the subquery
		subState := &types.State{
			Tables:       make(map[string]struct{}),
			Params:       s.Params, // Share params with parent
			CurrentTable: targetTable,
		}
		subState.Tables[targetTable] = struct{}{}

		// Build WHERE conditions for correlation
		var whereExpr ast.Expr

		if junctionTable == "" {
			// Direct relationship (belongs_to or has_many)
			whereExpr = &ast.BinaryExpr{
				Left: &ast.Path{
					Idents: []*ast.Ident{
						{Name: targetTable},
						{Name: keys.To},
					},
				},
				Op: ast.OpEqual,
				Right: &ast.Path{
					Idents: []*ast.Ident{
						{Name: baseAlias},
						{Name: keys.From},
					},
				},
			}
		} else {
			// Many-to-many through junction - no WHERE needed here since we'll use JOIN
			whereExpr = nil
		}

		// Apply options to add more WHERE conditions
		subQuery := &ast.Query{
			Query: &ast.Select{
				Results: []ast.SelectItem{
					&ast.Star{},
				},
				From: &ast.From{
					Source: &ast.TableName{
						Table: &ast.Ident{Name: targetTable},
					},
				},
			},
		}

		// Add WHERE clause only if we have a where expression
		if whereExpr != nil {
			subQuery.Query.(*ast.Select).Where = &ast.Where{
				Expr: whereExpr,
			}
		}

		// Apply options
		for _, opt := range opts {
			opt.Apply(subState, subQuery)
		}

		// Update parent state params
		s.Params = subState.Params

		// Create the final subquery expression
		var subqueryExpr ast.Expr
		subSelect := subQuery.Query.(*ast.Select)

		if isArray {
			// ARRAY for has_many and many_to_many
			var structSelect *ast.Select

			if junctionTable == "" {
				// Direct has_many relationship
				structSelect = &ast.Select{
					As:      &ast.AsStruct{},
					Results: []ast.SelectItem{&ast.Star{}},
					From: &ast.From{
						Source: &ast.TableName{
							Table: &ast.Ident{Name: targetTable},
						},
					},
					Where: subSelect.Where,
				}
			} else {
				// Many-to-many through junction table - use JOIN for efficiency
				// First, check if there are additional WHERE conditions from options
				var additionalWhere ast.Expr
				if subSelect.Where != nil {
					additionalWhere = subSelect.Where.Expr
				}

				structSelect = &ast.Select{
					As: &ast.AsStruct{},
					Results: []ast.SelectItem{
						&ast.DotStar{
							Expr: &ast.Path{
								Idents: []*ast.Ident{{Name: targetTable}},
							},
						},
					},
					From: &ast.From{
						Source: &ast.Join{
							Op: ast.InnerJoin,
							Left: &ast.TableName{
								Table: &ast.Ident{Name: targetTable},
							},
							Right: &ast.TableName{
								Table: &ast.Ident{Name: junctionTable},
							},
							Cond: &ast.On{
								Expr: &ast.BinaryExpr{
									Left: &ast.Path{
										Idents: []*ast.Ident{
											{Name: targetTable},
											{Name: junctionKeys.To},
										},
									},
									Op: ast.OpEqual,
									Right: &ast.Path{
										Idents: []*ast.Ident{
											{Name: junctionTable},
											{Name: junctionKeys.From},
										},
									},
								},
							},
						},
					},
					Where: &ast.Where{
						Expr: &ast.BinaryExpr{
							Left: &ast.Path{
								Idents: []*ast.Ident{
									{Name: junctionTable},
									{Name: keys.To},
								},
							},
							Op: ast.OpEqual,
							Right: &ast.Path{
								Idents: []*ast.Ident{
									{Name: baseAlias},
									{Name: keys.From},
								},
							},
						},
					},
				}

				// Apply additional WHERE conditions from options
				if additionalWhere != nil {
					structSelect.Where.Expr = &ast.BinaryExpr{
						Op:    ast.OpAnd,
						Left:  structSelect.Where.Expr,
						Right: additionalWhere,
					}
				}
			}

			// Use ArraySubQuery for cleaner SQL
			subqueryExpr = &ast.ArraySubQuery{
				Query: &ast.Query{
					Query: structSelect,
				},
			}
		} else {
			// Single STRUCT for belongs_to
			subqueryExpr = &ast.ScalarSubQuery{
				Query: &ast.Query{
					Query: &ast.Select{
						As:      &ast.AsStruct{},
						Results: subSelect.Results,
						From:    subSelect.From,
						Where:   subSelect.Where,
					},
				},
			}
		}

		// Add to state's subquery columns
		s.SubqueryColumns = append(s.SubqueryColumns, types.SubqueryColumn{
			Alias:    relationshipName,
			Subquery: subqueryExpr,
		})
	}
}

// WhereExists creates a WHERE EXISTS condition for filtering parent rows
// based on related child rows
func WhereExists[TBase types.Table, TTarget types.Table](
	targetTable string,
	keys KeyPair,
	junctionTable string, // empty for direct relationships
	junctionKeys KeyPair, // empty for direct relationships
	opts ...types.Option[TTarget],
) types.ExprOption[TBase] {
	return func(s *types.State, expr *ast.Expr) {
		baseAlias := s.CurrentAlias()

		// Create a new state for the subquery
		subState := &types.State{
			Tables:       make(map[string]struct{}),
			Params:       s.Params, // Share params with parent
			CurrentTable: targetTable,
		}
		subState.Tables[targetTable] = struct{}{}

		// Build WHERE conditions for correlation
		var whereExpr ast.Expr
		var junctionCond ast.Expr // Declare at function scope

		if junctionTable == "" {
			// Direct relationship (belongs_to or has_many)
			whereExpr = &ast.BinaryExpr{
				Left: &ast.Path{
					Idents: []*ast.Ident{
						{Name: targetTable},
						{Name: keys.To},
					},
				},
				Op: ast.OpEqual,
				Right: &ast.Path{
					Idents: []*ast.Ident{
						{Name: baseAlias},
						{Name: keys.From},
					},
				},
			}
		} else {
			// Many-to-many through junction
			// First create the junction condition
			junctionCond = &ast.BinaryExpr{
				Left: &ast.Path{
					Idents: []*ast.Ident{
						{Name: junctionTable},
						{Name: keys.To},
					},
				},
				Op: ast.OpEqual,
				Right: &ast.Path{
					Idents: []*ast.Ident{
						{Name: baseAlias},
						{Name: keys.From},
					},
				},
			}

			// Then create the target-junction join condition
			targetJoinCond := &ast.BinaryExpr{
				Left: &ast.Path{
					Idents: []*ast.Ident{
						{Name: targetTable},
						{Name: junctionKeys.To},
					},
				},
				Op: ast.OpEqual,
				Right: &ast.Path{
					Idents: []*ast.Ident{
						{Name: junctionTable},
						{Name: junctionKeys.From},
					},
				},
			}

			// Combine both conditions
			whereExpr = &ast.BinaryExpr{
				Op:    ast.OpAnd,
				Left:  junctionCond,
				Right: targetJoinCond,
			}
		}

		// Create EXISTS subquery
		subQuery := &ast.Query{
			Query: &ast.Select{
				Results: []ast.SelectItem{
					&ast.ExprSelectItem{
						Expr: &ast.IntLiteral{Value: "1"},
					},
				},
				From: &ast.From{
					Source: &ast.TableName{
						Table: &ast.Ident{Name: targetTable},
					},
				},
				Where: &ast.Where{
					Expr: whereExpr,
				},
			},
		}

		// Add junction table to FROM if needed
		if junctionTable != "" {
			subSelect := subQuery.Query.(*ast.Select)
			subSelect.From.Source = &ast.Join{
				Op: ast.InnerJoin,
				Left: &ast.TableName{
					Table: &ast.Ident{Name: targetTable},
				},
				Right: &ast.TableName{
					Table: &ast.Ident{Name: junctionTable},
				},
				Cond: &ast.On{
					Expr: &ast.BinaryExpr{
						Left: &ast.Path{
							Idents: []*ast.Ident{
								{Name: targetTable},
								{Name: junctionKeys.To},
							},
						},
						Op: ast.OpEqual,
						Right: &ast.Path{
							Idents: []*ast.Ident{
								{Name: junctionTable},
								{Name: junctionKeys.From},
							},
						},
					},
				},
			}

			// Update WHERE to only include the base-junction condition
			subSelect.Where.Expr = junctionCond
		}

		// Apply options
		for _, opt := range opts {
			opt.Apply(subState, subQuery)
		}

		// Update parent state params
		s.Params = subState.Params

		// Create EXISTS expression
		*expr = &ast.ExistsSubQuery{
			Exists: 0, // Token position will be set by memefish
			Query:  subQuery,
		}
	}
}
