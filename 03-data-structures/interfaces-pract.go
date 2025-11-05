package main

import "fmt"

// type PaymentProcessor interface {
// 	Pay(amount float64)
// }

// type Stripe struct {
// }

// func (s Stripe) Pay(amount float64) {
// 	fmt.Printf("Amount paid via Stripe method %f\n", amount)
// }

// type PayPal struct {
// }

// func (p PayPal) Pay(amount float64) {
// 	fmt.Printf("Amount paid via Paypal method %f\n", amount)
// }

// func ProcessPayment(p PaymentProcessor, amount float64) {
// 	p.Pay(amount)
// }

type Notifier interface {
	Notify(msg string)
}

type EmailNotifier struct{}

func (e EmailNotifier) Notify(message string) {
	fmt.Println("Email : ", message)
}

type SMSNotifier struct{}

func (s SMSNotifier) Notify(message string) {
	fmt.Println("SMS : ", message)
}

func SendNotification(n Notifier) {
	n.Notify("Hello!")
}

func main() {
	// ss := Stripe{}
	// pp := PayPal{}

	// ProcessPayment(ss, 500.0)
	// ProcessPayment(pp, 40.0)

	ee := EmailNotifier{}
	s := SMSNotifier{}

	SendNotification(ee)
	SendNotification(s)

}
