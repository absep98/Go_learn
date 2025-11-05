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

func demonstrateNilPointers() {
	fmt.Println("=== NIL POINTER DEMONSTRATION ===")

	// 1. DECLARING NIL POINTERS
	var p *User         // p is nil (zero value of pointer type)
	var num *int        // num is nil
	var slice *[]string // slice is nil

	fmt.Printf("p is nil: %t\n", p == nil)
	fmt.Printf("num is nil: %t\n", num == nil)
	fmt.Printf("slice is nil: %t\n", slice == nil)

	// 2. WHAT HAPPENS WHEN YOU DEREFERENCE NIL?
	fmt.Println("\n--- Attempting to dereference nil pointer ---")

	// ⚠️ THIS WILL PANIC! (Uncomment to see crash)
	fmt.Println("User age:", p.Age) // PANIC: runtime error: invalid memory address

	// Instead, we should check first:
	if p != nil {
		fmt.Println("User age:", p.Age)
	} else {
		fmt.Println("❌ Cannot access Age: pointer is nil")
	}

	// 3. PROPER NIL CHECKING PATTERNS
	fmt.Println("\n--- Safe NIL checking patterns ---")

	// Pattern 1: Check before use
	safelyPrintUser(p)   // nil pointer
	safelyPrintUser(nil) // explicitly nil

	// Pattern 2: Create valid pointer first
	validUser := &User{Name: "Alice", Age: 25}
	safelyPrintUser(validUser) // valid pointer

	// 4. COMMON NIL POINTER BUGS
	fmt.Println("\n--- Common NIL pointer bugs ---")

	// Bug 1: Forgetting to initialize
	var users []*User
	users = append(users, nil) // Adding nil pointer!

	for i, user := range users {
		if user != nil {
			fmt.Printf("User %d: %s\n", i, user.Name)
		} else {
			fmt.Printf("User %d: NIL POINTER!\n", i)
		}
	}

	// Bug 2: Function returns nil
	user := findUserByID(999) // Returns nil (not found)
	if user != nil {
		fmt.Printf("Found user: %s\n", user.Name)
	} else {
		fmt.Println("❌ User not found (nil returned)")
	}

	// 5. INTERFACE NIL TRAP (Advanced)
	fmt.Println("\n--- Interface NIL trap ---")
	demonstrateInterfaceNilTrap()
}

func safelyPrintUser(u *User) {
	if u == nil {
		fmt.Println("❌ Cannot print user: pointer is nil")
		return
	}
	fmt.Printf("✅ User: %s (age %d)\n", u.Name, u.Age)
}

func findUserByID(id int) *User {
	// Simulate database lookup that fails
	if id == 999 {
		return nil // User not found
	}
	return &User{Name: "Found User", Age: 30}
}

// 6. THE INTERFACE NIL TRAP (Interview Favorite!)
func demonstrateInterfaceNilTrap() {
	// This is the trap:
	var s *MyStruct = nil // s is nil pointer
	var i Printer = s     // i is NOT nil interface!

	fmt.Printf("s == nil: %t\n", s == nil) // true
	fmt.Printf("i == nil: %t\n", i == nil) // FALSE! (tricky!)

	// Why? Because interface stores (type, value)
	// i = (*MyStruct, nil) which is NOT nil interface
	// nil interface = (nil, nil)

	if i == nil {
		fmt.Println("Interface is nil")
	} else {
		fmt.Println("❌ Interface is NOT nil (even though pointer is nil!)")
		// Calling i.Print() would still panic because underlying value is nil
	}
}

// 7. DEFENSIVE PROGRAMMING PATTERNS
func processUser(u *User) error {
	// Always check nil first!
	if u == nil {
		return fmt.Errorf("user cannot be nil")
	}

	// Now safe to use
	fmt.Printf("Processing user: %s\n", u.Name)
	return nil
}

// 8. NIL RECEIVER METHODS (Advanced Pattern)
func (u *User) SafeName() string {
	if u == nil {
		return "Unknown User" // Handle nil receiver gracefully
	}
	return u.Name
}

func main() {
	fmt.Println("🚨 GO NIL POINTERS - COMPLETE GUIDE")
	fmt.Println("====================================")

	demonstrateNilPointers()

	// 9. TESTING NIL RECEIVER METHODS
	fmt.Println("\n--- NIL receiver method ---")
	var nilUser *User = nil
	fmt.Printf("Nil user name: %s\n", nilUser.SafeName()) // Won't panic!

	// 10. REAL-WORLD EXAMPLE
	fmt.Println("\n--- Real-world example ---")
	err := processUser(nil)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	}

	validUser := &User{Name: "Bob", Age: 30}
	err = processUser(validUser)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	}

	fmt.Println("\n🎯 KEY TAKEAWAYS:")
	fmt.Println("1. Always check 'if pointer != nil' before dereferencing")
	fmt.Println("2. Nil pointer dereference = instant panic")
	fmt.Println("3. Interface with nil pointer ≠ nil interface (tricky!)")
	fmt.Println("4. Use defensive programming - check nil in functions")
	fmt.Println("5. Methods can handle nil receivers gracefully")

	fmt.Println("\n💡 INTERVIEW TIP:")
	fmt.Println("When you see pointer code, ALWAYS ask: 'What if this is nil?'")
}
