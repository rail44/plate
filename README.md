# Plate - Type-Safe SQL Query Builder for Go

Plate is a type-safe SQL query builder for Go, designed to work with Google Cloud Spanner and the memefish SQL parser. It combines compile-time type safety with an intuitive API that prevents SQL injection and runtime errors.

## Overview

Plate provides strongly-typed query builders with compile-time safety for SQL query construction. The API uses a combination of method chaining for column operations and functional options for query structure.

```go
// Type-safe query construction with fluent Column API
sql, params := query.Select[tables.User](
    user.Name().Like("John%"),
    user.Email().Ne("admin@example.com"),
    user.ID().Between("100", "200"),
    user.OrderBy(user.Name(), ast.DirectionAsc),
    user.Limit(10),
)
// sql: SELECT * FROM user WHERE user.name LIKE @p0 AND user.email != @p1 AND user.id BETWEEN @p2 AND @p3 ORDER BY user.name ASC LIMIT 10
// params: []any{"John%", "admin@example.com", "100", "200"}
```

## Key Features

### 🔒 **Compile-Time Type Safety**
- **Column Type Constraints**: String-specific methods like `Like()` only work with string columns
- **Value Type Matching**: Cannot pass wrong type values to column operations
- **Table Type Isolation**: Cannot mix expressions from different tables

```go
// ✅ Compile-time safe
user.Name().Like("John%")        // Like works with string columns
user.ID().Between("1", "100")    // Correct string type matching

// ❌ Compile-time errors
// user.ID().Like("123")         // Like not available for non-string types
// user.Name().Eq(123)           // Type mismatch: string != int
```

### 🛡️ **SQL Injection Prevention**
- All values are automatically parameterized
- No string concatenation in SQL generation
- Safe handling of user input

### 🔧 **Rich Column Operations**
- **Comparison**: `Eq()`, `Ne()`, `Lt()`, `Gt()`, `Le()`, `Ge()`
- **Pattern Matching**: `Like()`, `NotLike()` (string columns only)
- **Range Queries**: `Between()`, `In()`
- **NULL Checking**: `IsNull()`, `IsNotNull()`

### 🔗 **Relationship Support**
- **One-to-Many**: `user.Posts()` with optional filtering
- **Many-to-Many**: `post.Tags()` through junction tables
- **Belongs-To**: `post.Author()` with INNER JOIN by default
- **JOIN Type Control**: `WithInnerJoin()` to override default JOIN types

## Design Philosophy

### Method Chaining + Functional Options
- **Column Operations**: Use method chaining for intuitive value-level operations
- **Query Structure**: Use functional options for flexible query construction
- **Type Safety**: Leverage Go's type system to prevent runtime errors

```go
// Column operations: method chaining
condition := user.Name().Like("J%").And(user.ID().Gt("10"))

// Query structure: functional options
sql, params := query.Select[tables.User](
    condition,
    user.OrderBy(user.Name(), ast.DirectionAsc),
    user.Limit(10),
)
```

## Usage Examples

### Basic Queries

```go
// Simple condition
sql, params := query.Select[tables.User](
    user.Email().Eq("john@example.com"),
)

// Multiple conditions (implicit AND)
sql, params := query.Select[tables.User](
    user.Name().Like("J%"),
    user.Email().IsNotNull(),
    user.ID().Gt("100"),
    user.OrderBy(user.Name(), ast.DirectionAsc),
)

// OR conditions
sql, params := query.Select[tables.User](
    user.Or(
        user.Name().Eq("John"),
        user.Name().Eq("Jane"),
        user.Name().Eq("Bob"),
    ),
    user.OrderBy(user.ID(), ast.DirectionDesc),
)

// Range and list queries
sql, params := query.Select[tables.User](
    user.ID().Between("100", "200"),
    user.Name().In("John", "Jane", "Bob"),
    user.Email().IsNotNull(),
)
```

### Complex Boolean Logic

```go
// Complex boolean: (a AND b) OR (c AND d)
sql, params := query.Select[tables.User](
    user.Or(
        user.And(
            user.Name().Eq("John"),
            user.Email().Like("%@example.com"),
        ),
        user.And(
            user.Name().Eq("Jane"),
            user.Email().Like("%@company.com"),
        ),
    ),
)
```

### Relationship Queries

```go
// One-to-Many: Users with their posts
sql, params := query.Select[tables.User](
    user.Posts(),  // LEFT OUTER JOIN by default
    user.OrderBy(user.Name(), ast.DirectionAsc),
)

// Filtered relationships
sql, params := query.Select[tables.User](
    user.Posts(post.Title().Like("%important%")),
    user.Name().Ne("Admin"),
)

// Many-to-Many: Posts with specific tags
sql, params := query.Select[tables.Post](
    post.Tags(tag.Name().Eq("Go")),
    post.Title().Like("%tutorial%"),
)

// Belongs-To: Posts with author information
sql, params := query.Select[tables.Post](
    post.Author(user.Email().Like("%@company.com")),
    post.Title().Like("Announcement:%"),
)

// Override JOIN type
sql, params := query.Select[tables.User](
    user.Posts(
        post.Title().Like("%important%"),
        post.WithInnerJoin(),  // Use INNER JOIN instead of LEFT OUTER JOIN
    ),
)
```

### Multi-level JOINs

```go
// User → Posts → Tags (3-level JOIN)
sql, params := query.Select[tables.User](
    user.Email().Eq("john@example.com"),
    user.Posts(
        post.Tags(),
        post.OrderBy(post.Title(), ast.DirectionDesc),
    ),
)

// Complex filtered multi-level JOIN
sql, params := query.Select[tables.User](
    user.Posts(
        post.Tags(tag.Name().Eq("Go")),
        post.Title().Like("%tutorial%"),
    ),
    user.OrderBy(user.Name(), ast.DirectionAsc),
)
```

## Type Safety Examples

### String-Specific Operations
```go
// ✅ These work (string columns)
user.Name().Like("John%")
user.Email().NotLike("%spam%")
post.Title().Like("Tutorial:%")
post.Content().NotLike("%deprecated%")

// ❌ These cause compile errors
// user.ID().Like("123")        // Like not available for non-string types
```

### Type-Safe Value Matching
```go
// ✅ Type-safe operations
user.Name().Eq("John")           // string = string
user.ID().Between("1", "100")    // string range
user.Name().In("John", "Jane")   // string list

// ❌ Compile-time type errors
// user.Name().Eq(123)           // Cannot pass int to string column
// user.ID().Between(1, 100)     // Type mismatch if ID is defined as string
```

## Project Structure

```
├── tables/         # Table type definitions
├── types/          # Core types (Column, State, Options)
├── user/           # User table query builder
├── post/           # Post table query builder  
├── tag/            # Tag table query builder
├── post_tag/       # Junction table definitions
├── query/          # Generic Select function and helpers
└── docs/plan/      # Design documents
```

## Future: Code Generation

Currently, the query builders are hand-written to validate the design and establish patterns. Once the API is stable, we plan to generate them automatically from Cloud Spanner schema definitions.

This template-based approach ensures that generated code follows consistent patterns and maintains the same level of type safety and usability.

See [docs/plan/](docs/plan/) for detailed design decisions and future plans.

## Dependencies

- [memefish](https://github.com/cloudspannerecosystem/memefish) - SQL parser for Cloud Spanner
- Go 1.18+ (generics are required for type-safe query construction)

## License

MIT