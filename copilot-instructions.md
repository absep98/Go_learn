# Go Learning Project - AI Instructions

> This file provides context to GitHub Copilot and other AI assistants about how to work with this codebase.

## Project Overview

This is a Go learning repository focused on mastering Go fundamentals through hands-on practice. The project includes:

- Basic concepts (types, functions, control flow)
- Data structures (arrays, slices, maps, structs)
- Concurrency (goroutines, channels)
- Practical projects (APIs, CLI tools)
- Interview preparation materials

## Code Conventions

### General Go Style

- Follow official Go conventions and `gofmt` formatting
- Use `goimports` for import organization
- Variable names: short and descriptive (prefer `i` over `index` in loops)
- Package names: lowercase, single word, no underscores
- Error handling: always check errors, return early

### File Organization

```
01-basics/          → Fundamental concepts
02-control-flow/    → Loops, conditionals, switch
03-data-structures/ → Arrays, slices, maps, structs, interfaces
05-concurrency/     → Goroutines, channels, patterns
06-projects/        → Small complete projects
07-Pointers/        → Pointer concepts and practices
08-errors/          → Error handling patterns
09-packages/        → Module and package management
11-interview-prep/  → Interview questions and tests
12-marshalling/     → JSON encoding/decoding
Project-Building-While-Learning/ → Real-world backend project
```

### Naming Conventions

**Files:**
- Learning files: `descriptive-name.go` (e.g., `arrays-slices.go`)
- Test files: `test-*.go` or `*-test.go`
- Practice files: `practice-*.go`

**Functions:**
- Exported (public): `PascalCase` - starts with uppercase
- Unexported (private): `camelCase` - starts with lowercase
- Test functions: `TestFunctionName(t *testing.T)`

**Variables:**
- Short names in small scopes: `i`, `n`, `err`, `ok`
- Descriptive names in larger scopes: `userCount`, `responseData`
- Constants: `MaxConnections`, `DefaultTimeout`

### Testing

- Use table-driven tests where appropriate
- Always test error cases
- Use meaningful test names: `TestParseInput_InvalidFormat_ReturnsError`
- Group related tests in subtests with `t.Run()`

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive numbers", 2, 3, 5},
        {"negative numbers", -1, -1, -2},
        {"zero", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.want {
                t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
            }
        })
    }
}
```

### Error Handling

- Always handle errors explicitly
- Return errors, don't panic (except in truly unrecoverable situations)
- Use `fmt.Errorf` with `%w` for error wrapping
- Check errors immediately after they occur

```go
// Good
result, err := doSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// Avoid
result, _ := doSomething() // Never ignore errors
```

### Comments

- Package comment: every package should have a comment
- Exported functions: comment what it does, not how
- Complex logic: explain why, not what
- Use `//` for single-line comments
- Use `/* */` for multi-line comments or temporary code disabling

```go
// CalculateTotal computes the total price including tax.
// It returns an error if the tax rate is negative.
func CalculateTotal(price, taxRate float64) (float64, error) {
    // Business rule: negative tax rates are invalid
    if taxRate < 0 {
        return 0, errors.New("tax rate cannot be negative")
    }
    return price * (1 + taxRate), nil
}
```

## Learning Approach

### When Helping with Concepts

1. **Explain first** - Provide clear explanations before showing code
2. **Show examples** - Use simple, isolated examples
3. **Compare with other languages** - When it helps understanding
4. **Common pitfalls** - Warn about typical mistakes
5. **Best practices** - Teach the idiomatic Go way

### When Generating Practice Code

- Create progressive difficulty levels
- Include comments explaining key concepts
- Add TODO sections for the learner to complete
- Provide test cases to verify understanding

### When Reviewing Code

- Identify non-idiomatic Go patterns
- Suggest more efficient approaches
- Explain the "Go way" of solving problems
- Point out potential bugs or edge cases

## Project-Specific Patterns

### Personal Analytics Backend

This is a production-style project in `Project-Building-While-Learning/personal-analytics-backend/`

**Architecture:**
- Clean architecture with handlers, models, db layers
- JWT authentication
- PostgreSQL database
- Redis caching
- Structured logging
- Metrics collection

**When working on this project:**
- Follow existing patterns in the codebase
- Add tests for new endpoints
- Update API documentation in `API-ENDPOINTS.md`
- Log important operations
- Handle errors gracefully with proper HTTP status codes

## Common Patterns

### Struct Initialization

```go
// Prefer designated initializers for clarity
user := User{
    Name:  "John",
    Email: "john@example.com",
    Age:   30,
}

// Not: user := User{"John", "john@example.com", 30}
```

### Concurrency

```go
// Use WaitGroups for goroutine synchronization
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        // work here
    }(i)
}
wg.Wait()

// Use channels for communication
results := make(chan Result, 10)
go producer(results)
go consumer(results)
```

### Interface Usage

- Keep interfaces small (ideally one or two methods)
- Define interfaces where they're used, not where they're implemented
- Accept interfaces, return structs

```go
// Good - small, focused interface
type Reader interface {
    Read(p []byte) (n int, err error)
}

// Function accepts interface
func ProcessData(r Reader) error {
    // implementation
}
```

## Dependencies and Tools

### Required Tools

- Go 1.21+ (or latest stable)
- `gofmt` for formatting
- `goimports` for import management
- `golangci-lint` for linting (if available)

### Common Commands

```bash
# Run a file
go run filename.go

# Format code
go fmt ./...

# Run tests
go test ./...

# Run specific test
go test -run TestFunctionName

# View test coverage
go test -cover ./...

# Build binary
go build

# Install dependencies
go mod download

# Tidy dependencies
go mod tidy
```

## AI Assistant Guidelines

### Do:

- ✅ Explain concepts clearly with examples
- ✅ Use idiomatic Go patterns
- ✅ Provide working, runnable code
- ✅ Include error handling
- ✅ Add helpful comments
- ✅ Suggest improvements to existing code
- ✅ Create table-driven tests
- ✅ Use standard library when possible

### Don't:

- ❌ Use non-standard libraries without explanation
- ❌ Ignore errors or use `panic` unnecessarily
- ❌ Write overly complex solutions for learning examples
- ❌ Skip error handling in examples
- ❌ Use deprecated Go features
- ❌ Assume prior knowledge of advanced concepts
- ❌ Generate code without explanations

## Questions to Ask When Unclear

When the task or requirement is ambiguous, ask:

1. **Complexity level**: Beginner, intermediate, or advanced?
2. **Purpose**: Learning exercise or production code?
3. **Constraints**: Any specific packages or patterns to use/avoid?
4. **Context**: Which directory should this go in?
5. **Testing**: Should tests be included?

## Examples of Good Interactions

### Example 1: Learning Request

**User:** "Explain Go interfaces"

**Good Response:**
1. Define what interfaces are
2. Show simple example (io.Reader)
3. Demonstrate interface implementation
4. Explain duck typing
5. Show practical use case
6. Mention common standard library interfaces

### Example 2: Code Review Request

**User:** Shows code with `result, _ := doSomething()`

**Good Response:**
1. Identify the error ignoring
2. Explain why it's problematic
3. Show correct error handling
4. Mention tools that can catch this (linters)

### Example 3: Project Request

**User:** "Create a file processing utility"

**Good Response:**
1. Ask clarifying questions (file type, operation, size?)
2. Suggest project structure
3. Implement with proper error handling
4. Include example usage
5. Add tests
6. Document edge cases

---

*This file helps AI assistants provide better, more contextual help for this Go learning project.*
