# Code Generation Approaches

This document explains our decision-making process for choosing a code generation strategy.

## Context

We need to generate Go code for SQL query builders from schema definitions. The generated code must be:
- Type-safe
- Maintainable
- Easy to understand
- Consistent across all tables

## Key Design Decision

We chose the **Template-Based Approach** for code generation because:

1. **Simplicity**: Templates look like the actual Go code they generate
2. **Maintainability**: Changes to patterns require only template modifications
3. **Transparency**: Easy to understand what code will be generated
4. **Debugging**: Generated code structure is predictable

## Why Not Other Approaches?

### Go AST Manipulation
- Too verbose and error-prone
- Requires deep AST knowledge
- Hard to visualize the generated code

### Code Builder Libraries (e.g., jennifer)
- Adds external dependencies
- Learning curve for the library API
- Less transparent than templates

### Intermediate Representation (IR)
- Adds unnecessary complexity
- Extra abstraction layer to maintain
- Overkill for our use case

## Template Strategy

Our templates follow these principles:

1. **One template per component**: Separate templates for columns, joins, etc.
2. **Minimal logic in templates**: Keep complex logic in the generator
3. **Use Go's text/template**: Standard library is sufficient
4. **Post-process with gofmt**: Ensures consistent formatting

This approach has proven successful in similar projects and provides the best balance of power and simplicity for our needs.