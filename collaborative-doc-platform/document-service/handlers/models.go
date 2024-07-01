package handlers

import "time"

type Document struct {
    ID        int       `json:"id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// Change represents a change made to a document
type Change struct {
    ID        int                    `json:"id"`
    DocumentID int                   `json:"document_id"`
    Delta     map[string]interface{} `json:"delta"`
    CreatedAt time.Time              `json:"created_at"`
}
