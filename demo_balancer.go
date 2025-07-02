package main

import (
	"fmt"
	"jetbrains-ai-proxy/internal/balancer"
	"jetbrains-ai-proxy/internal/config"
	"log"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== JWT负载均衡器演示 ===")
	
	// 演示轮询策略
	fmt.Println("\n1. 轮询策略演示:")
	demoRoundRobin()
	
	// 演示随机策略
	fmt.Println("\n2. 随机策略演示:")
	demoRandom()
	
	// 演示健康检查
	fmt.Println("\n3. 健康检查演示:")
	demoHealthCheck()
	
	// 演示并发访问
	fmt.Println("\n4. 并发访问演示:")
	demoConcurrent()
	
	fmt.Println("\n=== 演示完成 ===")
}

func demoRoundRobin() {
	tokens := []string{"JWT_TOKEN_1", "JWT_TOKEN_2", "JWT_TOKEN_3"}
	balancer := balancer.NewJWTBalancer(tokens, config.RoundRobin)
	
	fmt.Printf("配置了 %d 个JWT tokens，使用轮询策略\n", len(tokens))
	fmt.Println("获取token顺序:")
	
	for i := 0; i < 9; i++ {
		token, err := balancer.GetToken()
		if err != nil {
			log.Printf("错误: %v", err)
			continue
		}
		fmt.Printf("  第%d次: %s\n", i+1, token)
	}
}

func demoRandom() {
	tokens := []string{"JWT_TOKEN_A", "JWT_TOKEN_B", "JWT_TOKEN_C"}
	balancer := balancer.NewJWTBalancer(tokens, config.Random)
	
	fmt.Printf("配置了 %d 个JWT tokens，使用随机策略\n", len(tokens))
	fmt.Println("获取token顺序:")
	
	tokenCounts := make(map[string]int)
	for i := 0; i < 12; i++ {
		token, err := balancer.GetToken()
		if err != nil {
			log.Printf("错误: %v", err)
			continue
		}
		tokenCounts[token]++
		fmt.Printf("  第%d次: %s\n", i+1, token)
	}
	
	fmt.Println("使用统计:")
	for token, count := range tokenCounts {
		fmt.Printf("  %s: %d次\n", token, count)
	}
}

func demoHealthCheck() {
	tokens := []string{"JWT_HEALTHY_1", "JWT_HEALTHY_2", "JWT_UNHEALTHY_3"}
	balancer := balancer.NewJWTBalancer(tokens, config.RoundRobin)
	
	fmt.Printf("初始状态: %d/%d tokens健康\n", 
		balancer.GetHealthyTokenCount(), balancer.GetTotalTokenCount())
	
	// 标记一个token为不健康
	fmt.Println("标记 JWT_UNHEALTHY_3 为不健康...")
	balancer.MarkTokenUnhealthy("JWT_UNHEALTHY_3")
	
	fmt.Printf("更新后状态: %d/%d tokens健康\n", 
		balancer.GetHealthyTokenCount(), balancer.GetTotalTokenCount())
	
	fmt.Println("获取token（应该只返回健康的tokens）:")
	for i := 0; i < 6; i++ {
		token, err := balancer.GetToken()
		if err != nil {
			log.Printf("错误: %v", err)
			continue
		}
		fmt.Printf("  第%d次: %s\n", i+1, token)
	}
	
	// 恢复token健康状态
	fmt.Println("恢复 JWT_UNHEALTHY_3 为健康...")
	balancer.MarkTokenHealthy("JWT_UNHEALTHY_3")
	
	fmt.Printf("恢复后状态: %d/%d tokens健康\n", 
		balancer.GetHealthyTokenCount(), balancer.GetTotalTokenCount())
}

func demoConcurrent() {
	tokens := []string{"JWT_CONCURRENT_1", "JWT_CONCURRENT_2", "JWT_CONCURRENT_3", "JWT_CONCURRENT_4"}
	balancer := balancer.NewJWTBalancer(tokens, config.RoundRobin)
	
	fmt.Printf("使用 %d 个JWT tokens进行并发测试\n", len(tokens))
	
	var wg sync.WaitGroup
	numGoroutines := 5
	requestsPerGoroutine := 10
	
	tokenCounts := make(map[string]int)
	var mutex sync.Mutex
	
	startTime := time.Now()
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			
			for j := 0; j < requestsPerGoroutine; j++ {
				token, err := balancer.GetToken()
				if err != nil {
					log.Printf("Goroutine %d 错误: %v", goroutineID, err)
					continue
				}
				
				mutex.Lock()
				tokenCounts[token]++
				mutex.Unlock()
				
				// 模拟一些处理时间
				time.Sleep(time.Millisecond * 10)
			}
		}(i)
	}
	
	wg.Wait()
	duration := time.Since(startTime)
	
	fmt.Printf("并发测试完成，耗时: %v\n", duration)
	fmt.Printf("总请求数: %d\n", numGoroutines*requestsPerGoroutine)
	fmt.Println("Token使用分布:")
	
	for token, count := range tokenCounts {
		percentage := float64(count) / float64(numGoroutines*requestsPerGoroutine) * 100
		fmt.Printf("  %s: %d次 (%.1f%%)\n", token, count, percentage)
	}
}

// 演示配置加载
func demoConfigLoading() {
	fmt.Println("\n5. 配置加载演示:")
	
	// 模拟环境变量配置
	fmt.Println("模拟配置加载...")
	
	cfg := &config.Config{}
	
	// 设置多个JWT tokens
	cfg.SetJetbrainsTokens("token1,token2,token3")
	cfg.BearerToken = "bearer_token_example"
	cfg.LoadBalanceStrategy = config.RoundRobin
	
	fmt.Printf("JWT Tokens数量: %d\n", len(cfg.GetJetbrainsTokens()))
	fmt.Printf("Bearer Token: %s\n", cfg.BearerToken)
	fmt.Printf("负载均衡策略: %s\n", cfg.LoadBalanceStrategy)
	fmt.Printf("是否有JWT Tokens: %v\n", cfg.HasJetbrainsTokens())
	
	fmt.Println("JWT Tokens列表:")
	for i, token := range cfg.GetJetbrainsTokens() {
		fmt.Printf("  %d: %s\n", i+1, token)
	}
}
