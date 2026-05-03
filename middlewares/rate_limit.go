package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"gojo/global" // 引入你的 Redis 大坝

	"github.com/gin-gonic/gin"
)

// SubmitRateLimit 是专门针对代码提交接口的限流保安
func SubmitRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取玩家的唯一标识（这里用 IP 地址，如果你有 JWT，也可以换成解析出来的 UserID）
		//clientIP := c.ClientIP()
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "系统异常：无法获取当前用户身份"})
			return
		}

		// 2. 拼接他在 Redis 里的专属门牌号
		// 比如: rate_limit:submit:127.0.0.1
		redisKey := fmt.Sprintf("rate_limit:submit:%s", userID)

		// 3. 核心魔法：向 Redis 请求自增 1 (原子操作)
		// 如果 key 不存在，Redis 会自动创建并设为 1
		ctx := c.Request.Context()
		count, err := global.Rdb.Incr(ctx, redisKey).Result()
		if err != nil {
			// 如果 Redis 挂了，安全起见直接拦截（或者你也可以放行）
			c.JSON(http.StatusInternalServerError, gin.H{"error": "系统繁忙，限流器异常"})
			c.Abort() // 🛑 极其关键：拦截请求，不再往下传递！
			return
		}

		// 4. 判决时刻
		if count == 1 {
			// 如果是 1，说明他是这 5 秒内的第一次请求！
			// 立刻给这个 Key 设置 5 秒的寿命。5 秒后它会自动销毁。
			global.Rdb.Expire(ctx, redisKey, 5*time.Second)
		} else {
			// 如果大于 1，说明这个 Key 还没死（5秒还没过），他又来点提交了！
			// 直接一脚踢飞！
			c.JSON(http.StatusTooManyRequests, gin.H{ // 状态码 429
				"status":  "Error",
				"message": "手速太快啦！请 5 秒后再试！",
			})
			c.Abort() // 🛑 必须 Abort，否则请求还是会跑到 Controller 去！
			return
		}

		// 5. 检查通过，安检门放行！请求交接给你的 Controller
		c.Next()
	}
}
