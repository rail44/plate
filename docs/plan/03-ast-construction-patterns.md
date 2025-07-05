# AST Construction Patterns

This document outlines the patterns used in the query builder for constructing SQL ASTs using the memefish library.

## Overview

The query builders construct SQL queries by building Abstract Syntax Trees (ASTs) using the memefish library. This document details the specific patterns that need to be generated.

## Core Type Definitions

Each generated package requires two fundamental type definitions:

```go
type ExprOption func(*common.State, *ast.Expr)
type QueryOption func(*common.State, *ast.Query)
```

- `ExprOption`: Used for building SQL expressions (WHERE conditions, JOIN conditions)
- `QueryOption`: Used for modifying the overall query structure (adding JOINs, ORDER BY, LIMIT)

## Pattern 1: Column Filter Functions

Each column in a table generates a filter function following this pattern:

```go
func ColumnName(op ast.BinaryOp, value ColumnType) ExprOption {
    return func(s *common.State, expr *ast.Expr) {
        i := len(s.Params)
        s.Params = append(s.Params, value)

        *expr = &ast.BinaryExpr{
            Left: &ast.Path{
                Idents: []*ast.Ident{
                    {Name: s.WorkingTableAlias},
                    {Name: "column_name"},
                },
            },
            Op: op,
            Right: &ast.Param{
                Name: fmt.Sprintf("p%d", i),
            },
        }
    }
}
```

### Key Elements:
- Parameter handling: Appends value to `s.Params` and generates parameter name
- Table aliasing: Uses `s.WorkingTableAlias` to qualify column names
- Expression building: Creates a binary expression with column path, operator, and parameter

## Pattern 2: Select Function

The main entry point for each table:

```go
func Select(opts ...QueryOption) string {
    tableName := "table_name"

    s := &common.State{
        Tables: make(map[string]struct{}),
        Params: []any{},
        WorkingTableAlias: tableName,
    }
    s.Tables[tableName] = struct{}{}

    stmt := ast.Select{
        Results: []ast.SelectItem {
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

    query := &ast.Query{
        Query: &stmt,
    }

    for _, opt := range opts {
        opt(s, query)
    }

    return query.SQL()
}
```

### Key Elements:
- State initialization with table tracking
- Basic SELECT * query structure
- Option application pattern
- SQL string generation

## Pattern 3: Join Functions

For handling table relationships:

```go
func JoinTargetTable(whereOpt targetpackage.ExprOption) QueryOption {
    return func(s *common.State, q *ast.Query) {
        sl := q.Query.(*ast.Select)
        
        // Generate unique table alias
        baseTableName := "target_table"
        tableName := baseTableName
        counter := 1
        
        for {
            if _, exists := s.Tables[tableName]; !exists {
                break
            }
            tableName = fmt.Sprintf("%s%d", baseTableName, counter)
            counter++
        }
        
        s.Tables[tableName] = struct{}{}
        
        // Save and switch working table context
        previousAlias := s.WorkingTableAlias
        
        // Build JOIN structure
        join := &ast.Join{
            Left: sl.From.Source,
            Op: ast.InnerJoin,
            Right: &ast.TableName{
                Table: &ast.Ident{
                    Name: tableName,
                },
            },
            Cond: &ast.On{
                Expr: &ast.BinaryExpr{
                    Left: &ast.Path{
                        Idents: []*ast.Ident{
                            {Name: "source_table"},
                            {Name: "foreign_key_column"},
                        },
                    },
                    Op: ast.OpEqual,
                    Right: &ast.Path{
                        Idents: []*ast.Ident{
                            {Name: tableName},
                            {Name: "target_key_column"},
                        },
                    },
                },
            },
        }
        
        sl.From.Source = join
        
        // Apply optional WHERE condition
        if whereOpt != nil {
            s.WorkingTableAlias = tableName
            
            if sl.Where == nil {
                sl.Where = &ast.Where{}
            }
            
            if sl.Where.Expr != nil {
                existingExpr := sl.Where.Expr
                and := &ast.BinaryExpr{
                    Op: ast.OpAnd,
                    Left: existingExpr,
                }
                whereOpt(s, &and.Right)
                sl.Where.Expr = &ast.ParenExpr{
                    Expr: and,
                }
            } else {
                whereOpt(s, &sl.Where.Expr)
            }
            
            s.WorkingTableAlias = previousAlias
        }
    }
}
```

### Key Elements:
- Alias conflict resolution (profile, profile1, profile2, etc.)
- Table context switching for proper column qualification
- JOIN condition building
- Optional WHERE clause integration

## Pattern 4: Query Modifier Functions

### WHERE Clause
```go
func Where(opt ExprOption) QueryOption {
    return func(s *common.State, q *ast.Query) {
        sl := q.Query.(*ast.Select)
        if sl.Where == nil {
            sl.Where = &ast.Where{}
        }
        opt(s, &sl.Where.Expr)
    }
}
```

### LIMIT Clause
```go
func Limit(count int) QueryOption {
    return func(s *common.State, q *ast.Query) {
        q.Limit = &ast.Limit{
            Count: &ast.IntLiteral{
                Value: fmt.Sprintf("%d", count),
            },
        }
    }
}
```

### ORDER BY Clause
```go
func OrderBy(column string, dir ast.Direction) QueryOption {
    return func(s *common.State, q *ast.Query) {
        if q.OrderBy == nil {
            q.OrderBy = &ast.OrderBy{
                Items: []*ast.OrderByItem{},
            }
        }
        
        q.OrderBy.Items = append(q.OrderBy.Items, &ast.OrderByItem{
            Expr: &ast.Path{
                Idents: []*ast.Ident{
                    {Name: s.WorkingTableAlias},
                    {Name: column},
                },
            },
            Dir: dir,
        })
    }
}
```

## Pattern 5: Logical Operators

### AND Operator
```go
func And(left, right ExprOption) ExprOption {
    return Paren(func(s *common.State, expr *ast.Expr) {
        b := &ast.BinaryExpr{
            Op: ast.OpAnd,
        }
        left(s, &b.Left)
        right(s, &b.Right)
        *expr = b
    })
}
```

### OR Operator
```go
func Or(left, right ExprOption) ExprOption {
    return Paren(func(s *common.State, expr *ast.Expr) {
        b := &ast.BinaryExpr{
            Op: ast.OpOr,
        }
        left(s, &b.Left)
        right(s, &b.Right)
        *expr = b
    })
}
```

### Parentheses Wrapper
```go
func Paren(inner ExprOption) ExprOption {
    return func(s *common.State, expr *ast.Expr) {
        paren := ast.ParenExpr{}
        inner(s, &paren.Expr)
        *expr = &paren
    }
}
```

## Pattern 6: Constants

For sortable columns, generate constants:

```go
const (
    OrderByID = "id"
    OrderByName = "name"
    OrderByCreatedAt = "created_at"
)
```

## State Management

The `common.State` structure is crucial for maintaining query context:

```go
type State struct {
    Tables            map[string]struct{} // Track used table aliases
    Params            []any              // Query parameters
    WorkingTableAlias string             // Current table context
}
```

### Key Responsibilities:
1. **Table alias tracking**: Prevents naming conflicts in JOINs
2. **Parameter management**: Sequential parameter naming (p0, p1, p2...)
3. **Context switching**: Maintains correct table qualification for columns

## Generation Requirements

When generating code for a new table:

1. Generate type definitions (`ExprOption`, `QueryOption`)
2. Generate `Select` function
3. For each filterable column: Generate filter function
4. For each sortable column: Generate OrderBy constant
5. For each relation: Generate JOIN function
6. Generate standard query modifiers (Where, Limit, OrderBy)
7. Generate logical operators (And, Or, Paren)

## Template Structure

The template-based generator should create these patterns with proper:
- Package naming
- Import statements
- Function naming (snake_case to PascalCase conversion)
- Type mapping (database types to Go types)
- Relationship handling (foreign key resolution)