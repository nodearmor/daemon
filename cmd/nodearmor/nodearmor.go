package main

import (
	"fmt"
)

func main() {
	LoadConfig()

	fmt.Println("Hello World!")
	fmt.Printf("Backend: %s", Config.GetString("BackendURL"))
}
