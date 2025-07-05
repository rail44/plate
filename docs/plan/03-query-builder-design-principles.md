# Query Builder Design Principles

This document explains the key design decisions behind our SQL query builder architecture.

## Core Principles

### 1. Type Safety Above All

Every table package has its own `ExprOption` and `QueryOption` types. This prevents mixing expressions from different tables:

```go
// Compile error - cannot mix user and profile expressions
user.Select(profile.Where(...))  // ❌
```

This design choice trades some code duplication for compile-time safety.

### 2. Functional Options Pattern

We use functional options for query construction because:
- **Composability**: Options can be combined in any order
- **Extensibility**: New options can be added without breaking existing code
- **Readability**: Query construction reads naturally

### 3. State Management

The `common.State` structure serves three critical purposes:
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

## Trade-offs

### Code Duplication vs Type Safety
Each package has similar functions (And, Or, Where, etc.). We accept this duplication because:
- Type safety prevents critical runtime errors
- Code is generated, so maintenance burden is minimal
- Clear ownership and boundaries

### Simplicity vs Features
We prioritize a simple, predictable API over supporting every SQL feature because:
- Most queries use a common subset of SQL
- Complex queries can still be written with raw SQL
- Predictability is more valuable than completeness

## Future Considerations

These principles guide our decisions as we add features:
- New SQL constructs must maintain type safety
- API additions should follow the functional options pattern
- Helper functions should remain pure and testable