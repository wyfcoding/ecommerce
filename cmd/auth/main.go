package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"ecommerce/internal/auth/handler"
	"ecommerce/internal/auth/service"
	// 伪代码: 模拟 user gRPC 客户端
	// userpb "ecommerce/gen/user/v1"
)

func main() {
	// 1. 初始化配置
	v 