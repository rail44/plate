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

// buildAndApplyOptions builds basic subquery and applies options
func (sq *subquery) buildAndApplyOptions(opts []types.Option[types.Table]) *ast.Query {
	subQuery := sq.buildBasicSubquery([]ast.SelectItem{&ast.Star{}})
	sq.applyOptions(subQuery, opts)
	return subQuery
}

// addSubqueryColumn adds a subquery column to the parent state
func (sq *subquery) addSubqueryColumn(s *types.State, relationshipName string, subqueryExpr ast.Expr) {
	s.SubqueryColumns = append(s.SubqueryColumns, types.SubqueryColumn{
		Alias:    relationshipName,
		Subquery: subqueryExpr,
	})
}

// extractAdditionalWhere extracts WHERE conditions from options
func (sq *subquery) extractAdditionalWhere(subQuery *ast.Query) ast.Expr {
	subSelect := subQuery.Query.(*ast.Select)
	if subSelect.Where != nil {
		return subSelect.Where.Expr
	}
	return nil
}

// buildScalarSubquery builds a scalar subquery for belongs_to relationships
func (sq *subquery) buildScalarSubquery(subQuery *ast.Query) ast.Expr {
	subSelect := subQuery.Query.(*ast.Select)
	return &ast.ScalarSubQuery{
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

// buildArraySubquery builds an array subquery for has_many relationships
func (sq *subquery) buildArraySubquery(subQuery *ast.Query) ast.Expr {
	subSelect := subQuery.Query.(*ast.Select)
	structSelect := &ast.Select{
		As:      &ast.AsStruct{},
		Results: []ast.SelectItem{&ast.Star{}},
		From: &ast.From{
			Source: &ast.TableName{
				Table: &ast.Ident{Name: sq.targetTable},
			},
		},
		Where: subSelect.Where,
	}

	return &ast.ArraySubQuery{
		Query: &ast.Query{
			Query: structSelect,
		},
	}
}

// buildArrayThroughSubquery builds an array subquery for many_to_many relationships
func (sq *subquery) buildArrayThroughSubquery(subQuery *ast.Query, additionalWhere ast.Expr) ast.Expr {
	structSelect := &ast.Select{
		As: &ast.AsStruct{},
		Results: []ast.SelectItem{
			&ast.DotStar{
				Expr: &ast.Path{
					Idents: []*ast.Ident{{Name: sq.targetTable}},
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

	return &ast.ArraySubQuery{
		Query: &ast.Query{
			Query: structSelect,
		},
	}
}
