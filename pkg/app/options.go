package app

import "ecommerce/pkg/server"

// Option is an application option.
type Option func(*options)

// options is an application options.
type options struct {
	servers  []server.Server
	cleanups []func()
}

// WithServer adds servers to the application.
func WithServer(servers ...server.Server) Option {
	return func(o *options) {
		o.servers = append(o.servers, servers...)
	}
}

// WithCleanup adds cleanup functions to the application.
func WithCleanup(cleanup func()) Option {
	return func(o *options) {
		o.cleanups = append(o.cleanups, cleanup)
	}
}