package main

import "fmt"

func main() {
	fmt.Println("🔍 MAP ZERO VALUES DEMONSTRATION\n")

	// Example 1: map[string]int
	fmt.Println("1️⃣ map[string]int")
	m1 := make(map[string]int)
	value1 := m1["non_existent_key"]
	fmt.Printf("m1['non_existent_key'] = %v (type: %T)\n", value1, value1)
	fmt.Println("Zero value of int: 0\n")

	// Example 2: map[string]string
	fmt.Println("2️⃣ map[string]string")
	m2 := make(map[string]string)
	value2 := m2["non_existent_key"]
	fmt.Printf("m2['non_existent_key'] = '%v' (type: %T)\n", value2, value2)
	fmt.Println("Zero value of string: empty string \"\"\n")

	// Example 3: map[string]bool
	fmt.Println("3️⃣ map[string]bool")
	m3 := make(map[string]bool)
	value3 := m3["non_existent_key"]
	fmt.Printf("m3['non_existent_key'] = %v (type: %T)\n", value3, value3)
	fmt.Println("Zero value of bool: false\n")

	// Example 4: map[string][]int (slice)
	fmt.Println("4️⃣ map[string][]int")
	m4 := make(map[string][]int)
	value4 := m4["non_existent_key"]
	fmt.Printf("m4['non_existent_key'] = %v (type: %T)\n", value4, value4)
	fmt.Println("Zero value of slice: nil\n")

	// ===== KEY CONCEPT =====
	fmt.Println("📝 KEY CONCEPT:")
	fmt.Println("The zero value depends on the VALUE TYPE, not the KEY type!")
	fmt.Println("\nmap[keyType]valueType")
	fmt.Println("        ↑        ↑")
	fmt.Println("    ignored   THIS determines zero value")

	// ===== PRACTICAL USAGE =====
	fmt.Println("\n💡 PRACTICAL: Using ok to check existence")
	m := make(map[string]int)
	m["age"] = 25

	// Method 1: Without ok (can't distinguish missing vs 0)
	fmt.Println("\n❌ Without 'ok':")
	value := m["age"]
	fmt.Printf("m['age'] = %d\n", value)

	value = m["missing"]
	fmt.Printf("m['missing'] = %d\n", value)
	fmt.Println("Problem: Can't tell if 'missing' key exists or value is 0!")

	// Method 2: With ok (recommended!)
	fmt.Println("\n✅ With 'ok' (RECOMMENDED):")
	if value, ok := m["age"]; ok {
		fmt.Printf("Found 'age': %d\n", value)
	} else {
		fmt.Println("Key 'age' not found")
	}

	if value, ok := m["missing"]; ok {
		fmt.Printf("Found 'missing': %d\n", value)
	} else {
		fmt.Printf("Key 'missing' not found\n")
	}
}