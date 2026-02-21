package main

import "log/slog"

// Returns 业务已收敛到 aftersales 服务，此入口保留用于兼容部署编排。
func main() {
	slog.Info("returns entrypoint is retained for compatibility; use aftersales service for active return flows")
}
