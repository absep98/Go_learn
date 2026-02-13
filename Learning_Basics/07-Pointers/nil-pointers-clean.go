package main

import "fmt"

type User struct {
	Name string
	Age  int
}

type Printer interface {
	Print()
}

type MyStruct struct{}

func (m *MyStruct) Print() {
	fmt.Println("Printing...")
}

func main() {
	fmt.Println("🚨 GO NIL POINTERS - COMPLETE GUIDE")
	fmt.Println("====================================")

	// 1. WHAT IS NIL?
	fmt.Println("=== 1. WHAT IS NIL? ===")
	var p *User         // p is nil (zero value of pointer type)
	var num *int        // num is nil
	var slice *[]string // slice is nil

	fmt.Printf("p is nil: %t\n", p == nil)
	fmt.Printf("num is nil: %t\n", num == nil)
	fmt.Printf("slice is nil: %t\n", slice == nil)

	// 2. THE DANGER - DEREFERENCING NIL
	fmt.Println("\n=== 2. THE DANGER ===")
	fmt.Println("❌ This would panic: fmt.Println(p.Age)")
	fmt.Println("❌ This would panic: fmt.Println(*num)")

	// Instead, always check first:
	if p != nil {
		fmt.Println("User age:", p.Age)
	} else {
		fmt.Println("✅ Safely detected: pointer is nil")
	}

	// 3. SAFE NIL CHECKING PATTERNS
	fmt.Println("\n=== 3. SAFE PATTERNS ===")

	// Create a valid pointer
	validUser := &User{Name: "Alice", Age: 25}

	// Safe function
	printUserSafely(p)         // nil pointer
	printUserSafely(validUser) // valid pointer

	// 4. COMMON NIL BUGS
	fmt.Println("\n=== 4. COMMON BUGS ===")

	// Bug 1: Slice of nil pointers
	var users []*User
	users = append(users, nil, validUser, nil)

	fmt.Println("Users list:")
	for i, user := range users {
		if user != nil {
			fmt.Printf("  [%d]: %s\n", i, user.Name)
		} else {
			fmt.Printf("  [%d]: ❌ NIL POINTER\n", i)
		}
	}

	// Bug 2: Function returns nil
	foundUser := findUser("nonexistent")
	if foundUser != nil {
		fmt.Printf("Found: %s\n", foundUser.Name)
	} else {
		fmt.Println("❌ User not found (returned nil)")
	}

	// 5. THE INTERFACE NIL TRAP (Interview Favorite!)
	fmt.Println("\n=== 5. INTERFACE NIL TRAP ===")

	var s *MyStruct = nil // s is nil pointer
	var i Printer = s     // i is NOT nil interface!

	fmt.Printf("s == nil: %t\n", s == nil) // true
	fmt.Printf("i == nil: %t\n", i == nil) // FALSE! (tricky!)

	fmt.Println("Why? Interface stores (type, value)")
	fmt.Println("i = (*MyStruct, nil) ≠ nil interface")
	fmt.Println("nil interface = (nil, nil)")

	// 6. DEFENSIVE PROGRAMMING
	fmt.Println("\n=== 6. DEFENSIVE PROGRAMMING ===")

	err := processUser(nil)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	}

	err = processUser(validUser)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	}

	// 7. NIL RECEIVER METHODS (Advanced)
	fmt.Println("\n=== 7. NIL RECEIVER METHODS ===")
	var nilUser *User = nil
	fmt.Printf("Nil user safe name: %s\n", nilUser.SafeName())

	// 8. KEY TAKEAWAYS
	fmt.Println("\n🎯 KEY TAKEAWAYS:")
	fmt.Println("1. Always check 'if pointer != nil' before dereferencing")
	fmt.Println("2. Nil pointer dereference = INSTANT PANIC")
	fmt.Println("3. Interface with nil pointer ≠ nil interface")
	fmt.Println("4. Use defensive programming in functions")
	fmt.Println("5. Methods can handle nil receivers gracefully")

	fmt.Println("\n💡 INTERVIEW TIP:")
	fmt.Println("When you see pointer code, ALWAYS ask: 'What if this is nil?'")
}

func printUserSafely(u *User) {
	if u == nil {
		fmt.Println("❌ Cannot print: pointer is nil")
		return
	}
	fmt.Printf("✅ User: %s (age %d)\n", u.Name, u.Age)
}

func findUser(name string) *User {
	// Simulate lookup that fails
	if name == "nonexistent" {
		return nil
	}
	return &User{Name: name, Age: 25}
}

func processUser(u *User) error {
	if u == nil {
		return fmt.Errorf("user cannot be nil")
	}
	fmt.Printf("✅ Processing user: %s\n", u.Name)
	return nil
}

// NIL RECEIVER METHOD - handles nil gracefully
func (u *User) SafeName() string {
	if u == nil {
		return "Unknown User"
	}
	return u.Name
}
