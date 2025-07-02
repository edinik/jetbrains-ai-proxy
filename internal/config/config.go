package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var JetbrainsAiConfig *Config

// LoadBalanceStrategy 负载均衡策略
type LoadBalanceStrategy string

const (
	RoundRobin LoadBalanceStrategy = "round_robin"
	Random     LoadBalanceStrategy = "random"
)

type Config struct {
	JetbrainsTokens     []string            // 支持多个JWT tokens
	BearerToken         string              // Bearer token保持不变
	LoadBalanceStrategy LoadBalanceStrategy // 负载均衡策略
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	// 读取多个JWT tokens，支持逗号分隔
	jwtTokensStr := os.Getenv("JWT_TOKENS")
	if jwtTokensStr == "" {
		// 兼容旧的单token配置
		jwtTokensStr = os.Getenv("JWT_TOKEN")
	}

	var jwtTokens []string
	if jwtTokensStr != "" {
		// 按逗号分割并去除空白
		tokens := strings.Split(jwtTokensStr, ",")
		for _, token := range tokens {
			token = strings.TrimSpace(token)
			if token != "" {
				jwtTokens = append(jwtTokens, token)
			}
		}
	}

	// 读取负载均衡策略，默认为轮询
	strategy := LoadBalanceStrategy(os.Getenv("LOAD_BALANCE_STRATEGY"))
	if strategy != RoundRobin && strategy != Random {
		strategy = RoundRobin // 默认策略
	}

	JetbrainsAiConfig = &Config{
		JetbrainsTokens:     jwtTokens,
		BearerToken:         os.Getenv("BEARER_TOKEN"),
		LoadBalanceStrategy: strategy,
	}
	return JetbrainsAiConfig
}

// SetJetbrainsTokens 设置JWT tokens（用于命令行参数）
func (c *Config) SetJetbrainsTokens(tokensStr string) {
	if tokensStr == "" {
		return
	}

	var tokens []string
	tokenList := strings.Split(tokensStr, ",")
	for _, token := range tokenList {
		token = strings.TrimSpace(token)
		if token != "" {
			tokens = append(tokens, token)
		}
	}
	c.JetbrainsTokens = tokens
}

// GetJetbrainsTokens 获取所有可用的JWT tokens
func (c *Config) GetJetbrainsTokens() []string {
	return c.JetbrainsTokens
}

// HasJetbrainsTokens 检查是否有可用的JWT tokens
func (c *Config) HasJetbrainsTokens() bool {
	return len(c.JetbrainsTokens) > 0
}
