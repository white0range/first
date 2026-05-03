package controllers

/*
import (
	"context"
	"fmt"
	"gojo/judge"
	"gojo/models"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CompareOutput 极其严苛又极其包容的 OJ 级字符串比对算法
func CompareOutput(userOutput string, standardAnswer string) bool {
	// strings.Fields 会极其智能地把一段长文本，按照所有的空格、换行符 \n、制表符 \t 给切碎成数组
	// 比如 "  8   \n\n  " 会被切成 ["8"]
	// 比如 "3 \n 5" 会被切成 ["3", "5"]
	userTokens := strings.Fields(userOutput)
	stdTokens := strings.Fields(standardAnswer)

	// 如果切出来的词汇数量都不一样，那绝对是错的
	if len(userTokens) != len(stdTokens) {
		return false
	}

	// 挨个比对每一个词，哪怕错了一个字符，也是 WA
	for i := range userTokens {
		if userTokens[i] != stdTokens[i] {
			return false
		}
	}

	return true
}

// SubmitCode 接收玩家提交的代码
func SubmitCode(c *gin.Context) {
	// 1. 拿出提交表单
	var req models.SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误：缺少题目ID、语言或代码"})
		return
	}

	// 2. 极其核心的一步：从保安（中间件）那里拿到当前玩家的 ID
	// 绝不能让前端传 UserID，否则黑客可以伪造别人的身份交代码！
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "系统异常：无法获取当前用户身份"})
		return
	}

	// 3. 把收到的代码组装成要下硬盘的实体对象
	submission := models.Submission{
		// 注意：中间件里存的 userID 默认是 interface{}，需要断言回 uint
		UserID:    userID.(uint),
		ProblemID: req.ProblemID,
		Language:  req.Language,
		Code:      req.Code,
		// Status 不需要写，GORM 会自动触发 default:'Pending'
	}

	// 4. 落库保存！
	if err := models.DB.Create(&submission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常，代码提交失败"})
		return
	}

	// 5. 告诉玩家：代码已收到，正在排队评测中...
	c.JSON(http.StatusOK, gin.H{
		"message":       "代码提交成功！正在排队评测中...",
		"submission_id": submission.ID,
		"status":        "Pending",
	})

	// ==========================================
	// 🚀 5. 【真正的魔法】：开启 Goroutine 后台异步判题
	// ==========================================
	// 使用 go 关键字，启动一个极其轻量的后台幽灵线程。
	// 从这一行开始，大堂经理已经回去招待下一个客人了，
	// 而后台的包工头正在默默地干脏活累活！
	// ==========================================
	go func(subID uint, probID uint, code string) {
		// 🚨【核心新增】：模拟从数据库查询该题目的测试用例
		// 真实场景下，你会用 req.ProblemID 去 Problem 表里查出这道题的测试用例。
		// ==========================================
		// 利用 GORM 的 Preload("TestCases")，一句话把这道题和它所有的测试用例全捞出来！
		var problem models.Problem
		if err := models.DB.Preload("TestCases").First(&problem, probID).Error; err != nil {
			//c.JSON(http.StatusNotFound, gin.H{"error": "这道题不存在！"})
			return
		}
		if len(problem.TestCases) == 0 {
			//c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常：这道题没有配置任何测试用例！"})
			return
		}

		ctx := context.Background()

		// 【场地管理】：由 Controller 统一开辟一块专属小黑屋
		workDir, err := os.MkdirTemp("", "judge_*")
		if err != nil {
			//c.JSON(http.StatusInternalServerError, gin.H{"error": "系统异常：无法创建判题环境"})
			models.DB.Model(&models.Submission{}).Where("id = ?", subID).Updates(map[string]interface{}{
				"status":        "SE",
				"actual_output": "系统异常：无法创建判题环境",
			})
			return
		}
		// 无论判题如何结束，最后必将这里夷为平地！
		defer os.RemoveAll(workDir)

		// 【阶段一：编译特种兵入场】
		flag, info, err := judge.CompileCode(ctx, req.Code, workDir)

		// 防线一：系统崩了
		if err != nil {
			models.DB.Model(&models.Submission{}).Where("id = ?", subID).Updates(map[string]interface{}{
				"status":        "SE",
				"actual_output": "沙箱系统崩溃",
			})
			return
		}

		// 防线二：玩家代码有语法错误 (短路结束)
		if !flag {
			models.DB.Model(&models.Submission{}).Where("id = ?", submission.ID).Updates(map[string]interface{}{
				"status":        string(judge.StatusCompileError),
				"actual_output": info,
			})
			//c.JSON(http.StatusOK, gin.H{"status": "CE", "message": "编译错误", "output": info})
			return
		}

		var finalStatus string = string(judge.StatusAccepted) // 默认是 AC，只要错一个就改掉它
		var finalOutput string = ""

		// 遍历这道题的所有测试用例
		for i, tc := range problem.TestCases {
			fmt.Printf("🔍 正在测试第 %d/%d 个用例...\n", i+1, len(problem.TestCases))

			// 👉 核心在这里：把共享的 workDir 传进去！
			result := judge.RunTestCase(ctx, workDir, tc.Input)

			// 场景 A：代码死在里面了 (CE, RE, TLE, SE)
			if result.Status != judge.StatusAccepted {
				finalStatus = string(result.Status)
				finalOutput = fmt.Sprintf("在第 %d 个测试点崩溃！\n报错日志:\n%s", i+1, result.Output)
				break // 🛑 核心：短路！直接退出循环！
			}

			// 场景 B：代码活着出来了，开始严苛对答案
			isCorrect := CompareOutput(result.Output, tc.ExpectedOutput)
			if !isCorrect {
				finalStatus = string(judge.StatusWrongAnswer)
				// 贴心地告诉玩家错在了哪里
				finalOutput = fmt.Sprintf("❌ 在第 %d 个测试点答案错误！\n输入:\n%s\n预期输出:\n%s\n你的输出:\n%s",
					i+1, tc.Input, tc.ExpectedOutput, result.Output)
				break // 🛑 核心：短路！直接退出循环！
			}
		}

		// 如果循环顺利跑完没有被 break，那么 finalStatus 依然是光荣的 AC！
		if finalStatus == string(judge.StatusAccepted) {
			finalOutput = "🎉 所有的测试用例全部通过！"
		}

		// 防御机制：截断太长的恶搞输出
		if len(finalOutput) > 2000 {
			finalOutput = finalOutput[:2000] + "\n...[输出过长，已被系统截断]..."
		}

		// 4. 【更新战果】把最终状态和输出日志一起写回数据库
		models.DB.Model(&models.Submission{}).Where("id = ?", submission.ID).Updates(map[string]interface{}{
			"status":        finalStatus,
			"actual_output": finalOutput, // 【新增】：落库保存！
		})
		fmt.Printf("✅ 提交记录 %d 的异步判题任务已完成！结果：%s\n", subID, finalStatus)
		//// 5. 【给前端发捷报/讣告】
		//c.JSON(http.StatusOK, gin.H{
		//	"message":       "评测完毕！",
		//	"submission_id": submission.ID,
		//	"status":        finalStatus, // AC, WA, TLE, SE 等
		//	"actual_output": finalOutput, // 方便前端页面展示调试信息
		//})

		// ==========================================
		// 🚀 【新增：联动大厅业务】自动计算题目通过率
		// ==========================================

		// 1. 只要提交了，不管对错，这道题的“总提交次数”必须 +1
		// ⚠️ 极其硬核的细节：这里使用了 gorm.Expr 进行“原子更新”。
		// 千万不能先查出来 +1 再存进去，高并发下数据绝对错乱！直接让 MySQL 物理执行 +1！
		models.DB.Model(&models.Problem{}).
			Where("id = ?", probID).
			UpdateColumn("submit_count", gorm.Expr("submit_count + ?", 1))

		// 2. 如果判题结果是光荣的 AC，那“通过次数”也要 +1
		if finalStatus == string(judge.StatusAccepted) {
			models.DB.Model(&models.Problem{}).
				Where("id = ?", probID).
				UpdateColumn("accepted_count", gorm.Expr("accepted_count + ?", 1))
		}
	}(submission.ID, req.ProblemID, req.Code)
}

// GetSubmissionResult 获取某次代码提交的最终战果（查快递接口）
func GetSubmissionResult(c *gin.Context) {
	// 1. 从 URL 路径中拿到物流单号 (例如 /api/submissions/5 中的 "5")
	id := c.Param("id")

	// 2. 去数据库里查这条记录
	var submission models.Submission
	// First 方法会自动生成类似 SELECT * FROM submissions WHERE id = 5 的 SQL
	if err := models.DB.First(&submission, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "客官，找不到这条提交记录！"})
		return
	}

	// 3. 将当前的最新状态返回给前端
	c.JSON(http.StatusOK, gin.H{
		"message":       "获取状态成功",
		"submission_id": submission.ID,
		"problem_id":    submission.ProblemID,
		"status":        submission.Status,       // 可能是 Pending, AC, WA, CE, SE 等
		"actual_output": submission.ActualOutput, // 如果报错了，这里有详细日志
		"language":      submission.Language,
	})
}

// GetMySubmissions 获取当前玩家的提交历史（带分页）
func GetMySubmissions(c *gin.Context) {
	// 1. 极其关键的身份校验：只能查自己的！
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "系统异常：无法获取当前用户身份"})
		return
	}

	// 2. 提取分页参数 (和大厅列表一样)
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	offset := (page - 1) * limit

	var submissions []models.Submission
	var total int64

	// 3. 递空名片，查总数 (注意：要加上 Where 条件，只查当前用户的！)
	models.DB.Model(&models.Submission{}).
		Where("user_id = ?", userID).
		Count(&total)

	// 4. 查具体的战绩列表
	// 【大厂细节 1】：Order("created_at desc") 按照时间倒序，最新提交的排在最前面！
	// 【大厂细节 2】：Omit("code", "actual_output") 如果不需要在列表页展示几百行的代码和长篇报错日志，可以用 Omit 剔除这些大字段，极大地节省网络带宽。
	models.DB.Where("user_id = ?", userID).
		Order("created_at desc").      // desc 是 descending (降序) 的缩写
		Omit("code", "actual_output"). // 列表页不查大文本（可选项，看你的前端设计）
		Limit(limit).
		Offset(offset).
		Find(&submissions)

	// 5. 组装并返回
	c.JSON(http.StatusOK, gin.H{
		"message": "获取提交历史成功",
		"data": gin.H{
			"items": submissions,
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}
*/
