package main

import "fmt"

func main() {
	// --- Value copy vs pointer alias ---

	// Value copy: changing y does NOT affect x
	x := 10
	y := x
	y = 999
	fmt.Println("Value copy:")
	fmt.Println("  x =", x) // 10 (unchanged)
	fmt.Println("  y =", y) // 999

	// Pointer: changing *p DOES affect x
	a := 10
	p := &a   // p holds ADDRESS of a
	*p = 999  // write to the memory p points at
	fmt.Println("\nPointer alias:")
	fmt.Println("  a  =", a)  // 999 (changed via pointer)
	fmt.Println("  *p =", *p) // 999

	// --- Why flag.Int returns *int ---
	// flag.Int("start", 8080, "...") returns *int
	// because flag.Parse() needs to WRITE into that
	// memory later. If it returned plain int, the
	// parsed value couldn't update the caller's variable.

	// --- Two pointer operations ---
	val := 42
	ptr := &val

	*ptr = 100     // write THROUGH the pointer (changes val)
	fmt.Println("\nWrite through pointer:")
	fmt.Println("  val =", val) // 100

	other := 77
	ptr = &other   // redirect the pointer itself (val unchanged)
	fmt.Println("\nRedirect pointer:")
	fmt.Println("  val =", val)  // still 100
	fmt.Println("  *ptr =", *ptr) // 77 (now points to other)
}
