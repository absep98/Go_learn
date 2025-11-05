package main

import (
	"fmt"
	"myapp/models"
	"myapp/utils"
)

func main() {
	u := models.User{Name: "Aks", Age: 26}
	fmt.Println(utils.Add(2, 3))
	fmt.Println(u.Name)
}
