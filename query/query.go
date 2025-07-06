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
	}
	s.Tables[tableName] = struct{}{}

	stmt := ast.Select{
		Results: []ast.SelectItem{
			&ast.Star{},
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

	for _, opt := range opts {
		opt.Apply(s, q)
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
		Keys: KeyPair{
			From: config.BaseToJunction.From,
			To:   config.BaseToJunction.To,
		},
		JoinType: config.JoinType,
	})

	// Then create target table join using junction join as source
	return join(joinConfig{
		Source:      junctionJoin,
		BaseTable:   config.JunctionAlias,
		TargetTable: config.TargetTable,
		TargetAlias: config.TargetAlias,
		Keys: KeyPair{
			From: config.JunctionToTarget.From,
			To:   config.JunctionToTarget.To,
		},
		JoinType: config.JoinType,
	})
}

// FindLastJoin recursively finds the last JOIN in the FROM clause
func FindLastJoin(source ast.TableExpr) *ast.Join {
	if join, ok := source.(*ast.Join); ok {
		// Check if the right side has more joins
		if rightJoin := FindLastJoin(join.Right); rightJoin != nil {
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

		// Wrap with NOT
		*expr = &ast.UnaryExpr{
			Op:   ast.OpNot,
			Expr: *expr,
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
			if join := FindLastJoin(sl.From.Source); join != nil {
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
