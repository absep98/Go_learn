package main

import "fmt"

type BankAccount struct {
	Owner   string
	Balance int
}

// CORRECT: Pointer receiver - modifies original
func (bank *BankAccount) DepositCorrect(amount int) {
	bank.Balance += amount
	fmt.Printf("Inside DepositCorrect: Balance is now %d\n", bank.Balance)
}

// WRONG: Value receiver - modifies copy only
func (bank BankAccount) DepositWrong(amount int) {
	bank.Balance += amount
	fmt.Printf("Inside DepositWrong: Balance appears to be %d\n", bank.Balance)
}

func main() {
	fmt.Println("=== POINTER RECEIVER DEMO ===")

	account1 := BankAccount{Owner: "Alice", Balance: 100}
	fmt.Printf("Before DepositCorrect: %d\n", account1.Balance)
	account1.DepositCorrect(50)
	fmt.Printf("After DepositCorrect: %d\n", account1.Balance)

	fmt.Println("\n=== VALUE RECEIVER DEMO ===")

	account2 := BankAccount{Owner: "Bob", Balance: 100}
	fmt.Printf("Before DepositWrong: %d\n", account2.Balance)
	account2.DepositWrong(50)
	fmt.Printf("After DepositWrong: %d (UNCHANGED!)\n", account2.Balance)
}
