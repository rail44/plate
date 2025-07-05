# Future Code Generation Plan

This document outlines our plan for implementing template-based code generation for the query builders.

## Current State

The query builders are currently hand-written to:
- Validate the API design
- Ensure the patterns work well in practice
- Discover edge cases and requirements

## Code Generation Goals

Once the hand-written implementation is stable, we will:

1. **Extract patterns into templates**
   - One template per package component
   - Minimal logic in templates
   - Use Go's text/template

2. **Create a generator tool**
   - Reads schema definitions from Spanner-generated models
   - Applies templates to generate query builders
   - Handles import management and formatting

3. **Integrate with existing workflow**
   - Plugs into the current Spanner → Domain Model pipeline
   - Triggered automatically on schema changes
   - Generates consistent code every time

## Implementation Strategy

### Phase 1: Template Creation
- Extract patterns from hand-written code
- Create templates for each component type
- Validate templates generate identical code

### Phase 2: Generator Development
- Build CLI tool for code generation
- Implement schema reading from domain models
- Add proper error handling and validation

### Phase 3: Integration
- Add to CI/CD pipeline
- Create documentation for generator usage
- Establish testing strategy for generated code

## Success Criteria

The code generation is successful when:
- Generated code is identical to hand-written code
- All tests pass without modification
- Adding new tables requires zero manual code
- Schema changes automatically propagate

## Benefits

This approach will provide:
- **Consistency**: All query builders follow the same patterns
- **Maintainability**: Bug fixes and improvements apply everywhere
- **Productivity**: New tables get query builders automatically
- **Type Safety**: Maintained through code generation