package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"jetbrains-ai-proxy/internal/apiserver"
	"jetbrains-ai-proxy/internal/config"
	"log"
	"net/http"
)

func main() {
	// 定义命令行参数
	port := flag.Int("p", 8080, "服务器监听端口")
	host := flag.String("h", "0.0.0.0", "服务器监听地址")
	jtwToken := flag.String("c", "", "JWT Token值 (JWT_TOKEN)")
	bearerToken := flag.String("k", "", "Bearer Token值 (BEARER_TOKEN)")

	flag.Usage = func() {
		fmt.Printf("用法: %s [选项]\n\n", flag.CommandLine.Name())
		fmt.Println("选项:")
		flag.PrintDefaults()
		fmt.Println("\n示例: ./jetbrains-ai-proxy -p 8080 -c \"cookie\" -k \"token\" -i=false")
	}

	flag.Parse()

	cfg := config.LoadConfig()
	if *jtwToken != "" {
		cfg.JetbrainsToken = *jtwToken
	}
	if *bearerToken != "" {
		cfg.BearerToken = *bearerToken
	}

	// 检查必要的配置
	if cfg.JetbrainsToken == "" {
		log.Fatal("Jetbrains Ai Token is required. Please set it via -c flag or JWT_TOKEN environment variable")
	}
	if cfg.BearerToken == "" {
		log.Fatal("Bearer Token is required. Please set it via -k flag or BEARER_TOKEN environment variable")
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	apiserver.RegisterRoutes(e)

	addr := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("Server starting on %s", addr)
	if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("start server error: %v", err)
	}
}
