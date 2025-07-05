# Plate - Type-Safe SQL Query Builder for Go

Plate is a type-safe SQL query builder generator for Go, designed to work with Google Cloud Spanner and the memefish SQL parser.

## Overview

Plate generates strongly-typed query builders from your database schema, ensuring compile-time safety for SQL query construction.

```go
// Type-safe query construction
sql, params := user.Select(
    user.JoinProfile(profile.Bio(ast.OpEqual, "Engineer")),
    user.Where(user.Name(ast.OpEqual, "John")),
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

## Design Principles

1. **Type safety above all else** - Each table has its own types to prevent mixing
2. **Functional options pattern** - Flexible and composable query construction
3. **Pure functions where possible** - Easier to test and reason about
4. **Generated code should be readable** - As if hand-written

## Project Structure

```
├── user/           # User table query builder
├── profile/        # Profile table query builder
├── query/          # Shared helper functions
├── common/         # Common types (State)
└── docs/plan/      # Design documents
```

## Usage Example

```go
// Simple select
sql, params := user.Select(
    user.Where(user.Email(ast.OpEqual, "john@example.com")),
)

// Complex query with JOIN
sql, params := user.Select(
    user.JoinProfile(
        profile.And(
            profile.Bio(ast.OpEqual, "Engineer"),
            profile.UserID(ast.OpNotEqual, "123"),
        ),
    ),
    user.Where(
        user.Or(
            user.Name(ast.OpEqual, "John"),
            user.Name(ast.OpEqual, "Jane"),
        ),
    ),
    user.OrderBy(user.OrderByID, ast.DirectionDesc),
    user.Limit(20),
)
```

## Future: Code Generation

Currently, the query builders are hand-written to validate the design. Once stable, we plan to generate them automatically from Spanner schema definitions.

See [docs/plan/](docs/plan/) for detailed design decisions and future plans.

## Dependencies

- [memefish](https://github.com/cloudspannerecosystem/memefish) - SQL parser for Cloud Spanner
- Go 1.18+ (no generics used currently, but may be added)

## License

MIT