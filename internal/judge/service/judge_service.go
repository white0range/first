package service

import (
	"context"
	"fmt"
	"gojo/infrastructure/websocket"
	"gojo/internal/judge/dto"
	"gojo/internal/judge/model"
	"gojo/internal/judge/repository"
	"gojo/internal/judge/sandbox"

	"gojo/pkg/compare"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type JudgeService struct {
	repo repository.JudgeRepository
}

func NewJudgeService(r repository.JudgeRepository) *JudgeService {
	return &JudgeService{repo: r}
}

// Process 核心指挥官
func (s *JudgeService) Process(ctx context.Context, task dto.JudgeTask) {
	// 1. 找仓管拿题目和测试用例
	problem, err := s.repo.GetProblemWithCases(ctx, task.ProblemID)
	if err != nil || len(problem.TestCases) == 0 {
		s.repo.UpdateJudgeResult(ctx, task.SubmissionID, task.ProblemID, task.UserID, "SE", "题目数据异常")
		return
	}

	// 2. 准备场地
	workDir, err := os.MkdirTemp("", "judge_*")
	if err != nil {
		s.repo.UpdateJudgeResult(ctx, task.SubmissionID, task.ProblemID, task.UserID, "SE", "创建场地失败")
		return
	}
	defer os.RemoveAll(workDir)

	// 3. 编译
	// 建议后续将 old.CompileCode 移到 sandbox 包下
	flag, info, err := sandbox.CompileCode(ctx, task.Code, workDir)
	if err != nil || !flag {
		s.repo.UpdateJudgeResult(ctx, task.SubmissionID, task.ProblemID, task.UserID, string(model.StatusCompileError), info)
		return
	}

	// 4. 启动常驻沙箱
	containerID, err := sandbox.StartPersistentSandbox(ctx, workDir)
	if err != nil {
		s.repo.UpdateJudgeResult(ctx, task.SubmissionID, task.ProblemID, task.UserID, "SE", "启动沙箱失败")
		return
	}
	// 💡 注意：你需要导入 docker 包，或者把 Remove 逻辑封装进 sandbox.go 里，这更符合设计
	defer sandbox.RemoveSandbox(ctx, containerID)

	// 5. 循环对决
	finalStatus := string(model.StatusAccepted)
	finalOutput := "🎉 所有的测试用例全部通过！"

	for i, tc := range problem.TestCases {
		timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		result := sandbox.ExecTestCase(timeoutCtx, containerID, tc.Input)
		cancel()

		if result.Status != model.StatusAccepted {
			finalStatus = string(result.Status)
			finalOutput = fmt.Sprintf("在第 %d 个测试点崩溃！\n日志:\n%s", i+1, result.Output)
			break
		}

		// 假设 compare 移到了 pkg/compare 里
		if !compare.CompareOutput(result.Output, tc.ExpectedOutput) {
			finalStatus = string(model.StatusWrongAnswer)
			finalOutput = fmt.Sprintf("❌ 第 %d 个测试点错误！\n你的输出:\n%s", i+1, result.Output)
			break
		}
	}

	// 6. 呼叫仓管，落库并结算积分！
	_ = s.repo.UpdateJudgeResult(ctx, task.SubmissionID, task.ProblemID, task.UserID, finalStatus, finalOutput)

	// 7. 发送实时通知
	websocket.SendWsMessage(fmt.Sprintf("%d", task.UserID), gin.H{
		"type":          "JUDGE_RESULT",
		"submission_id": task.SubmissionID,
		"status":        finalStatus,
	})
}
