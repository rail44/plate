# Import Path Management in Popular Go Code Generation Tools

This document summarizes how popular Go code generation tools handle import path management, particularly in the context of Go modules.

## 1. protoc-gen-go (Protocol Buffers)

### Import Path Determination
- **Primary Method**: Uses the `go_package` option in `.proto` files
- **Fallback**: Command-line `M` flag for import path mapping
- **Module Support**: Added `module=` flag for module-aware generation

### Configuration Options
```proto
// In .proto file
option go_package = "github.com/example/project/pb";
```

```bash
# Command-line options
protoc --go_out=. --go_opt=paths=source_relative file.proto
protoc --go_out=. --go_opt=module=github.com/yourmodule proto/*.proto
protoc --go_opt=Mfile.proto=github.com/example/pkg file.proto
```

### Output Path Strategies
- `paths=import` (default): Places files according to Go import path
- `paths=source_relative`: Places files relative to source `.proto` files

### Key Features
- Explicit configuration required (no automatic detection)
- Module-aware with dedicated flags
- Supports both GOPATH and module modes

## 2. sqlc

### Import Path Determination
- **Primary Method**: Configuration file (`sqlc.yaml`)
- **Module Support**: Fully module-aware by default

### Configuration
```yaml
version: "2"
packages:
  - name: "db"
    path: "internal/db"  # Output path relative to module root
    queries: "./sql/query/"
    schema: "./sql/schema/"
    engine: "postgresql"
```

### Advanced Import Configuration
```yaml
overrides:
  - db_type: "uuid"
    go_type:
      import: "github.com/google/uuid"
      type: "UUID"
```

### Key Features
- Explicit configuration via YAML file
- Automatic module detection
- Supports custom type imports with full import paths
- Works relative to the location where `sqlc generate` is run

## 3. mockgen (gomock)

### Import Path Determination
- **Source Mode**: Analyzes source file directly
- **Reflect Mode**: Requires explicit import path
- **Module Support**: Uses `--build_flags=--mod=mod` for module compatibility

### Usage Patterns
```bash
# Source mode (automatic import detection)
mockgen -source=interface.go -destination=mocks/mock.go

# Reflect mode (explicit import path)
mockgen github.com/example/package Interface

# Current directory shorthand
mockgen -destination=mocks/mock.go . Interface

# Module-aware generation
mockgen --build_flags=--mod=mod github.com/example/pkg Interface
```

### Key Features
- Source mode provides automatic detection
- Reflect mode requires explicit import paths
- Special handling for vendor directories
- Module support via build flags

## 4. stringer

### Import Path Determination
- **Primary Method**: Automatic detection using `golang.org/x/tools/go/packages`
- **Context**: Analyzes entire package from working directory
- **Module Support**: Fully module-aware through package loading

### Usage Pattern
```go
//go:generate stringer -type=Status
type Status int
```

### Key Features
- No explicit import path configuration needed
- Automatically detects package context
- Works seamlessly with `go generate`
- Analyzes whole package regardless of file location
- Module-aware by default

## 5. wire (Dependency Injection)

### Import Path Determination
- **Primary Method**: Automatic detection during code generation
- **Build Constraint**: Uses `//+build wireinject` tag
- **Module Support**: Fully module-aware

### Usage Pattern
```go
//+build wireinject

package main

import "github.com/google/wire"

func InitializeApp() (*App, error) {
    wire.Build(ProviderSet)
    return nil, nil
}
```

### Key Features
- Automatic import path detection
- Generates `wire_gen.go` with proper imports
- Module-aware by analyzing source at compile time
- No explicit import configuration needed

## Common Patterns and Best Practices

### 1. Explicit vs Automatic Detection
- **Explicit**: protoc-gen-go, sqlc, mockgen (reflect mode)
- **Automatic**: stringer, wire, mockgen (source mode)

### 2. Configuration Methods
- **Config Files**: sqlc (YAML)
- **Source Annotations**: protoc-gen-go (`go_package`), stringer/wire (`//go:generate`)
- **Command-line**: All tools support various flags

### 3. Module-Aware Strategies
- **Package Analysis**: stringer, wire use Go's package loading
- **Configuration Options**: protoc-gen-go (`module=`), sqlc (path in config)
- **Build Flags**: mockgen (`--build_flags=--mod=mod`)

### 4. Output Path Control
- **Relative to Config**: sqlc
- **Relative to Source**: protoc-gen-go with `paths=source_relative`
- **Import Path Based**: protoc-gen-go default
- **Specified via Flags**: All tools via `-destination` or similar

### 5. Import Path Resolution
- Tools that analyze Go source (stringer, wire, mockgen source mode) can automatically determine import paths
- Tools that work with non-Go source (protoc-gen-go, sqlc) require explicit configuration
- Most tools respect the module context when run within a Go module

## Recommendations for New Code Generation Tools

1. **Prefer Automatic Detection**: When working with Go source files, use `golang.org/x/tools/go/packages` for automatic import path detection
2. **Provide Explicit Options**: For non-Go sources or special cases, allow explicit import path configuration
3. **Support Module Mode**: Ensure the tool works correctly within Go modules
4. **Flexible Output Paths**: Support both relative and import-path-based output locations
5. **Clear Documentation**: Document how import paths are determined and can be configured