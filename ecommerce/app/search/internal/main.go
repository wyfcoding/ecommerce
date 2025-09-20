package main

import (
	"context"
	"log"
	"net"

	"github.com/elastic/go-elasticsearch/v8"
	"google.golang.org/grpc"

	"ecommerce/ecommerce/app/search/internal/consumer"
	v1 "ecommerce/ecommerce/api/search/v1"
	"ecommerce/ecommerce/app/search/internal/biz"
	"ecommerce/ecommerce/app/search/internal/data"
	"ecommerce/ecommerce/app/search/internal/service"

const (
	searchServicePort = ":9004"
)

var (
	kafkaBrokers = []string{"localhost:9092"}
	kafkaTopic   = "product_updates"
	esAddresses  = []string{"http://localhost:9200"}
)

func main() {
	// --- 1. 初始化依賴 ---

	// 初始化 Elasticsearch 客戶端
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{Addresses: esAddresses})
	if err != nil {
		log.Fatalf("Error creating the Elasticsearch client: %s", err)
	}

	// --- 2. 啟動 Kafka 消費者 (作為後台任務) ---
	productConsumer := consumer.NewProductConsumer(kafkaBrokers, kafkaTopic, esClient)
	go func() {
		log.Println("Starting Kafka consumer for product updates...")
		productConsumer.Run(context.Background())
	}()

	// --- 3. 依賴注入並啟動 gRPC 服務 ---
	// (這部分程式碼將在下一步實現搜索接口時填充)

	searchRepo := data.NewSearchRepo(esClient)
	searchUsecase := biz.NewSearchUsecase(searchRepo)
	searchService := service.NewSearchService(searchUsecase)

	lis, err := net.Listen("tcp", searchServicePort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	searchV1.RegisterSearchServer(s, searchService)

	log.Printf("gRPC server (search-service) listening at %s", searchServicePort)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
