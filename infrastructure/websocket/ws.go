package websocket

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// 1. 魔法升级器：把普通的一锤子买卖 HTTP 协议，升级成一条永远不断的 WebSocket 水管
var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 极其关键：开发阶段允许跨域，不然前端连不上！
	},
}

// 2. 呼叫器总机（极其硬核的并发安全 Map）
// 记录格式：[玩家 UserID 字符串] -> [*websocket.Conn 呼叫器本体]
var WsClients sync.Map

// 3. 全局广播/单发大招：给指定玩家发送消息
func SendWsMessage(userID string, message interface{}) {
	// 去总机里查，这哥们儿有没有领呼叫器？
	if client, ok := WsClients.Load(userID); ok {
		conn := client.(*websocket.Conn)
		// 顺着网线把 JSON 数据砸过去！
		err := conn.WriteJSON(message)
		if err != nil {
			fmt.Printf("❌ 给玩家 %s 推送消息失败: %v\n", userID, err)
			conn.Close()
			WsClients.Delete(userID) // 发送失败说明线断了，强制销毁
		}
	}
}
