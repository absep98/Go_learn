package main

import (
	"fmt"
)

type Entry struct {
    ID       int
    UserID   int
    Text     string
    Mood     int
    Category string
}

func filterEntries(entries []Entry, category string, minMood int) []Entry {
	filteredEntries := []Entry{}

	for _, entry := range entries {
		if category != "" && entry.Category != category {
			continue
		}

		if minMood != 0 && entry.Mood < minMood {
			continue
		}
		filteredEntries = append(filteredEntries, entry)
	}
	return filteredEntries
}


func main() {

	entries := []Entry{
		{ID: 1, UserID: 1, Text: "meeting", Mood: 8, Category: "work"},
		{ID: 2, UserID: 2, Text: "fun", Mood: 10, Category: "joy"},
		{ID: 3, UserID: 3, Text: "games", Mood: 10, Category: "games"},
		{ID: 4, UserID: 4, Text: "emotional", Mood: 5, Category: "work"},
	}

	result := filterEntries(entries, "work", 0)
	fmt.Println(result)
}