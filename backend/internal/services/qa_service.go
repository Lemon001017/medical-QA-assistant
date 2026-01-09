package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"medical-qa-assistant/internal/logger"

	openai "github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

// QAService handles question answering via cloud LLM providers and integrates RAG when available.
type QAService struct {
	client *openai.Client
	model  string
	rag    *RAGService
}

func NewQAService(apiKey, model, baseURL string, rag *RAGService) *QAService {
	if apiKey == "" {
		// Keep nil client; Ask will return clear error.
		return &QAService{model: model, rag: rag}
	}
	cfg := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		cfg.BaseURL = baseURL
	}
	return &QAService{
		client: openai.NewClientWithConfig(cfg),
		model:  model,
		rag:    rag,
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

	messages, err := s.buildMessagesWithContext(ctx, userID, trimmed)
	if err != nil {
		logger.L.Error("failed to build messages with context",
			zap.Error(err),
			zap.Uint("user_id", userID),
		)
		return nil, err
	}

	logger.L.Info("sending LLM request",
		zap.Uint("user_id", userID),
		zap.String("model", s.model),
	)
	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       s.model,
		Messages:    messages,
		Temperature: 0.2,
	})
	if err != nil {
		logger.L.Error("LLM request failed",
			zap.Error(err),
			zap.Uint("user_id", userID),
			zap.String("model", s.model),
		)
		return nil, fmt.Errorf("llm request failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no answer returned")
	}

	answer := resp.Choices[0].Message.Content
	logger.L.Info("LLM answer generated",
		zap.Uint("user_id", userID),
		zap.Int("answer_length", len(answer)),
	)
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

	messages, err := s.buildMessagesWithContext(ctx, userID, trimmed)
	if err != nil {
		logger.L.Error("failed to build messages with context (stream)",
			zap.Error(err),
			zap.Uint("user_id", userID),
		)
		return err
	}

	req := openai.ChatCompletionRequest{
		Model:       s.model,
		Messages:    messages,
		Temperature: 0.2,
		Stream:      true,
	}

	stream, err := s.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		logger.L.Error("failed to create LLM stream",
			zap.Error(err),
			zap.Uint("user_id", userID),
			zap.String("model", s.model),
		)
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

// buildMessagesWithContext constructs the chat messages including retrieved document context when RAG is enabled.
func (s *QAService) buildMessagesWithContext(ctx context.Context, userID uint, question string) ([]openai.ChatCompletionMessage, error) {
	// Default system prompt.
	systemPrompt := "你是一个智能医学问答助手，请根据用户的问题，给出简洁明了且安全的医学回答。"

	var contextText string
	// if s.rag != nil && s.rag.IsEnabled() {
	// 	chunks, err := s.rag.RetrieveRelevantChunks(ctx, userID, question, 5)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if len(chunks) > 0 {
	// 		var sb strings.Builder
	// 		sb.WriteString("以下是与用户问题相关的医学文档片段，请优先根据这些内容回答问题：\n\n")
	// 		for i, ch := range chunks {
	// 			sb.WriteString(fmt.Sprintf("【片段 %d】:\n%s\n\n", i+1, ch.Content))
	// 		}
	// 		sb.WriteString("回答时请：\n- 仅基于上述片段中的信息进行推理；\n- 如果文档中没有足够信息，请明确说明\"根据已提供文档无法确定\"，不要编造；\n- 用中文回答。\n")
	// 		contextText = sb.String()
	// 	}
	// }

	systemContent := systemPrompt
	if contextText != "" {
		systemContent = systemContent + "\n\n" + contextText
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemContent,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: question,
		},
	}
	return messages, nil
}
