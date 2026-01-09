package models

// Chunk represents a chunk of a document retrieved from Chroma.
// This is a lightweight data structure used for passing chunk data between services.
type Chunk struct {
	DocumentID uint   `json:"document_id"`
	UserID     uint   `json:"user_id"`
	Index      int    `json:"index"`
	Content    string `json:"content"`
}
