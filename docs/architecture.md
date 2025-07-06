# Architecture

This document explains the key architectural decisions and design patterns used in Plate.

## Table Type Isolation

To prevent circular imports while supporting bidirectional JOINs, all table type definitions are centralized in the `tables` package:

```go
// tables/tables.go
package tables

type User struct{}
func (User) TableName() string { return "user" }

type Profile struct{}
func (Profile) TableName() string { return "profile" }
```

This architecture provides:
- **No circular imports**: Table types are in a shared package
- **Type safety**: Options are still parameterized by table type
- **Bidirectional JOINs**: Both packages can reference each other's table types
- **Clean architecture**: Clear separation of concerns

Example usage:
```go
// User to Profile JOIN
query.Select[tables.User](
    user.JoinProfile(profile.Bio().Eq("Engineer")),
)

// Profile to User JOIN  
query.Select[tables.Profile](
    profile.JoinUser(user.Name().Eq("John")),
)
```

## Multi-Layered Type Safety

The architecture uses multiple layers of type safety to prevent errors at compile time:

### Table Type Isolation
A centralized `tables` package with type-parameterized options prevents mixing expressions from different tables:

```go
// Compile error - cannot mix user and post expressions
query.Select[tables.User](
    post.Title().Eq("Test"),  // ❌ Type error!
)
```

### Column Value Type Safety
Column types include both table and value type parameters `Column[T Table, V any]`:

```go
// Type-safe column operations
user.Name().Eq("John")        // ✅ string = string
user.ID().Between("1", "10")  // ✅ string range

// Compile errors - type mismatches
// user.Name().Eq(123)        // ❌ string != int
// user.ID().Like("123")      // ❌ Like only for string columns
```

### Method Constraint by Type
String-specific operations are only available for string columns:

```go
// ✅ Available for Column[T, string]
user.Name().Like("John%")
user.Email().NotLike("%spam%")

// ❌ Compile error for non-string columns
// user.CreatedAt().Like("2023%")  // time.Time column
```

## State Management

The `types.State` structure serves three critical purposes:
1. **Parameter tracking**: Sequential parameter naming (p0, p1, p2...)
2. **Table alias management**: Prevents naming conflicts in JOINs
3. **Context preservation**: Maintains current table context for column qualification

## Trade-offs

### Code Duplication vs Type Safety
Each table package has similar functions (column builders, Or, Where, etc.). We accept this duplication because:
- Type safety prevents critical runtime errors
- Code is generated, so maintenance burden is minimal
- Clear ownership and boundaries
- Go's type inference limitations prevent fully generic implementations

### Simplicity vs Features
We prioritize a simple, predictable API over supporting every SQL feature because:
- Most queries use a common subset of SQL
- Complex queries can still be written with raw SQL
- Predictability is more valuable than completeness

### Value Semantics
Helper functions return values rather than modifying pointers because:
- AST construction is not a configuration task
- Pure functions are easier to test and reason about
- Follows Go's preference for value semantics where appropriate