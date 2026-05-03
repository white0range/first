package routes

import (
	"gojo/controllers"
	"gojo/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRouter 负责配置所有的 API 路由地址
func SetupRouter() *gin.Engine {
	r := gin.Default()
	// ====================
	// 公共区域：不需要手环谁都能进
	// ====================

	// 先写一个测试接口，验证咱们的服务器通没通
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "欢迎来到我的 OJ 平台，核心引擎运转正常！",
		})
	})

	// 以后咱们的 r.POST("/api/register") 等等都会统一写在这个文件里
	r.POST("api/register", controllers.Register)

	r.POST("api/login", controllers.Login)

	// GET 请求：获取题目列表
	r.GET("/api/problems", controllers.GetProblemList)
	// GET 请求：获取题目详情。注意这个 ":id" 是 Gin 特有的动态路由语法
	r.GET("/api/problems/:id", controllers.GetProblemDetail)
	// ====================
	// 核心区域：必须通过安检 (使用中间件)
	// ====================
	// 创建一个路由组，叫 "protected"，给它强行绑定 AuthMiddleware 保安
	protected := r.Group("/api")
	protected.Use(middlewares.AuthMiddleware())
	{
		// 这个括号里所有的路由，都会被保安死死守住！
		protected.GET("/profile", func(c *gin.Context) {
			// 能走到这里，说明百分之百是合法顾客。
			// 我们直接从他后背上把刚才保安贴的标签撕下来用！
			userID, _ := c.Get("userID")
			username, _ := c.Get("username")

			c.JSON(http.StatusOK, gin.H{
				"message": "尊贵的 VIP 用户，欢迎来到核心区域！",
				"my_id":   userID,
				"my_name": username,
			})
		})

		protected.POST("/problems", controllers.CreateProblem)
		// 👉 重点：给 /submit 接口单独套上限流保护！
		// 注意看写法：在路径和真实逻辑之间，插入 middlewares.SubmitRateLimit()
		protected.POST("/submit", middlewares.SubmitRateLimit(), controllers.SubmitCode)
		// 👇 【新增】查结果的接口 (注意是用 GET 请求，且路径带动态参数 :id)
		protected.GET("/submissions/:id", controllers.GetSubmissionResult)

		// 👇 新增这个接口 (比如叫 /my-submissions)
		protected.GET("/my-submissions", controllers.GetMySubmissions)
	}

	return r
}
