// cmd/recommendation/main.go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"C:\Users\15849\Downloads\ecommerce\internal\recommendation" // 假设内部逻辑在此
	"C:\Users\15849\Downloads\ecommerce\pkg\config"             // 假设配置包在此
	"C:\Users\15849\Downloads\ecommerce\pkg\logging"            // 假设日志包在此
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
		configPath = "C:\Users\15849\Downloads\ecommerce\configs\recommendation.toml" // 默认配置路径
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

	// 5. 定义 HTTP 路由
	http.HandleFunc("/v1/recommendation", recommendationService.HandleGetRecommendation)
	http.HandleFunc("/healthz", healthCheckHandler) // 健康检查端点
	http.HandleFunc("/readyz", readinessCheckHandler) // 就绪检查端点

	// 6. 启动 HTTP 服务器
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Recommendation Service listening on %s\n", addr)
	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", addr, err)
	}
}

// healthCheckHandler 简单的健康检查
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

// readinessCheckHandler 简单的就绪检查
func readinessCheckHandler(w http.ResponseWriter, r *http.Request) {
	// 在这里可以添加更复杂的逻辑，例如检查数据库连接、依赖服务等
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Ready")
}