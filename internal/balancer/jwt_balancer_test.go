package balancer

import (
	"jetbrains-ai-proxy/internal/config"
	"sync"
	"testing"
	"time"
)

func TestNewJWTBalancer(t *testing.T) {
	tokens := []string{"token1", "token2", "token3"}
	
	// 测试轮询策略
	balancer := NewJWTBalancer(tokens, config.RoundRobin)
	if balancer == nil {
		t.Fatal("Expected balancer to be created")
	}
	
	if balancer.GetTotalTokenCount() != 3 {
		t.Errorf("Expected 3 tokens, got %d", balancer.GetTotalTokenCount())
	}
	
	if balancer.GetHealthyTokenCount() != 3 {
		t.Errorf("Expected 3 healthy tokens, got %d", balancer.GetHealthyTokenCount())
	}
}

func TestRoundRobinStrategy(t *testing.T) {
	tokens := []string{"token1", "token2", "token3"}
	balancer := NewJWTBalancer(tokens, config.RoundRobin)
	
	// 测试轮询顺序
	expectedOrder := []string{"token1", "token2", "token3", "token1", "token2", "token3"}
	
	for i, expected := range expectedOrder {
		token, err := balancer.GetToken()
		if err != nil {
			t.Fatalf("Unexpected error at iteration %d: %v", i, err)
		}
		
		if token != expected {
			t.Errorf("At iteration %d, expected %s, got %s", i, expected, token)
		}
	}
}

func TestRandomStrategy(t *testing.T) {
	tokens := []string{"token1", "token2", "token3"}
	balancer := NewJWTBalancer(tokens, config.Random)
	
	// 测试随机策略 - 多次获取token，确保都是有效的
	tokenCounts := make(map[string]int)
	iterations := 100
	
	for i := 0; i < iterations; i++ {
		token, err := balancer.GetToken()
		if err != nil {
			t.Fatalf("Unexpected error at iteration %d: %v", i, err)
		}
		
		// 检查token是否在预期列表中
		found := false
		for _, expectedToken := range tokens {
			if token == expectedToken {
				found = true
				break
			}
		}
		
		if !found {
			t.Errorf("Got unexpected token: %s", token)
		}
		
		tokenCounts[token]++
	}
	
	// 确保所有token都被使用过（随机策略下应该都有机会被选中）
	for _, token := range tokens {
		if tokenCounts[token] == 0 {
			t.Errorf("Token %s was never selected", token)
		}
	}
}

func TestMarkTokenUnhealthy(t *testing.T) {
	tokens := []string{"token1", "token2", "token3"}
	balancer := NewJWTBalancer(tokens, config.RoundRobin)
	
	// 标记一个token为不健康
	balancer.MarkTokenUnhealthy("token2")
	
	if balancer.GetHealthyTokenCount() != 2 {
		t.Errorf("Expected 2 healthy tokens, got %d", balancer.GetHealthyTokenCount())
	}
	
	// 获取token，应该只返回健康的token
	for i := 0; i < 10; i++ {
		token, err := balancer.GetToken()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if token == "token2" {
			t.Errorf("Got unhealthy token: %s", token)
		}
	}
}

func TestMarkTokenHealthy(t *testing.T) {
	tokens := []string{"token1", "token2", "token3"}
	balancer := NewJWTBalancer(tokens, config.RoundRobin)
	
	// 先标记为不健康，再标记为健康
	balancer.MarkTokenUnhealthy("token2")
	if balancer.GetHealthyTokenCount() != 2 {
		t.Errorf("Expected 2 healthy tokens after marking unhealthy, got %d", balancer.GetHealthyTokenCount())
	}
	
	balancer.MarkTokenHealthy("token2")
	if balancer.GetHealthyTokenCount() != 3 {
		t.Errorf("Expected 3 healthy tokens after marking healthy, got %d", balancer.GetHealthyTokenCount())
	}
}

func TestNoHealthyTokens(t *testing.T) {
	tokens := []string{"token1", "token2"}
	balancer := NewJWTBalancer(tokens, config.RoundRobin)
	
	// 标记所有token为不健康
	balancer.MarkTokenUnhealthy("token1")
	balancer.MarkTokenUnhealthy("token2")
	
	// 尝试获取token应该返回错误
	_, err := balancer.GetToken()
	if err == nil {
		t.Error("Expected error when no healthy tokens available")
	}
}

func TestConcurrentAccess(t *testing.T) {
	tokens := []string{"token1", "token2", "token3", "token4", "token5"}
	balancer := NewJWTBalancer(tokens, config.RoundRobin)
	
	var wg sync.WaitGroup
	numGoroutines := 10
	tokensPerGoroutine := 100
	
	// 并发获取tokens
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < tokensPerGoroutine; j++ {
				_, err := balancer.GetToken()
				if err != nil {
					t.Errorf("Unexpected error in concurrent access: %v", err)
				}
			}
		}()
	}
	
	// 并发标记tokens健康状态
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			token := tokens[index%len(tokens)]
			for j := 0; j < 10; j++ {
				if j%2 == 0 {
					balancer.MarkTokenUnhealthy(token)
				} else {
					balancer.MarkTokenHealthy(token)
				}
				time.Sleep(time.Millisecond)
			}
		}(i)
	}
	
	wg.Wait()
	
	// 确保最终状态正常
	if balancer.GetTotalTokenCount() != len(tokens) {
		t.Errorf("Expected %d total tokens, got %d", len(tokens), balancer.GetTotalTokenCount())
	}
}

func TestRefreshTokens(t *testing.T) {
	tokens := []string{"token1", "token2"}
	balancer := NewJWTBalancer(tokens, config.RoundRobin)
	
	if balancer.GetTotalTokenCount() != 2 {
		t.Errorf("Expected 2 tokens initially, got %d", balancer.GetTotalTokenCount())
	}
	
	// 刷新tokens
	newTokens := []string{"token3", "token4", "token5"}
	balancer.RefreshTokens(newTokens)
	
	if balancer.GetTotalTokenCount() != 3 {
		t.Errorf("Expected 3 tokens after refresh, got %d", balancer.GetTotalTokenCount())
	}
	
	if balancer.GetHealthyTokenCount() != 3 {
		t.Errorf("Expected 3 healthy tokens after refresh, got %d", balancer.GetHealthyTokenCount())
	}
	
	// 验证新tokens可以被获取
	for i := 0; i < 6; i++ { // 两轮完整轮询
		token, err := balancer.GetToken()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		found := false
		for _, newToken := range newTokens {
			if token == newToken {
				found = true
				break
			}
		}
		
		if !found {
			t.Errorf("Got unexpected token after refresh: %s", token)
		}
	}
}
