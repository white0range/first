package service

import (
	"context"
	"encoding/json"
	"fmt"
	"gojo/infrastructure/cache"
	"gojo/internal/problem/dto"
	"gojo/internal/problem/model"
	"gojo/internal/problem/repository"
	"log"
	"time"
)

// 🧱 1. 定义 Service 大脑结构体（注意 P 和 S 必须大写，暴露给外部使用）
type ProblemService struct {
	// 这里以后会放 Repository (数据仓管)
	repo       repository.ProblemRepository       // 👈 依赖注入的核心
	searchRepo repository.ProblemSearchRepository // 👈 专门负责 ES 的仓管
}

// 🧱 2. 提供一个实例化方法 (给 main.go 总装车间用的)
func NewProblemService(r repository.ProblemRepository, sr repository.ProblemSearchRepository) *ProblemService {
	return &ProblemService{
		repo:       r,
		searchRepo: sr,
	}
}

func (s *ProblemService) CreateProblem(ctx context.Context, req dto.ProblemRequest) (*model.Problem, error) {
	// 1. 业务逻辑：处理默认值
	timeLimit := req.TimeLimit
	if timeLimit == 0 {
		timeLimit = 1000
	}
	memoryLimit := req.MemoryLimit
	if memoryLimit == 0 {
		memoryLimit = 256
	}

	// 2. 组装实体
	problem := model.Problem{
		Title:       req.Title,
		Description: req.Description,
		TimeLimit:   timeLimit,
		MemoryLimit: memoryLimit,
	}

	// 3. 🚨 修正：呼叫仓管拿标签，不再直连 DB
	if len(req.TagIDs) > 0 {
		tags, err := s.repo.GetTagsByIDs(ctx, req.TagIDs)
		if err == nil {
			problem.Tags = tags
		}
	}

	// 4. 组装测试用例
	if len(req.TestCases) > 0 {
		var cases []model.TestCase
		for _, tcReq := range req.TestCases {
			cases = append(cases, model.TestCase{
				Input:          tcReq.Input,
				ExpectedOutput: tcReq.ExpectedOutput,
			})
		}
		problem.TestCases = cases
	}

	// 5. 🚨 修正：呼叫仓管保存（含级联插入）
	// 执行完这一步，problem.ID 就被 GORM 自动填上了！
	if err := s.repo.CreateProblem(ctx, &problem); err != nil {
		return nil, err
	}

	// ==========================================
	// 🚀 核心新增：同步到 Elasticsearch (双写)
	// ==========================================
	var tagNames []string
	for _, t := range problem.Tags {
		tagNames = append(tagNames, t.Name)
	}

	// 呼叫咱们刚才写的 ES 仓管！
	_ = s.searchRepo.UpsertProblemToES(ctx, model.EsProblem{
		ID:          problem.ID,
		Title:       problem.Title,
		Description: problem.Description,
		Tags:        tagNames,
	})

	// 6. 缓存编排：撕毁 Redis 缓存
	s.clearProblemCache(ctx)

	return &problem, nil
}

// 辅助方法：抽取出来复用
func (s *ProblemService) clearProblemCache(ctx context.Context) {
	keys, _ := cache.Rdb.Keys(ctx, "cache:problems:page:*").Result()
	if len(keys) > 0 {
		cache.Rdb.Del(ctx, keys...)
	}
}

func (s *ProblemService) GetProblemList(ctx context.Context, page int, limit int, tagIDStr string, uid uint) (*dto.ProblemListResponse, error) {
	// 防止恶意参数（比如传个 page=-1 搞崩数据库）
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	} // 每页最多不超100条
	// 初始化返回对象
	res := &dto.ProblemListResponse{
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
	cachedData, err := cache.Rdb.Get(ctx, cacheKey).Result()
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

		items, total, err := s.repo.GetProblemList(ctx, page, limit, tagIDStr)
		if err != nil {
			return nil, err
		}
		res.Items = items
		res.Total = total
		res.Message = "获取题目列表成功"

		// ==========================================
		// 🎁 阶段三：让秘书复印一份记住！（写入 Redis）
		// ==========================================
		// 把 Go 语言的 Map/Struct 序列化成 JSON 字符串
		jsonBytes, marshalErr := json.Marshal(res)
		if marshalErr == nil {
			// 塞进 Redis，并给它定一个 1 小时的存活时间（过期自动销毁）
			cache.Rdb.Set(ctx, cacheKey, jsonBytes, 1*time.Hour)
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
		// 🚨 核心改动：呼叫 Repo 获取状态，Service 只负责比对映射
		userACList, err := s.repo.GetUserACProblemIDs(ctx, uid, pageProblemIDs)
		if err == nil {
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

	}
	return res, nil
}

func (s *ProblemService) GetProblemDetail(ctx context.Context, problemID string) (*model.Problem, error) {
	// 1. 🔑 极其关键：拼接这道题专属的 Redis 门牌号
	// 注意：这个 Key 必须和我们刚才在 Update 和 Delete 里撕毁的 Key 保持完全一致！
	cacheKey := fmt.Sprintf("cache:problem:detail:%s", problemID)
	// ==========================================
	// 🛡️ 阶段一：先问 Redis 秘书！
	// ==========================================
	cachedData, err := cache.Rdb.Get(ctx, cacheKey).Result()
	problem := &model.Problem{}
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
	//err = mysql.DB.Preload("Tags").First(problem, problemID).Error
	//if err != nil {
	//	return nil, err
	//}

	problem, err = s.repo.GetProblemByID(ctx, problemID)
	if err != nil {
		return nil, err
	}
	// 🎁 阶段三：让秘书复印一份记住！（写入 Redis）
	// 将整个 response 序列化成 JSON 字节流
	jsonBytes, marshalErr := json.Marshal(problem)
	if marshalErr == nil {
		// 塞进 Redis！因为题目详情很少变动，我们可以给它定一个更长的存活时间，比如 24 小时
		cache.Rdb.Set(ctx, cacheKey, jsonBytes, 24*time.Hour)
	}
	return problem, nil
}

// UpdateProblem 修改题目基础信息，并清理相关缓存
func (s *ProblemService) UpdateProblem(ctx context.Context, problemID string, req dto.ProblemRequest) error {
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
	if err := s.repo.UpdateProblem(ctx, problemID, updateData); err != nil {
		return err
	}

	// ==========================================
	// 🚨 极其关键：撕毁两层缓存！
	// ==========================================
	// 1. 撕毁这道题的【详情缓存】
	detailCacheKey := fmt.Sprintf("cache:problem:detail:%s", problemID)
	cache.Rdb.Del(ctx, detailCacheKey)

	// 2. 极其暴力的模糊匹配，撕毁所有【列表分页缓存】
	keys, err := cache.Rdb.Keys(ctx, "cache:problems:page:*").Result()
	if err == nil && len(keys) > 0 {
		cache.Rdb.Del(ctx, keys...)
		fmt.Printf("🗑️ 题目 %s 被修改，已强制清空相关缓存！\n", problemID)
	}

	return nil
}

// services/problem_service.go

// DeleteProblem 删除题目，清理关联数据，并撕毁缓存
func (s *ProblemService) DeleteProblem(ctx context.Context, problemID string) error {
	// 这一行代码，在底层极其安全地执行了 4 条 SQL 语句（带事务保护）
	if err := s.repo.DeleteProblem(ctx, problemID); err != nil {
		return err // 直接把错误（如 404）扔回给 Handler 门卫去处理
	}

	// ==========================================
	// 🚨 撕毁缓存
	// ==========================================
	// 1. 撕毁详情缓存
	cache.Rdb.Del(ctx, fmt.Sprintf("cache:problem:detail:%s", problemID))

	// 2. 暴力清空列表缓存
	keys, err := cache.Rdb.Keys(ctx, "cache:problems:page:*").Result()
	if err == nil && len(keys) > 0 {
		cache.Rdb.Del(ctx, keys...)
	}

	fmt.Printf("🗡️ 题目 %s 及其关联数据已被物理超度！\n", problemID)
	return nil
}

// services/problem_service.go

// UpdateProblemTags 更新题目的标签关联，并清理缓存
func (s *ProblemService) UpdateProblemTags(ctx context.Context, problemID string, tagIDs []uint) error {
	if err := s.repo.UpdateProblemTags(ctx, problemID, tagIDs); err != nil {
		return err // 直接把错误（如 404）扔回给 Handler 门卫去处理
	}

	// ==========================================
	// 🚨 架构师补漏：撕毁缓存！(因为标签也会在详情和列表展示)
	// ==========================================
	// 1. 撕毁这道题的详情缓存
	detailCacheKey := fmt.Sprintf("cache:problem:detail:%s", problemID)
	cache.Rdb.Del(ctx, detailCacheKey)

	// 2. 暴力清空列表缓存
	keys, err := cache.Rdb.Keys(ctx, "cache:problems:page:*").Result()
	if err == nil && len(keys) > 0 {
		cache.Rdb.Del(ctx, keys...)
		fmt.Printf("🗑️ 题目 %s 的标签被修改，已强制清空相关缓存！\n", problemID)
	}

	return nil
}

// 2. 纯净无比的搜索方法
func (s *ProblemService) SearchProblems(ctx context.Context, req dto.SearchRequest) (*dto.SearchResult, error) {
	// 直接呼叫 ES 仓管，不碰任何 HTTP 请求和 JSON 拼接
	total, data, err := s.searchRepo.SearchProblems(ctx, req.Keyword, req.Tags)
	if err != nil {
		return nil, err
	}

	return &dto.SearchResult{
		Total: total,
		Data:  data,
	}, nil
}

// 3. 极其优雅的全量同步方法 (老板发话大搬家！)
func (s *ProblemService) SyncAllProblemsToES(ctx context.Context) error {
	// 步骤 1：呼叫 MySQL 仓管把仓库搬空
	problems, err := s.repo.GetAllProblemsWithTags(ctx)
	if err != nil {
		return fmt.Errorf("从 MySQL 获取数据失败: %w", err)
	}

	fmt.Printf("📦 准备将 %d 道题目同步至 Elasticsearch...\n", len(problems))
	successCount := 0

	// 步骤 2：在 Service 内存中做数据转换（洗菜）
	for _, p := range problems {
		var tagNames []string
		for _, t := range p.Tags {
			tagNames = append(tagNames, t.Name)
		}

		doc := model.EsProblem{
			ID:          p.ID,
			Title:       p.Title,
			Description: p.Description,
			Tags:        tagNames,
		}

		// 步骤 3：把洗好的菜交给 ES 仓管入库
		if err := s.searchRepo.UpsertProblemToES(ctx, doc); err != nil {
			log.Printf("⚠️ 题目 ID %d 同步失败: %v\n", p.ID, err)
			continue
		}
		successCount++
	}

	fmt.Printf("✅ ES 数据初始化完毕！共成功注入 %d 道题目数据！\n", successCount)
	return nil
}
