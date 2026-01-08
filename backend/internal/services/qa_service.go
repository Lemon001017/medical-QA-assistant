package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// QAService handles question answering via cloud LLM providers.
type QAService struct {
	client *openai.Client
	model  string
}

func NewQAService(apiKey, model, baseURL string) *QAService {
	if apiKey == "" {
		// Keep nil client; Ask will return clear error.
		return &QAService{model: model}
	}
	cfg := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	return &QAService{
		client: openai.NewClientWithConfig(cfg),
		model:  model,
	}
}

type AskRequest struct {
	Question string `json:"question" binding:"required,min=1"`
}

type AskResponse struct {
	Answer string `json:"answer"`
}

func (s *QAService) Ask(ctx context.Context, userID uint, question string) (*AskResponse, error) {
	if userID == 0 {
		return nil, errors.New("invalid user")
	}
	if s.client == nil {
		return nil, errors.New("llm client not configured (missing LLM API key)")
	}
	trimmed := strings.TrimSpace(question)
	if trimmed == "" {
		return nil, errors.New("question is empty")
	}

	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: s.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "你是一个智能医学问答助手，请根据用户的问题，给出简洁明了的回答。",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: trimmed,
			},
		},
		Temperature: 0.2,
	})
	if err != nil {
		return nil, fmt.Errorf("llm request failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no answer returned")
	}

	answer := resp.Choices[0].Message.Content
	return &AskResponse{Answer: answer}, nil
}

// AskStream handles streaming question answering via SSE.
// It writes chunks to the provided writer function as they arrive.
func (s *QAService) AskStream(ctx context.Context, userID uint, question string, writeChunk func(string) error) error {
	if userID == 0 {
		return errors.New("invalid user")
	}
	if s.client == nil {
		return errors.New("llm client not configured (missing LLM API key)")
	}
	trimmed := strings.TrimSpace(question)
	if trimmed == "" {
		return errors.New("question is empty")
	}

	req := openai.ChatCompletionRequest{
		Model: s.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "你是一个智能医学问答助手，请根据用户的问题，给出简洁明了的回答。",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: trimmed,
			},
		},
		Temperature: 0.2,
		Stream:      true,
	}

	stream, err := s.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			// Stream ended
			return nil
		}
		if err != nil {
			return fmt.Errorf("stream error: %w", err)
		}

		if len(response.Choices) > 0 {
			delta := response.Choices[0].Delta.Content
			if delta != "" {
				if err := writeChunk(delta); err != nil {
					return fmt.Errorf("failed to write chunk: %w", err)
				}
			}
		}
	}
}
