package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"jetbrains-ai-proxy/internal/apiserver"
	"jetbrains-ai-proxy/internal/config"
	"jetbrains-ai-proxy/internal/jetbrains"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// 定义命令行参数
	port := flag.Int("p", 8080, "服务器监听端口")
	host := flag.String("h", "0.0.0.0", "服务器监听地址")
	jwtTokens := flag.String("c", "", "JWT Tokens值，多个token用逗号分隔 (JWT_TOKENS)")
	bearerToken := flag.String("k", "", "Bearer Token值 (BEARER_TOKEN)")
	loadBalanceStrategy := flag.String("s", "round_robin", "负载均衡策略: round_robin 或 random")

	flag.Usage = func() {
		fmt.Printf("用法: %s [选项]\n\n", flag.CommandLine.Name())
		fmt.Println("选项:")
		flag.PrintDefaults()
		fmt.Println("\n示例:")
		fmt.Println("  单个JWT: ./jetbrains-ai-proxy -p 8080 -c \"jwt_token\" -k \"bearer_token\"")
		fmt.Println("  多个JWT: ./jetbrains-ai-proxy -p 8080 -c \"jwt1,jwt2,jwt3\" -k \"bearer_token\" -s \"random\"")
		fmt.Println("\n负载均衡策略:")
		fmt.Println("  round_robin: 轮询策略（默认）")
		fmt.Println("  random: 随机策略")
	}

	flag.Parse()

	cfg := config.LoadConfig()

	// 设置JWT tokens
	if *jwtTokens != "" {
		cfg.SetJetbrainsTokens(*jwtTokens)
	}

	// 设置Bearer token
	if *bearerToken != "" {
		cfg.BearerToken = *bearerToken
	}

	// 检查必要的配置
	if !cfg.HasJetbrainsTokens() {
		log.Fatal("JWT Tokens are required. Please set them via -c flag or JWT_TOKENS/JWT_TOKEN environment variable")
	}
	if cfg.BearerToken == "" {
		log.Fatal("Bearer Token is required. Please set it via -k flag or BEARER_TOKEN environment variable")
	}

	// 初始化JWT负载均衡器
	tokens := cfg.GetJetbrainsTokens()
	if err := jetbrains.InitializeBalancer(tokens, *loadBalanceStrategy); err != nil {
		log.Fatalf("Failed to initialize JWT balancer: %v", err)
	}

	// 设置优雅关闭
	setupGracefulShutdown()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 添加健康检查端点
	e.GET("/health", func(c echo.Context) error {
		healthy, total := jetbrains.GetBalancerStats()
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status":        "ok",
			"healthy_tokens": healthy,
			"total_tokens":   total,
			"strategy":       *loadBalanceStrategy,
		})
	})

	apiserver.RegisterRoutes(e)

	addr := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("Server starting on %s", addr)
	log.Printf("JWT tokens configured: %d", len(tokens))
	log.Printf("Load balance strategy: %s", *loadBalanceStrategy)

	if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("start server error: %v", err)
	}
}

// setupGracefulShutdown 设置优雅关闭
func setupGracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Shutting down gracefully...")
		jetbrains.StopBalancer()
		os.Exit(0)
	}()
}
