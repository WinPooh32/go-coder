package architector

import (
	"context"
	"fmt"

	"github.com/WinPooh32/go-coder/internal/developer"
	"github.com/WinPooh32/go-coder/internal/project"
	"github.com/WinPooh32/go-coder/pkg/tasktracker"
)

type Architector struct {
	project project.Config
	tracker tasktracker.Tracker

	analyzedTasks []developer.TaskAnalyze
}

func New(projcfg project.Config, tracker tasktracker.Tracker) (*Architector, error) {
	return &Architector{
		project:       projcfg,
		tracker:       tracker,
		analyzedTasks: nil,
	}, nil
}

func (arch *Architector) AnalyzeTasks(ctx context.Context) ([]developer.TaskAnalyze, error) {
	done := true

	doneTasks, err := arch.tracker.List(ctx, &done)
	if err != nil {
		return nil, fmt.Errorf("list done tasks: %w", err)
	}

	done = false

	tasks, err := arch.tracker.List(ctx, &done)
	if err != nil {
		return nil, fmt.Errorf("list active tasks: %w", err)
	}

	if len(doneTasks) == 0 && len(tasks) == 0 {
		if err := arch.generateInitialTasks(ctx); err != nil {
			return nil, fmt.Errorf("generate initial tasks: %w", err)
		}
	}

	analyze, err := arch.analyzeTasks(ctx, tasks)
	if err != nil {
		return nil, err
	}

	arch.analyzedTasks = analyze

	return analyze, nil
}

func (arch *Architector) NextTask(ctx context.Context) (developer.TaskExecute, error) {
	panic("TODO: Implement")
}

func (arch *Architector) Exec(ctx context.Context, task developer.Task) error {
	panic("TODO: Implement")
}

func (arch *Architector) generateInitialTasks(ctx context.Context) error {
	panic("TODO: Implement")
}

func (arch *Architector) analyzeTasks(ctx context.Context, tasks []tasktracker.Task) ([]developer.TaskAnalyze, error) {
	panic("TODO: Implement")
}
