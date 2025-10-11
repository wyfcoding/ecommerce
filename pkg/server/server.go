package server

import "context"

// Server 是一个通用的服务器接口。
type Server interface {
	Start(context.Context) error
	Stop(context.Context) error
}
