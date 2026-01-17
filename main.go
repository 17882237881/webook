package main

import (
	"webook/config"
)

func main() {
	// 使用 Wire 生成的依赖注入代码初始化 Web 服务器
	server := InitWebServer()

	// 加载配置获取端口
	cfg := config.Load()

	// 启动服务器
	server.Run(cfg.Server.Port)
}
