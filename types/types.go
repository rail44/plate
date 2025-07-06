package types

import (
	"strings"
	"github.com/cloudspannerecosystem/memefish/ast"
)

type State struct {
	Tables           map[string]struct{}
	Params           []any
	RelationshipPath []string // リレーションシップのパスを記録
}

// RegisterRelationship リレーションシップを辿ってエイリアスを登録
func (s *State) RegisterRelationship(relationshipName string) string {
	// パスに追加
	s.RelationshipPath = append(s.RelationshipPath, relationshipName)
	
	// CurrentAliasを使用してエイリアスを取得
	alias := s.CurrentAlias()
	
	s.Tables[alias] = struct{}{}
	return alias
}

// RegisterJunctionTable 中間テーブル用（パスを進めずにエイリアスだけ生成）
func (s *State) RegisterJunctionTable(tableName string) string {
	// 現在のエイリアスに基づいた中間テーブル名を生成
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

// CurrentAlias 現在のパスからエイリアスを取得
func (s *State) CurrentAlias() string {
	if len(s.RelationshipPath) == 0 {
		return ""
	}
	if len(s.RelationshipPath) == 1 {
		// ルートテーブルの場合
		return s.RelationshipPath[0]
	}
	// 2要素以上の場合は先頭を除いて結合
	return strings.Join(s.RelationshipPath[1:], "_")
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
