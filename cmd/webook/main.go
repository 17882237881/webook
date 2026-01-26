package main

import (
	"context"
	"webook/config"
)

func main() {
	// 加载配置获取端口
	cfg := config.Load()

	// 使用 Wire 生成的依赖注入代码初始化 Web 服务器
	server := InitWebServer(cfg)

	statsWorker := InitPostStatsWorker(cfg)
	statsWorker.Start(context.Background())

	// 启动服务器
	server.Run(cfg.Server.Port)
}
