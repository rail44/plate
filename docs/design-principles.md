# Design Principles

This document outlines the core design principles that guide Plate's development.

## Core Principles

### 1. Type Safety Above All

Plate leverages Go's type system to provide comprehensive compile-time safety, preventing runtime errors through:
- Table type isolation to prevent mixing expressions from different tables
- Column value type parameters for type-safe operations
- Method constraints based on column types

### 2. Developer Experience First

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

### 3. Schema as Single Source of Truth

Plate integrates with existing schema generation workflows:
- Spanner schema remains the authoritative source
- Schema changes automatically reflect in query builders
- Domain models and query builders use the same type definitions
- Leverages already generated metadata

### 4. Code Generation Strategy

We chose template-based code generation for:
- **Simplicity**: Templates look like the actual Go code they generate
- **Maintainability**: Changes to patterns require only template modifications
- **Transparency**: Easy to understand what code will be generated
- **Debugging**: Generated code structure is predictable

### 5. Functional Options Pattern

We use functional options for query construction because:
- **Composability**: Options can be combined in any order
- **Extensibility**: New options can be added without breaking existing code
- **Readability**: Query construction reads naturally

### 6. Pure Functions for AST Construction

The `query` package provides pure functions that build AST nodes. This separation:
- Keeps business logic centralized
- Makes testing easier (pure functions with deterministic output)
- Allows code reuse without compromising type safety

## API Design Philosophy

### No Where() Needed
Column conditions integrate seamlessly with Select - no explicit Where() method required.

### Implicit AND, Explicit OR
- Multiple conditions are joined with AND by default
- OR conditions require explicit Or() function call
- This matches SQL's natural precedence rules

### Type Safety Through Value Parameters
Operations are naturally constrained to appropriate types:
- String operations only available for string columns
- Numeric operations work with any comparable type
- Prevents runtime type errors at compile time

## Future Considerations

These principles guide our decisions as we add features:
- New SQL constructs must maintain multi-layered type safety
- API additions should follow the dual pattern (method chaining + functional options)
- Helper functions should remain pure and testable
- Type constraints should prevent errors at the earliest possible stage (compile-time)