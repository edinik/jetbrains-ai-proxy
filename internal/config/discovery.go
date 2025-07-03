package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ConfigDiscovery 配置发现器
type ConfigDiscovery struct {
	searchPaths []string
	manager     *Manager
}

// NewConfigDiscovery 创建配置发现器
func NewConfigDiscovery(manager *Manager) *ConfigDiscovery {
	return &ConfigDiscovery{
		manager: manager,
		searchPaths: []string{
			// 当前目录
			"config.json",
			"jetbrains-ai-proxy.json",
			".jetbrains-ai-proxy.json",
			
			// config 目录
			"config/config.json",
			"config/jetbrains-ai-proxy.json",
			"configs/config.json",
			"configs/jetbrains-ai-proxy.json",
			
			// 隐藏配置目录
			".config/config.json",
			".config/jetbrains-ai-proxy.json",
			
			// 用户主目录
			os.ExpandEnv("$HOME/.config/jetbrains-ai-proxy/config.json"),
			os.ExpandEnv("$HOME/.jetbrains-ai-proxy/config.json"),
			os.ExpandEnv("$HOME/.jetbrains-ai-proxy.json"),
			
			// 系统配置目录 (Linux/macOS)
			"/etc/jetbrains-ai-proxy/config.json",
			"/usr/local/etc/jetbrains-ai-proxy/config.json",
		},
	}
}

// DiscoverAndLoad 发现并加载配置文件
func (cd *ConfigDiscovery) DiscoverAndLoad() error {
	log.Println("Starting configuration discovery...")
	
	// 1. 尝试从环境变量指定的配置文件加载
	if configPath := os.Getenv("CONFIG_FILE"); configPath != "" {
		if err := cd.loadConfigFile(configPath); err != nil {
			log.Printf("Failed to load config from CONFIG_FILE=%s: %v", configPath, err)
		} else {
			log.Printf("Successfully loaded config from CONFIG_FILE: %s", configPath)
			return nil
		}
	}
	
	// 2. 搜索预定义路径
	for _, path := range cd.searchPaths {
		if cd.fileExists(path) {
			if err := cd.loadConfigFile(path); err != nil {
				log.Printf("Failed to load config from %s: %v", path, err)
				continue
			}
			log.Printf("Successfully loaded config from: %s", path)
			return nil
		}
	}
	
	// 3. 尝试从当前目录的 .env 文件加载
	if cd.fileExists(".env") {
		log.Println("Found .env file, loading environment variables...")
		return nil // .env 文件会在 LoadConfig 中自动加载
	}
	
	// 4. 如果没有找到配置文件，生成示例配置
	log.Println("No configuration file found, generating example config...")
	return cd.generateDefaultConfig()
}

// loadConfigFile 加载指定的配置文件
func (cd *ConfigDiscovery) loadConfigFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}
	
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}
	
	// 验证配置
	if err := cd.validateLoadedConfig(&config); err != nil {
		return fmt.Errorf("invalid config: %v", err)
	}
	
	// 合并到管理器
	cd.manager.mutex.Lock()
	cd.manager.mergeConfig(&config)
	cd.manager.configPath = path
	cd.manager.mutex.Unlock()
	
	return nil
}

// validateLoadedConfig 验证加载的配置
func (cd *ConfigDiscovery) validateLoadedConfig(config *Config) error {
	if len(config.JetbrainsTokens) == 0 {
		return fmt.Errorf("no JWT tokens found in config")
	}
	
	// 验证每个JWT token
	for i, tokenConfig := range config.JetbrainsTokens {
		if tokenConfig.Token == "" {
			return fmt.Errorf("JWT token %d is empty", i+1)
		}
		if len(tokenConfig.Token) < 10 {
			return fmt.Errorf("JWT token %d appears to be invalid (too short)", i+1)
		}
	}
	
	if config.BearerToken == "" {
		log.Println("Warning: No bearer token found in config file")
	}
	
	return nil
}

// generateDefaultConfig 生成默认配置
func (cd *ConfigDiscovery) generateDefaultConfig() error {
	configDir := "config"
	configPath := filepath.Join(configDir, "config.json")
	
	// 创建配置目录
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}
	
	// 生成示例配置
	if err := cd.manager.GenerateExampleConfig(configPath); err != nil {
		return fmt.Errorf("failed to generate example config: %v", err)
	}
	
	// 同时生成 .env 示例文件
	envPath := ".env.example"
	if err := cd.generateEnvExample(envPath); err != nil {
		log.Printf("Warning: Failed to generate .env example: %v", err)
	}
	
	log.Printf("Generated example configuration files:")
	log.Printf("  - %s (JSON format)", configPath)
	log.Printf("  - %s (Environment variables)", envPath)
	log.Printf("Please edit these files with your actual JWT tokens and restart the application.")
	
	return fmt.Errorf("no valid configuration found, example files generated")
}

// generateEnvExample 生成 .env 示例文件
func (cd *ConfigDiscovery) generateEnvExample(path string) error {
	envContent := `# JetBrains AI Proxy Configuration
# Copy this file to .env and fill in your actual values

# Multiple JWT tokens (comma-separated)
JWT_TOKENS=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...,eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...

# Or single JWT token (for backward compatibility)
# JWT_TOKEN=eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...

# Bearer token for API authentication
BEARER_TOKEN=your_bearer_token_here

# Load balancing strategy: round_robin or random
LOAD_BALANCE_STRATEGY=round_robin

# Server configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Alternative: specify config file path
# CONFIG_FILE=config/config.json
`
	
	return ioutil.WriteFile(path, []byte(envContent), 0644)
}

// fileExists 检查文件是否存在
func (cd *ConfigDiscovery) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// WatchConfig 监控配置文件变化（简单实现）
func (cd *ConfigDiscovery) WatchConfig() {
	if cd.manager.configPath == "" {
		return
	}
	
	go func() {
		var lastModTime time.Time
		
		// 获取初始修改时间
		if stat, err := os.Stat(cd.manager.configPath); err == nil {
			lastModTime = stat.ModTime()
		}
		
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		
		for range ticker.C {
			stat, err := os.Stat(cd.manager.configPath)
			if err != nil {
				continue
			}
			
			if stat.ModTime().After(lastModTime) {
				log.Printf("Config file changed, reloading: %s", cd.manager.configPath)
				if err := cd.loadConfigFile(cd.manager.configPath); err != nil {
					log.Printf("Failed to reload config: %v", err)
				} else {
					log.Println("Config reloaded successfully")
				}
				lastModTime = stat.ModTime()
			}
		}
	}()
}

// ListAvailableConfigs 列出可用的配置文件
func (cd *ConfigDiscovery) ListAvailableConfigs() []string {
	var available []string
	
	for _, path := range cd.searchPaths {
		if cd.fileExists(path) {
			available = append(available, path)
		}
	}
	
	return available
}

// ValidateConfigFile 验证配置文件格式
func (cd *ConfigDiscovery) ValidateConfigFile(path string) error {
	if !cd.fileExists(path) {
		return fmt.Errorf("config file does not exist: %s", path)
	}
	
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}
	
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("invalid JSON format: %v", err)
	}
	
	return cd.validateLoadedConfig(&config)
}

// GetConfigSummary 获取配置摘要信息
func (cd *ConfigDiscovery) GetConfigSummary() map[string]interface{} {
	config := cd.manager.GetConfig()
	
	// 隐藏敏感信息
	tokenSummary := make([]map[string]interface{}, len(config.JetbrainsTokens))
	for i, token := range config.JetbrainsTokens {
		tokenSummary[i] = map[string]interface{}{
			"name":        token.Name,
			"description": token.Description,
			"priority":    token.Priority,
			"token_preview": token.Token[:min(len(token.Token), 20)] + "...",
		}
	}
	
	return map[string]interface{}{
		"jwt_tokens_count":      len(config.JetbrainsTokens),
		"jwt_tokens":           tokenSummary,
		"bearer_token_set":     config.BearerToken != "",
		"load_balance_strategy": config.LoadBalanceStrategy,
		"health_check_interval": config.HealthCheckInterval.String(),
		"server_host":          config.ServerHost,
		"server_port":          config.ServerPort,
		"config_file":          cd.manager.configPath,
	}
}
