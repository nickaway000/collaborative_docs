package main

import (

	"fmt"

	"log"
	"net/http"


	"github.com/nikhil/collaborative-doc-platform/user-service/handlers"
)

func main() {
	err := handlers.InitDB()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	fileServer := http.FileServer(http.Dir("static"))
	http.Handle("/", fileServer)

	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)

	fmt.Printf("Starting server at port 9000\n")
	log.Fatal(http.ListenAndServe(":9000", nil))
}
