package architector

import (
	"context"

	"github.com/WinPooh32/go-coder/internal/developer"
	"github.com/WinPooh32/go-coder/pkg/tasktracker"
)

type spec struct {
	Title       string
	Description string
}

func (arch *Architector) analyzeTasks(ctx context.Context, tasks []tasktracker.Task) ([]developer.TaskAnalyze, error) {
	// prompt, ok := arch.prompts["analyze_task_schema"]
	// if !ok {
	// 	return nil, errors.New("not found for analyze_task_schema")
	// }
	panic("todo")
}

func (arch *Architector) generateInitialTasks(ctx context.Context) error {
	// arch.project.
	panic("todo")
}
