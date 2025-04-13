package main

import (
    "fmt"
    "log"
    "net/http"
    "os"

    "Bank-Management-System/config"
    "Bank-Management-System/routes"
    "github.com/joho/godotenv"
)

func main() {
    // Load .env file for local development
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    os.MkdirAll("Uploads/photos", 0755)
    os.MkdirAll("Uploads/ids", 0755)

    config.InitDB()
    router := routes.Routes()

    fmt.Println("Server running on http://localhost:9000")
    log.Fatal(http.ListenAndServe(":9000", router))
}