// Package utils 提供了通用的工具函数集合。
// 这个雪花算法(Snowflake)实现用于生成全局唯一的ID。
// 节点ID(nodeID)可以通过环境变量 `SNOWFLAKE_NODE_ID` 进行配置，以支持分布式部署。
// 如果不配置，默认为1。
package utils

import (
	"os"
	"strconv"
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	// node 是全局唯一的雪花算法节点实例。
	node *snowflake.Node
	// once 用于确保雪花算法节点的初始化过程只执行一次。
	once sync.Once
)

// initNode 使用单例模式(sync.Once)初始化雪花算法节点。
// 节点ID(nodeID)会尝试从环境变量 `SNOWFLAKE_NODE_ID` 读取。
// 如果环境变量未设置或解析失败，则使用默认值 1。
// 这种设计允许多个实例在不同环境下运行时，通过配置不同的节点ID来避免ID冲突。
func initNode() error {
	var err error
	once.Do(func() {
		nodeID := int64(1) // 默认 nodeID = 1

		// 尝试从环境变量读取: SNOWFLAKE_NODE_ID
		if v := os.Getenv("SNOWFLAKE_NODE_ID"); v != "" {
			if n, e := strconv.ParseInt(v, 10, 64); e == nil {
				nodeID = n
			}
		}

		node, err = snowflake.NewNode(nodeID)
	})

	return err
}

// GenerateID 生成一个线程安全的、全局唯一的雪花ID。
// 如果内部的雪花节点尚未初始化，它会自动调用 initNode() 进行初始化。
// 如果初始化失败，该函数会触发 panic，因为ID生成是系统的关键功能，失败通常意味着严重的环境配置错误。
func GenerateID() int64 {
	if node == nil {
		if err := initNode(); err != nil {
			panic("failed to init snowflake node: " + err.Error())
		}
	}
	return node.Generate().Int64()
}
