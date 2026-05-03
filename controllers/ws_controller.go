package controllers

import (
	"fmt"
	"gojo/global"

	"github.com/gin-gonic/gin"
)

// ConnectWS 处理玩家连接 WebSocket 的请求
func ConnectWS(c *gin.Context) {
	// 直接白嫖中间件解析好的 userID！极其优雅！
	userIDAny, _ := c.Get("userID")
	userID := fmt.Sprintf("%v", userIDAny)
	// 👇 加上这一行！！！
	fmt.Printf("🔴 准备发呼叫器，当前解析到的 userID 是: [%s]\n", userID)
	// 协议升级 (HTTP -> WebSocket)
	conn, err := global.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("协议升级失败:", err)
		return
	}

	// 登记呼叫器并保持死循环 (和之前完全一样)
	global.WsClients.Store(userID, conn)
	// 3. 【极其关键的防御机制】：玩家关掉网页时的善后处理
	defer func() {
		conn.Close()
		global.WsClients.Delete(userID)
	}()
	// 4. 死循环：死死盯住这条连接，接收心跳包
	// 如果不写这个死循环，这个函数瞬间就结束了，连接就会断开！
	for {
		// 我们其实不需要前端发什么复杂内容，主要靠这个动作判断线断没断
		_, _, readErr := conn.ReadMessage()
		if readErr != nil {
			break // 前端网络断了/关了网页，直接跳出循环执行 defer 销毁
		}
	}
}
