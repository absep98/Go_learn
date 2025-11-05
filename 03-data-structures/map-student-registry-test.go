package main

import "fmt"

type Student struct {
	Name  string
	Marks []int
}

func main() {
	st1 := Student{Name: "Aks", Marks: []int{97, 90, 100}}
	st2 := Student{Name: "Dhanda", Marks: []int{90, 90, 100}}
	mp := map[string][]int{}
	mp[st1.Name] = st1.Marks
	mp[st2.Name] = st2.Marks

	for name, marks := range mp {
		sum := 0
		n := len(marks)
		for i := range n {
			sum += marks[i]
		}
		fmt.Printf("Name : %s and average of marks : %d\n", name, sum/n)
	}
}
