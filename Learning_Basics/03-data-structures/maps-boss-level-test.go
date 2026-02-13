package main

import "fmt"

type Stats struct {
	HP    int
	Level int
}

func main() {
	// FIX #1: Initialize with make
	playerBase := make(map[string]*Stats)
	playerBase["Hero"] = &Stats{HP: 100, Level: 1}
	playerBase["Duke"] = &Stats{HP: 400, Level: 1}

	fmt.Println("=== PLAYER BASE ===")
	for name, stats := range playerBase {
		fmt.Printf("%s: HP=%d, Level=%d\n", name, stats.HP, stats.Level)
	}

	// FIX #2: Use pointer type so fields are addressable
	fmt.Println("\n=== TEAM ===")
	team := make(map[string]*Stats)
	team["Archer"] = &Stats{HP: 50, Level: 5}

	// This now works! We are modifying the actual data at the address.
	team["Archer"].HP = 60

	for name, stats := range team {
		fmt.Printf("%s: HP=%d, Level=%d\n", name, stats.HP, stats.Level)
	}

	fmt.Printf("\n=== COMPARISON ===\n")
	fmt.Printf("Hero HP: %d, Archer HP: %d\n", playerBase["Hero"].HP, team["Archer"].HP)
}
