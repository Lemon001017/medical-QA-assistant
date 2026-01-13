package config

import (
	"os"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
	Port       string

	// LLM 配置
	LLMProvider      string
	OpenAIKey        string
	OpenAIModel      string
	OpenAIBaseURL    string
	DeepSeekKey      string
	DeepSeekModel    string
	DeepSeekBaseURL  string

	// Chroma 向量数据库配置
	ChromaBaseURL    string
	ChromaCollection string

	// embedding 配置
	AliyunEmbeddingModel string
	AliyunEmbeddingKey string
	AliyunEmbeddingBaseURL string
}

func Load() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "medical_qa"),
		JWTSecret:  getEnv("JWT_SECRET", "dev-secret-change-me"),
		Port:       getEnv("PORT", "8081"),

		LLMProvider:      getEnv("LLM_PROVIDER", "openai"), // openai | deepseek
		OpenAIKey:        getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:      getEnv("OPENAI_MODEL", "gpt-3.5-turbo"),
		OpenAIBaseURL:    getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		DeepSeekKey:      getEnv("DEEPSEEK_API_KEY", ""),
		DeepSeekModel:    getEnv("DEEPSEEK_MODEL", "deepseek-chat"),
		DeepSeekBaseURL:  getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com/v1"),

		ChromaBaseURL:    getEnv("CHROMA_BASE_URL", "http://localhost:8000"),
		ChromaCollection: getEnv("CHROMA_COLLECTION", "medical_documents"),

		AliyunEmbeddingModel: getEnv("ALIYUN_EMBEDDING_MODEL", "text-embedding-v4"),
		AliyunEmbeddingKey: getEnv("ALIYUN_EMBEDDING_KEY", ""),
		AliyunEmbeddingBaseURL: getEnv("ALIYUN_EMBEDING_BASEURL", "https://dashscope.aliyuncs.com/compatible-mode/v1"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
