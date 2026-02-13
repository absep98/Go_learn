package main

import (
	"fmt"
)

type User struct {
	Name string
	Age  int
}

func FilterByMinAge(users []User, minAge int) []User {
	result := []User{}

	for _, u := range users {
		if u.Age >= minAge {
			result = append(result, u)
		}
	}

	return result
}

func main() {
	users := []User{
		{Name: "Abhishek", Age: 26},
		{Name: "Rahul", Age: 17},
		{Name: "Priya", Age: 30},
		{Name: "Ankit", Age: 15},
	}

	filtered := FilterByMinAge(users, 18)

	fmt.Println("Filtered Users:")
	for _, u := range filtered {
		fmt.Println(u.Name, "-", u.Age)
	}
}
