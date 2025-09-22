package metrics

import (
	"log"
	"net/http"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// OrdersCreatedCounter 是一个自定义的业务指标
var OrdersCreatedCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "ecommerce_orders_created_total",
		Help: "Total number of created orders.",
	},
	[]string{"status"}, // 标签：例如 "success", "failure"
)

// ProductsCreatedCounter 是一个自定义的业务指标，用于统计产品创建数量
var ProductsCreatedCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "ecommerce_products_created_total",
		Help: "Total number of created products.",
	},
	[]string{"status"}, // 标签：例如 "success", "failure"
)

// InitMetrics 注册所有需要采集的指标
func InitMetrics() {
	// 注册 gRPC 默认指标
	grpc_prometheus.EnableHandlingTimeHistogram()

	// 注册自定义指标
	prometheus.MustRegister(OrdersCreatedCounter)
	prometheus.MustRegister(ProductsCreatedCounter)
}

// ExposeHttp 启动一个独立的 HTTP 服务器，用于暴露 /metrics 端点
func ExposeHttp(port string) {
	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: promhttp.Handler(),
	}
	log.Printf("Metrics server listening on :%s", port)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("failed to start metrics server: %v", err)
	}
}
