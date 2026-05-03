package controllers

import (
	"encoding/json"
	"fmt"
	"gojo/global"
	"gojo/models"
	"net/http"
	"strconv" //是 Go 标准库里做字符串和基本类型转换的包
	"time"

	"github.com/gin-gonic/gin"
)

// CreateProblem 接收前端发来的题目数据，并存入数据库
func CreateProblem(c *gin.Context) {
	// 1. 拿出咱们的“接客表单”（DTO）
	var req models.ProblemRequest

	// 2. 门卫查验：解析 JSON 并触发 `binding:"required"` 校验
	// 如果前端没传 Title 或 Description，这里会直接报错拦截
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误，请检查必填项：标题(title)和描述(description)"})
		return
	}

	// 手动做零值兜底，绝对安全
	if req.TimeLimit == 0 {
		req.TimeLimit = 1000 // 默认 1000ms
	}
	if req.MemoryLimit == 0 {
		req.MemoryLimit = 256 // 默认 256MB
	}

	// 2. 将 DTO 里的测试用例，转换成准备下硬盘的 DAO 对象
	var testCasesDAO []models.TestCase
	for _, tcReq := range req.TestCases {
		testCasesDAO = append(testCasesDAO, models.TestCase{
			Input:          tcReq.Input,
			ExpectedOutput: tcReq.ExpectedOutput,
		})
	}

	// 我们直接把 DTO 里的值塞进 DAO。
	// 如果前端没传 TimeLimit，req.TimeLimit 就是 0。
	// 传给 Problem 后，GORM 发现是 0，就会自动触发 default:1000 的机制！
	problem := models.Problem{
		Title:       req.Title,
		Description: req.Description,
		TimeLimit:   req.TimeLimit,
		MemoryLimit: req.MemoryLimit,

		TestCases: testCasesDAO, // 把组装好的小弟们塞进去
	}

	// 4. 呼叫包工头（GORM），执行 INSERT 语句存入数据库！
	// 它会自动执行 Transaction -> INSERT problems -> 获取 ID -> 批量 INSERT test_cases
	if err := models.DB.Create(&problem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常，题目发布失败"})
		return
	}

	// 💥【新增】：数据改了！一脚踢翻秘书的桌子！
	// 因为分页有很多个 Key（page:1, page:2...），我们为了绝对安全，直接用极其暴力的模糊匹配删掉所有列表缓存
	ctx := c.Request.Context()

	// 找出所有以 "cache:problems:page:" 开头的 key
	keys, _ := global.Rdb.Keys(ctx, "cache:problems:page:*").Result()
	if len(keys) > 0 {
		// 无情抹除！下一次有人查列表时，系统就会重新去 MySQL 查最新数据了！
		global.Rdb.Del(ctx, keys...)
		fmt.Println("🗑️ 题目数据有变，已强制清空所有相关分页缓存！")
	}
	// 5. 喜报！成功返回
	c.JSON(http.StatusOK, gin.H{
		"message":    "题目发布成功！",
		"problem_id": problem.ID, // 极其贴心的细节：GORM 存完后，会自动把自增的主键 ID 塞回这个对象里
	})
}

// GetProblemList 获取题目列表（菜单）
func GetProblemList(c *gin.Context) {
	git - v
	// 1. 从 URL 的问号后面拿分页参数，如果没传，就给默认值
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	// 转换成纯数字
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	// 防止恶意参数（比如传个 page=-1 搞崩数据库）
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	} // 每页最多不超100条
	// 2. 🔑 极其关键：拼接这页数据专属的 Redis 门牌号
	// 比如：cache:problems:page:1:size:10
	cacheKey := fmt.Sprintf("cache:problems:page:%s:limit:%s", page, limit)
	ctx := c.Request.Context()
	// ==========================================
	// 🛡️ 阶段一：先问 Redis 秘书！
	// ==========================================
	cachedData, err := global.Rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		// 命中缓存！极其狂暴的性能提升！
		// 秘书直接把之前背下来的 JSON 字符串甩给前端，根本不碰 MySQL！
		fmt.Println("⚡ 触发 Redis 缓存！0.1 毫秒极速返回！")

		// ⚠️ 注意这里的写法：因为存进去的就是 JSON 字符串，所以直接用 c.Data 返回
		c.Data(http.StatusOK, "application/json", []byte(cachedData))
		return
	}
	// ==========================================
	// 🐌 阶段二：秘书不知道，老老实实去 MySQL 查
	// ==========================================
	fmt.Println("🐢 缓存未命中，开始苦逼地查询 MySQL...")
	// 2. 计算跳过多少条 (Offset)
	// 比如第 1 页跳过 0 条，第 2 页跳过 10 条
	offset := (page - 1) * limit
	// 在 Java 里这相当于 List<Problem>
	var problems []models.Problem

	var total int64 // 极其关键：告诉前端总共有多少道题，前端才能画出底部的“共100页”按钮
	// 3. 先查总数 (注意：Count的时候千万不能带 limit 和 offset)
	models.DB.Model(&models.Problem{}).Count(&total)
	// 【大厂性能优化点】：Select
	// 列表页通常只需要展示 ID、标题、通过率。
	// 题目的 Description（描述）通常是一大段长文本，如果列表页也全查出来，会极大地浪费服务器带宽！
	// 所以我们用 Select 指定只捞取需要的列，这就是 GORM 的优雅之处。
	models.DB.Select("id", "title", "submit_count", "accepted_count").
		Limit(limit).
		Offset(offset).
		Find(&problems)

	response := gin.H{
		"message": "获取题目列表成功",
		"items":   problems, // 当前页的数组
		"total":   total,    // 题目总数
		"page":    page,     // 当前页码
		"limit":   limit,    // 每页条数
	}
	// ==========================================
	// 🎁 阶段三：让秘书复印一份记住！（写入 Redis）
	// ==========================================
	// 把 Go 语言的 Map/Struct 序列化成 JSON 字符串
	jsonBytes, marshalErr := json.Marshal(response)
	if marshalErr == nil {
		// 塞进 Redis，并给它定一个 1 小时的存活时间（过期自动销毁）
		global.Rdb.Set(ctx, cacheKey, jsonBytes, 1*time.Hour)
	}

	// 正常返回给前端
	c.JSON(http.StatusOK, response)
}

// GetProblemDetail 获取单个题目详情（菜品详情）
func GetProblemDetail(c *gin.Context) {
	// 获取 URL 路径里的动态参数，比如 /api/problems/1 里的 "1"
	id := c.Param("id")

	var problem models.Problem
	// First 方法会去数据库里找主键匹配的第一条记录
	// 对应 Java MyBatis-Plus 里的 getById()
	if err := models.DB.First(&problem, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "客官，这道题不存在或已被删除！"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "获取题目详情成功",
		"data":    problem,
	})
}
