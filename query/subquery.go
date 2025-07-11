package query

import (
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/plate/types"
)

// subquery holds common data for building subqueries
type subquery struct {
	parentState *types.State
	subState    *types.State
	baseAlias   string
	targetTable string
	keys        KeyPair
}

// newSubquery creates a new subquery builder
func newSubquery(parentState *types.State, targetTable string, keys KeyPair) *subquery {
	sq := &subquery{
		parentState: parentState,
		baseAlias:   parentState.CurrentAlias(),
		targetTable: targetTable,
		keys:        keys,
	}
	sq.subState = parentState.NewSubqueryState(targetTable)
	return sq
}

// buildJunctionCorrelation builds the WHERE clause for junction table correlation
func (sq *subquery) buildJunctionCorrelation(junctionTable string) ast.Expr {
	return &ast.BinaryExpr{
		Left: &ast.Path{
			Idents: []*ast.Ident{
				{Name: junctionTable},
				{Name: sq.keys.To},
			},
		},
		Op: ast.OpEqual,
		Right: &ast.Path{
			Idents: []*ast.Ident{
				{Name: sq.baseAlias},
				{Name: sq.keys.From},
			},
		},
	}
}

// buildJunctionJoin builds the JOIN clause for junction tables
func (sq *subquery) buildJunctionJoin(junctionTable string, junctionKeys KeyPair) *ast.Join {
	return &ast.Join{
		Op: ast.InnerJoin,
		Left: &ast.TableName{
			Table: &ast.Ident{Name: sq.targetTable},
		},
		Right: &ast.TableName{
			Table: &ast.Ident{Name: junctionTable},
		},
		Cond: &ast.On{
			Expr: &ast.BinaryExpr{
				Left: &ast.Path{
					Idents: []*ast.Ident{
						{Name: sq.targetTable},
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
}

// buildBasicSubquery creates a basic subquery with FROM and WHERE
func (sq *subquery) buildBasicSubquery(selectItems []ast.SelectItem) *ast.Query {
	query := &ast.Query{
		Query: &ast.Select{
			Results: selectItems,
			From: &ast.From{
				Source: &ast.TableName{
					Table: &ast.Ident{Name: sq.targetTable},
				},
			},
		},
	}

	// Add WHERE clause for direct relationships
	whereExpr := &ast.BinaryExpr{
		Left: &ast.Path{
			Idents: []*ast.Ident{
				{Name: sq.targetTable},
				{Name: sq.keys.To},
			},
		},
		Op: ast.OpEqual,
		Right: &ast.Path{
			Idents: []*ast.Ident{
				{Name: sq.baseAlias},
				{Name: sq.keys.From},
			},
		},
	}
	query.Query.(*ast.Select).Where = &ast.Where{
		Expr: whereExpr,
	}

	return query
}

// applyOptions applies the given options to the subquery
func (sq *subquery) applyOptions(subQuery *ast.Query, opts []types.Option[types.Table]) {
	for _, opt := range opts {
		opt.Apply(sq.subState, subQuery)
	}
	// Update parent state params
	sq.parentState.Params = sq.subState.Params
}

// addSubqueryColumn adds a subquery column to the parent state
func (sq *subquery) addSubqueryColumn(s *types.State, relationshipName string, subqueryExpr ast.Expr) {
	s.SubqueryColumns = append(s.SubqueryColumns, types.SubqueryColumn{
		Alias:    relationshipName,
		Subquery: subqueryExpr,
	})
}
