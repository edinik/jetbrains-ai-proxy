package balancer

import (
	"fmt"
	"jetbrains-ai-proxy/internal/config"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// JWTBalancer JWT负载均衡器接口
type JWTBalancer interface {
	GetToken() (string, error)
	MarkTokenUnhealthy(token string)
	MarkTokenHealthy(token string)
	GetHealthyTokenCount() int
	GetTotalTokenCount() int
	RefreshTokens(tokens []string)
}

// TokenStatus token状态
type TokenStatus struct {
	Token     string
	Healthy   bool
	LastUsed  time.Time
	ErrorCount int64
}

// BaseBalancer 基础负载均衡器
type BaseBalancer struct {
	tokens   map[string]*TokenStatus
	strategy config.LoadBalanceStrategy
	mutex    sync.RWMutex
	counter  int64 // 用于轮询计数
	rand     *rand.Rand
}

// NewJWTBalancer 创建JWT负载均衡器
func NewJWTBalancer(tokens []string, strategy config.LoadBalanceStrategy) JWTBalancer {
	balancer := &BaseBalancer{
		tokens:   make(map[string]*TokenStatus),
		strategy: strategy,
		rand:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	
	// 初始化tokens
	for _, token := range tokens {
		balancer.tokens[token] = &TokenStatus{
			Token:     token,
			Healthy:   true,
			LastUsed:  time.Now(),
			ErrorCount: 0,
		}
	}
	
	return balancer
}

// GetToken 获取一个可用的token
func (b *BaseBalancer) GetToken() (string, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	
	// 获取所有健康的tokens
	healthyTokens := make([]*TokenStatus, 0)
	for _, status := range b.tokens {
		if status.Healthy {
			healthyTokens = append(healthyTokens, status)
		}
	}
	
	if len(healthyTokens) == 0 {
		return "", fmt.Errorf("no healthy JWT tokens available")
	}
	
	var selectedToken *TokenStatus
	
	switch b.strategy {
	case config.RoundRobin:
		// 轮询策略
		index := atomic.AddInt64(&b.counter, 1) % int64(len(healthyTokens))
		selectedToken = healthyTokens[index]
	case config.Random:
		// 随机策略
		index := b.rand.Intn(len(healthyTokens))
		selectedToken = healthyTokens[index]
	default:
		// 默认使用轮询
		index := atomic.AddInt64(&b.counter, 1) % int64(len(healthyTokens))
		selectedToken = healthyTokens[index]
	}
	
	// 更新最后使用时间
	selectedToken.LastUsed = time.Now()
	
	return selectedToken.Token, nil
}

// MarkTokenUnhealthy 标记token为不健康
func (b *BaseBalancer) MarkTokenUnhealthy(token string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	
	if status, exists := b.tokens[token]; exists {
		status.Healthy = false
		atomic.AddInt64(&status.ErrorCount, 1)
		fmt.Printf("JWT token marked as unhealthy: %s (errors: %d)\n", 
			token[:min(len(token), 10)]+"...", status.ErrorCount)
	}
}

// MarkTokenHealthy 标记token为健康
func (b *BaseBalancer) MarkTokenHealthy(token string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	
	if status, exists := b.tokens[token]; exists {
		status.Healthy = true
		atomic.StoreInt64(&status.ErrorCount, 0)
		fmt.Printf("JWT token marked as healthy: %s\n", 
			token[:min(len(token), 10)]+"...")
	}
}

// GetHealthyTokenCount 获取健康token数量
func (b *BaseBalancer) GetHealthyTokenCount() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	
	count := 0
	for _, status := range b.tokens {
		if status.Healthy {
			count++
		}
	}
	return count
}

// GetTotalTokenCount 获取总token数量
func (b *BaseBalancer) GetTotalTokenCount() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	
	return len(b.tokens)
}

// RefreshTokens 刷新token列表
func (b *BaseBalancer) RefreshTokens(tokens []string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	
	// 清空现有tokens
	b.tokens = make(map[string]*TokenStatus)
	
	// 添加新tokens
	for _, token := range tokens {
		b.tokens[token] = &TokenStatus{
			Token:     token,
			Healthy:   true,
			LastUsed:  time.Now(),
			ErrorCount: 0,
		}
	}
	
	fmt.Printf("JWT tokens refreshed, total: %d\n", len(tokens))
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
