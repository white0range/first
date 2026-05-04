package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminCheck 超级管理员专属拦截器
func AdminCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文中撕下刚才 AuthMiddleware 贴上的标签
		roleAny, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未获取到权限信息"})
			c.Abort()
			return
		}

		// 2. 🚨 极其关键：因为你在 AuthMiddleware 存的是 uint，这里必须断言为 uint！
		role, ok := roleAny.(uint)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "权限数据格式严重异常"})
			c.Abort()
			return
		}

		if role != 1 {
			// 极其冷酷的 403 拒绝
			c.JSON(http.StatusForbidden, gin.H{"error": "僭越！你没有管理员权限！"})
			c.Abort()
			return
		}

		c.Next() // 验明正身，放行！
	}
}
