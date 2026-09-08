package main

import "fmt"

// Package-level variables (like class variables in Java)
var globalCounter int = 0
var appName string = "MyApp"

// Package-level constants
const Version = "1.0.0"

func incrementCounter() {
	globalCounter++ // Can access package-level variables
	fmt.Println("Counter:", globalCounter)
}

func main() {
	// Local variables (inside function)
	var localVar int = 42
	
	fmt.Println("App Name:", appName)     // Access package-level variable
	fmt.Println("Version:", Version)      // Access package-level constant
	fmt.Println("Local Var:", localVar)  // Access local variable
	
	incrementCounter() // Counter: 1
	incrementCounter() // Counter: 2
	
	// localVar is only available inside main()
	// globalCounter and appName are available to all functions in this package
}