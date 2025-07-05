# Schema Definition Approaches

This document outlines our analysis of different approaches for defining table schemas and relationships for query builder code generation.

## Context

We are building a code generator that produces type-safe SQL query builders from schema definitions. This generator will integrate with existing Spanner schema generation workflows.

## Key Design Decision

We chose the **Schema Definition DSL approach** because:

1. **Single Source of Truth**: Spanner schema remains the authoritative source
2. **Automatic Synchronization**: Schema changes automatically reflect in query builders
3. **Type Consistency**: Domain models and query builders use the same type definitions
4. **Leverage Existing Assets**: Reuses already generated metadata from Spanner

## Integration Strategy

The existing workflow:
```
Spanner Schema → Domain Model Generation
      ↓
   Schema Info
      ↓
Plate Code Generation → Query Builders
```

This approach allows us to:
- Avoid manual schema definition duplication
- Ensure consistency across the stack
- Reduce maintenance burden

## Alternative Approaches Considered

1. **Struct Tag Based**: Similar to GORM, but limited expressiveness for complex relationships
2. **Fluent API**: More flexible but requires runtime definition
3. **Configuration Files**: Language agnostic but lacks type safety

Each was evaluated but ultimately rejected in favor of leveraging the existing Spanner schema generation pipeline.