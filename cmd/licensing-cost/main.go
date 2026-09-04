package main

import (
	"fmt"
	"licensing-cost/internal/api"
	"log"
)

func main() {
	fmt.Println("y")
	log.Println("Application start!")
	api.StartServer()
	log.Println("Application terminated!")
}
