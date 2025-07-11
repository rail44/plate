# API Reference

This document provides comprehensive reference for Plate's type-safe query builder API.

## Core Types

### Column[T Table, V any]

The `Column` type represents a typed table column that provides compile-time type safety.

```go
type Column[T Table, V any] struct {
    Name string
}
```

**Type Parameters:**
- `T`: Table type (e.g., `tables.User`, `tables.Post`)
- `V`: Value type (e.g., `string`, `int`, `time.Time`)

**Type Safety:**
- Column operations are constrained by the value type `V`
- String-specific methods are only available for `Column[T, string]`
- Compile-time type checking prevents mismatched value types

## Column Operations

### Universal Methods
Available for all column types `Column[T, V]`:

#### Comparison Operations
```go
func (c Column[T, V]) Eq(value V) ExprOption[T]      // =
func (c Column[T, V]) Ne(value V) ExprOption[T]      // !=
func (c Column[T, V]) Lt(value V) ExprOption[T]      // <
func (c Column[T, V]) Gt(value V) ExprOption[T]      // >
func (c Column[T, V]) Le(value V) ExprOption[T]      // <=
func (c Column[T, V]) Ge(value V) ExprOption[T]      // >=
```

**Examples:**
```go
user.ID().Eq("123")           // user.id = @p0
user.Name().Ne("Admin")       // user.name != @p0
post.ID().Gt("100")          // post.id > @p0
```

#### Range and List Operations
```go
func (c Column[T, V]) Between(min, max V) ExprOption[T]
func (c Column[T, V]) In(values ...V) ExprOption[T]
```

**Examples:**
```go
user.ID().Between("100", "200")              // user.id BETWEEN @p0 AND @p1
user.Name().In("John", "Jane", "Bob")        // user.name IN (@p0, @p1, @p2)
```

#### NULL Operations
```go
func (c Column[T, V]) IsNull() ExprOption[T]
func (c Column[T, V]) IsNotNull() ExprOption[T]
```

**Examples:**
```go
user.Email().IsNull()        // user.email IS NULL
user.Email().IsNotNull()     // user.email IS NOT NULL
```

### String-Only Methods
Available only for string columns `Column[T, string]`:

```go
func (c Column[T, string]) Like(pattern string) ExprOption[T]
func (c Column[T, string]) NotLike(pattern string) ExprOption[T]
```

**Examples:**
```go
user.Name().Like("John%")           // user.name LIKE @p0
user.Email().NotLike("%spam%")      // user.email NOT LIKE @p0

// ❌ Compile error - Like not available for non-string columns
// user.CreatedAt().Like("2023%")   // time.Time column
```

## Query Options

### Ordering

Each generated table package provides its own OrderBy function:

```go
// In user package
func OrderBy[V any](column types.Column[tables.User, V], dir ast.Direction) types.QueryOption[tables.User]
```

**Direction Constants:**
- `ast.DirectionAsc` - Ascending order  
- `ast.DirectionDesc` - Descending order

**Examples:**
```go
user.OrderBy(user.Name(), ast.DirectionAsc)      // ORDER BY user.name ASC
post.OrderBy(post.Title(), ast.DirectionDesc)    // ORDER BY post.title DESC
```

### Limiting

Each generated table package provides its own Limit function:

```go
// In user package
func Limit(count int) types.QueryOption[tables.User]
```

**Example:**
```go
user.Limit(10)    // LIMIT 10
```

## Boolean Logic

### AND Operations
Multiple conditions passed to `Select` are combined with AND by default:

```go
user.Select(
    user.Name().Eq("John"),        // Condition 1
    user.Email().IsNotNull(),      // AND Condition 2
    user.ID().Gt("100"),          // AND Condition 3
)
// WHERE user.name = @p0 AND user.email IS NOT NULL AND user.id > @p1
```

### OR Operations

Each generated table package provides its own Or function:

```go
// In user package
func Or(opts ...types.ExprOption[tables.User]) types.ExprOption[tables.User]
```

**Examples:**
```go
// Simple OR
user.Or(
    user.Name().Eq("John"),
    user.Name().Eq("Jane"),
)
// WHERE (user.name = @p0 OR user.name = @p1)

// Complex boolean logic: (a AND b) OR (c AND d)
user.Or(
    user.And(
        user.Name().Eq("John"),
        user.Email().Like("%@example.com"),
    ),
    user.And(
        user.Name().Eq("Jane"),
        user.Email().Like("%@company.com"),
    ),
)
```

### AND Operations (Explicit)

Each generated table package provides its own And function:

```go
// In user package
func And(opts ...types.ExprOption[tables.User]) types.ExprOption[tables.User]
```

Used for grouping conditions within OR operations.

### NOT Operations

Each generated table package provides its own Not function:

```go
// In user package
func Not(opt types.ExprOption[tables.User]) types.ExprOption[tables.User]
```

Creates a logical NOT condition that wraps any ExprOption. This allows negation of complex expressions including And() and Or() combinations.

**Examples:**
```go
// Simple NOT
user.Not(user.Name().Like("Admin%"))
// WHERE NOT (user.name LIKE @p0)

// NOT with OR
user.Not(user.Or(
    user.Name().Eq("John"),
    user.Name().Eq("Jane"),
))
// WHERE NOT (user.name = @p0 OR user.name = @p1)

// Complex NOT with AND/OR
user.Not(user.And(
    user.Name().Like("Admin%"),
    user.Email().Like("%@admin.com"),
))
// WHERE NOT (user.name LIKE @p0 AND user.email LIKE @p1)
```

## Relationship Methods

Plate generates two types of relationship methods:

### WithXxx Methods (Subqueries)
These methods add related data to the SELECT clause as nested structures.

#### One-to-Many Relationships

```go
// In user package
func WithPosts(opts ...types.Option[tables.Post]) types.QueryOption[tables.User]
```

**Behavior:** Adds posts as nested array using ARRAY(SELECT AS STRUCT ...)

**Examples:**
```go
// All users with their posts
user.WithPosts()
// Generates: SELECT user.*, ARRAY(SELECT AS STRUCT * FROM post WHERE post.user_id = user.id) AS posts FROM user

// Users with filtered posts
user.WithPosts(post.Title().Like("%important%"))
// Generates: SELECT user.*, ARRAY(SELECT AS STRUCT * FROM post WHERE post.user_id = user.id AND post.title LIKE @p0) AS posts FROM user
```

#### Many-to-Many Relationships

```go
// In post package
func WithTags(opts ...types.Option[tables.Tag]) types.QueryOption[tables.Post]
```

Joins through junction table (`post_tag`).

**Examples:**
```go
// Posts with any tags
post.WithTags()
// Generates: SELECT post.*, ARRAY(SELECT AS STRUCT tag.* FROM tag INNER JOIN post_tag ON tag.id = post_tag.tag_id WHERE post_tag.post_id = post.id) AS tags FROM post

// Posts with specific tags
post.WithTags(tag.Name().Eq("Go"))
```

#### Belongs-To Relationships

```go
// In post package
func WithAuthor(opts ...types.Option[tables.User]) types.QueryOption[tables.Post]
```

**Behavior:** Adds single related record as nested struct

**Examples:**
```go
// Posts with author information
post.WithAuthor()
// Generates: SELECT post.*, (SELECT AS STRUCT * FROM user WHERE user.id = post.user_id) AS author FROM post

// Posts by specific authors
post.WithAuthor(user.Email().Like("%@company.com"))
```

### WhereXxx Methods (EXISTS filtering)
These methods filter parent records based on child conditions.

```go
// In user package
func WherePosts(opts ...types.Option[tables.Post]) types.ExprOption[tables.User]

// In post package
func WhereAuthor(opts ...types.Option[tables.User]) types.ExprOption[tables.Post]
func WhereTags(opts ...types.Option[tables.Tag]) types.ExprOption[tables.Post]
```

**Examples:**
```go
// Users who have posts with "important" in title
user.Select(
    user.WherePosts(post.Title().Like("%important%")),
)
// Generates: SELECT user.* FROM user WHERE EXISTS(SELECT 1 FROM post WHERE post.user_id = user.id AND post.title LIKE @p0)

// Posts that have "Go" tag
post.Select(
    post.WhereTags(tag.Name().Eq("Go")),
)
// Generates: SELECT post.* FROM post WHERE EXISTS(SELECT 1 FROM tag INNER JOIN post_tag ON tag.id = post_tag.tag_id WHERE post_tag.post_id = post.id AND tag.name = @p0)
```

## Multi-level Relationships

Relationships can be nested for complex queries:

```go
// User → Posts (with tags loaded)
user.Select(
    user.WithPosts(
        post.WithTags(tag.Name().Eq("Go")),
        post.Title().Like("%tutorial%"),
    ),
)
// This loads users with their posts, and each post includes its tags
```

## Type Safety Features

### Compile-Time Constraints

1. **Method Availability**: String methods only work with string columns
2. **Value Type Matching**: Column operations require matching value types
3. **Table Isolation**: Cannot mix expressions from different tables

### Common Compile Errors

```go
// ❌ Method not available for type
user.ID().Like("123")           // Like only for string columns

// ❌ Type mismatch
user.Name().Eq(123)             // Cannot pass int to string column

// ❌ Table type mismatch  
user.Select(
    post.Title().Eq("Test"),    // Cannot use Post expression in User query
)
```

## Select Function

Each generated table package provides its own Select function:

```go
// In user package
func Select(opts ...types.Option[tables.User]) (string, []any)

// In post package  
func Select(opts ...types.Option[tables.Post]) (string, []any)
```

**Returns:** 
- `string` - Generated SQL query
- `[]any` - Parameter values for prepared statement

**Examples:**
```go
// Basic select
sql, params := user.Select(
    user.Name().Eq("John"),
)

// Complex select with multiple options
sql, params := user.Select(
    user.Name().Like("J%"),
    user.WithPosts(post.Title().Like("%important%")),
    user.OrderBy(user.Name(), ast.DirectionAsc),
    user.Limit(10),
)
```

## Best Practices

### 1. Use Type Constraints
Leverage the type system to catch errors at compile time:
```go
// Good: Type-safe string operations
user.Name().Like("John%")

// Avoid: Raw SQL strings
// WHERE name LIKE 'John%'  -- Not type-safe
```

### 2. Explicit OR, Implicit AND
```go
// Clear intent with explicit OR
user.Or(user.Name().Eq("John"), user.Name().Eq("Jane"))

// Implicit AND for multiple conditions
user.Select(
    user.Name().Like("J%"),     // AND
    user.Email().IsNotNull(),   // AND  
    user.ID().Gt("100"),       // AND
)
```

### 3. Subqueries vs EXISTS Filtering
```go
// Load related data (subquery in SELECT)
user.WithPosts()  // Returns users with posts array

// Filter by relationship (EXISTS in WHERE)
user.WherePosts(post.Title().Like("%important%"))  // Only users who have important posts
```

### 4. Complex Conditions
Build complex boolean logic with explicit grouping:
```go
user.Or(
    user.And(condition1, condition2),
    user.And(condition3, condition4),
)
```

## Code Generator API

### GeneratorConfig

```go
type GeneratorConfig struct {
    Tables    []TableConfig
    Junctions []JunctionConfig
}
```

The main configuration structure for the code generator.

### TableConfig

```go
type TableConfig struct {
    Schema    TableSchema
    Relations []Relation
}
```

Represents a table that needs a query builder.

### JunctionConfig

```go
type JunctionConfig struct {
    Schema    TableSchema
    Relations []Relation // Must have exactly 2 relations
}
```

Represents a junction table for many-to-many relationships.

### TableSchema

```go
type TableSchema struct {
    TableName string      // Database table name (e.g., "user")
    Model     interface{} // Model instance (e.g., models.User{})
}
```

Contains the basic information about a table.

### Relation

```go
type Relation struct {
    Name        string // Relationship name (e.g., "Author")
    Target      string // Target table name (e.g., "User")
    From        string // Source column (e.g., "UserID")
    To          string // Target column (e.g., "ID")
    ReverseName string // Optional: Name for the reverse relation
}
```

Represents a relationship between tables.

### Generator Methods

```go
func NewGenerator() *Generator
```

Creates a new Generator instance.

```go
func (g *Generator) Generate(config GeneratorConfig, outputDir string) (GeneratedFiles, error)
```

Generates query builder code based on the provided configuration.

**Parameters:**
- `config`: The generator configuration
- `outputDir`: Directory where generated files will be written

**Returns:**
- `GeneratedFiles`: Map of relative file paths to file contents
- `error`: Any error that occurred during generation

### Usage Example

```go
generator := plate.NewGenerator()
files, err := generator.Generate(config, "./generated")
if err != nil {
    log.Fatal(err)
}

// Write generated files to disk
for path, content := range files.Files {
    // Write content to path
}
```