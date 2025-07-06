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

```go
func OrderBy[V any](column Column[T, V], dir ast.Direction) QueryOption[T]
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

```go
func Limit(count int) QueryOption[T]
```

**Example:**
```go
user.Limit(10)    // LIMIT 10
```

### JOIN Type Control

```go
func WithInnerJoin() QueryOption[T]
```

Changes the most recent JOIN from LEFT OUTER JOIN to INNER JOIN.

**Example:**
```go
user.Posts(
    post.Title().Like("%important%"),
    post.WithInnerJoin(),  // Use INNER JOIN instead of LEFT OUTER JOIN
)
```

## Boolean Logic

### AND Operations
Multiple conditions passed to `Select` are combined with AND by default:

```go
query.Select[tables.User](
    user.Name().Eq("John"),        // Condition 1
    user.Email().IsNotNull(),      // AND Condition 2
    user.ID().Gt("100"),          // AND Condition 3
)
// WHERE user.name = @p0 AND user.email IS NOT NULL AND user.id > @p1
```

### OR Operations

```go
func Or(opts ...ExprOption[T]) ExprOption[T]
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

```go
func And(opts ...ExprOption[T]) ExprOption[T]
```

Used for grouping conditions within OR operations.

## Relationship Methods

### One-to-Many Relationships

```go
func Posts(opts ...Option[tables.Post]) QueryOption[tables.User]
```

**Default Behavior:** LEFT OUTER JOIN
**Examples:**
```go
// All users with their posts
user.Posts()

// Users with filtered posts
user.Posts(post.Title().Like("%important%"))

// Users with posts (INNER JOIN - only users who have posts)
user.Posts(post.WithInnerJoin())
```

### Many-to-Many Relationships

```go
func Tags(opts ...Option[tables.Tag]) QueryOption[tables.Post]
```

Joins through junction table (`post_tag`).

**Examples:**
```go
// Posts with any tags
post.Tags()

// Posts with specific tags
post.Tags(tag.Name().Eq("Go"))

// Posts with multiple tag conditions
post.Tags(tag.Or(
    tag.Name().Eq("Go"),
    tag.Name().Eq("Tutorial"),
))
```

### Belongs-To Relationships

```go
func Author(opts ...Option[tables.User]) QueryOption[tables.Post]
```

**Default Behavior:** INNER JOIN (foreign key relationship)

**Examples:**
```go
// Posts with author information
post.Author()

// Posts by specific authors
post.Author(user.Email().Like("%@company.com"))
```

## Multi-level JOINs

Relationships can be nested for complex queries:

```go
// User → Posts → Tags (3-level JOIN)
query.Select[tables.User](
    user.Posts(
        post.Tags(tag.Name().Eq("Go")),
        post.Title().Like("%tutorial%"),
    ),
)
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
query.Select[tables.User](
    post.Title().Eq("Test"),    // Cannot use Post expression in User query
)
```

## Select Function

```go
func Select[T Table](opts ...Option[T]) (string, []any)
```

**Generic Parameter:** `T` - Table type for type safety
**Returns:** 
- `string` - Generated SQL query
- `[]any` - Parameter values for prepared statement

**Examples:**
```go
// Basic select
sql, params := query.Select[tables.User](
    user.Name().Eq("John"),
)

// Complex select with multiple options
sql, params := query.Select[tables.User](
    user.Name().Like("J%"),
    user.Posts(post.Title().Like("%important%")),
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
query.Select[tables.User](
    user.Name().Like("J%"),     // AND
    user.Email().IsNotNull(),   // AND  
    user.ID().Gt("100"),       // AND
)
```

### 3. JOIN Type Control
```go
// Default: LEFT OUTER JOIN (includes users without posts)
user.Posts()

// Explicit: INNER JOIN (only users with posts)
user.Posts(post.WithInnerJoin())
```

### 4. Complex Conditions
Build complex boolean logic with explicit grouping:
```go
user.Or(
    user.And(condition1, condition2),
    user.And(condition3, condition4),
)
```