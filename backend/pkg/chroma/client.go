package chroma

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client 是与 Chroma 向量数据库交互的客户端
type Client struct {
	baseURL    string
	httpClient *http.Client
	collection string
}

// NewClient 创建一个新的 Chroma 客户端
func NewClient(baseURL, collection string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}
	if collection == "" {
		collection = "medical_documents"
	}
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		collection: collection,
	}
}

// CollectionRequest 表示创建/获取集合的请求
type CollectionRequest struct {
	Name              string                 `json:"name"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	EmbeddingFunction string                 `json:"embedding_function,omitempty"`
}

// CollectionResponse 表示 Chroma 集合响应
type CollectionResponse struct {
	Name     string                 `json:"name"`
	ID       string                 `json:"id"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Float32Slice 是用于 JSON 序列化 float32 切片的自定义类型
type Float32Slice []float32

// MarshalJSON 将 float32 切片转换为 float64 切片用于 JSON 编码
func (f Float32Slice) MarshalJSON() ([]byte, error) {
	float64Slice := make([]float64, len(f))
	for i, v := range f {
		float64Slice[i] = float64(v)
	}
	return json.Marshal(float64Slice)
}

// AddRequest 表示向 Chroma 添加文档的请求
type AddRequest struct {
	IDs        []string                 `json:"ids"`
	Embeddings []Float32Slice           `json:"embeddings"`
	Documents  []string                 `json:"documents"`
	Metadatas  []map[string]interface{} `json:"metadatas,omitempty"`
}

// QueryRequest 表示查询 Chroma 的请求
type QueryRequest struct {
	QueryEmbeddings []Float32Slice         `json:"query_embeddings"`
	NResults        int                    `json:"n_results"`
	Where           map[string]interface{} `json:"where,omitempty"`
	Include         []string               `json:"include,omitempty"`
}

// QueryResponse 表示来自 Chroma 的查询响应
type QueryResponse struct {
	IDs       [][]string                 `json:"ids"`
	Distances [][]float64                `json:"distances"`
	Documents [][]string                 `json:"documents"`
	Metadatas [][]map[string]interface{} `json:"metadatas"`
}

// EnsureCollection 如果集合不存在则创建它，如果存在则获取它
func (c *Client) EnsureCollection(ctx context.Context) error {
	// 首先尝试获取集合
	url := fmt.Sprintf("%s/api/v1/collections/%s", c.baseURL, c.collection)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}
	defer resp.Body.Close()

	// 如果集合存在（200），完成
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	// 如果未找到（404），创建它
	if resp.StatusCode == http.StatusNotFound {
		return c.createCollection(ctx)
	}

	// 其他错误
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
}

// createCollection 在 Chroma 中创建一个新集合
func (c *Client) createCollection(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v1/collections", c.baseURL)
	reqBody := CollectionRequest{
		Name: c.collection,
		Metadata: map[string]interface{}{
			"description": "Medical documents collection",
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create collection: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Add 将带有嵌入向量的文档添加到 Chroma
func (c *Client) Add(ctx context.Context, ids []string, embeddings [][]float32, documents []string, metadatas []map[string]interface{}) error {
	if len(ids) != len(embeddings) || len(ids) != len(documents) {
		return fmt.Errorf("ids, embeddings, and documents must have the same length")
	}

	url := fmt.Sprintf("%s/api/v1/collections/%s/add", c.baseURL, c.collection)

	// 将 [][]float32 转换为 []Float32Slice 用于 JSON 序列化
	embeddingSlices := make([]Float32Slice, len(embeddings))
	for i, emb := range embeddings {
		embeddingSlices[i] = Float32Slice(emb)
	}

	reqBody := AddRequest{
		IDs:        ids,
		Embeddings: embeddingSlices,
		Documents:  documents,
		Metadatas:  metadatas,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add documents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add documents: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Query 查询 Chroma 以查找相似文档
func (c *Client) Query(ctx context.Context, queryEmbedding []float32, nResults int, where map[string]interface{}) (*QueryResponse, error) {
	url := fmt.Sprintf("%s/api/v1/collections/%s/query", c.baseURL, c.collection)
	reqBody := QueryRequest{
		QueryEmbeddings: []Float32Slice{Float32Slice(queryEmbedding)},
		NResults:        nResults,
		Where:           where,
		Include:         []string{"documents", "metadatas", "distances"},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to query: status %d, body: %s", resp.StatusCode, string(body))
	}

	var queryResp QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &queryResp, nil
}

// Delete 通过 ID 从 Chroma 删除文档
func (c *Client) Delete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	url := fmt.Sprintf("%s/api/v1/collections/%s/delete", c.baseURL, c.collection)
	reqBody := map[string]interface{}{
		"ids": ids,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete documents: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
