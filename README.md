# Plate - Type-Safe SQL Query Builder for Go

Plate is a type-safe SQL query builder generator for Go, designed to work with Google Cloud Spanner and the memefish SQL parser.

## Overview

Plate generates strongly-typed query builders from your database schema, ensuring compile-time safety for SQL query construction.

```go
// Type-safe query construction with implicit AND conditions
sql, params := query.Select[tables.User](
    user.JoinProfile(profile.Bio(ast.OpEqual, "Engineer")),
    user.Name(ast.OpEqual, "John"),  // No Where() needed!
    user.OrderBy(user.OrderByName, ast.DirectionAsc),
    user.Limit(10),
)
// sql: SELECT * FROM user INNER JOIN profile ON user.id = profile.user_id WHERE profile.bio = @p0 AND user.name = @p1 ORDER BY user.name ASC LIMIT 10
// params: []any{"Engineer", "John"}
```

## Key Features

- **Type Safety**: Cannot mix expressions from different tables
- **SQL Injection Prevention**: All values are parameterized
- **Composable**: Build complex queries with functional options
- **Zero Runtime Reflection**: All code is generated at compile time
- **Intuitive API**: Direct condition passing without Where(), implicit AND, explicit OR

## Design Principles

1. **Type safety above all else** - Each table has its own types to prevent mixing
2. **Functional options pattern** - Flexible and composable query construction
3. **Pure functions where possible** - Easier to test and reason about
4. **Generated code should be readable** - As if hand-written
5. **Developer experience first** - Prioritize intuitive API over implementation complexity

## Project Structure

```
├── tables/         # Centralized table type definitions
├── types/          # Core type definitions (Table, State, Options)
├── user/           # User table query builder
├── profile/        # Profile table query builder
├── query/          # Shared helper functions and generic Select
└── docs/plan/      # Design documents
```

## Usage Example

```go
// Simple select - no Where() needed!
sql, params := query.Select[tables.User](
    user.Email(ast.OpEqual, "john@example.com"),
)

// Multiple conditions - implicit AND
sql, params := query.Select[tables.User](
    user.Name(ast.OpEqual, "John"),
    user.Email(ast.OpLike, "%@example.com"),
    user.Limit(10),
)

// Explicit OR conditions
sql, params := query.Select[tables.User](
    user.Or(
        user.Name(ast.OpEqual, "John"),
        user.Name(ast.OpEqual, "Jane"),
        user.Name(ast.OpEqual, "Bob"),
    ),
    user.OrderBy(user.OrderByID, ast.DirectionDesc),
)

// Complex query with JOIN
sql, params := query.Select[tables.User](
    user.JoinProfile(profile.Bio(ast.OpEqual, "Engineer")),
    user.Email(ast.OpLike, "%@company.com"),
    user.Or(
        user.Name(ast.OpEqual, "Admin"),
        user.ID(ast.OpEqual, "1"),
    ),
    user.Limit(20),
)

// Bidirectional JOIN - Profile to User
sql, params := query.Select[tables.Profile](
    profile.JoinUser(user.Name(ast.OpEqual, "John")),
    profile.Bio(ast.OpEqual, "Engineer"),
)
```

## Future: Code Generation

Currently, the query builders are hand-written to validate the design. Once stable, we plan to generate them automatically from Spanner schema definitions.

See [docs/plan/](docs/plan/) for detailed design decisions and future plans.

## Dependencies

- [memefish](https://github.com/cloudspannerecosystem/memefish) - SQL parser for Cloud Spanner
- Go 1.18+ (generics are used for type-safe query construction)

## License

MIT