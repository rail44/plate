package query

import (
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/plate/types"
)

// subquery holds common data for building subqueries
type subquery struct {
	parentState   *types.State
	subState      *types.State
	baseAlias     string
	targetTable   string
	keys          KeyPair
	junctionTable string
	junctionKeys  KeyPair
}

// newSubquery creates a new subquery builder
func newSubquery(parentState *types.State, targetTable string, keys KeyPair, junctionTable string, junctionKeys KeyPair) *subquery {
	sq := &subquery{
		parentState:   parentState,
		baseAlias:     parentState.CurrentAlias(),
		targetTable:   targetTable,
		keys:          keys,
		junctionTable: junctionTable,
		junctionKeys:  junctionKeys,
	}
	sq.subState = parentState.NewSubqueryState(targetTable)
	return sq
}

// buildDirectRelationshipWhere builds WHERE clause for direct relationships
func (sq *subquery) buildDirectRelationshipWhere() ast.Expr {
	if sq.junctionTable != "" {
		return nil // Junction relationships don't use direct WHERE
	}
	return buildDirectRelationshipWhere(sq.targetTable, sq.baseAlias, sq.keys)
}

// buildJunctionCorrelation builds the WHERE clause for junction table correlation
func (sq *subquery) buildJunctionCorrelation() ast.Expr {
	if sq.junctionTable == "" {
		return nil
	}
	return &ast.BinaryExpr{
		Left: &ast.Path{
			Idents: []*ast.Ident{
				{Name: sq.junctionTable},
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
func (sq *subquery) buildJunctionJoin() *ast.Join {
	if sq.junctionTable == "" {
		return nil
	}
	return &ast.Join{
		Op: ast.InnerJoin,
		Left: &ast.TableName{
			Table: &ast.Ident{Name: sq.targetTable},
		},
		Right: &ast.TableName{
			Table: &ast.Ident{Name: sq.junctionTable},
		},
		Cond: &ast.On{
			Expr: &ast.BinaryExpr{
				Left: &ast.Path{
					Idents: []*ast.Ident{
						{Name: sq.targetTable},
						{Name: sq.junctionKeys.To},
					},
				},
				Op: ast.OpEqual,
				Right: &ast.Path{
					Idents: []*ast.Ident{
						{Name: sq.junctionTable},
						{Name: sq.junctionKeys.From},
					},
				},
			},
		},
	}
}

// buildBasicSubquery creates a basic subquery with FROM and WHERE
func (sq *subquery) buildBasicSubquery(selectItems []ast.SelectItem) *ast.Query {
	whereExpr := sq.buildDirectRelationshipWhere()

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

	if whereExpr != nil {
		query.Query.(*ast.Select).Where = &ast.Where{
			Expr: whereExpr,
		}
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
