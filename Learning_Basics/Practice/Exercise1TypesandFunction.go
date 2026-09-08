package main

import (
	"fmt"
	"strings"
	"math"
	"errors"
	"encoding/json"
	// "os"
	// "strconv"
	// "bufio"
)

func clamp(value, min, max int) int {
	if value < min {
		return min
	} else if value > max {
		return max
	}
	return value
}

func sum(nums []int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

func wordCount(s string) map[string]int {
	mp := make(map[string]int)

	str := strings.Split(s, " ")
	for i := 0 ; i < len(str) ; i++ {
		mp[str[i]]++
	}
	return mp
}

type Rectangle struct {
	Width float64
	Height float64
}

func (rect Rectangle) Area() float64 {
	return rect.Width * rect.Height
}

func (rect Rectangle) Perimeter() float64 {
	return 2 * (rect.Height + rect.Width)
}

func double(n *int) {
	*n =  *n * 2
}

type Shape interface {
	Area() float64
}

type Circle struct {
	Radius float64
}

func printArea(s Shape) {
	fmt.Println("The area :", s.Area())
}

func (c Circle) Area() float64 {
    return math.Pi * c.Radius * c.Radius
}


func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("Division by zero is not allowed")
	}
	return a/b, nil
}

func generate(nums []int) <-chan int {
	ch := make(chan int)

	go func() {
		for _, n := range nums {
			ch <- n
		}
		close(ch)
	}()

	return ch
}

type User struct {
	Name string `json:"name"`
	Email string `json:"email"`
	Age int `json:"age"`
}

func main() {
	
	v1 := clamp(5, 1, 10)
	v2 := clamp(-3, 1, 10)
	v3 := clamp(15, 1, 10)

	fmt.Println(v1, v2, v3)

	nums := []int{1, 2, 3, 4, 5}
	empty := []int{}
	v4 := sum(nums)
	v5 := sum(empty)
	
	fmt.Println(v4)
	fmt.Println(v5)

	ans := wordCount("the cat sat on the mat")
	fmt.Println(ans)

	rect := Rectangle{
		Width: 5,
		Height: 3,
	}

	fmt.Printf("Area: %.0f\n", rect.Area())
	fmt.Printf("Perimeter: %.0f\n", rect.Perimeter())

	x := 7
	double(&x)
	fmt.Println(x)

	printArea(Circle{Radius: 4})
	printArea(Rectangle{Width: 5, Height: 3}) 


	result, err := divide(10, 2)
	
	if err != nil {
		fmt.Println("Error: ", err)
	} else {
		fmt.Println(result)
	}

	result, err = divide(10, 0)
	if err != nil {
		fmt.Println("Error: ", err)
	} else {
		fmt.Println(result)
	}

	ch := generate(nums)
	for n := range ch {
		fmt.Println(n)
	}

	user := User{Name: "Aks", Email: "aks@gmail.com", Age: 27}
	jsonData, err := json.Marshal(user)

	if err != nil {
		fmt.Println("Error marshalling: ", err)
		return
	}

	var decoded User
	err = json.Unmarshal(jsonData, &decoded)

	if err != nil {
		fmt.Println("Error unmarshalling : ", err)
		return
	}

	fmt.Println(string(jsonData))
	fmt.Println(decoded)
}