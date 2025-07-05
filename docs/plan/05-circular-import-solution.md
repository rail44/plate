# Circular Import Resolution

This document explains how we solved the circular import problem that arose when implementing bidirectional JOINs.

## The Problem

We wanted to support bidirectional JOINs:
- `user.JoinProfile()` - JOIN from user to profile table
- `profile.JoinUser()` - JOIN from profile to user table

However, implementing this naively would create circular imports:
- `user` package would need to import `profile` types
- `profile` package would need to import `user` types

## Solutions Considered

### 1. Type Parameters with Interfaces
Define interfaces like `CanJoinProfile` and `CanJoinUser`, but Go's type inference limitations made this unwieldy.

### 2. Generic Join Functions
Create generic `Join[From, To]()` functions, but Go cannot infer type parameters from return types, requiring explicit type annotations.

### 3. Separate Join Package
Move all JOIN logic to a separate package, but this would lose compile-time type safety for which JOINs are valid.

### 4. Centralized Tables Package (Chosen Solution)
Move all table type definitions to a central `tables` package:

```go
// tables/tables.go
package tables

type User struct{}
func (User) TableName() string { return "user" }

type Profile struct{}
func (Profile) TableName() string { return "profile" }
```

## Implementation

The chosen solution provides:
- **No circular imports**: Table types are in a shared package
- **Type safety**: Options are still parameterized by table type
- **Bidirectional JOINs**: Both packages can reference each other's table types
- **Clean architecture**: Clear separation of concerns

## Benefits

1. **Type Safety Preserved**: The type system still prevents mixing expressions
2. **Simple Implementation**: No complex type gymnastics required
3. **Extensible**: Easy to add new tables and relationships
4. **Developer Experience**: Clean, intuitive API without type annotations

## Example Usage

```go
// User to Profile JOIN
query.Select[tables.User](
    user.JoinProfile(profile.Bio(ast.OpEqual, "Engineer")),
)

// Profile to User JOIN  
query.Select[tables.Profile](
    profile.JoinUser(user.Name(ast.OpEqual, "John")),
)
```

This architecture successfully enables bidirectional JOINs while maintaining type safety and avoiding circular dependencies.