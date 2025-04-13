package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"Bank-Management-System/config"
	"Bank-Management-System/routes"
)

func main() {
	// Create upload directories
	os.MkdirAll("uploads/photos", 0755)
	os.MkdirAll("uploads/ids", 0755)

	config.InitDB()
	router := routes.Routes()

	fmt.Println("Server running on http://localhost:9000")
	log.Fatal(http.ListenAndServe(":9000", router))
}
