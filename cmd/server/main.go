package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

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

	jw := judgeWorker.NewJudgeWorker(judgeService, subR)
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
	server := &http.Server{
		Addr:              addr,
		Handler:           r,
		ReadHeaderTimeout: secondsDuration(config.GlobalConfig.Server.ReadHeaderTimeoutSeconds),
		ReadTimeout:       secondsDuration(config.GlobalConfig.Server.ReadTimeoutSeconds),
		WriteTimeout:      secondsDuration(config.GlobalConfig.Server.WriteTimeoutSeconds),
		IdleTimeout:       secondsDuration(config.GlobalConfig.Server.IdleTimeoutSeconds),
		MaxHeaderBytes:    config.GlobalConfig.Server.MaxHeaderBytes,
	}
	fmt.Printf("server listening on %s\n", addr)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server start failed: %v", err)
	}
}

func secondsDuration(seconds int) time.Duration {
	if seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}
