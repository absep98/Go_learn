# Go Quick Start Guide 🚀

## Step 1: Install Go
1. Download Go from [golang.org/dl](https://golang.org/dl/)
2. Follow the installation instructions for Windows
3. Verify installation by opening PowerShell and running:
   ```powershell
   go version
   ```

## Step 2: Run Your First Program
Navigate to the `01-basics` folder and run:
```powershell
go run hello.go
```

## Step 3: Learning Path

### 🎯 Recommended Order:
1. **01-basics/** - Start here!
   - `hello.go` - Your first Go program
   - `types.go` - Learn about data types
   - `functions.go` - Understanding functions

2. **02-control-flow/**
   - `loops.go` - For loops and control structures
   - `switch.go` - Switch statements

3. **03-data-structures/**
   - `arrays-slices.go` - Arrays and dynamic slices
   - `maps.go` - Key-value pairs
   - `structs.go` - Custom data types

4. **06-projects/**
   - `number-guessing-game.go` - Fun interactive game
   - `todo-list.go` - Practice with structs and methods

## Step 4: Key Go Concepts

### 🔥 Important Features:
- **Fast compilation** - Go compiles very quickly
- **Static typing** - Types are checked at compile time
- **Garbage collection** - Automatic memory management
- **Concurrency** - Built-in support for concurrent programming
- **Simple syntax** - Easy to read and write

### 📝 Go Syntax Highlights:
```go
package main        // Every Go file starts with a package declaration
import "fmt"        // Import packages you need

func main() {       // main function is the entry point
    fmt.Println("Hello, World!")
}
```

### 🎨 Variable Declarations:
```go
var name string = "Go"     // Explicit type
age := 25                  // Type inference
const pi = 3.14159         // Constants
```

## Step 5: Common Commands

```powershell
# Run a Go program
go run filename.go

# Build an executable
go build filename.go

# Format your code (very important in Go!)
go fmt filename.go

# Get help
go help

# Create a new module (for larger projects)
go mod init myproject
```

## Step 6: Next Steps

1. Complete all the examples in order
2. Try modifying the code to see what happens
3. Build your own small projects
4. Learn about Go modules and packages
5. Explore goroutines and channels for concurrency

## 💡 Pro Tips:
- Go has very strong opinions about code formatting - use `go fmt`!
- Error handling is explicit in Go - always check errors
- Go prefers composition over inheritance
- Keep your code simple and readable

Happy coding! 🎉