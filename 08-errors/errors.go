package main

import (
	"errors"
	"fmt"
	"os"
)

func SafeDivide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("division by zero.")
	}
	return a / b, nil
}

func ValidateUsername(name string) error {
	if name == "" {
		return errors.New("username cannot be empty")
	}
	if len(name) < 3 {
		return errors.New("username too short.")
	}
	return nil
}

func LoadConfig(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("load config failed: %w", err)
	}
	return string(data), nil
}

func ReadAndValidate(path string) (string, error) {
	data, err := LoadConfig(path)
	if err != nil {
		return "", err
	}
	if data == "" {
		return "", errors.New("config is empty")
	}
	return data, nil
}

func main() {
	if res, err := SafeDivide(10, 0); err != nil {
		fmt.Println("SafeDivide error: ", err)
	} else {
		fmt.Println("SafeDivide result: ", res)
	}

	if err := ValidateUsername("ab"); err != nil {
		fmt.Println("ValidateUsername error:", err)
	}

	if _, err := LoadConfig("missing.txt"); err != nil {
		fmt.Println("LoadConfig error:", err)
	}
}
