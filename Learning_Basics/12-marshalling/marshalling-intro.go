package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	AppName string `json:"app_name"`
	Port    int    `json:"port"`
	Debug   bool   `json:"debug"`
}

func LoadConfig(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {
	cfg, err := LoadConfig("config.json")
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	fmt.Println("App:", cfg.AppName)
	fmt.Println("Port:", cfg.Port)
	fmt.Println("Debug:", cfg.Debug)
}
