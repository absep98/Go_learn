package main

import "fmt"

// Define structs
type Person struct {
	Name string
	Age  int
	City string
}

// Struct with embedded fields
type Address struct {
	Street  string
	City    string
	ZipCode string
}

type Employee struct {
	Person  // Embedded struct
	Address // Embedded struct
	ID      int
	Salary  float64
}

// Methods on structs
func (p Person) Greet() string {
	return fmt.Sprintf("Hi, I'm %s from %s", p.Name, p.City)
}

// Method with pointer receiver (can modify the struct)
func (p *Person) HaveBirthday() {
	p.Age++
}

// Method that returns calculated value
func (e Employee) GetAnnualSalary() float64 {
	return e.Salary * 12
}

func main() {
	fmt.Println("=== Structs ===")

	// Creating structs
	fmt.Println("\n--- Creating Structs ---")

	// Struct literal
	alice := Person{
		Name: "Alice",
		Age:  25,
		City: "New York",
	}
	fmt.Printf("Alice: %+v\n", alice)

	// Positional initialization
	bob := Person{"Bob", 30, "San Francisco"}
	fmt.Printf("Bob: %+v\n", bob)

	// Zero value struct
	var charlie Person
	fmt.Printf("Charlie (zero value): %+v\n", charlie)

	// Accessing fields
	fmt.Println("\n--- Accessing Fields ---")
	fmt.Printf("Alice's name: %s\n", alice.Name)
	fmt.Printf("Bob's age: %d\n", bob.Age)

	// Modifying fields
	charlie.Name = "Charlie"
	charlie.Age = 28
	charlie.City = "Chicago"
	fmt.Printf("Charlie after assignment: %+v\n", charlie)

	// Pointer to struct
	fmt.Println("\n--- Pointers to Structs ---")
	alicePtr := &alice
	fmt.Printf("Alice via pointer: %+v\n", *alicePtr)

	// Modifying through pointer
	alicePtr.Age = 26 // Automatic dereferencing
	fmt.Printf("Alice after pointer modification: %+v\n", alice)

	// Methods
	fmt.Println("\n--- Methods ---")
	fmt.Println(alice.Greet())
	fmt.Println(bob.Greet())

	fmt.Printf("Alice's age before birthday: %d\n", alice.Age)
	alice.HaveBirthday()
	fmt.Printf("Alice's age after birthday: %d\n", alice.Age)

	// Embedded structs
	fmt.Println("\n--- Embedded Structs ---")
	emp := Employee{
		Person: Person{
			Name: "John Doe",
			Age:  35,
			City: "Boston",
		},
		Address: Address{
			Street:  "123 Main St",
			City:    "Boston", // Note: this shadows Person.City
			ZipCode: "02101",
		},
		ID:     12345,
		Salary: 5000.0,
	}

	fmt.Printf("Employee: %+v\n", emp)

	// Accessing embedded fields
	fmt.Printf("Employee name: %s\n", emp.Name)                 // From Person
	fmt.Printf("Employee street: %s\n", emp.Street)             // From Address
	fmt.Printf("Employee person city: %s\n", emp.Person.City)   // Specific field
	fmt.Printf("Employee address city: %s\n", emp.Address.City) // Specific field

	// Methods on embedded structs
	fmt.Println(emp.Greet()) // Method from Person
	fmt.Printf("Annual salary: $%.2f\n", emp.GetAnnualSalary())

	// Anonymous structs
	fmt.Println("\n--- Anonymous Structs ---")
	config := struct {
		Host string
		Port int
		SSL  bool
	}{
		Host: "localhost",
		Port: 8080,
		SSL:  false,
	}

	fmt.Printf("Server config: %+v\n", config)

	// Slice of structs
	fmt.Println("\n--- Slice of Structs ---")
	people := []Person{alice, bob, charlie}

	fmt.Println("All people:")
	for i, person := range people {
		fmt.Printf("%d: %s (%d) from %s\n", i+1, person.Name, person.Age, person.City)
	}
}
