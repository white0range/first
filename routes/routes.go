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

	r.GET("/api/tags", controllers.GetTagList) // 👈 新增：公共获取标签接口
	r.GET("/api/leaderboard", middlewares.OptionalAuth(), controllers.GetGlobalLeaderboard)

	// 搜索题目，使用 POST 是因为需要传复杂的 JSON 过滤条件
	r.POST("/problems/search", controllers.SearchProblems)
	// ====================
	// 核心区域：必须通过安检 (使用中间件)
	// ====================
	// 创建一个路由组，叫 "protected"，给它强行绑定 AuthMiddleware 保安
	protected := r.Group("/api")
	protected.Use(middlewares.AuthMiddleware())
	{
		// ==========================================
		// 👑 皇家禁地：管理员专属操作台
		// ==========================================
		adminGroup := protected.Group("/admin")
		adminGroup.Use(middlewares.AdminCheck()) // 禁卫军：查是不是管理员！
		{
			adminGroup.POST("/problems", controllers.CreateProblem)
			// 👉 重点：给 /submit 接口单独套上限流保护！
			// 注意看写法：在路径和真实逻辑之间，插入 middlewares.SubmitRateLimit()
			adminGroup.PUT("/problems/:id", controllers.UpdateProblem)
			adminGroup.DELETE("/problems/:id", controllers.DeleteProblem)

			// 1. 获取列表：GET /api/admin/problems/1/cases 👈 【我们刚加的！】
			adminGroup.GET("/problems/:id/cases", controllers.GetTestCases)
			// 给题目加样例：POST /api/admin/problems/1/cases
			adminGroup.POST("/problems/:id/cases", controllers.AddTestCase)
			// 删掉某个样例：DELETE /api/admin/problems/cases/10
			// (注意这里的路径，因为删除只需要样例的 ID 即可，不需要题目的 ID)
			adminGroup.DELETE("/problems/cases/:case_id", controllers.DeleteTestCase)
			adminGroup.POST("/tags", controllers.CreateTag) // 👈 新增：超管创建标签
			// 👇 新增这两个门牌号
			adminGroup.DELETE("/tags/:id", controllers.DeleteTag)
			adminGroup.PUT("/problems/:id/tags", controllers.UpdateProblemTags) // 注意这里是 PUT，表示全量替换
		}
		// 这个括号里所有的路由，都会被保安死死守住！
		protected.GET("/profile", controllers.GetProfile)

		protected.POST("/submit", middlewares.SubmitRateLimit(), controllers.SubmitCode)
		// 👇 【新增】查结果的接口 (注意是用 GET 请求，且路径带动态参数 :id)
		protected.GET("/submissions/:id", controllers.GetSubmissionResult)

		// 👇 新增这个接口 (比如叫 /my-submissions)
		protected.GET("/my-submissions", controllers.GetMySubmissions)

		protected.GET("/ws", controllers.ConnectWS)
		//ai助手
		protected.GET("/submissions/:id/ai-help", middlewares.AIRateLimit(), controllers.GetAIAssistance)
	}

	return r
}
