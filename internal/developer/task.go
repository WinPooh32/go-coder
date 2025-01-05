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
	Title       string
	Description string
}
