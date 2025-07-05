# Code Generation Approaches

This document outlines different approaches for generating query builder code from schema definitions.

## Overview

Once we have schema definitions, we need to generate the actual Go code for the query builders. This document explores various code generation strategies.

## Approach 1: Direct Go AST Manipulation

### Description
Use Go's `go/ast` package to programmatically build the Abstract Syntax Tree and generate code.

### Implementation Example
```go
// generator/ast_builder.go
import (
    "go/ast"
    "go/token"
)

func buildSelectFunction(table Table) *ast.FuncDecl {
    return &ast.FuncDecl{
        Name: ast.NewIdent("Select"),
        Type: &ast.FuncType{
            Params: &ast.FieldList{
                List: []*ast.Field{{
                    Names: []*ast.Ident{ast.NewIdent("options")},
                    Type: &ast.Ellipsis{
                        Elt: ast.NewIdent("SelectOption"),
                    },
                }},
            },
            Results: &ast.FieldList{
                List: []*ast.Field{{
                    Type: ast.NewIdent("string"),
                }},
            },
        },
        Body: buildSelectBody(table),
    }
}
```

### Advantages
- **Full control**: Complete control over generated AST
- **Type safe**: Compile-time verification of AST structure
- **No external dependencies**: Uses only standard library

### Disadvantages
- **Verbose**: Requires lots of boilerplate code
- **Error prone**: Easy to make mistakes in AST construction
- **Steep learning curve**: Requires deep understanding of Go AST

## Approach 2: Template Engine Based (Recommended)

### Description
Use Go's `text/template` package to generate code from templates.

### Implementation Example
```go
// templates/query_builder.tmpl
func Select(options ...SelectOption) string {
    state := &common.State{
        Tables: map[string]string{},
        Params: []any{},
    }
    
    query := &ast.Select{
        Results: {{.SelectResults}},
        From: &ast.From{
            Source: &ast.TableName{
                Table: &ast.Ident{Name: "{{.TableName}}"},
            },
        },
    }
    
    for _, option := range options {
        option(query, state)
    }
    
    return query.SQL()
}
```

### Advantages
- **Readable**: Templates look like actual Go code
- **Easy to modify**: Changes to generated code are straightforward
- **Familiar**: Most developers understand template syntax
- **Separation of concerns**: Template logic separate from generation logic

### Disadvantages
- **String based**: No compile-time checking of generated code
- **Formatting required**: Need to run `gofmt` on output
- **Complex logic**: Advanced logic in templates can become unwieldy

## Approach 3: Intermediate Representation (IR) Based

### Description
Build an intermediate representation of the code structure, then generate code from the IR.

### Implementation Example
```go
// generator/ir.go
type QueryBuilder struct {
    Table    Table
    Columns  []Column
    Joins    []JoinDef
    Filters  []FilterDef
}

func (qb *QueryBuilder) GenerateAST() *ast.File {
    file := &ast.File{
        Name: ast.NewIdent(qb.Table.Name),
    }
    
    // Generate each function
    file.Decls = append(file.Decls, 
        qb.generateSelect(),
        qb.generateFilters(),
        qb.generateJoins(),
        qb.generateOrderBy(),
    )
    
    return file
}
```

### Advantages
- **Abstraction**: Separates code structure from generation details
- **Testable**: Can test IR transformation independently
- **Flexible**: Easy to support multiple output formats

### Disadvantages
- **Additional complexity**: Extra layer of abstraction
- **More code**: Need to maintain IR structures

## Approach 4: Code Builder Libraries

### Description
Use third-party libraries designed for code generation.

### Jennifer Library Example
```go
import "github.com/dave/jennifer/jen"

func generateColumnFilter(col ColumnIR) jen.Code {
    return jen.Func().Id(col.GoName).Params(
        jen.Id("op").Qual("github.com/cloudspannerecosystem/memefish/ast", "BinaryOp"),
        jen.Id("value").Id(col.GoType),
    ).Id("ExprOption").Block(
        jen.Return(jen.Func().Params(
            jen.Id("s").Op("*").Qual("common", "State"),
            jen.Id("expr").Op("*").Op("*").Qual("ast", "Expr"),
        ).Block(
            // Function body
        )),
    )
}
```

### Advantages
- **Type safe API**: Fluent API for building code
- **Import management**: Automatic handling of imports
- **Formatting**: Automatic code formatting

### Disadvantages
- **External dependency**: Adds third-party dependency
- **Learning curve**: Need to learn library API
- **Less transparent**: Generated code structure less obvious

## Hybrid Approach

Combine templates for simple patterns with code generation for complex structures:

```go
type Builder struct {
    // Use jennifer for complex structures
    file *jen.File
    
    // Use templates for repetitive patterns
    templates map[string]*template.Template
}

func (b *Builder) GenerateTable(table TableIR) error {
    // Package initialization
    b.file = jen.NewFile(table.PackageName)
    
    // Complex functions with jennifer
    b.file.Add(b.generateSelect(table))
    
    // Pattern-based functions with templates
    for _, col := range table.Columns {
        code := b.executeTemplate("columnFilter", col)
        b.file.Add(jen.Id(code))
    }
    
    return b.file.Save(table.PackageName + "/" + table.TableName + ".go")
}
```

## Recommendation

For this project, the **Template Engine Based approach** is recommended because:

1. **Simplicity**: Easy to understand and modify
2. **Visibility**: Generated code structure is clear in templates
3. **Maintenance**: Non-Go developers can understand templates
4. **Flexibility**: Easy to adjust generated code patterns
5. **Integration**: Works well with existing Spanner schema generation

The template approach provides the best balance of simplicity, maintainability, and flexibility for generating query builder code.