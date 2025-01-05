package developer

import "context"

type TaskExecute struct {
	Task
	Executor TaskExecutor
}

type TaskAnalyze struct {
	Task
	Feedback            string
	ClarificationNeeded bool
	Done                bool
}

type Architector interface {
	Executor
	AnalyzeTasks(ctx context.Context) ([]TaskAnalyze, error)
	NextTask(ctx context.Context) (TaskExecute, error)
}
