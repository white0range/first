package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"gojo/judge"
	"gojo/models"
	"gojo/utils"
	"log"
	"os"
	"time"

	// 引入你自己的包名，比如 "gojo/global" 和 "gojo/models" 等
	"gojo/global"

	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// JudgeTask 定义从 Redis 里拿出来的 JSON 纸条长什么样
type JudgeTask struct {
	SubmissionID uint   `json:"submission_id"`
	ProblemID    uint   `json:"problem_id"`
	Code         string `json:"code"`
	UserID       uint   `json:"user_id"`
}

// StartWorkerPool 是包工头，负责一次性招募并启动 N 个工人
func StartWorkerPool(workerCount int) {
	fmt.Printf("🚀 正在启动判题工作池，共招募 %d 个后台工人...\n", workerCount)
	for i := 1; i <= workerCount; i++ {
		go worker(i) // 极其轻量！开启一个 Goroutine 就是启动一个工人
	}
}

// worker 是每个打工人真实的工作循环
func worker(id int) {
	fmt.Printf("👷 工人 %d 号已就绪，死死盯住传送带...\n", id)
	ctx := context.Background()

	// 开启打工人的无限循环（007工作制，永不退出）
	for {
		// 1. 【魔法代码 BRPop】：从 judge_queue 右侧拿任务。
		// 这里的 "0" 极其关键，它代表 "Block"（阻塞）。
		// 如果队列里没任务，工人会在这里原地睡觉，绝对不消耗 CPU！只要有任务进来，瞬间被唤醒。
		result, err := global.Rdb.BRPop(ctx, 0, "judge_queue").Result()
		if err != nil {
			log.Printf("⚠️ 工人 %d 号从队列拿取任务异常: %v\n", id, err)
			continue
		}

		// result 是一个数组：result[0] 是队列名，result[1] 是真正的 JSON 数据
		taskJSON := result[1]
		fmt.Printf("\n📦 工人 %d 号抢到了一个新任务！准备拆包...\n", id)

		// 2. 拆包裹（反序列化）：把 JSON 字符串还原成 Go 语言的结构体
		var task JudgeTask
		if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
			log.Printf("❌ 致命错误：工人 %d 号看不懂这张纸条 (JSON解析失败): %v\n", id, err)
			continue
		}

		// 3. 开始干硬活！拿着遥控器去启动 Docker (把你之前写在 go func 里的核心判题逻辑搬到这里)
		fmt.Printf("⚙️ 工人 %d 号正在用 Docker 编译运行 提交ID: %d, 题目ID: %d\n", id, task.SubmissionID, task.ProblemID)
		processSingleTask(id, task)

	}
}

// 📦 【新增】专门处理单个订单的车间！
func processSingleTask(workerID int, task JudgeTask) {
	var problem models.Problem
	if err := models.DB.Preload("TestCases").First(&problem, task.ProblemID).Error; err != nil {
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
		models.DB.Model(&models.Submission{}).Where("id = ?", task.SubmissionID).Updates(map[string]interface{}{
			"status":        "SE",
			"actual_output": "系统异常：无法创建判题环境",
		})
		return
	}
	// 无论判题如何结束，最后必将这里夷为平地！
	defer os.RemoveAll(workDir)

	// 【阶段一：编译特种兵入场】
	flag, info, err := judge.CompileCode(ctx, task.Code, workDir)
	// 防线一：系统崩了
	if err != nil {
		// 🚨 【新增】：在终端把血淋淋的真实报错打印出来！
		log.Printf("❌ 致命异常：沙箱编译阶段崩溃，错误详情: %v\n", err)
		models.DB.Model(&models.Submission{}).Where("id = ?", task.SubmissionID).Updates(map[string]interface{}{
			"status":        "SE",
			"actual_output": "沙箱系统崩溃",
		})
		return
	}

	// 防线二：玩家代码有语法错误 (短路结束)
	if !flag {
		models.DB.Model(&models.Submission{}).Where("id = ?", task.SubmissionID).Updates(map[string]interface{}{
			"status":        string(judge.StatusCompileError),
			"actual_output": info,
		})
		//c.JSON(http.StatusOK, gin.H{"status": "CE", "message": "编译错误", "output": info})
		return
	}

	var finalStatus string = string(judge.StatusAccepted) // 默认是 AC，只要错一个就改掉它
	var finalOutput string = ""

	containerID, err := judge.StartPersistentSandbox(ctx, workDir)
	if err != nil {
		models.DB.Model(&models.Submission{}).Where("id = ?", task.SubmissionID).Updates(map[string]interface{}{
			"status": "SE", "actual_output": "启动沙箱失败",
		})
		return
	}
	// 💣 必须的收尾：不管下面跑成什么样，函数退出时直接把这个常驻容器物理销毁！
	defer judge.DockerClient.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
	// 遍历这道题的所有测试用例
	for i, tc := range problem.TestCases {
		fmt.Printf("🔍 正在测试第 %d/%d 个用例...\n", i+1, len(problem.TestCases))

		// 【新增】：给单次用例加个时间紧箍咒（防 TLE 死循环）
		timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		// 👉 核心在这里：把共享的 workDir 传进去！
		//result := judge.RunTestCase(ctx, workDir, tc.Input)
		// 极其轻量的 Exec 执行！把刚才的 containerID 传进去
		result := judge.ExecTestCase(timeoutCtx, containerID, tc.Input)
		cancel() // 跑完立刻释放 context
		// 场景 A：代码死在里面了 (CE, RE, TLE, SE)
		if result.Status != judge.StatusAccepted {
			finalStatus = string(result.Status)
			finalOutput = fmt.Sprintf("在第 %d 个测试点崩溃！\n报错日志:\n%s", i+1, result.Output)
			break // 🛑 核心：短路！直接退出循环！
		}

		// 场景 B：代码活着出来了，开始严苛对答案
		isCorrect := utils.CompareOutput(result.Output, tc.ExpectedOutput)
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
	models.DB.Model(&models.Submission{}).Where("id = ?", task.SubmissionID).Updates(map[string]interface{}{
		"status":        finalStatus,
		"actual_output": finalOutput, // 【新增】：落库保存！
	})
	fmt.Printf("✅ 提交记录 %d 的异步判题任务已完成！结果：%s\n", task.SubmissionID, finalStatus)
	fmt.Printf("✅ 工人 %d 号完工！继续盯传送带...\n", workerID)

	// 1. 只要提交了，不管对错，这道题的“总提交次数”必须 +1
	// ⚠️ 极其硬核的细节：这里使用了 gorm.Expr 进行“原子更新”。
	// 千万不能先查出来 +1 再存进去，高并发下数据绝对错乱！直接让 MySQL 物理执行 +1！
	models.DB.Model(&models.Problem{}).
		Where("id = ?", task.ProblemID).
		UpdateColumn("submit_count", gorm.Expr("submit_count + ?", 1))

	// 2. 如果判题结果是光荣的 AC，那“通过次数”也要 +1
	if finalStatus == string(judge.StatusAccepted) {
		models.DB.Model(&models.Problem{}).
			Where("id = ?", task.ProblemID).
			UpdateColumn("accepted_count", gorm.Expr("accepted_count + ?", 1))
	}

	// ==========================================
	// 💥 【新增】：拿起对讲机，顺着网线拍在玩家屏幕上！
	// ==========================================
	// 假设你的 task 里能拿到提交者的 userID。如果没有，你需要查一下数据库
	userIDStr := fmt.Sprintf("%d", task.UserID)

	global.SendWsMessage(userIDStr, gin.H{
		"type":          "JUDGE_RESULT",    // 告诉前端这是一个判题结果通知
		"submission_id": task.SubmissionID, // 哪一次提交
		"status":        finalStatus,       // 是绿色的 AC，还是红色的 WA
		"message":       "你的代码判完了，赶紧看结果！",
	})

	fmt.Printf("🚀 实时判题结果已光速推送到玩家 %s 的呼叫器！\n", userIDStr)
}
