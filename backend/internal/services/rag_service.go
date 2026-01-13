package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"medical-qa-assistant/internal/logger"
	"medical-qa-assistant/internal/models"
	"medical-qa-assistant/pkg/chroma"

	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

// RAGService 封装了文档分块、嵌入向量生成和使用 Chroma 进行检索的功能
type RAGService struct {
	embedClient  *openai.Client
	embedModel   string
	chromaClient *chroma.Client
}

// NewRAGService 创建一个新的 RAGService。如果 apiKey 为空，服务将被禁用
func NewRAGService(apiKey, baseURL, embedModel, chromaBaseURL, chromaCollection string) *RAGService {
	rag := &RAGService{
		embedModel: embedModel,
	}

	if apiKey != "" {
		cfg := openai.DefaultConfig(apiKey)
		if baseURL != "" {
			cfg.BaseURL = baseURL
		}
		rag.embedClient = openai.NewClientWithConfig(cfg)
	}

	// 初始化 Chroma 客户端
	rag.chromaClient = chroma.NewClient(chromaBaseURL, chromaCollection)

	// 确保集合存在
	if rag.IsEnabled() {
		if err := rag.chromaClient.EnsureCollection(context.Background()); err != nil {
			// 记录错误但不中断初始化
			logger.L.Warn("failed to ensure Chroma collection",
				zap.Error(err),
				zap.String("chroma_base_url", chromaBaseURL),
				zap.String("chroma_collection", chromaCollection),
			)
		}
	}

	return rag
}

// IsEnabled 返回是否可以生成嵌入向量
func (s *RAGService) IsEnabled() bool {
	return s != nil && s.embedClient != nil
}

// IndexDocument 对文档进行分块，生成嵌入向量并存储到 Chroma
func (s *RAGService) IndexDocument(ctx context.Context, doc *models.Document) error {
	if !s.IsEnabled() {
		logger.L.Info("RAG disabled, skipping document indexing",
			zap.Uint("document_id", doc.ID),
			zap.Uint("user_id", doc.UserID),
		)
		return nil
	}
	if doc == nil || doc.ID == 0 || doc.UserID == 0 {
		return errors.New("invalid document for indexing")
	}

	chunks := chunkText(doc.Content, 800) // 简单的基于字符的分块
	if len(chunks) == 0 {
		logger.L.Info("no chunks generated for document, skipping indexing",
			zap.Uint("document_id", doc.ID),
			zap.Uint("user_id", doc.UserID),
		)
		return nil
	}

	// 批量生成嵌入向量
	resp, err := s.embedClient.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Model: openai.EmbeddingModel(s.embedModel),
		Input: chunks,
	})
	if err != nil {
		logger.L.Error("failed to create embeddings for document",
			zap.Error(err),
			zap.Uint("document_id", doc.ID),
			zap.Uint("user_id", doc.UserID),
			zap.Int("chunk_count", len(chunks)),
		)
		return fmt.Errorf("failed to create embeddings: %w", err)
	}
	if len(resp.Data) != len(chunks) {
		return fmt.Errorf("embeddings count mismatch: got %d, want %d", len(resp.Data), len(chunks))
	}

	// 准备 Chroma 数据
	ids := make([]string, len(chunks))
	embeddings := make([][]float32, len(chunks))
	documents := make([]string, len(chunks))
	metadatas := make([]map[string]interface{}, len(chunks))

	for i, chunk := range chunks {
		// 生成唯一 ID：document_id-chunk_index-user_id
		ids[i] = fmt.Sprintf("%d-%d-%d", doc.ID, i, doc.UserID)
		embeddings[i] = resp.Data[i].Embedding
		documents[i] = chunk
		metadatas[i] = map[string]interface{}{
			"document_id": int(doc.ID),
			"user_id":     int(doc.UserID),
			"chunk_index": i,
			"title":       doc.Title,
		}
	}

	// 存储到 Chroma
	if err := s.chromaClient.Add(ctx, ids, embeddings, documents, metadatas); err != nil {
		logger.L.Error("failed to add document chunks to Chroma",
			zap.Error(err),
			zap.Uint("document_id", doc.ID),
			zap.Uint("user_id", doc.UserID),
			zap.Int("chunk_count", len(chunks)),
		)
		return fmt.Errorf("failed to add documents to Chroma: %w", err)
	}

	return nil
}

// RetrieveRelevantChunks 从 Chroma 返回给定问题和用户的前 k 个相关文档块
func (s *RAGService) RetrieveRelevantChunks(ctx context.Context, userID uint, question string, topK int) ([]models.Chunk, error) {
	if !s.IsEnabled() {
		logger.L.Info("RAG disabled, skipping retrieval",
			zap.Uint("user_id", userID),
		)
		return nil, nil
	}
	if userID == 0 {
		return nil, errors.New("invalid user")
	}
	trimmed := strings.TrimSpace(question)
	if trimmed == "" {
		return nil, errors.New("question is empty")
	}
	if topK <= 0 {
		topK = 5
	}

	// 将问题转换为嵌入向量
	logger.L.Info("creating question embedding",
		zap.Uint("user_id", userID),
		zap.String("model", s.embedModel),
	)
	embedResp, err := s.embedClient.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Model: openai.EmbeddingModel(s.embedModel),
		Input: []string{trimmed},
	})
	if err != nil {
		logger.L.Error("failed to create question embedding",
			zap.Error(err),
			zap.Uint("user_id", userID),
			zap.String("model", s.embedModel),
		)
		return nil, fmt.Errorf("failed to create question embedding: %w", err)
	}
	if len(embedResp.Data) == 0 {
		return nil, errors.New("no embedding returned for question")
	}

	queryVec := embedResp.Data[0].Embedding

	// 使用用户过滤器查询 Chroma
	where := map[string]interface{}{
		"user_id": int(userID),
	}

	queryResp, err := s.chromaClient.Query(ctx, queryVec, topK, where)
	if err != nil {
		return nil, fmt.Errorf("failed to query Chroma: %w", err)
	}

	if len(queryResp.Documents) == 0 || len(queryResp.Documents[0]) == 0 {
		return nil, nil
	}

	// 将 Chroma 响应转换为 Chunk 模型
	chunks := make([]models.Chunk, 0, len(queryResp.Documents[0]))
	for i, doc := range queryResp.Documents[0] {
		if i >= len(queryResp.Metadatas[0]) {
			continue
		}

		metadata := queryResp.Metadatas[0][i]
		chunk := models.Chunk{
			Content: doc,
		}

		// 提取元数据
		if docID, ok := metadata["document_id"].(float64); ok {
			chunk.DocumentID = uint(docID)
		}
		if uid, ok := metadata["user_id"].(float64); ok {
			chunk.UserID = uint(uid)
		}
		if idx, ok := metadata["chunk_index"].(float64); ok {
			chunk.Index = int(idx)
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// DeleteDocument 从 Chroma 中删除指定文档的所有向量数据
func (s *RAGService) DeleteDocument(ctx context.Context, docID, userID uint) error {
	if !s.IsEnabled() {
		logger.L.Info("RAG disabled, skipping document deletion from Chroma",
			zap.Uint("document_id", docID),
			zap.Uint("user_id", userID),
		)
		return nil
	}
	if docID == 0 || userID == 0 {
		return errors.New("invalid document or user for deletion")
	}

	where := map[string]interface{}{
		"$and": []map[string]interface{}{
			{"document_id": int(docID)},
			{"user_id": int(userID)},
		},
	}

	ids, err := s.chromaClient.GetIDsByMetadata(ctx, where)
	if err != nil {
		logger.L.Error("failed to get document chunk ids for deletion",
			zap.Error(err),
			zap.Uint("document_id", docID),
			zap.Uint("user_id", userID),
		)
		return fmt.Errorf("failed to get document chunk ids: %w", err)
	}

	if len(ids) == 0 {
		logger.L.Info("no chunks found for document deletion",
			zap.Uint("document_id", docID),
			zap.Uint("user_id", userID),
		)
		return nil
	}

	if err := s.chromaClient.Delete(ctx, ids); err != nil {
		logger.L.Error("failed to delete document chunks from Chroma",
			zap.Error(err),
			zap.Uint("document_id", docID),
			zap.Uint("user_id", userID),
			zap.Int("chunk_count", len(ids)),
		)
		return fmt.Errorf("failed to delete chunks from Chroma: %w", err)
	}

	logger.L.Info("document chunks deleted from Chroma",
		zap.Uint("document_id", docID),
		zap.Uint("user_id", userID),
		zap.Int("chunk_count", len(ids)),
	)

	return nil
}


// chunkText 是一个简单的辅助函数，将文本分割成大约 maxLen 字符的块
func chunkText(text string, maxLen int) []string {
	text = strings.TrimSpace(text)
	if text == "" || maxLen <= 0 {
		return nil
	}

	var chunks []string
	runes := []rune(text)
	for start := 0; start < len(runes); start += maxLen {
		end := start + maxLen
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[start:end]))
	}
	return chunks
}
