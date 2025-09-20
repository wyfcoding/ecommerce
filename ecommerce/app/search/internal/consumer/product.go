package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/segmentio/kafka-go"
)

type ProductConsumer struct {
	reader *kafka.Reader
	es     *elasticsearch.Client
}

func NewProductConsumer(kafkaBrokers []string, topic string, esClient *elasticsearch.Client) *ProductConsumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  kafkaBrokers,
		Topic:    topic,
		GroupID:  "genesis-search-consumer-group", // 消費者組ID
		MinBytes: 10e3,                            // 10KB
		MaxBytes: 10e6,                            // 10MB
	})
	return &ProductConsumer{reader: r, es: esClient}
}

// Run 啟動消費者循環
func (c *ProductConsumer) Run(ctx context.Context) {
	defer c.reader.Close()

	for {
		// 讀取消息
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			log.Printf("could not fetch message: %v", err)
			break
		}

		// 解析 Canal 消息
		var canalMsg CanalMessage
		if err := json.Unmarshal(msg.Value, &canalMsg); err != nil {
			log.Printf("could not unmarshal message: %v. Raw: %s", err, string(msg.Value))
			c.reader.CommitMessages(ctx, msg) // 即使解析失敗也提交，防止壞消息阻塞隊列
			continue
		}

		// 處理消息
		if err := c.processMessage(ctx, canalMsg); err != nil {
			log.Printf("could not process message: %v", err)
			// 處理失敗，不提交 offset，以便後續重試
		} else {
			// 成功處理後，提交 offset
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				log.Printf("failed to commit messages: %v", err)
			}
		}
	}
}

// processMessage 根據消息類型處理數據
func (c *ProductConsumer) processMessage(ctx context.Context, msg CanalMessage) error {
	if msg.Table != "product_spu" {
		return nil // 我們只關心 product_spu 表的變更
	}

	switch msg.Type {
	case "INSERT", "UPDATE":
		for _, row := range msg.Data {
			doc := c.mapToDocument(row)
			if err := c.indexDocument(ctx, doc); err != nil {
				return err
			}
			log.Printf("Indexed document SPU ID: %d", doc.SpuID)
		}
	case "DELETE":
		for _, row := range msg.Data {
			spuID, ok := row["spu_id"].(float64) // Canal JSON 數字默認為 float64
			if !ok {
				continue
			}
			if err := c.deleteDocument(ctx, uint64(spuID)); err != nil {
				return err
			}
			log.Printf("Deleted document SPU ID: %d", uint64(spuID))
		}
	}
	return nil
}

// indexDocument 索引文檔到 ES
func (c *ProductConsumer) indexDocument(ctx context.Context, doc *ProductDocument) error {
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      "products",
		DocumentID: strconv.FormatUint(doc.SpuID, 10),
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing document ID %d: %s", doc.SpuID, res.String())
	}
	return nil
}

// deleteDocument 從 ES 刪除文檔
func (c *ProductConsumer) deleteDocument(ctx context.Context, spuID uint64) error {
	req := esapi.DeleteRequest{
		Index:      "products",
		DocumentID: strconv.FormatUint(spuID, 10),
		Refresh:    "true",
	}
	// ... 執行請求和錯誤處理 ...
}

// mapToDocument 將從 Canal 來的 map 數據轉換為 ES 文檔結構
func (c *ProductConsumer) mapToDocument(data map[string]interface{}) *ProductDocument {
	// ... 類型斷言和轉換邏輯 ...
	// 注意 Canal 過來的數據類型，例如數字可能是 float64，字符串可能是 string
	// 需要做安全的類型轉換
	return &ProductDocument{
		// ...
	}
}
