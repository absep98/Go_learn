package main

import "fmt"

type Speaker interface {
	Speak() string
}

type Dog struct{}

// ✅ CORRECT: Method syntax
func (d Dog) Speak() string {
	return "Woof!"
}

type Robot struct{}

// ✅ CORRECT: Method syntax
func (r Robot) Speak() string {
	return "Beep Beep"
}

func main() {
	var s Speaker
	fmt.Println("s == nil:", s == nil) // true (Both Type and Value are nil)

	var d *Dog = nil
	s = d
	fmt.Println("s = nil pointer, s == nil:", s == nil) // false! (Type is *Dog, even if Value is nil)

	// ✅ CORRECT USAGE: Create actual instances
	dog := Dog{}
	robot := Robot{}

	s = dog
	fmt.Println("s = dog:", s.Speak())

	s = robot
	fmt.Println("s = robot:", s.Speak())
}
