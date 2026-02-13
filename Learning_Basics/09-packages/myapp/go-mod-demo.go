package main

import (
	"encoding/json"
	"fmt"
	"os"
)

/*
=================================================================
PRACTICAL: go mod tidy DEMONSTRATION
=================================================================
*/

func main() {
	fmt.Println("🔧 UNDERSTANDING go mod tidy")
	fmt.Println("============================\n")

	// STEP 1: What happens without go.mod?
	fmt.Println("❌ PROBLEM: Without go.mod")
	fmt.Println("Error: 'go: go.mod file not found...'")
	fmt.Println("Reason: Go doesn't know it's a module\n")

	// STEP 2: Create go.mod
	fmt.Println("✅ SOLUTION: Create go.mod")
	fmt.Println("Command: go mod init myapp")
	fmt.Println("Creates: go.mod file\n")

	// STEP 3: What go.mod looks like
	fmt.Println("📄 SAMPLE go.mod file:")
	fmt.Println(`module myapp
go 1.21

require (
    github.com/sirupsen/logrus v1.9.2
)`)
	fmt.Println()

	// STEP 4: What does go mod tidy do?
	fmt.Println("🧹 WHAT go mod tidy DOES:")
	fmt.Println("1. ✓ Adds missing dependencies")
	fmt.Println("2. ✓ Removes unused dependencies")
	fmt.Println("3. ✓ Updates go.sum with checksums")
	fmt.Println("4. ✓ Ensures go.mod is clean")
	fmt.Println()

	// STEP 5: Example output
	demonstrateModTidy()

	// STEP 6: Built-in vs external packages
	fmt.Println("\n📦 BUILT-IN vs EXTERNAL PACKAGES:")
	fmt.Println("Built-in (no go.mod entry needed):")
	fmt.Println("  - fmt")
	fmt.Println("  - os")
	fmt.Println("  - encoding/json")
	fmt.Println("External (needs go.mod entry):")
	fmt.Println("  - github.com/sirupsen/logrus")
	fmt.Println("  - github.com/google/uuid")

	// STEP 7: Common commands
	fmt.Println("\n🔨 COMMON go mod COMMANDS:")
	fmt.Println("go mod init <name>      - Initialize new module")
	fmt.Println("go mod tidy             - Clean up dependencies")
	fmt.Println("go get <package>        - Add/update dependency")
	fmt.Println("go list -m all          - List all dependencies")
	fmt.Println("go mod why <package>    - Why is package needed?")
	fmt.Println("go mod vendor           - Copy deps to vendor/ folder")

	// Using the packages we imported
	fmt.Println("\n✅ Using imported packages:")

	// Using encoding/json
	data := map[string]interface{}{
		"name":    "Go",
		"version": "1.21",
	}
	jsonBytes, _ := json.Marshal(data)
	fmt.Printf("JSON: %s\n", jsonBytes)

	// Using os
	fmt.Printf("Current user: %s\n", os.Getenv("USERNAME"))

	fmt.Println("\n💡 KEY TAKEAWAY:")
	fmt.Println("Always run 'go mod tidy' before git commit!")
}

func demonstrateModTidy() {
	fmt.Println("📋 EXAMPLE: Before and After go mod tidy\n")

	fmt.Println("BEFORE go mod tidy:")
	fmt.Println("──────────────────")
	fmt.Println(`go.mod content:
module myapp
go 1.21
require github.com/unused/pkg v1.0.0          ← Not used in code
require github.com/sirupsen/logrus v1.9.2
// Missing: github.com/google/uuid (imported but not listed)
`)

	fmt.Println("main.go imports:")
	fmt.Println("import (")
	fmt.Println(`    "fmt"
    "github.com/sirupsen/logrus"
    "github.com/google/uuid"`)
	fmt.Println(")")

	fmt.Println("\nAFTER go mod tidy:")
	fmt.Println("─────────────────")
	fmt.Println(`go.mod content:
module myapp
go 1.21
require (
    github.com/sirupsen/logrus v1.9.2
    github.com/google/uuid v1.3.0               ← Added!
)
// ✓ Removed: github.com/unused/pkg

go.sum gets updated with checksums
`)
}
