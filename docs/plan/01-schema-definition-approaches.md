# Schema Definition Approaches

This document outlines different approaches for defining table schemas and relationships for the query builder code generation.

## Overview

The goal is to provide a way to define database table structures, columns, and relationships that can be used as input for generating type-safe SQL query builders.

## Approach 1: Schema Definition DSL + go generate (Recommended)

### Description
Define schemas using Go structs with a Domain-Specific Language (DSL) approach, then use `go generate` to create the query builders.

### Implementation Example
```go
// schema.go
type Schema struct {
    Tables []Table
}

type Table struct {
    Name    string
    Columns []Column
    Relations []Relation
}

type Column struct {
    Name string
    Type ColumnType
    Tags []string // primary, unique, nullable, etc.
}

type Relation struct {
    Type       RelationType // HasOne, HasMany, BelongsTo
    Target     string       // table name
    ForeignKey string
    LocalKey   string
}
```

### Advantages
- **Centralized schema management**: All table definitions in one place
- **Flexible extensibility**: Can design DSL freely to express complex relationships
- **Migration integration**: Can generate migration files from schema definitions
- **Language agnostic**: Schema definitions can be reused for other language code generation
- **Version control friendly**: Easy to track schema change history

### Disadvantages
- **Learning curve**: Need to learn custom DSL syntax
- **Limited editor support**: Custom DSL lacks code completion and type checking
- **Double management risk**: Need to sync actual DB schema with code definitions

## Approach 2: Struct Tag Based

### Description
Use Go struct tags to define schema information, similar to popular ORMs like GORM.

### Implementation Example
```go
type User struct {
    ID    int64  `plate:"primary"`
    Name  string `plate:"index"`
    Email string `plate:"unique"`
    
    // Relation definitions
    Profile *Profile `plate:"has_one,fk:user_id"`
}

type Profile struct {
    ID     int64  `plate:"primary"`
    UserID int64  `plate:"fk:users.id"`
    Bio    string
}
```

### Advantages
- **Go idiomatic**: Same pattern as existing tools
- **Type safe**: Compile-time error detection
- **Rich IDE support**: Auto-completion, refactoring, go-to-definition work

### Disadvantages
- **Scattered definitions**: Table definitions spread across multiple files
- **Limited expressiveness**: Complex relationships hard to express with tags alone
- **Circular reference issues**: Difficult to define mutually referencing structs

## Approach 3: Fluent API Definition

### Description
Define schemas using a fluent API with method chaining.

### Implementation Example
```go
// definition.go
var UserTable = plate.DefineTable("users").
    Column("id", plate.Int64, plate.Primary()).
    Column("name", plate.String).
    Column("email", plate.String, plate.Unique()).
    HasOne("profile", "profiles", "user_id")

var ProfileTable = plate.DefineTable("profiles").
    Column("id", plate.Int64, plate.Primary()).
    Column("user_id", plate.Int64).
    Column("bio", plate.Text).
    BelongsTo("user", "users", "user_id")
```

### Advantages
- **Type-safe DSL**: Leverages Go's type system
- **Intuitive definition**: Method chaining feels natural

### Disadvantages
- **Runtime definition**: Schema determined at runtime rather than compile time
- **Verbosity**: Each column definition can become lengthy

## Approach 4: Configuration File Based

### Description
Define schemas in YAML, TOML, or JSON configuration files.

### Implementation Example
```yaml
# schema.yaml
tables:
  - name: users
    columns:
      - name: id
        type: int64
        primary: true
      - name: name
        type: string
      - name: email
        type: string
        unique: true
    relations:
      - type: has_one
        name: profile
        table: profiles
        foreign_key: user_id
```

### Advantages
- **Language independent**: YAML/JSON readable by anyone
- **External tool integration**: Easy to share schema with other tools

### Disadvantages
- **No type safety**: Typos and type errors only detected at runtime
- **Difficult refactoring**: No IDE support

## Hybrid Approach (Best of Both Worlds)

Combine the advantages of Schema Definition DSL while mitigating disadvantages:

```go
// schema/schema.go
package schema

import "github.com/rail44/plate/dsl"

var Schema = dsl.Schema{
    Tables: []dsl.Table{
        {
            Name: "users",
            Columns: []dsl.Column{
                {Name: "id", Type: dsl.Int64, Primary: true},
                {Name: "name", Type: dsl.String, Index: true},
                {Name: "email", Type: dsl.String, Unique: true},
            },
            Relations: []dsl.Relation{
                {Type: dsl.HasOne, Name: "profile", Target: "profiles", ForeignKey: "user_id"},
            },
        },
    },
}
```

This approach provides:
- **Type safety**: Defined as Go structs
- **IDE support**: Code completion and error checking work
- **Centralized management**: Entire schema in one place
- **Extensibility**: Can add methods to DSL structs for additional functionality

## Recommendation for Spanner Integration

Given the existing Spanner schema definition → domain model generation flow, the pure Schema Definition DSL approach (Approach 1) is most suitable because:

1. **Single Source of Truth**: Spanner schema remains the authoritative source
2. **Automatic synchronization**: Schema changes automatically reflect in query builders
3. **Type consistency**: Domain models and query builders use same type definitions
4. **Leverage existing assets**: Reuse already generated metadata

The integration would look like:
```
Spanner Schema → Domain Model Generation
      ↓
   Schema Info
      ↓
Plate Code Generation → Query Builders
```