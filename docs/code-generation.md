# Code Generation

This document describes Plate's code generation system and its design decisions.

## Overview

Plate uses a template-based code generation system that:
- Automatically generates type-safe query builders from table schemas
- Maintains the same level of type safety as hand-written code
- Handles relationships including one-to-many, many-to-many, and belongs-to
- Integrates with existing Spanner schema generation workflows

## Generated Code Structure

```
generated/
├── tables/         # Table type definitions
│   └── tables.go
├── user/           # Query builder for each table
│   └── user.go
├── post/
│   └── post.go
└── tag/
    └── tag.go
```

## Key Features

### Relationship Generation

The generator automatically creates relationship methods based on configuration:

- **BelongsTo**: Creates methods with INNER JOIN by default
- **HasMany**: Creates methods with LEFT OUTER JOIN by default  
- **ManyToMany**: Creates methods that JOIN through junction tables
- **Reverse Relations**: Automatically generated when `ReverseName` is specified

### Type Safety

Generated code maintains full type safety:
- Column types are preserved from model structs
- String-specific methods only available for string columns
- Compile-time checking of value types

### Import Management

The generator intelligently manages imports:
- Detects when `time` package is needed
- Calculates correct import paths for generated packages
- Handles both local and vendored dependencies

## Template Strategy

Our templates follow these principles:

1. **One template per component**: Separate templates for columns, joins, etc.
2. **Minimal logic in templates**: Keep complex logic in the generator
3. **Use Go's text/template**: Standard library is sufficient
4. **Post-process with gofmt**: Ensures consistent formatting

## Benefits

- **Consistency**: All query builders follow identical patterns
- **Maintainability**: Updates to templates propagate to all generated code
- **Productivity**: New tables get query builders automatically
- **Type Safety**: Fully maintained through code generation
- **Zero Manual Code**: Adding new tables requires only configuration

## Usage Example

```go
generator := plate.NewGenerator()
files, err := generator.Generate(config, "./generated")
if err != nil {
    log.Fatal(err)
}

// Write generated files to disk
for path, content := range files.Files {
    // Write content to path
}
```