/*
Build a small CLI package manager.
Define a Package struct with Name, Version, Dependencies []string.
Implement methods: AddDependency(dep string), String() string (returns "name@version").
Create an interface Installer with Install() error.
Make Package implement Installer (print installation steps and return nil).
In main(), create a package, add dependencies, and call Install().
Rules:
Use proper error handling (no panic unless truly unrecoverable).
Use pointer receivers where mutation or large structs are involved.
Format your code with go fmt (or gofmt).
*/
// package main

// import (
// 	"fmt"
// )


// type Package struct {
// 	Name string
// 	Version string
// 	Dependencies []string
// }

// type Installer interface {
// 	Install() error
// }

// func (p *Package) Install() error {
// 	for _, name := range p.Dependencies {
// 		fmt.Println(name)
// 	}
// 	return nil
// }

// func (p *Package) String() string {
// 	return fmt.Sprintf("%s@%s", p.Name, p.Version)
// }

// func (p *Package) AddDependency(dep string) {
// 	p.Dependencies = append(p.Dependencies, dep)
// }

// func main() {
// 	pkg := Package{Name: "encoding/json", Version: "1.0.0", Dependencies: []string{"json"}}
// 	pkg.AddDependency("net/http")
	
// 	var i Installer  // i can hold anything that has Install() error
// 	i = &pkg

// 	err := i.Install()
// 	if err != nil {
// 		fmt.Println("Something went wrong so ", err)
// 	}
// 	fmt.Println(&pkg)
// }


package main

import (
	"fmt"
	"strings"
	"os"
)

// Package represents a software package with dependencies.
type Package struct {
	Name         string
	Version      string
	Dependencies []string
}

// Installer defines the contract for installable entities.
type Installer interface {
	Install() error
}

// NewPackage is a constructor — idiomatic Go for complex initialization.
func NewPackage(name, version string, deps ...string) *Package {
	return &Package{
		Name:         name,
		Version:      version,
		Dependencies: deps,
	}
}

// String satisfies fmt.Stringer. Value receiver — safe for read-only.
func (p Package) String() string {
	return fmt.Sprintf("%s@%s", p.Name, p.Version)
}

// AddDependency adds a dependency. Pointer receiver — mutates state.
func (p *Package) AddDependency(deps ...string) {
	p.Dependencies = append(p.Dependencies, deps...)
}

// Install satisfies Installer. Pointer receiver — may mutate in future.
func (p *Package) Install() error {
	if p.Name == "" {
		return fmt.Errorf("package name cannot be empty")
	}

	fmt.Printf("Installing %s...\n", p)
	for _, dep := range p.Dependencies {
		fmt.Printf("  → resolving %s\n", dep)
	}
	fmt.Printf("✓ %s installed successfully\n", p)
	return nil
}

// DependencyGraph returns a formatted dependency tree.
func (p Package) DependencyGraph() string {
	if len(p.Dependencies) == 0 {
		return fmt.Sprintf("%s (no dependencies)", p)
	}
	return fmt.Sprintf("%s requires: %s", p, strings.Join(p.Dependencies, ", "))
}

func main() {
	// Constructor usage — clear, explicit, flexible
	pkg := NewPackage("github.com/gorilla/mux", "v1.8.1",
		"github.com/gorilla/context",
		"github.com/gorilla/securecookie",
	)
	pkg.AddDependency("golang.org/x/net")

	// Interface usage — decoupled, testable
	var installer Installer = pkg

	if err := installer.Install(); err != nil {
		fmt.Fprintf(os.Stderr, "installation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(pkg.DependencyGraph())
}