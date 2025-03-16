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

	fmt.Println("Server running on http://localhost:9000")
	log.Fatal(http.ListenAndServe(":9000", router))
}
