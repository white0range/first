package main

import (
	"context"
	"fmt"
	"log"

	"gojo/config"
	"gojo/infrastructure/cache"
	"gojo/infrastructure/mysql"
	"gojo/infrastructure/search"
	"gojo/internal/app"
	chatHandler "gojo/internal/chat/handler"
	chatRepo "gojo/internal/chat/repository"
	chatSvc "gojo/internal/chat/service"
	chatWorker "gojo/internal/chat/worker"
	"gojo/internal/judge/docker"
	judgeRepo "gojo/internal/judge/repository"
	judgeSvc "gojo/internal/judge/service"
	judgeWorker "gojo/internal/judge/worker"
	leaderboardHandler "gojo/internal/leaderboard/handler"
	leaderboardRepo "gojo/internal/leaderboard/repository"
	leaderboardSvc "gojo/internal/leaderboard/service"
	problemHandler "gojo/internal/problem/handler"
	problemRepo "gojo/internal/problem/repository"
	problemSvc "gojo/internal/problem/service"
	subHandler "gojo/internal/submission/handler"
	subRepo "gojo/internal/submission/repository"
	subSvc "gojo/internal/submission/service"
	"gojo/internal/syncer"
	userHandler "gojo/internal/user/handler"
	userRepo "gojo/internal/user/repository"
	userSvc "gojo/internal/user/service"
)

func main() {
	fmt.Println("starting Gojo backend...")

	config.InitConfig()
	mysql.InitDB()
	cache.InitRedis()
	search.InitElasticsearch()

	if err := docker.InitDockerClient(); err != nil {
		log.Fatalf("docker client init failed: %v", err)
	}

	ur := userRepo.NewUserRepository()
	usr := userRepo.NewRefreshSessionRepository()
	pr := problemRepo.NewProblemRepository()
	sr := problemRepo.NewProblemSearchRepository()
	subR := subRepo.NewSubmissionRepository()
	syncManager := syncer.NewManager(pr, sr)
	syncManager.Start(context.Background())

	jr := judgeRepo.NewJudgeRepository(syncManager)
	lr := leaderboardRepo.NewLeaderboardRepository()
	cr := chatRepo.NewChatRepository()

	judgeService := judgeSvc.NewJudgeService(jr)

	submissionService := subSvc.NewSubmissionService(subR)
	userService := userSvc.NewUserService(ur, usr, submissionService)
	problemService := problemSvc.NewProblemService(pr, sr, syncManager)
	tagService := problemSvc.NewTagService(problemRepo.NewTagRepository(), syncManager)
	testCaseService := problemSvc.NewTestCaseService(problemRepo.NewTestCaseRepository(), syncManager)
	leaderboardService := leaderboardSvc.NewLeaderboardService(lr, userService)
	chatService := chatSvc.NewChatService(cr, userService, subR, pr)

	jw := judgeWorker.NewJudgeWorker(judgeService)
	jw.StartWorkerPool(config.GlobalConfig.Judge.WorkerCount)

	cw, err := chatWorker.NewChatWorker(cr)
	if err != nil {
		log.Fatalf("chat worker init failed: %v", err)
	}
	cw.StartTurnWorkerPool(config.GlobalConfig.Chat.WorkerCount)

	uHandler := userHandler.NewUserHandler(userService)
	pHandler := problemHandler.NewProblemHandler(problemService)
	sHandler := subHandler.NewSubmissionHandler(submissionService)
	lHandler := leaderboardHandler.NewLeaderboardHandler(leaderboardService)
	tHandler := problemHandler.NewTagHandler(tagService)
	tcHandler := problemHandler.NewTestCaseHandler(testCaseService)
	searchHandler := problemHandler.NewSearchHandler(problemService)
	cHandler := chatHandler.NewChatHandler(chatService)

	r := app.SetupRouter(
		uHandler,
		pHandler,
		sHandler,
		lHandler,
		tHandler,
		tcHandler,
		searchHandler,
		cHandler,
	)

	addr := fmt.Sprintf(":%d", config.GlobalConfig.Server.Port)
	fmt.Printf("server listening on %s\n", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("server start failed: %v", err)
	}
}
