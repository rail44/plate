# Query Builder Design Principles

This document explains the key design decisions behind our SQL query builder architecture.

## Core Principles

### 1. Type Safety Above All

The architecture uses multiple layers of type safety to prevent errors at compile time:

#### Table Type Isolation
A centralized `tables` package with type-parameterized options prevents mixing expressions from different tables:

```go
// Compile error - cannot mix user and post expressions
query.Select[tables.User](
    post.Title().Eq("Test"),  // ❌ Type error!
)
```

#### Column Value Type Safety
Column types include both table and value type parameters `Column[T Table, V any]`:

```go
// Type-safe column operations
user.Name().Eq("John")        // ✅ string = string
user.ID().Between("1", "10")  // ✅ string range

// Compile errors - type mismatches
// user.Name().Eq(123)        // ❌ string != int
// user.ID().Like("123")      // ❌ Like only for string columns
```

#### Method Constraint by Type
String-specific operations are only available for string columns:

```go
// ✅ Available for Column[T, string]
user.Name().Like("John%")
user.Email().NotLike("%spam%")

// ❌ Compile error for non-string columns
// user.CreatedAt().Like("2023%")  // time.Time column
```

This multi-layered approach provides comprehensive compile-time safety while maintaining an intuitive API.

### 2. Functional Options Pattern

We use functional options for query construction because:
- **Composability**: Options can be combined in any order
- **Extensibility**: New options can be added without breaking existing code
- **Readability**: Query construction reads naturally

### 3. State Management

The `types.State` structure serves three critical purposes:
1. **Parameter tracking**: Sequential parameter naming (p0, p1, p2...)
2. **Table alias management**: Prevents naming conflicts in JOINs
3. **Context preservation**: Maintains current table context for column qualification

### 4. Helper Functions for AST Construction

The `query` package provides pure functions that build AST nodes. This separation:
- Keeps business logic centralized
- Makes testing easier (pure functions with deterministic output)
- Allows code reuse without compromising type safety

### 5. Value Semantics

Helper functions return values rather than modifying pointers because:
- AST construction is not a configuration task
- Pure functions are easier to test and reason about
- Follows Go's preference for value semantics where appropriate

### 6. Developer Experience First

The API is designed for intuitive usage through two complementary patterns:

#### Method Chaining for Column Operations
Column operations use fluent method chaining for natural value-level operations:
```go
// Natural, readable column operations
condition := user.Name().Like("John%")
rangeQuery := user.ID().Between("100", "200")
nullCheck := user.Email().IsNotNull()
```

#### Functional Options for Query Structure
Query construction uses functional options for flexible structural operations:
```go
// Flexible, composable query building
sql, params := query.Select[tables.User](
    condition,                                    // ExprOption
    user.OrderBy(user.Name(), ast.DirectionAsc), // QueryOption
    user.Limit(10),                              // QueryOption
)
```

#### Combined Benefits
- **No Where() needed**: Column conditions integrate seamlessly with Select
- **Implicit AND**: Multiple conditions are joined with AND by default
- **Explicit OR**: OR conditions require explicit Or() function call
- **Type Safety**: Both patterns maintain compile-time type checking
- **Mixed options**: Select accepts both QueryOption and ExprOption types

This dual approach optimizes each use case: method chaining for value operations and functional options for structural flexibility.

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

### Enhanced Type Safety Through Value Parameters

Our implementation uses value type parameters to provide multiple layers of compile-time safety:

#### Benefit 1: Runtime Error Prevention
Type constraints prevent common SQL errors at compile time:
```go
// Prevents runtime type errors
user.Name().Eq("string_value")     // ✅ Matches string type
user.Age().Between(18, 65)         // ✅ Matches int type

// Compile-time errors instead of runtime failures
// user.Name().Eq(123)             // ❌ Type mismatch caught early
// user.Age().Like("18")           // ❌ Invalid operation caught early
```

#### Benefit 2: Method Availability by Type
Operations are naturally constrained to appropriate types:
```go
// String operations only available for string columns
user.Name().Like("pattern")        // ✅ Natural string operation
user.Email().NotLike("spam")       // ✅ Prevents invalid queries

// Numeric operations work with any comparable type
user.Age().Between(18, 65)         // ✅ Range queries
user.Score().Gt(100)              // ✅ Comparison operations
```

#### Benefit 3: IDE Integration
Type constraints enable better development experience:
- IntelliSense shows only valid methods for each column type
- Compile-time feedback prevents invalid SQL generation
- Refactoring tools understand type relationships

This approach aligns with Go's philosophy of catching errors early and using the type system to enforce correctness.

## Future Considerations

These principles guide our decisions as we add features:
- New SQL constructs must maintain multi-layered type safety
- API additions should follow the dual pattern (method chaining + functional options)
- Helper functions should remain pure and testable
- Type constraints should prevent errors at the earliest possible stage (compile-time)