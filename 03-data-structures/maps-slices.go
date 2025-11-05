package main

import (
	"fmt"
	"sort"
	"strings"
)

// CountWords returns a map with word -> count.
// It lowercases words and splits on whitespace (basic).
func CountWords(sentence string) map[string]int {
	mp := make(map[string]int)
	words := strings.Fields(sentence)

	for _, word := range words {
		word = strings.ToLower(word)
		mp[word]++
	}
	return mp
}

// helper max for ints
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	// ----------------------
	// Problem 1: simple map
	// ----------------------
	people := map[string]int{
		"Aks":   26,
		"Rahul": 17,
		"Priya": 30,
	}

	fmt.Println("Users with age >= 20:")
	// note: map iteration order is random
	for name, age := range people {
		if age >= 20 {
			fmt.Printf("- %s: %d\n", name, age)
		}
	}

	// ----------------------
	// Problem 2: products map
	// ----------------------
	type Product struct {
		Name  string
		Price int
	}

	products := map[string]Product{
		"Apple":  {Name: "MacBook Air", Price: 80_000},
		"Lenovo": {Name: "Ideapad 330", Price: 45_000},
		"Dell":   {Name: "Inspiron 15", Price: 145_000},
	}

	// find most expensive product
	expensive := 0
	var expensiveBrand string
	for brand, product := range products {
		if product.Price > expensive {
			expensive = product.Price
			expensiveBrand = brand
		}
	}

	// if you want stable output, sort the keys
	brands := make([]string, 0, len(products))
	for b := range products {
		brands = append(brands, b)
	}
	sort.Strings(brands)

	fmt.Println("\nProducts:")
	for _, brand := range brands {
		p := products[brand]
		fmt.Printf("%s: %s - ₹%d\n", brand, p.Name, p.Price)
	}
	fmt.Printf("Most expensive: %s costing ₹%d\n", expensiveBrand, expensive)

	// ----------------------
	// Problem 3: CountWords
	// ----------------------
	text := "hi hello hi test"
	counts := CountWords(text)
	fmt.Println("\nWord counts for:", text)
	// print in deterministic order
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%q: %d\n", k, counts[k])
	}
}
