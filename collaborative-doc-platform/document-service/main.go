package main

import (
    "log"
    "net/http"

    "github.com/nikhil/collaborative-doc-platform/document-service/handlers"
    "github.com/gorilla/mux"
)

func main() {
    // Initialize the database connection
    err := handlers.InitDB()
    if err != nil {
        log.Fatalf("Error initializing database: %v", err)
        
    }
    go handlers.BroadcastChanges()

    // Create a new router
    router := mux.NewRouter()

    // Define routes
    router.HandleFunc("/documents", handlers.CreateDocumentHandler).Methods("POST")
    router.HandleFunc("/documents/{id}", handlers.GetDocumentHandler).Methods("GET")
    router.HandleFunc("/documents", handlers.ListDocumentsHandler).Methods("GET")
    router.HandleFunc("/documents/{id}", handlers.UpdateDocumentHandler).Methods("PUT")
    router.HandleFunc("/ws", handlers.WebSocketHandler)

    // Serve static files
    fs := http.FileServer(http.Dir("./static"))
    router.PathPrefix("/").Handler(fs)

    log.Println("Starting server on :8081")
    log.Fatal(http.ListenAndServe(":8081", router))
}
