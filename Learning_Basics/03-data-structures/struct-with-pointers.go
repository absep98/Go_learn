package main

import "fmt"

type BankAccount struct {
	Owner   string
	Balance int
}

func (bank *BankAccount) Deposit(amount int) {
	bank.Balance += amount
}

func (bank *BankAccount) WithDraw(amount int) bool {
	if bank.Balance < amount {
		fmt.Println("Not enough balance")
		return false
	}
	bank.Balance -= amount
	return true
}

func (bank *BankAccount) Info() {
	fmt.Printf("%s user have current balance : %d", bank.Owner, bank.Balance)
}

func main() {
	bank := BankAccount{
		Owner:   "Aks",
		Balance: 100}
	bank.Deposit(10)
	bank.WithDraw(30)
	bank.Info()
}
