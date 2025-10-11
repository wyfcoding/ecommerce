// cmd/recommendation/main.go
package main

import (
	"log"
	"os"

	"ecommerce/internal/recommendation"                               // 假设内部逻辑在此
	recommendationhandler "ecommerce/internal/recommendation/handler" // Added this line
	"ecommerce/pkg/config"                                            // 假设配置包在此
	"ecommerce/pkg/logging"                                           // 假设日志包在此
)

// Config 结构体定义了服务的配置
type Config struct {
	Port string `toml:"port"`
	// 其他服务特定配置...
}

func main() {
	// 1. 初始化日志
	logging.InitLogger()
	log.Println("Recommendation Service starting...")

	// 2. 加载配置
	var cfg Config
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./configs/recommendation.toml" // 默认配置路径
	}
	if err := config.LoadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Configuration loaded: %+v\n", cfg)

	// 3. 初始化服务依赖 (例如数据库连接、外部客户端等)
	// db, err := recommendation.NewDatabaseClient(...)
	// if err != nil {
	// 	log.Fatalf("Failed to connect to database: %v", err)
	// }
	// defer db.Close()

	// 4. 初始化推荐服务处理器
	recommendationService := recommendation.NewService() // 假设 recommendation 包提供了 NewService 函数

	// 5. 启动 HTTP 服务器
	recommendationhandler.StartHTTPServer(recommendationService, cfg.Port)
}
