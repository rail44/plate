package query

import (
	"github.com/cloudspannerecosystem/memefish/ast"
	"github.com/rail44/plate/types"
)

// subqueryContext holds common data for building subqueries
type subqueryContext struct {
	parentState   *types.State
	baseAlias     string
	targetTable   string
	keys          KeyPair
	junctionTable string
	junctionKeys  KeyPair
}

// newSubqueryContext creates a new subquery context
func newSubqueryContext(parentState *types.State, targetTable string, keys KeyPair, junctionTable string, junctionKeys KeyPair) *subqueryContext {
	return &subqueryContext{
		parentState:   parentState,
		baseAlias:     parentState.CurrentAlias(),
		targetTable:   targetTable,
		keys:          keys,
		junctionTable: junctionTable,
		junctionKeys:  junctionKeys,
	}
}

// createSubState creates a new state for the subquery
func (ctx *subqueryContext) createSubState() *types.State {
	return ctx.parentState.NewSubqueryState(ctx.targetTable)
}

// buildDirectRelationshipWhere builds WHERE clause for direct relationships
func (ctx *subqueryContext) buildDirectRelationshipWhere() ast.Expr {
	if ctx.junctionTable != "" {
		return nil // Junction relationships don't use direct WHERE
	}
	return buildDirectRelationshipWhere(ctx.targetTable, ctx.baseAlias, ctx.keys)
}

// buildJunctionCorrelation builds the WHERE clause for junction table correlation
func (ctx *subqueryContext) buildJunctionCorrelation() ast.Expr {
	if ctx.junctionTable == "" {
		return nil
	}
	return &ast.BinaryExpr{
		Left: &ast.Path{
			Idents: []*ast.Ident{
				{Name: ctx.junctionTable},
				{Name: ctx.keys.To},
			},
		},
		Op: ast.OpEqual,
		Right: &ast.Path{
			Idents: []*ast.Ident{
				{Name: ctx.baseAlias},
				{Name: ctx.keys.From},
			},
		},
	}
}

// buildJunctionJoin builds the JOIN clause for junction tables
func (ctx *subqueryContext) buildJunctionJoin() *ast.Join {
	if ctx.junctionTable == "" {
		return nil
	}
	return &ast.Join{
		Op: ast.InnerJoin,
		Left: &ast.TableName{
			Table: &ast.Ident{Name: ctx.targetTable},
		},
		Right: &ast.TableName{
			Table: &ast.Ident{Name: ctx.junctionTable},
		},
		Cond: &ast.On{
			Expr: &ast.BinaryExpr{
				Left: &ast.Path{
					Idents: []*ast.Ident{
						{Name: ctx.targetTable},
						{Name: ctx.junctionKeys.To},
					},
				},
				Op: ast.OpEqual,
				Right: &ast.Path{
					Idents: []*ast.Ident{
						{Name: ctx.junctionTable},
						{Name: ctx.junctionKeys.From},
					},
				},
			},
		},
	}
}

// buildBasicSubquery creates a basic subquery with FROM and WHERE
func (ctx *subqueryContext) buildBasicSubquery(selectItems []ast.SelectItem) *ast.Query {
	whereExpr := ctx.buildDirectRelationshipWhere()

	query := &ast.Query{
		Query: &ast.Select{
			Results: selectItems,
			From: &ast.From{
				Source: &ast.TableName{
					Table: &ast.Ident{Name: ctx.targetTable},
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
func (ctx *subqueryContext) applyOptions(subState *types.State, subQuery *ast.Query, opts []types.Option[types.Table]) {
	for _, opt := range opts {
		opt.Apply(subState, subQuery)
	}
	// Update parent state params
	ctx.parentState.Params = subState.Params
}
