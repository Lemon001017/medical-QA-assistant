package models

// Chunk 表示从 Chroma 检索到的文档块
type Chunk struct {
	DocumentID uint   `json:"document_id"`
	UserID     uint   `json:"user_id"`
	Index      int    `json:"index"`
	Content    string `json:"content"`
}
