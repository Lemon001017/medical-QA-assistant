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

// QAService 通过云 LLM 提供商处理问答，并在可用时集成 RAG
type QAService struct {
	client *openai.Client
	model  string
	rag    *RAGService
}

func NewQAService(apiKey, model, baseURL string, rag *RAGService) *QAService {
	if apiKey == "" {
		// 保持客户端为 nil；Ask 将返回明确的错误
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

// AskStream 通过 SSE 处理流式问答
// 当数据块到达时，将它们写入提供的写入函数
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
			// 流结束
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

// buildMessagesWithContext 构建聊天消息，包括在启用 RAG 时检索到的文档上下文
func (s *QAService) buildMessagesWithContext(ctx context.Context, userID uint, question string) ([]openai.ChatCompletionMessage, error) {
	// 默认系统提示词
	systemPrompt := `
	你是一名专业、谨慎的医学问答助手，仅用于提供医学知识层面的信息支持。

	【核心约束】
	1. 你只回答医学及健康相关的问题（包括疾病、症状、检查、治疗方式的通用医学知识）。
	2. 如果问题与医学无关，请直接说明你无法回答。
	3. 你不能进行疾病诊断、个体化治疗方案制定或处方用药建议。
	4. 不得编造医学事实、研究结论或指南内容。

	【知识来源优先级】
	1. 若提供了“医学文档片段”，必须优先基于片段内容进行回答。
	2. 只有在文档片段信息不足时，才可以补充医学常识性知识。
	3. 若文档与常识均不足以支持回答，应明确说明“不确定”或“无法给出可靠结论”。

	【回答规范】
	- 使用中文作答
	- 语言专业、清晰、简洁
	- 可使用条列或小标题
	- 避免绝对化表述（如“一定”“必须”“完全治愈”）
	- 必要时提醒用户咨询专业医生

	你的目标是：**在保证安全与准确的前提下，帮助用户理解医学问题，而不是替代医生。**
	`

	var contextText string
	if s.rag != nil && s.rag.IsEnabled() {
		chunks, err := s.rag.RetrieveRelevantChunks(ctx, userID, question, 5)
		if err != nil {
			return nil, err
		}
		if len(chunks) > 0 {
			var sb strings.Builder
			sb.WriteString(`
			以下是与用户问题相关的医学文档片段（可能来自指南、教材或医学资料）：

			请严格按照以下规则回答：
			1. 回答时必须优先基于文档片段中的信息进行推理和总结；
			2. 如果文档片段无法直接回答问题，可补充通用医学知识；
			3. 如果无法基于可靠医学信息回答，请明确说明“文档和医学常识均不足以支持明确结论”；
			4. 禁止编造文档中不存在的结论或数据；
			5. 回答应保持医学审慎性，避免诊断式或处方式表述。

			医学文档片段如下：
			`)

			for i, ch := range chunks {
				sb.WriteString(fmt.Sprintf("【片段 %d】:\n%s\n\n", i+1, ch.Content))
			}
			sb.WriteString("回答时请：\n- 优先基于上述片段中的信息进行推理；\n- 如果文档中没有足够信息，可以查找网上相关的医学知识，但是请记住不要编造；\n- 用中文回答。\n")
			contextText = sb.String()
		}
	}

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
