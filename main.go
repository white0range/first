package main

import (
	"fmt"
	"gojo/config"
	"gojo/global"
	"gojo/judge"
	"gojo/models"
	"gojo/routes" // 导入咱们自己写的路由包 (注意这里的 go-oj 必须和你 go.mod 里的第一行 module 名字一致)
	"gojo/workers"
	"log"
)

func main() {
	fmt.Println("正在启动 OJ 平台核心服务器...")
	// 必须放在最前面！先加载配置，再去初始化 MySQL、Redis
	config.InitConfig()
	// 第一步：先让模型部门把数据库连上，建好表
	models.InitDB()

	// 初始化并测试 Docker 引擎连接
	// 如果连不上 Docker，整个 OJ 的判题功能就废了，所以连不上我们干脆不要启动服务器
	err := judge.InitDockerClient()
	if err != nil {
		log.Fatalf("❌ 致命错误：Docker 引擎未准备就绪, 启动失败: %v", err)
	}
	//log.Fatalf 等同于打印完日志后，立刻执行 os.Exit(1)，直接拔掉整个 Go 程序的电源，强行退出系统！

	// 2. 👇 【新增】初始化 Redis
	global.InitRedis()
	// ==========================================
	// 🚀 4. 【新增】启动后台工作池！我们招募 3 个工人就足够应付几千并发了。
	// ==========================================
	workers.StartWorkerPool(3)
	// 第二步：呼叫路由部门，拿到路由管家
	r := routes.SetupRouter()

	// 第三步：启动服务器监听
	if err := r.Run(":8080"); err != nil {
		fmt.Println("服务器启动失败: ", err)
	}
}
