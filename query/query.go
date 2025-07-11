package query

import (
	"fmt"

	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/plate/types"
)

// logicalOp creates a logical operator condition (AND/OR) that groups multiple conditions
func logicalOp[T types.Table](op ast.BinaryOp, opts ...types.ExprOption[T]) types.ExprOption[T] {
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
					Op:    op,
					Left:  left,
					Right: right,
				},
			}
		}

		*expr = left
	}
}

// And creates an AND condition that groups multiple conditions
func And[T types.Table](opts ...types.ExprOption[T]) types.ExprOption[T] {
	return logicalOp(ast.OpAnd, opts...)
}

// Or creates an OR condition from multiple conditions
func Or[T types.Table](opts ...types.ExprOption[T]) types.ExprOption[T] {
	return logicalOp(ast.OpOr, opts...)
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

// buildDirectRelationshipWhere builds WHERE clause for direct relationships
func buildDirectRelationshipWhere(targetTable, baseAlias string, keys KeyPair) ast.Expr {
	return &ast.BinaryExpr{
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
}

// convertOptions converts typed options to untyped options for subquery context
func convertOptions[T types.Table](opts []types.Option[T]) []types.Option[types.Table] {
	result := make([]types.Option[types.Table], len(opts))
	for i, opt := range opts {
		result[i] = types.Option[types.Table](opt)
	}
	return result
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

// WithOne adds a single-value subquery column (for belongs_to relationships)
// Generates: (SELECT AS STRUCT t.* FROM t WHERE t.id = parent.foreign_key)
func WithOne[TBase types.Table, TTarget types.Table](
	relationshipName string,
	targetTable string,
	keys KeyPair,
	opts ...types.Option[TTarget],
) types.QueryOption[TBase] {
	return withSingleSubquery[TBase, TTarget](relationshipName, targetTable, keys, "", KeyPair{}, opts...)
}

// WithMany adds an array subquery column (for has_many relationships)
// Generates: ARRAY(SELECT AS STRUCT t.* FROM t WHERE t.foreign_key = parent.id)
func WithMany[TBase types.Table, TTarget types.Table](
	relationshipName string,
	targetTable string,
	keys KeyPair,
	opts ...types.Option[TTarget],
) types.QueryOption[TBase] {
	return withManySubquery[TBase, TTarget](relationshipName, targetTable, keys, opts...)
}

// WithManyThrough adds an array subquery column (for many_to_many relationships through a junction table)
// Generates: ARRAY(SELECT AS STRUCT t.* FROM t JOIN junction ON ... WHERE junction.foreign_key = parent.id)
func WithManyThrough[TBase types.Table, TTarget types.Table](
	relationshipName string,
	targetTable string,
	keys KeyPair,
	junctionTable string,
	junctionKeys KeyPair,
	opts ...types.Option[TTarget],
) types.QueryOption[TBase] {
	return withManyThroughSubquery[TBase, TTarget](relationshipName, targetTable, keys, junctionTable, junctionKeys, opts...)
}

// withSingleSubquery handles single-value subqueries (belongs_to relationships)
func withSingleSubquery[TBase types.Table, TTarget types.Table](
	relationshipName string,
	targetTable string,
	keys KeyPair,
	junctionTable string,
	junctionKeys KeyPair,
	opts ...types.Option[TTarget],
) types.QueryOption[TBase] {
	return func(s *types.State, q *ast.Query) {
		sq := newSubquery(s, targetTable, keys, junctionTable, junctionKeys)

		// Build basic subquery
		subQuery := sq.buildBasicSubquery([]ast.SelectItem{&ast.Star{}})

		// Apply options
		sq.applyOptions(subQuery, convertOptions(opts))

		// Create scalar subquery for single value
		subSelect := subQuery.Query.(*ast.Select)
		subqueryExpr := &ast.ScalarSubQuery{
			Query: &ast.Query{
				Query: &ast.Select{
					As:      &ast.AsStruct{},
					Results: subSelect.Results,
					From:    subSelect.From,
					Where:   subSelect.Where,
				},
			},
		}

		// Add to state's subquery columns
		s.SubqueryColumns = append(s.SubqueryColumns, types.SubqueryColumn{
			Alias:    relationshipName,
			Subquery: subqueryExpr,
		})
	}
}

// withManySubquery handles direct has_many array subqueries
func withManySubquery[TBase types.Table, TTarget types.Table](
	relationshipName string,
	targetTable string,
	keys KeyPair,
	opts ...types.Option[TTarget],
) types.QueryOption[TBase] {
	return func(s *types.State, q *ast.Query) {
		sq := newSubquery(s, targetTable, keys, "", KeyPair{})

		// Build basic subquery
		subQuery := sq.buildBasicSubquery([]ast.SelectItem{&ast.Star{}})

		// Apply options
		sq.applyOptions(subQuery, convertOptions(opts))

		// Create direct has_many array subquery
		subSelect := subQuery.Query.(*ast.Select)
		structSelect := &ast.Select{
			As:      &ast.AsStruct{},
			Results: []ast.SelectItem{&ast.Star{}},
			From: &ast.From{
				Source: &ast.TableName{
					Table: &ast.Ident{Name: targetTable},
				},
			},
			Where: subSelect.Where,
		}

		subqueryExpr := &ast.ArraySubQuery{
			Query: &ast.Query{
				Query: structSelect,
			},
		}

		// Add to state's subquery columns
		s.SubqueryColumns = append(s.SubqueryColumns, types.SubqueryColumn{
			Alias:    relationshipName,
			Subquery: subqueryExpr,
		})
	}
}

// withManyThroughSubquery handles many-to-many array subqueries through junction tables
func withManyThroughSubquery[TBase types.Table, TTarget types.Table](
	relationshipName string,
	targetTable string,
	keys KeyPair,
	junctionTable string,
	junctionKeys KeyPair,
	opts ...types.Option[TTarget],
) types.QueryOption[TBase] {
	return func(s *types.State, q *ast.Query) {
		sq := newSubquery(s, targetTable, keys, junctionTable, junctionKeys)

		// Build basic subquery
		subQuery := sq.buildBasicSubquery([]ast.SelectItem{&ast.Star{}})

		// Apply options
		sq.applyOptions(subQuery, convertOptions(opts))

		// Create many-to-many array subquery with JOIN
		subSelect := subQuery.Query.(*ast.Select)

		// Check for additional WHERE conditions from options
		var additionalWhere ast.Expr
		if subSelect.Where != nil {
			additionalWhere = subSelect.Where.Expr
		}

		structSelect := &ast.Select{
			As: &ast.AsStruct{},
			Results: []ast.SelectItem{
				&ast.DotStar{
					Expr: &ast.Path{
						Idents: []*ast.Ident{{Name: targetTable}},
					},
				},
			},
			From: &ast.From{
				Source: sq.buildJunctionJoin(),
			},
			Where: &ast.Where{
				Expr: sq.buildJunctionCorrelation(),
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

		subqueryExpr := &ast.ArraySubQuery{
			Query: &ast.Query{
				Query: structSelect,
			},
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
		sq := newSubquery(s, targetTable, keys, junctionTable, junctionKeys)

		// Create EXISTS subquery with SELECT 1
		selectItems := []ast.SelectItem{
			&ast.ExprSelectItem{
				Expr: &ast.IntLiteral{Value: "1"},
			},
		}

		var subQuery *ast.Query
		if junctionTable == "" {
			// Direct relationship
			subQuery = sq.buildBasicSubquery(selectItems)
		} else {
			// Junction relationship - need to handle differently
			subQuery = &ast.Query{
				Query: &ast.Select{
					Results: selectItems,
					From: &ast.From{
						Source: sq.buildJunctionJoin(),
					},
					Where: &ast.Where{
						Expr: sq.buildJunctionCorrelation(),
					},
				},
			}
		}

		// Apply options
		sq.applyOptions(subQuery, convertOptions(opts))

		// Create EXISTS expression
		*expr = &ast.ExistsSubQuery{
			Exists: 0, // Token position will be set by memefish
			Query:  subQuery,
		}
	}
}
