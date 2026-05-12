package services

import (
	"context"
	"encoding/json"
	"fmt"
	"gojo/global"
	"gojo/models"
	"gojo/utils"
	"time"
)

// CreateProblem 负责创建题目、关联标签、关联测试用例，并清理缓存
func CreateProblem(ctx context.Context, req models.ProblemRequest) (*models.Problem, error) {
	// 1. 默认值兜底
	timeLimit := req.TimeLimit
	if timeLimit == 0 {
		timeLimit = 1000
	}
	memoryLimit := req.MemoryLimit
	if memoryLimit == 0 {
		memoryLimit = 256
	}

	// 2. 组装核心实体
	problem := models.Problem{
		Title:       req.Title,
		Description: req.Description,
		TimeLimit:   timeLimit,
		MemoryLimit: memoryLimit,
	}

	// 3. 处理可选的 Tags
	if len(req.TagIDs) > 0 {
		var tags []models.Tag
		models.DB.Find(&tags, req.TagIDs)
		problem.Tags = tags // 绑定给题目
	}

	// 4. 处理可选的 TestCases (💥 极其优雅的修复：交由 GORM 级联处理)
	if len(req.TestCases) > 0 {
		var cases []models.TestCase
		for _, tcReq := range req.TestCases {
			cases = append(cases, models.TestCase{
				Input:          tcReq.Input,
				ExpectedOutput: tcReq.ExpectedOutput,
				// 🚨 根本不需要手动写 ProblemID: problem.ID，GORM 等会知道怎么做！
			})
		}
		problem.TestCases = cases // 也是直接绑定给题目
	}

	// 5. 呼叫包工头！一行代码搞定 Transaction -> INSERT problem -> 获取 ID -> 批量 INSERT test_cases 和 problem_tags
	if err := models.DB.Create(&problem).Error; err != nil {
		return nil, err
	}

	// 6. 缓存一致性：暴力清空分页缓存
	keys, _ := global.Rdb.Keys(ctx, "cache:problems:page:*").Result()
	if len(keys) > 0 {
		global.Rdb.Del(ctx, keys...)
		fmt.Println("🗑️ 题目数据有变，已强制清空所有相关分页缓存！")
	}

	return &problem, nil
}

// 建议将这个结构体移到 models 包，这里暂且定义在 service 里演示
type ProblemListResponse struct {
	Total   int64            `json:"total"`
	Page    int              `json:"page"`
	Limit   int              `json:"limit"`
	TagID   string           `json:"tag_id"`
	Message string           `json:"message"`
	Items   []models.Problem `json:"items"`
}

func GetProblemList(ctx context.Context, page int, limit int, tagIDStr string, uid uint) (*ProblemListResponse, error) {
	// 防止恶意参数（比如传个 page=-1 搞崩数据库）
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	} // 每页最多不超100条
	// 初始化返回对象
	res := &ProblemListResponse{
		Page:  page,
		Limit: limit,
		TagID: tagIDStr,
	}

	// 2. 🔑 极其关键：拼接这页数据专属的 Redis 门牌号
	// 比如：cache:problems:page:1:size:10
	cacheKey := fmt.Sprintf("cache:problems:page:%d:limit:%d:tag:%s", res.Page, res.Limit, res.TagID)
	// 在 Java 里这相当于 List<Problem>
	//var problems []models.Problem

	// ==========================================
	// 🛡️ 阶段一：先问 Redis 秘书！
	// ==========================================
	cachedData, err := global.Rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		// 命中缓存！极其狂暴的性能提升！
		// 秘书直接把之前背下来的 JSON 字符串甩给前端，根本不碰 MySQL！
		fmt.Println("⚡ 触发 Redis 缓存！0.1 毫秒极速返回！")
		json.Unmarshal([]byte(cachedData), res)
	} else {
		// ==========================================
		// 🐌 阶段二：秘书不知道，老老实实去 MySQL 查
		// ==========================================
		fmt.Println("🐢 缓存未命中，开始苦逼地查询 MySQL...")
		// 👈 修改开始：构建动态的 GORM 查询器
		query := models.DB.Model(&models.Problem{})
		// 如果前端传了 tag_id，立刻启动 JOIN 联表过滤！
		if res.TagID != "" {
			query = query.Joins("JOIN problem_tags ON problem_tags.problem_id = problems.id").
				Where("problem_tags.tag_id = ?", res.TagID)
		}
		// 3. 先查总数 (注意：Count的时候千万不能带 limit 和 offset)
		//models.DB.Model(&models.Problem{}).Count(&total)
		query.Count(&res.Total)
		// 💥 弃用 utils.Paginate(c)，在 Service 层手动计算 Offset（极其简单）
		offset := (page - 1) * limit
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
			Scopes(utils.Paginate(offset, limit)).
			Find(&res.Items)
		// 👈 修改结束

		res.Message = "获取题目列表成功"

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
	if uid != 0 && len(res.Items) > 0 {
		// 2. 把当前这页 10 道题的 ID 抽成一个数组
		var pageProblemIDs []uint
		for _, p := range res.Items {
			pageProblemIDs = append(pageProblemIDs, p.ID)
		}

		// 3. 去数据库查：这 10 道题里，当前用户 AC 了哪些？
		var userACList []uint
		if len(pageProblemIDs) > 0 {
			models.DB.Model(&models.Submission{}).
				Where("user_id = ? AND status = 'AC' AND problem_id IN ?", uid, pageProblemIDs).
				Distinct("problem_id").
				Pluck("problem_id", &userACList)
		}

		// 4. 把 userACList 变成一个 Map（查询速度是 O(1)，极其狂暴）
		acMap := make(map[uint]bool)
		for _, id := range userACList {
			acMap[id] = true
		}

		// 5. 遍历这 10 道题，挨个染色！
		for i := range res.Items {
			if acMap[res.Items[i].ID] {
				res.Items[i].IsAC = true // 染上绿勾勾！
			}
		}

		// ⚠️ 极其关键：因为 Go 语言的 slice 里的 struct 是值传递，
		// 必须用 problems[i].IsAC = true，不能用 _, p := range 的 p.IsAC

		// 6. 把染好色的 problems 重新塞回 response 准备发给前端
		//res.Items = problems

	}
	return res, nil
}

func GetProblemDetail(ctx context.Context, problemID string) (*models.Problem, error) {
	// 1. 🔑 极其关键：拼接这道题专属的 Redis 门牌号
	// 注意：这个 Key 必须和我们刚才在 Update 和 Delete 里撕毁的 Key 保持完全一致！
	cacheKey := fmt.Sprintf("cache:problem:detail:%s", problemID)
	// ==========================================
	// 🛡️ 阶段一：先问 Redis 秘书！
	// ==========================================
	cachedData, err := global.Rdb.Get(ctx, cacheKey).Result()
	problem := &models.Problem{}
	if err == nil {
		// 命中缓存！极其狂暴的性能提升！
		fmt.Printf("⚡ 触发 Redis 缓存！题目 %s 详情 0.1 毫秒极速返回！\n", problemID)
		json.Unmarshal([]byte(cachedData), problem)
		// 因为存进去的就是提前拼好的 JSON 字符串，直接用 c.Data 原封不动地甩给前端
		//c.Data(http.StatusOK, "application/json", []byte(cachedData))
		return problem, nil
	}
	// 🐌 阶段二：秘书不知道，老老实实去 MySQL 查
	// First 方法会去数据库里找主键匹配的第一条记录
	// 对应 Java MyBatis-Plus 里的 getById()
	err = models.DB.Preload("Tags").First(problem, problemID).Error
	if err != nil {
		return nil, err
	}
	// 🎁 阶段三：让秘书复印一份记住！（写入 Redis）
	// 将整个 response 序列化成 JSON 字节流
	jsonBytes, marshalErr := json.Marshal(problem)
	if marshalErr == nil {
		// 塞进 Redis！因为题目详情很少变动，我们可以给它定一个更长的存活时间，比如 24 小时
		global.Rdb.Set(ctx, cacheKey, jsonBytes, 24*time.Hour)
	}
	return problem, nil
}

// UpdateProblem 修改题目基础信息，并清理相关缓存
func UpdateProblem(ctx context.Context, problemID string, req models.ProblemRequest) error {
	// 1. 构建需要更新的字段 Map (极其精准的按需更新)
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

	// 2. 如果没有需要更新的字段，直接返回成功
	if len(updateData) == 0 {
		return nil
	}

	// 3. 执行 MySQL 更新
	if err := models.DB.Model(&models.Problem{}).Where("id = ?", problemID).Updates(updateData).Error; err != nil {
		return err
	}

	// ==========================================
	// 🚨 极其关键：撕毁两层缓存！
	// ==========================================
	// 1. 撕毁这道题的【详情缓存】
	detailCacheKey := fmt.Sprintf("cache:problem:detail:%s", problemID)
	global.Rdb.Del(ctx, detailCacheKey)

	// 2. 极其暴力的模糊匹配，撕毁所有【列表分页缓存】
	keys, err := global.Rdb.Keys(ctx, "cache:problems:page:*").Result()
	if err == nil && len(keys) > 0 {
		global.Rdb.Del(ctx, keys...)
		fmt.Printf("🗑️ 题目 %s 被修改，已强制清空相关缓存！\n", problemID)
	}

	return nil
}

// services/problem_service.go

// DeleteProblem 删除题目，清理关联数据，并撕毁缓存
func DeleteProblem(ctx context.Context, problemID string) error {
	// 1. 查验是否存在
	var problem models.Problem
	if err := models.DB.First(&problem, problemID).Error; err != nil {
		return err // 直接返回错误，让 Controller 处理 404
	}

	// 2. 执行删除
	// 💡 架构师技巧：如果你的模型里定义了 HasMany 关系
	// 可以使用 Select("TestCases") 实现关联删除，
	// 但为了保持你逻辑的直观性，我们这里继续使用显式删除
	if err := models.DB.Delete(&problem).Error; err != nil {
		return err
	}

	// 3. 🚨 斩草除根：干掉关联的测试用例
	// 即使你通过其他接口管理用例，但题目死了，用例留着就是浪费空间的“脏数据”
	models.DB.Where("problem_id = ?", problemID).Delete(&models.TestCase{})

	// ==========================================
	// 🚨 撕毁缓存
	// ==========================================
	// 1. 撕毁详情缓存
	global.Rdb.Del(ctx, fmt.Sprintf("cache:problem:detail:%s", problemID))

	// 2. 暴力清空列表缓存
	keys, err := global.Rdb.Keys(ctx, "cache:problems:page:*").Result()
	if err == nil && len(keys) > 0 {
		global.Rdb.Del(ctx, keys...)
	}

	fmt.Printf("🗡️ 题目 %s 及其关联数据已被物理超度！\n", problemID)
	return nil
}

// services/problem_service.go

// UpdateProblemTags 更新题目的标签关联，并清理缓存
func UpdateProblemTags(ctx context.Context, problemID string, tagIDs []uint) error {
	// 1. 找到这道题
	var problem models.Problem
	if err := models.DB.First(&problem, problemID).Error; err != nil {
		return err // 交给 Controller 判断是否是 404
	}

	// 2. 把前端传来的新标签 ID，去数据库换成真实的 Tag 实体
	var tags []models.Tag
	if len(tagIDs) > 0 {
		models.DB.Find(&tags, tagIDs)
	}

	// 3. 💥 GORM 魔法：关联替换 (Replace)
	if err := models.DB.Model(&problem).Association("Tags").Replace(tags); err != nil {
		return err
	}

	// ==========================================
	// 🚨 架构师补漏：撕毁缓存！(因为标签也会在详情和列表展示)
	// ==========================================
	// 1. 撕毁这道题的详情缓存
	detailCacheKey := fmt.Sprintf("cache:problem:detail:%s", problemID)
	global.Rdb.Del(ctx, detailCacheKey)

	// 2. 暴力清空列表缓存
	keys, err := global.Rdb.Keys(ctx, "cache:problems:page:*").Result()
	if err == nil && len(keys) > 0 {
		global.Rdb.Del(ctx, keys...)
		fmt.Printf("🗑️ 题目 %s 的标签被修改，已强制清空相关缓存！\n", problemID)
	}

	return nil
}
