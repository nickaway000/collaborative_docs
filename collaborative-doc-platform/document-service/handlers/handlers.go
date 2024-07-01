package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

var (
	clients       = make(map[int]map[*websocket.Conn]bool)
	clientsMutex  sync.Mutex
	broadcastChan = make(chan WebSocketMessage)
)

type WebSocketMessage struct {
	DocumentID int                    `json:"document_id"`
	Message    map[string]interface{} `json:"message"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

func SaveChange(documentID int, delta interface{}) error {
	deltaJSON, err := json.Marshal(delta)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO changes (document_id, delta, created_at) VALUES ($1, $2, $3)",
		documentID, string(deltaJSON), time.Now())
	return err
}

func CreateDocumentHandler(w http.ResponseWriter, r *http.Request) {
    var doc Document
    err := json.NewDecoder(r.Body).Decode(&doc)
    if err != nil {
        fmt.Println("Error decoding JSON:", err)
        http.Error(w, "Invalid input", http.StatusBadRequest)
        return
    }

    fmt.Println("Received document:", doc)

    doc.CreatedAt = time.Now()
    doc.UpdatedAt = time.Now()

    err = db.QueryRow("INSERT INTO documents (title, content, created_at, updated_at) VALUES ($1, $2, $3, $4) RETURNING id",
        doc.Title, doc.Content, doc.CreatedAt, doc.UpdatedAt).Scan(&doc.ID)
    if err != nil {
        fmt.Println("Error inserting document:", err)
        http.Error(w, "Server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(doc)
}



func GetDocumentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	var doc Document
	err = db.QueryRow("SELECT id, title, content, created_at, updated_at FROM documents WHERE id = $1", id).
		Scan(&doc.ID, &doc.Title, &doc.Content, &doc.CreatedAt, &doc.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Document not found", http.StatusNotFound)
		} else {
			http.Error(w, "Server error", http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(doc)
}

func ListDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, title, content, created_at, updated_at FROM documents")
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var doc Document
		err = rows.Scan(&doc.ID, &doc.Title, &doc.Content, &doc.CreatedAt, &doc.UpdatedAt)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		documents = append(documents, doc)
	}

	json.NewEncoder(w).Encode(documents)
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	documentIDStr := r.URL.Query().Get("doc")
	documentID, err := strconv.Atoi(documentIDStr)
	if err != nil {
		log.Println("Invalid document ID:", documentIDStr)
		return
	}

	clientsMutex.Lock()
	if _, ok := clients[documentID]; !ok {
		clients[documentID] = make(map[*websocket.Conn]bool)
	}
	clients[documentID][conn] = true
	clientsMutex.Unlock()

	defer func() {
		clientsMutex.Lock()
		delete(clients[documentID], conn)
		if len(clients[documentID]) == 0 {
			delete(clients, documentID)
		}
		clientsMutex.Unlock()
	}()

	initialContent, title, err := getInitialDocumentContent(documentID)
	if err != nil {
		log.Println("Error loading document:", err)
		return
	}

	err = conn.WriteJSON(map[string]interface{}{
		"type":    "initial",
		"title":   title,
		"content": initialContent,
	})
	if err != nil {
		log.Println("Error sending initial content:", err)
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		var msg map[string]interface{}
		err = json.Unmarshal(message, &msg)
		if err != nil {
			log.Println("Unmarshal error:", err)
			continue
		}

		msgType := msg["type"].(string)
		if msgType == "edit" {
			delta := msg["delta"]
			err = SaveChange(documentID, delta)
			if err != nil {
				log.Println("Error saving change:", err)
			}

			broadcastChan <- WebSocketMessage{
				DocumentID: documentID,
				Message:    msg,
			}
		}
	}
}

func getInitialDocumentContent(documentID int) (interface{}, string, error) {
	var content string
	var title string
	err := db.QueryRow("SELECT title, content FROM documents WHERE id = $1", documentID).Scan(&title, &content)
	if err != nil {
		return nil, "", err
	}

	var delta interface{}
	if err = json.Unmarshal([]byte(content), &delta); err != nil {
		return nil, title, nil // Returning empty delta if parsing fails
	}
	return delta, title, nil
}

func BroadcastChanges() {
	for {
		msg := <-broadcastChan
		clientsMutex.Lock()
		for conn := range clients[msg.DocumentID] {
			err := conn.WriteJSON(msg.Message)
			if err != nil {
				log.Println("Error broadcasting message:", err)
				conn.Close()
				delete(clients[msg.DocumentID], conn)
			}
		}
		clientsMutex.Unlock()
	}
}

func UpdateDocumentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid document ID", http.StatusBadRequest)
		return
	}

	var doc struct {
		Title   string      `json:"title"`
		Content interface{} `json:"content"`
	}
	err = json.NewDecoder(r.Body).Decode(&doc)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	contentJSON, err := json.Marshal(doc.Content)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("UPDATE documents SET title = $1, content = $2, updated_at = $3 WHERE id = $4",
		doc.Title, string(contentJSON), time.Now(), id)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
