package main

import (
	"fmt"
	"bufio"
	"os"
)

func reverse(s string) string {
	runes := []rune(s)
	reversedRunes := make([]rune, len(s))
	for i := len(runes)-1 ; i >= 0 ; i-- {
		reversedRunes[len(runes) - 1 - i] = runes[i]
	}
	return string(reversedRunes)
}

func isPalindrome(s string) bool {
	orgStr := s

	reversedStr := reverse(s)

	if orgStr == reversedStr {
		return true
	}
	return false
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	input := scanner.Text()
	v1 := reverse(input)
	fmt.Println(v1)

	v2 := isPalindrome(input)
	fmt.Println(v2)
}