package controllers

import (
	"encoding/json"
	"fmt"
	"gojo/global"
	"gojo/models"
	"gojo/utils"
	"net/http"
	"strconv" //是 Go 标准库里做字符串和基本类型转换的包
	"strings"
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

	//// 2. 将 DTO 里的测试用例，转换成准备下硬盘的 DAO 对象
	//var testCasesDAO []models.TestCase
	//for _, tcReq := range req.TestCases {
	//	testCasesDAO = append(testCasesDAO, models.TestCase{
	//		Input:          tcReq.Input,
	//		ExpectedOutput: tcReq.ExpectedOutput,
	//	})
	//}

	// 我们直接把 DTO 里的值塞进 DAO。
	// 如果前端没传 TimeLimit，req.TimeLimit 就是 0。
	// 传给 Problem 后，GORM 发现是 0，就会自动触发 default:1000 的机制！
	problem := models.Problem{
		Title:       req.Title,
		Description: req.Description,
		TimeLimit:   req.TimeLimit,
		MemoryLimit: req.MemoryLimit,
	}
	// 🎨 魔法一：处理可选的 Tags
	// ==========================
	if len(req.TagIDs) > 0 {
		var tags []models.Tag
		// 去数据库把对应的标签实体捞出来
		models.DB.Find(&tags, req.TagIDs)
		problem.Tags = tags // 绑定给题目
	}
	// ==========================
	// 📦 魔法二：处理可选的 TestCases
	// ==========================
	if len(req.TestCases) > 0 {
		var cases []models.TestCase
		for _, tcReq := range req.TestCases {
			cases = append(cases, models.TestCase{
				ProblemID:      problem.ID, // 拿刚刚新建成功的题目 ID
				Input:          tcReq.Input,
				ExpectedOutput: tcReq.ExpectedOutput,
			})
		}
		// 批量插入测试用例（极速）
		models.DB.Create(&cases)
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

// 专门为了这个接口定义一个返回结构体，方便我们从 Redis 里解冻数据
type ProblemListResponse struct {
	Message string           `json:"message"`
	Items   []models.Problem `json:"items"`
	Total   int64            `json:"total"`
	Page    int              `json:"page"`
	Limit   int              `json:"limit"`
}

// GetProblemList 获取题目列表（菜单）
func GetProblemList(c *gin.Context) {
	// 1. 从 URL 的问号后面拿分页参数，如果没传，就给默认值
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	tagIDStr := c.Query("tag_id") // 👈 新增：尝试获取标签 ID，如果没传就是空字符串 ""
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
	cacheKey := fmt.Sprintf("cache:problems:page:%d:limit:%d:tag:%s", page, limit, tagIDStr)
	ctx := c.Request.Context()
	authHeader := c.GetHeader("Authorization")
	// 在 Java 里这相当于 List<Problem>
	var problems []models.Problem
	var res ProblemListResponse
	var total int64 // 极其关键：告诉前端总共有多少道题，前端才能画出底部的“共100页”按钮
	// ==========================================
	// 🛡️ 阶段一：先问 Redis 秘书！
	// ==========================================
	cachedData, err := global.Rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		// 命中缓存！极其狂暴的性能提升！
		// 秘书直接把之前背下来的 JSON 字符串甩给前端，根本不碰 MySQL！
		fmt.Println("⚡ 触发 Redis 缓存！0.1 毫秒极速返回！")

		if authHeader == "" {
			// ⚠️ 注意这里的写法：因为存进去的就是 JSON 字符串，所以直接用 c.Data 返回
			c.Data(http.StatusOK, "application/json", []byte(cachedData))
			return
		}
		json.Unmarshal([]byte(cachedData), &res)
		problems = res.Items
		total = res.Total
	} else {
		// ==========================================
		// 🐌 阶段二：秘书不知道，老老实实去 MySQL 查
		// ==========================================
		fmt.Println("🐢 缓存未命中，开始苦逼地查询 MySQL...")
		// 👈 修改开始：构建动态的 GORM 查询器
		query := models.DB.Model(&models.Problem{})
		// 如果前端传了 tag_id，立刻启动 JOIN 联表过滤！
		if tagIDStr != "" {
			query = query.Joins("JOIN problem_tags ON problem_tags.problem_id = problems.id").
				Where("problem_tags.tag_id = ?", tagIDStr)
		}
		// 3. 先查总数 (注意：Count的时候千万不能带 limit 和 offset)
		//models.DB.Model(&models.Problem{}).Count(&total)
		query.Count(&total)
		// 【大厂性能优化点】：Select
		// 列表页通常只需要展示 ID、标题、通过率。
		// 题目的 Description（描述）通常是一大段长文本，如果列表页也全查出来，会极大地浪费服务器带宽！
		// 所以我们用 Select 指定只捞取需要的列，这就是 GORM 的优雅之处。
		//models.DB.Select("id", "title", "submit_count", "accepted_count").
		//Limit(limit).
		//Offset(offset).
		//Find(&problems)

		// 💥 见证奇迹的时刻：干掉繁琐的 Offset 计算，直接挂载分页插件！
		// ⚠️ 极其关键：因为用了 JOIN，Select 里的字段必须加上 "problems." 前缀，否则 MySQL 会报 Ambiguous column name (字段歧义)
		query.Select("problems.id", "problems.title", "problems.submit_count", "problems.accepted_count").
			Preload("Tags"). // 👈 新增：把题目关联的标签小尾巴也带上发给前端！
			Scopes(utils.Paginate(c)).
			Find(&problems)
		// 👈 修改结束

		res = ProblemListResponse{
			Message: "获取题目列表成功",
			Items:   problems,
			Total:   total,
			Page:    page,
			Limit:   limit,
		}
		// ==========================================
		// 🎁 阶段三：让秘书复印一份记住！（写入 Redis）
		// ==========================================
		// 把 Go 语言的 Map/Struct 序列化成 JSON 字符串
		jsonBytes, marshalErr := json.Marshal(res)
		if marshalErr == nil {
			// 塞进 Redis，并给它定一个 1 小时的存活时间（过期自动销毁）
			global.Rdb.Set(ctx, cacheKey, jsonBytes, 1*time.Hour)
		}
	}

	// ==========================================
	// 🎨 阶段四：为当前登录用户“染色”（画绿勾勾）
	// ==========================================
	// 1. 尝试从 Header 里拿 Token (因为是公共接口，我们不能强制要求有 Token，只能悄悄试探)
	if authHeader != "" {
		// 🚪 走大门 (Header)：必须严格遵守 "Bearer xxx" 的行业规范
		parts := strings.SplitN(authHeader, " ", 2)
		// 🚨 修复2：如果手环坏了，不要 Abort()，直接跳过染色，把他当游客处理即可！
		if len(parts) == 2 && parts[0] == "Bearer" {
			claims, err := utils.ParseToken(parts[1])
			if err == nil {
				currentUserID, ok := (*claims)["user_id"].(float64)
				if ok {
					UserID := uint(currentUserID)
					// 2. 把当前这页 10 道题的 ID 抽成一个数组
					var pageProblemIDs []uint
					for _, p := range problems {
						pageProblemIDs = append(pageProblemIDs, p.ID)
					}

					// 3. 去数据库查：这 10 道题里，当前用户 AC 了哪些？
					var userACList []uint
					if len(pageProblemIDs) > 0 {
						models.DB.Model(&models.Submission{}).
							Where("user_id = ? AND status = 'AC' AND problem_id IN ?", UserID, pageProblemIDs).
							Distinct("problem_id").
							Pluck("problem_id", &userACList)
					}

					// 4. 把 userACList 变成一个 Map（查询速度是 O(1)，极其狂暴）
					acMap := make(map[uint]bool)
					for _, id := range userACList {
						acMap[id] = true
					}

					// 5. 遍历这 10 道题，挨个染色！
					for i := range problems {
						if acMap[problems[i].ID] {
							problems[i].IsAC = true // 染上绿勾勾！
						}
					}

					// ⚠️ 极其关键：因为 Go 语言的 slice 里的 struct 是值传递，
					// 必须用 problems[i].IsAC = true，不能用 _, p := range 的 p.IsAC

					// 6. 把染好色的 problems 重新塞回 response 准备发给前端
					res.Items = problems
				}
			}
		}

	}

	// 正常返回给前端 (带有 is_ac 字段了！)
	c.JSON(http.StatusOK, res)
}

// GetProblemDetail 获取单个题目详情（菜品详情）
func GetProblemDetail(c *gin.Context) {
	// 获取 URL 路径里的动态参数，比如 /api/problems/1 里的 "1"
	id := c.Param("id")
	// 1. 🔑 极其关键：拼接这道题专属的 Redis 门牌号
	// 注意：这个 Key 必须和我们刚才在 Update 和 Delete 里撕毁的 Key 保持完全一致！
	cacheKey := fmt.Sprintf("cache:problem:detail:%s", id)
	ctx := c.Request.Context()
	// ==========================================
	// 🛡️ 阶段一：先问 Redis 秘书！
	// ==========================================
	cachedData, err := global.Rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		// 命中缓存！极其狂暴的性能提升！
		fmt.Printf("⚡ 触发 Redis 缓存！题目 %s 详情 0.1 毫秒极速返回！\n", id)

		// 因为存进去的就是提前拼好的 JSON 字符串，直接用 c.Data 原封不动地甩给前端
		c.Data(http.StatusOK, "application/json", []byte(cachedData))
		return
	}
	// 🐌 阶段二：秘书不知道，老老实实去 MySQL 查
	var problem models.Problem
	// First 方法会去数据库里找主键匹配的第一条记录
	// 对应 Java MyBatis-Plus 里的 getById()
	if err := models.DB.First(&problem, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "客官，这道题不存在或已被删除！"})
		return
	}

	// 🎁 阶段三：让秘书复印一份记住！（写入 Redis）
	response := gin.H{
		"message": "获取题目详情成功",
		"data":    problem,
	}

	// 将整个 response 序列化成 JSON 字节流
	jsonBytes, marshalErr := json.Marshal(response)
	if marshalErr == nil {
		// 塞进 Redis！因为题目详情很少变动，我们可以给它定一个更长的存活时间，比如 24 小时
		global.Rdb.Set(ctx, cacheKey, jsonBytes, 24*time.Hour)
	}
	c.JSON(http.StatusOK, response)
}

// UpdateProblem 修改题目（管理员专属）
func UpdateProblem(c *gin.Context) {
	// 1. 获取要修改的题目 ID
	problemID := c.Param("id")

	// 2. 拿出“接客表单”（这里我们复用创建时的请求体结构）
	var req models.ProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误，请检查传入的数据格式"})
		return
	}

	// 3. 构建需要更新的字段 Map
	// 为什么不用 Struct 更新？
	// 因为如果是 Struct，req.TimeLimit 如果是 0，GORM 会认为你不想更新这个字段（零值忽略机制）。
	// 用 Map 可以实现极其精准的按需更新。
	updateData := make(map[string]interface{})
	if req.Title != "" {
		updateData["title"] = req.Title
	}
	if req.Description != "" {
		updateData["description"] = req.Description
	}
	if req.TimeLimit > 0 {
		updateData["time_limit"] = req.TimeLimit
	}
	if req.MemoryLimit > 0 {
		updateData["memory_limit"] = req.MemoryLimit
	}

	// 2. 更新数据库
	if len(updateData) > 0 {
		if err := models.DB.Model(&models.Problem{}).Where("id = ?", problemID).Updates(updateData).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常，题目更新失败"})
			return
		}
	}

	// ==========================================
	// 🚨 极其关键：撕毁两层缓存！
	// ==========================================
	ctx := c.Request.Context()

	// 1. 撕毁这道题的【详情缓存】(如果你给 GetProblemDetail 也加了缓存的话)
	detailCacheKey := fmt.Sprintf("cache:problem:detail:%s", problemID)
	global.Rdb.Del(ctx, detailCacheKey)

	// 2. 极其暴力的模糊匹配，撕毁所有【列表分页缓存】
	keys, _ := global.Rdb.Keys(ctx, "cache:problems:page:*").Result()
	if len(keys) > 0 {
		global.Rdb.Del(ctx, keys...)
		fmt.Printf("🗑️ 题目 %s 被修改，已强制清空相关缓存！\n", problemID)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "题目修改成功，天下大吉！",
	})
}

// DeleteProblem 删除题目（管理员专属）
func DeleteProblem(c *gin.Context) {
	problemID := c.Param("id")

	// 1. 查验是否存在（避免空删）
	var problem models.Problem
	if err := models.DB.First(&problem, problemID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "斩杀失败：这道题根本不存在！"})
		return
	}

	// 2. 执行删除
	// 如果 problem 结构体里嵌套了 gorm.Model，这里会自动变成 UPDATE problems SET deleted_at = NOW() WHERE id = ?
	// 如果没有 gorm.Model，就是物理删除 DELETE FROM problems WHERE id = ?
	if err := models.DB.Delete(&problem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常，题目删除失败"})
		return
	}

	// 3. 🚨 同样极其关键：既然题目都没了，相关的小弟（测试用例）也必须干掉！
	// 执行：DELETE FROM test_cases WHERE problem_id = ?
	models.DB.Where("problem_id = ?", problemID).Delete(&models.TestCase{})

	// ==========================================
	// 🚨 撕毁缓存（与更新同理）
	// ==========================================
	ctx := c.Request.Context()
	global.Rdb.Del(ctx, fmt.Sprintf("cache:problem:detail:%s", problemID))
	keys, _ := global.Rdb.Keys(ctx, "cache:problems:page:*").Result()
	if len(keys) > 0 {
		global.Rdb.Del(ctx, keys...)
	}

	fmt.Printf("🗡️ 管理员下达处决令，题目 %s 及其测试用例已被物理超度！\n", problemID)
	c.JSON(http.StatusOK, gin.H{"message": "题目已成功删除"})
}
