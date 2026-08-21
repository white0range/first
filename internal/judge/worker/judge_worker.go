package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"gojo/internal/judge/dto"
	judgeQueue "gojo/internal/judge/queue"
	"gojo/internal/judge/service"
	submissionModel "gojo/internal/submission/model"
)

const (
	maxRetryCount        = 5
	taskLeaseDuration    = 30 * time.Minute
	queuePollInterval    = 200 * time.Millisecond
	leaseRenewInterval   = time.Minute
	retryPollInterval    = 5 * time.Second
	recoveryPollInterval = 30 * time.Second
	reconcileInterval    = time.Minute
	pendingThreshold     = 5 * time.Minute
)

type pendingSubmissionProvider interface {
	ListPendingSubmissionsBefore(ctx context.Context, before time.Time, limit int) ([]submissionModel.Submission, error)
}

// JudgeWorker consumes judge tasks with explicit acknowledgement. A task is
// only removed after its result has been persisted to MySQL.
type JudgeWorker struct {
	svc         *service.JudgeService
	queue       *judgeQueue.Queue
	submissions pendingSubmissionProvider
}

func NewJudgeWorker(svc *service.JudgeService, submissions pendingSubmissionProvider) *JudgeWorker {
	return &JudgeWorker{svc: svc, queue: judgeQueue.New(), submissions: submissions}
}

func (w *JudgeWorker) StartWorkerPool(workerCount int) {
	ctx := context.Background()
	if err := w.queue.RecoverExpired(ctx, time.Now()); err != nil {
		log.Printf("recover expired judge tasks at startup failed: %v", err)
	}
	w.reconcilePending(ctx)

	log.Printf("starting judge worker pool, workers=%d", workerCount)
	for i := 1; i <= workerCount; i++ {
		go w.run(ctx, i)
	}
	go w.runRetryScheduler(ctx)
	go w.runRecoveryWorker(ctx)
	go w.runPendingReconciler(ctx)
}

func (w *JudgeWorker) run(ctx context.Context, id int) {
	ticker := time.NewTicker(queuePollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			raw, err := w.queue.Claim(ctx, time.Now().Add(taskLeaseDuration))
			if err != nil {
				log.Printf("judge worker %d claim task failed: %v", id, err)
				continue
			}
			if raw == "" {
				continue
			}
			w.process(ctx, id, raw)
		}
	}
}

func (w *JudgeWorker) process(ctx context.Context, workerID int, raw string) {
	var task dto.JudgeTask
	if err := json.Unmarshal([]byte(raw), &task); err != nil || !task.Valid() {
		log.Printf("judge worker %d move malformed task to dead letter: %v", workerID, err)
		if deadErr := w.queue.DeadLetterRaw(ctx, raw); deadErr != nil {
			log.Printf("judge worker %d dead letter malformed task failed: %v", workerID, deadErr)
		}
		return
	}

	log.Printf("judge worker %d processing submission_id=%d", workerID, task.SubmissionID)
	leaseCtx, cancelLease := context.WithCancel(ctx)
	defer cancelLease()
	go w.renewLease(leaseCtx, raw, task.SubmissionID)

	if err := w.svc.Process(ctx, task); err != nil {
		w.handleFailure(ctx, raw, task, err)
		return
	}
	if err := w.queue.Acknowledge(ctx, raw, task.SubmissionID); err != nil {
		log.Printf("acknowledge judge task submission_id=%d failed: %v", task.SubmissionID, err)
	}
}

func (w *JudgeWorker) renewLease(ctx context.Context, raw string, submissionID uint) {
	ticker := time.NewTicker(leaseRenewInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.queue.RenewLease(ctx, raw, time.Now().Add(taskLeaseDuration)); err != nil {
				log.Printf("renew judge task lease submission_id=%d failed: %v", submissionID, err)
			}
		}
	}
}
func (w *JudgeWorker) handleFailure(ctx context.Context, raw string, task dto.JudgeTask, cause error) {
	task.RetryCount++
	task.LastError = judgeQueue.TruncateError(cause)
	if task.RetryCount > maxRetryCount {
		log.Printf("judge task submission_id=%d exhausted retries: %v", task.SubmissionID, cause)
		if err := w.svc.MarkFailed(ctx, task, cause); err != nil {
			log.Printf("mark exhausted judge task submission_id=%d failed: %v", task.SubmissionID, err)
		}
		if err := w.queue.DeadLetter(ctx, raw, task); err != nil {
			log.Printf("move judge task submission_id=%d to dead letter failed: %v", task.SubmissionID, err)
		}
		return
	}

	retryAt := time.Now().Add(backoff(task.RetryCount))
	log.Printf("judge task submission_id=%d failed (attempt=%d), retry_at=%s: %v", task.SubmissionID, task.RetryCount, retryAt.Format(time.RFC3339), cause)
	if err := w.queue.Retry(ctx, raw, task, retryAt); err != nil {
		log.Printf("schedule judge task submission_id=%d retry failed: %v", task.SubmissionID, err)
	}
}

func (w *JudgeWorker) runRetryScheduler(ctx context.Context) {
	ticker := time.NewTicker(retryPollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			if err := w.queue.PromoteRetries(ctx, now); err != nil {
				log.Printf("promote judge retries failed: %v", err)
			}
		}
	}
}

func (w *JudgeWorker) runRecoveryWorker(ctx context.Context) {
	ticker := time.NewTicker(recoveryPollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			if err := w.queue.RecoverExpired(ctx, now); err != nil {
				log.Printf("recover expired judge tasks failed: %v", err)
			}
		}
	}
}

func (w *JudgeWorker) runPendingReconciler(ctx context.Context) {
	ticker := time.NewTicker(reconcileInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.reconcilePending(ctx)
		}
	}
}

func (w *JudgeWorker) reconcilePending(ctx context.Context) {
	if w.submissions == nil {
		return
	}

	submissions, err := w.submissions.ListPendingSubmissionsBefore(ctx, time.Now().Add(-pendingThreshold), 100)
	if err != nil {
		log.Printf("list stale pending submissions failed: %v", err)
		return
	}
	for _, submission := range submissions {
		task := dto.NewJudgeTask(submission.ID, submission.ProblemID, submission.UserID, submission.Code)
		queued, err := w.queue.Enqueue(ctx, task)
		if err != nil {
			log.Printf("requeue pending submission_id=%d failed: %v", submission.ID, err)
			continue
		}
		if queued {
			log.Printf("requeued stale pending submission_id=%d", submission.ID)
		}
	}
}

func backoff(retryCount int) time.Duration {
	delays := []time.Duration{time.Minute, 5 * time.Minute, 30 * time.Minute, 2 * time.Hour, 12 * time.Hour}
	if retryCount <= 0 {
		return delays[0]
	}
	if retryCount > len(delays) {
		return delays[len(delays)-1]
	}
	return delays[retryCount-1]
}
