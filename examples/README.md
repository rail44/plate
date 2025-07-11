# Plate Examples

This directory contains examples of using the Plate query builder.

## Setup

Before running the examples or tests, you need to generate the query builders:

```bash
go generate ./...
```

This will create the `generated/` directory with type-safe query builders for the example models.

## Running Tests

```bash
go test
```

Note: The tests will automatically regenerate the code before running.

## Generated Files

The `generated/` directory is ignored by git. It contains:
- `tables/` - Table type definitions
- `user/`, `post/`, `tag/` - Query builders for each table

These files are regenerated from the schema defined in `generate.go`.