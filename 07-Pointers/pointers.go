package main

import "fmt"

type User struct {
	Name string
	Age  int
}

func Update(u *User) {
	u.Age++
}

func main() {
	user := User{"Aks", 26}
	Update(&user)
	fmt.Println(user.Name, user.Age)
}
