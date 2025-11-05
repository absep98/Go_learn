package main

import "fmt"

/*
=================================================================
                    INTERFACES IN GO
=================================================================

WHAT IS AN INTERFACE?
- A set of method signatures
- Any type that implements those methods satisfies the interface
- Like a "contract" - if you have these methods, you're this type
- Go's way of achieving polymorphism

KEY POINTS:
1. Interfaces are implicit (no "implements" keyword needed)
2. No explicit inheritance
3. Duck typing: "If it walks like a duck and quacks like a duck..."
*/

// ==========================================
// STEP 1: DEFINE AN INTERFACE
// ==========================================

type Speaker interface {
	Speak() string
}

// ==========================================
// STEP 2: DEFINE TYPES THAT IMPLEMENT IT
// ==========================================

type Dog struct {
	Name string
}

// ✅ CORRECT: Implement Speak() for Dog
func (d Dog) Speak() string {
	return fmt.Sprintf("%s says: Woof!", d.Name)
}

type Robot struct {
	Model string
}

// ✅ CORRECT: Implement Speak() for Robot
func (r Robot) Speak() string {
	return fmt.Sprintf("%s says: Beep Beep", r.Model)
}

type Person struct {
	Name string
}

// ✅ CORRECT: Implement Speak() for Person
func (p Person) Speak() string {
	return fmt.Sprintf("%s says: Hello!", p.Name)
}

// ==========================================
// PAYMENT INTERFACE & IMPLEMENTATIONS (defined outside main)
// ==========================================

type PaymentMethod interface {
	Pay(amount float64) string
}

type CreditCard struct {
	CardNumber string
}

func (c CreditCard) Pay(amount float64) string {
	return fmt.Sprintf("Paid $%.2f with credit card ending in %s", amount, c.CardNumber[12:])
}

type PayPal struct {
	Email string
}

func (p PayPal) Pay(amount float64) string {
	return fmt.Sprintf("Paid $%.2f via PayPal to %s", amount, p.Email)
}

// ==========================================
// STEP 3: POLYMORPHISM - USE THE INTERFACE
// ==========================================

// This function accepts ANY type that implements Speaker
func MakeSpeakTwice(s Speaker) {
	fmt.Println("First time:", s.Speak())
	fmt.Println("Second time:", s.Speak())
	fmt.Println()
}

func main() {
	fmt.Println("🎯 INTERFACES DEMONSTRATION")
	fmt.Println("===========================\n")

	// ==========================================
	// 1. BASIC USAGE
	// ==========================================
	fmt.Println("1️⃣ BASIC INTERFACE USAGE:")
	fmt.Println("--------------------------")

	dog := Dog{Name: "Rex"}
	robot := Robot{Model: "T-800"}
	person := Person{Name: "Alice"}

	// All three can be treated as Speaker
	MakeSpeakTwice(dog)
	MakeSpeakTwice(robot)
	MakeSpeakTwice(person)

	// ==========================================
	// 2. STORING MULTIPLE TYPES IN A SLICE
	// ==========================================
	fmt.Println("2️⃣ SLICE OF INTERFACE TYPE:")
	fmt.Println("--------------------------")

	// A slice of Speaker can hold ANY type that implements Speaker
	var speakers []Speaker
	speakers = append(speakers, dog, robot, person)

	for i, speaker := range speakers {
		fmt.Printf("[%d] %s\n", i, speaker.Speak())
	}

	// ==========================================
	// 3. TYPE ASSERTION (Getting the original type back)
	// ==========================================
	fmt.Println("\n3️⃣ TYPE ASSERTION:")
	fmt.Println("-------------------")

	var s Speaker = dog

	// Type assertion: s.(Dog) gets the Dog back
	if d, ok := s.(Dog); ok {
		fmt.Printf("✅ It's a Dog: %s\n", d.Name)
	}

	// What if we assert wrong type?
	if r, ok := s.(Robot); ok {
		fmt.Printf("It's a Robot: %s\n", r.Model)
	} else {
		fmt.Println("❌ It's NOT a Robot")
	}

	// ==========================================
	// 4. TYPE SWITCH (Match multiple types)
	// ==========================================
	fmt.Println("\n4️⃣ TYPE SWITCH:")
	fmt.Println("----------------")

	for i, speaker := range speakers {
		switch v := speaker.(type) {
		case Dog:
			fmt.Printf("[%d] Found a Dog: %s\n", i, v.Name)
		case Robot:
			fmt.Printf("[%d] Found a Robot: %s\n", i, v.Model)
		case Person:
			fmt.Printf("[%d] Found a Person: %s\n", i, v.Name)
		default:
			fmt.Printf("[%d] Unknown speaker\n", i)
		}
	}

	// ==========================================
	// 5. NIL INTERFACE TRAP (Important!)
	// ==========================================
	fmt.Println("\n5️⃣ NIL INTERFACE TRAP:")
	fmt.Println("----------------------")

	var speaker Speaker
	fmt.Printf("speaker == nil: %t (Both type and value are nil)\n", speaker == nil)

	// Assign a nil pointer
	var nilDog *Dog = nil
	speaker = nilDog
	fmt.Printf("speaker = nil pointer: speaker == nil is %t ❌ TRICKY!\n", speaker == nil)
	fmt.Println("Why? Interface stores (type, value) = (*Dog, nil) ≠ nil interface")

	// ==========================================
	// 6. EMPTY INTERFACE (can hold ANY type)
	// ==========================================
	fmt.Println("\n6️⃣ EMPTY INTERFACE interface{}:")
	fmt.Println("---------------------------------")

	var anything interface{}
	anything = 42
	fmt.Printf("anything = 42: %v (type: %T)\n", anything, anything)

	anything = "hello"
	fmt.Printf("anything = \"hello\": %v (type: %T)\n", anything, anything)

	anything = dog
	fmt.Printf("anything = dog: %v (type: %T)\n", anything, anything)

	// ==========================================
	// 7. INTERFACE COMPOSITION
	// ==========================================
	fmt.Println("\n7️⃣ INTERFACE COMPOSITION:")
	fmt.Println("-------------------------")

	type Mover interface {
		Move() string
	}

	type Actor interface {
		Speak() string
		Move() string
	}

	fmt.Println("Actor interface = Speak() + Move()")
	fmt.Println("(Would need to implement both methods)")

	// ==========================================
	// 8. PRACTICAL EXAMPLE - Payment Processing
	// ==========================================
	fmt.Println("\n8️⃣ PRACTICAL EXAMPLE - PAYMENT PROCESSING:")
	fmt.Println("------------------------------------------")

	// Create different payment methods
	creditCard := CreditCard{CardNumber: "1234-5678-9012-3456"}
	paypal := PayPal{Email: "user@example.com"}

	// Process multiple payments with different methods
	payments := []PaymentMethod{creditCard, paypal}

	for _, payment := range payments {
		fmt.Println(payment.Pay(99.99))
	}
}
