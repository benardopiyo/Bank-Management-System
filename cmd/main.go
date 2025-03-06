package main

import (
	"fmt"
	"log"
	"net/http"

	"Bank-Management-System/config"
	"Bank-Management-System/routes"
)

func main() {
	config.InitDB()
	router := routes.Routes()

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
