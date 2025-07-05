# Template-Based Implementation Plan

This document outlines the implementation plan for the template-based code generation approach.

## Overview

Based on our analysis, we'll implement a template-based code generator that creates query builders from schema definitions. This approach balances simplicity, maintainability, and flexibility.

## Project Structure

```
plate-generator/
├── generator/
│   ├── generator.go      # Main generator logic
│   ├── templates.go      # Template management
│   └── helpers.go        # Template helper functions
├── templates/
│   ├── package.tmpl      # Main package template
│   ├── select.tmpl       # Select function template
│   ├── filter.tmpl       # Column filter template
│   ├── join.tmpl         # Join function template
│   ├── common.tmpl       # Common query options
│   └── logical.tmpl      # Logical operators
└── example/
    └── schema.go         # Usage example
```

## Implementation Steps

### Step 1: Define Template Structure

#### Main Package Template (`package.tmpl`)
```go
package {{.PackageName}}

import (
    "fmt"
    "github.com/cloudspannerecosystem/memefish/ast"
    "github.com/rail44/hoge/common"
    {{- range .Imports}}
    "{{.}}"
    {{- end}}
)

type ExprOption func(*common.State, *ast.Expr)
type QueryOption func(*common.State, *ast.Query)

{{template "constants" .}}
{{template "select" .}}
{{range .Columns}}
{{- if .Filterable}}
{{template "filter" .}}
{{- end}}
{{- end}}
{{range .Relations}}
{{template "join" .}}
{{end}}
{{template "common" .}}
{{template "logical" .}}
```

### Step 2: Create Generator Interface

```go
// generator/types.go
type TableDefinition interface {
    PackageName() string
    TableName() string
    Columns() []ColumnDef
    Relations() []RelationDef
}

type ColumnDef interface {
    Name() string
    GoName() string
    GoType() string
    IsFilterable() bool
    IsSortable() bool
}

type RelationDef interface {
    Type() RelationType
    Name() string
    TargetTable() string
    LocalKey() string
    ForeignKey() string
}
```

### Step 3: Implement Template Engine

```go
// generator/generator.go
type Generator struct {
    templates *template.Template
}

func New() (*Generator, error) {
    tmpl := template.New("main").Funcs(template.FuncMap{
        "toGoName": toGoName,
        "sortableColumns": filterSortableColumns,
    })
    
    tmpl, err := tmpl.ParseGlob("templates/*.tmpl")
    if err != nil {
        return nil, err
    }
    
    return &Generator{templates: tmpl}, nil
}

func (g *Generator) Generate(table TableDefinition) ([]byte, error) {
    data := struct {
        PackageName string
        TableName   string
        Columns     []ColumnDef
        Relations   []RelationDef
        Imports     []string
    }{
        PackageName: table.PackageName(),
        TableName:   table.TableName(),
        Columns:     table.Columns(),
        Relations:   table.Relations(),
        Imports:     g.collectImports(table),
    }
    
    var buf bytes.Buffer
    if err := g.templates.ExecuteTemplate(&buf, "package.tmpl", data); err != nil {
        return nil, err
    }
    
    return format.Source(buf.Bytes())
}
```

### Step 4: Helper Functions

```go
// generator/helpers.go
func toGoName(s string) string {
    // Convert snake_case to PascalCase
    // user_id → UserID
    // created_at → CreatedAt
    parts := strings.Split(s, "_")
    for i, part := range parts {
        if isCommonAcronym(part) {
            parts[i] = strings.ToUpper(part)
        } else {
            parts[i] = strings.Title(part)
        }
    }
    return strings.Join(parts, "")
}

func filterSortableColumns(columns []ColumnDef) []ColumnDef {
    var sortable []ColumnDef
    for _, col := range columns {
        if col.IsSortable() {
            sortable = append(sortable, col)
        }
    }
    return sortable
}

func isCommonAcronym(s string) bool {
    acronyms := map[string]bool{
        "id": true,
        "url": true,
        "api": true,
        "uuid": true,
    }
    return acronyms[strings.ToLower(s)]
}
```

## Integration with Existing Spanner Schema

### Step 1: Create Adapter

```go
// adapter/spanner_adapter.go
type SpannerTableAdapter struct {
    table *generated.Table // From existing Spanner schema generation
}

func (s *SpannerTableAdapter) PackageName() string {
    return strings.ToLower(s.table.Name)
}

func (s *SpannerTableAdapter) TableName() string {
    return s.table.Name
}

func (s *SpannerTableAdapter) Columns() []generator.ColumnDef {
    var columns []generator.ColumnDef
    for _, col := range s.table.Columns {
        columns = append(columns, &SpannerColumnAdapter{col})
    }
    return columns
}
```

### Step 2: Usage Example

```go
// cmd/generate/main.go
func main() {
    gen, err := generator.New()
    if err != nil {
        log.Fatal(err)
    }
    
    // Use existing Spanner schema definitions
    tables := []*SpannerTableAdapter{
        {table: schema.UserTable},
        {table: schema.ProfileTable},
    }
    
    for _, table := range tables {
        code, err := gen.Generate(table)
        if err != nil {
            log.Fatal(err)
        }
        
        outputPath := filepath.Join("generated", table.PackageName(), table.TableName()+".go")
        if err := os.WriteFile(outputPath, code, 0644); err != nil {
            log.Fatal(err)
        }
    }
}
```

## Template Examples

### Column Filter Template (`filter.tmpl`)
```go
{{define "filter"}}
func {{.GoName}}(op ast.BinaryOp, value {{.GoType}}) ExprOption {
    return func(s *common.State, expr *ast.Expr) {
        i := len(s.Params)
        s.Params = append(s.Params, value)
        
        *expr = &ast.BinaryExpr{
            Left: &ast.Path{
                Idents: []*ast.Ident{
                    {Name: s.WorkingTableAlias},
                    {Name: "{{.Name}}"},
                },
            },
            Op: op,
            Right: &ast.Param{
                Name: fmt.Sprintf("p%d", i),
            },
        }
    }
}
{{end}}
```

### Constants Template (`constants.tmpl`)
```go
{{define "constants"}}
{{- $columns := sortableColumns .Columns}}
{{- if $columns}}
const (
{{- range $columns}}
    OrderBy{{.GoName}} = "{{.Name}}"
{{- end}}
)
{{- end}}
{{end}}
```

## Testing Strategy

1. **Unit Tests**: Test individual template functions
2. **Integration Tests**: Test full code generation
3. **Compilation Tests**: Ensure generated code compiles
4. **Example Tests**: Create example schemas and verify output

```go
// generator/generator_test.go
func TestGenerateUserTable(t *testing.T) {
    gen, err := New()
    require.NoError(t, err)
    
    table := &MockTableDefinition{
        name: "users",
        columns: []ColumnDef{
            {name: "id", goType: "int64", filterable: true, sortable: true},
            {name: "name", goType: "string", filterable: true, sortable: true},
        },
    }
    
    code, err := gen.Generate(table)
    require.NoError(t, err)
    
    // Verify generated code compiles
    _, err = format.Source(code)
    require.NoError(t, err)
    
    // Verify expected functions exist
    assert.Contains(t, string(code), "func Select(")
    assert.Contains(t, string(code), "func ID(")
    assert.Contains(t, string(code), "func Name(")
}
```

## Benefits of This Approach

1. **Simplicity**: Templates are easy to understand and modify
2. **Visibility**: Generated code structure is clear
3. **Maintainability**: Changes to patterns are straightforward
4. **Flexibility**: Easy to add new patterns or modify existing ones
5. **Integration**: Works seamlessly with existing Spanner schema generation

## Next Steps

1. Implement the basic generator with core templates
2. Add support for complex relationships (many-to-many, self-referential)
3. Add validation for generated code
4. Create comprehensive test suite
5. Document template customization options