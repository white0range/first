package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"gojo/infrastructure/cache"
	"gojo/internal/judge/dto"
	"gojo/internal/judge/service"
	"log"
)

// 结构体注入 Service 大脑
type JudgeWorker struct {
	svc *service.JudgeService
}

func NewJudgeWorker(svc *service.JudgeService) *JudgeWorker {
	return &JudgeWorker{svc: svc}
}

// StartWorkerPool 负责启动并发
func (w *JudgeWorker) StartWorkerPool(workerCount int) {
	fmt.Printf("🚀 启动判题工作池，共 %d 个工人...\n", workerCount)
	for i := 1; i <= workerCount; i++ {
		go w.run(i)
	}
}

// run 是单个工人的生命周期
func (w *JudgeWorker) run(id int) {
	ctx := context.Background()
	for {
		// 1. 阻塞获取任务
		result, err := cache.Rdb.BRPop(ctx, 0, "judge_queue").Result()
		if err != nil {
			log.Printf("⚠️ 工人 %d 拿取任务异常: %v\n", id, err)
			continue
		}

		// 2. 解析 JSON
		var task dto.JudgeTask
		if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
			log.Printf("❌ 工人 %d JSON解析失败: %v\n", id, err)
			continue
		}

		fmt.Printf("⚙️ 工人 %d 正在处理 提交ID: %d\n", id, task.SubmissionID)

		// 3. 呼叫大脑干活！
		w.svc.Process(ctx, task)
	}
}
