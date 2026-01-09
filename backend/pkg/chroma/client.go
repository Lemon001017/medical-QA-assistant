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

// Client is a client for interacting with Chroma vector database.
type Client struct {
	baseURL    string
	httpClient *http.Client
	collection string
}

// NewClient creates a new Chroma client.
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

// CollectionRequest represents a request to create/get a collection.
type CollectionRequest struct {
	Name              string                 `json:"name"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	EmbeddingFunction string                 `json:"embedding_function,omitempty"`
}

// CollectionResponse represents a Chroma collection response.
type CollectionResponse struct {
	Name     string                 `json:"name"`
	ID       string                 `json:"id"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Float32Slice is a custom type for JSON marshaling of float32 slices.
type Float32Slice []float32

// MarshalJSON converts float32 slice to float64 slice for JSON encoding.
func (f Float32Slice) MarshalJSON() ([]byte, error) {
	float64Slice := make([]float64, len(f))
	for i, v := range f {
		float64Slice[i] = float64(v)
	}
	return json.Marshal(float64Slice)
}

// AddRequest represents a request to add documents to Chroma.
type AddRequest struct {
	IDs        []string                 `json:"ids"`
	Embeddings []Float32Slice           `json:"embeddings"`
	Documents  []string                 `json:"documents"`
	Metadatas  []map[string]interface{} `json:"metadatas,omitempty"`
}

// QueryRequest represents a query request to Chroma.
type QueryRequest struct {
	QueryEmbeddings []Float32Slice         `json:"query_embeddings"`
	NResults        int                    `json:"n_results"`
	Where           map[string]interface{} `json:"where,omitempty"`
	Include         []string               `json:"include,omitempty"`
}

// QueryResponse represents a query response from Chroma.
type QueryResponse struct {
	IDs       [][]string                 `json:"ids"`
	Distances [][]float64                `json:"distances"`
	Documents [][]string                 `json:"documents"`
	Metadatas [][]map[string]interface{} `json:"metadatas"`
}

// EnsureCollection creates a collection if it doesn't exist, or gets it if it exists.
func (c *Client) EnsureCollection(ctx context.Context) error {
	// Try to get the collection first
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

	// If collection exists (200), we're done
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	// If not found (404), create it
	if resp.StatusCode == http.StatusNotFound {
		return c.createCollection(ctx)
	}

	// Other error
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
}

// createCollection creates a new collection in Chroma.
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

// Add adds documents with embeddings to Chroma.
func (c *Client) Add(ctx context.Context, ids []string, embeddings [][]float32, documents []string, metadatas []map[string]interface{}) error {
	if len(ids) != len(embeddings) || len(ids) != len(documents) {
		return fmt.Errorf("ids, embeddings, and documents must have the same length")
	}

	url := fmt.Sprintf("%s/api/v1/collections/%s/add", c.baseURL, c.collection)

	// Convert [][]float32 to []Float32Slice for JSON marshaling
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

// Query queries Chroma for similar documents.
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

// Delete deletes documents from Chroma by IDs.
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
