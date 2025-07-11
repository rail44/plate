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
		Tables:           make(map[string]struct{}),
		Params:           []any{},
		RelationshipPath: []string{tableName}, // Start from root table
		SubqueryColumns:  []types.SubqueryColumn{},
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

// joinConfig contains configuration for building JOIN clauses
type joinConfig struct {
	Source      ast.TableExpr
	BaseTable   string
	TargetTable string
	TargetAlias string
	Keys        KeyPair
	JoinType    ast.JoinOp // JOIN type (INNER, LEFT OUTER, etc.)
}

// joinThroughConfig contains configuration for building many-to-many JOIN through a junction table
type joinThroughConfig struct {
	Source           ast.TableExpr
	BaseTable        string
	JunctionTable    string
	JunctionAlias    string
	TargetTable      string
	TargetAlias      string
	BaseToJunction   KeyPair
	JunctionToTarget KeyPair
	JoinType         ast.JoinOp
}

// joinThrough creates a many-to-many JOIN through a junction table
func joinThrough(config joinThroughConfig) *ast.Join {
	// First create junction table join
	junctionJoin := join(joinConfig{
		Source:      config.Source,
		BaseTable:   config.BaseTable,
		TargetTable: config.JunctionTable,
		TargetAlias: config.JunctionAlias,
		Keys:        config.BaseToJunction,
		JoinType:    config.JoinType,
	})

	// Then create target table join using junction join as source
	return join(joinConfig{
		Source:      junctionJoin,
		BaseTable:   config.JunctionAlias,
		TargetTable: config.TargetTable,
		TargetAlias: config.TargetAlias,
		Keys:        config.JunctionToTarget,
		JoinType:    config.JoinType,
	})
}

// findLastJoin recursively finds the last JOIN in the FROM clause
func findLastJoin(source ast.TableExpr) *ast.Join {
	if join, ok := source.(*ast.Join); ok {
		// Check if the right side has more joins
		if rightJoin := findLastJoin(join.Right); rightJoin != nil {
			return rightJoin
		}
		return join
	}
	return nil
}

// join creates a JOIN AST node
func join(config joinConfig) *ast.Join {
	return &ast.Join{
		Left: config.Source,
		Op:   config.JoinType,
		Right: &ast.TableName{
			Table: &ast.Ident{
				Name: config.TargetTable,
			},
			As: &ast.AsAlias{
				Alias: &ast.Ident{
					Name: config.TargetAlias,
				},
			},
		},
		Cond: &ast.On{
			Expr: &ast.BinaryExpr{
				Left: &ast.Path{
					Idents: []*ast.Ident{
						{Name: config.BaseTable},
						{Name: config.Keys.From},
					},
				},
				Op: ast.OpEqual,
				Right: &ast.Path{
					Idents: []*ast.Ident{
						{Name: config.TargetAlias},
						{Name: config.Keys.To},
					},
				},
			},
		},
	}
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

// WithInnerJoinOption changes the JOIN type to INNER JOIN for any table type
func WithInnerJoinOption[T types.Table]() types.QueryOption[T] {
	return func(s *types.State, q *ast.Query) {
		// Find the last JOIN and change its type
		if sl, ok := q.Query.(*ast.Select); ok {
			if join := findLastJoin(sl.From.Source); join != nil {
				join.Op = ast.InnerJoin
			}
		}
	}
}

// DirectJoin creates a direct JOIN relationship (one-to-many or belongs-to)
func DirectJoin[TBase types.Table, TTarget types.Table](
	relationshipName string,
	targetTable string,
	keys KeyPair,
	joinType ast.JoinOp,
	opts ...types.Option[TTarget],
) types.QueryOption[TBase] {
	return func(s *types.State, q *ast.Query) {
		sl := q.Query.(*ast.Select)
		baseAlias := s.CurrentAlias()

		s.WithRelationship(relationshipName, func(alias string) {
			// Create and apply JOIN
			sl.From.Source = join(joinConfig{
				Source:      sl.From.Source,
				BaseTable:   baseAlias,
				TargetTable: targetTable,
				TargetAlias: alias,
				Keys:        keys,
				JoinType:    joinType,
			})

			// Apply options
			for _, opt := range opts {
				opt.Apply(s, q)
			}
		})
	}
}

// JunctionJoin creates a many-to-many JOIN through a junction table
func JunctionJoin[TBase types.Table, TTarget types.Table](
	relationshipName string,
	junctionTable string,
	baseToJunction KeyPair,
	targetTable string,
	junctionToTarget KeyPair,
	joinType ast.JoinOp,
	opts ...types.Option[TTarget],
) types.QueryOption[TBase] {
	return func(s *types.State, q *ast.Query) {
		sl := q.Query.(*ast.Select)
		baseAlias := s.CurrentAlias()
		junctionAlias := s.RegisterJunction(junctionTable)

		s.WithRelationship(relationshipName, func(targetAlias string) {
			// Create and apply JOIN
			sl.From.Source = joinThrough(joinThroughConfig{
				Source:           sl.From.Source,
				BaseTable:        baseAlias,
				JunctionTable:    junctionTable,
				JunctionAlias:    junctionAlias,
				TargetTable:      targetTable,
				TargetAlias:      targetAlias,
				BaseToJunction:   baseToJunction,
				JunctionToTarget: junctionToTarget,
				JoinType:         joinType,
			})

			// Apply options
			for _, opt := range opts {
				opt.Apply(s, q)
			}
		})
	}
}

// WithSubquery adds a subquery column to the SELECT clause
// For belongs_to: (SELECT AS STRUCT t.* FROM t WHERE ...)
// For has_many/many_to_many: (SELECT ARRAY_AGG(AS STRUCT t.*) FROM t WHERE ...)
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
			Tables:           make(map[string]struct{}),
			Params:           s.Params, // Share params with parent
			RelationshipPath: []string{targetTable},
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
			// Many-to-many through junction
			whereExpr = &ast.InExpr{
				Left: &ast.Path{
					Idents: []*ast.Ident{
						{Name: targetTable},
						{Name: junctionKeys.To},
					},
				},
				Right: &ast.SubQueryInCondition{
					Query: &ast.Query{
						Query: &ast.Select{
							Results: []ast.SelectItem{
								&ast.ExprSelectItem{
									Expr: &ast.Path{
										Idents: []*ast.Ident{
											{Name: junctionTable},
											{Name: junctionKeys.From},
										},
									},
								},
							},
							From: &ast.From{
								Source: &ast.TableName{
									Table: &ast.Ident{Name: junctionTable},
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
						},
					},
				},
			}
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
				Where: &ast.Where{
					Expr: whereExpr,
				},
			},
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
			// ARRAY_AGG for has_many and many_to_many
			// Create SELECT AS STRUCT for the inner query
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

			// Create ARRAY_AGG call
			arrayAggExpr := &ast.CallExpr{
				Func: &ast.Path{
					Idents: []*ast.Ident{{Name: "ARRAY_AGG"}},
				},
				Args: []ast.Arg{
					&ast.ExprArg{
						Expr: &ast.ScalarSubQuery{
							Query: &ast.SubQuery{
								Query: &ast.Query{
									Query: structSelect,
								},
							},
						},
					},
				},
			}

			// Wrap in scalar subquery
			subqueryExpr = &ast.ScalarSubQuery{
				Query: &ast.SubQuery{
					Query: &ast.Query{
						Query: &ast.Select{
							Results: []ast.SelectItem{
								&ast.ExprSelectItem{
									Expr: arrayAggExpr,
								},
							},
						},
					},
				},
			}
		} else {
			// Single STRUCT for belongs_to
			subqueryExpr = &ast.ScalarSubQuery{
				Query: &ast.SubQuery{
					Query: &ast.Query{
						Query: &ast.Select{
							As:      &ast.AsStruct{},
							Results: subSelect.Results,
							From:    subSelect.From,
							Where:   subSelect.Where,
						},
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
			Tables:           make(map[string]struct{}),
			Params:           s.Params, // Share params with parent
			RelationshipPath: []string{targetTable},
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
			Query: &ast.SubQuery{
				Query: subQuery,
			},
		}
	}
}
