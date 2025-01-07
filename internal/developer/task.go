package developer

type TaskExecutor int

const (
	TaskExecutorArchitector TaskExecutor = iota
	TaskExecutorCoder
	TaskExecutorDebugger
	TaskExecutorFixer
	TaskExecutorTester
)

type Task struct {
	ID          string
	Title       string
	Description string
}
