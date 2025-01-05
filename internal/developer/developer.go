package developer

import (
	"context"
	"errors"
	"fmt"
)

type Executor interface {
	Exec(ctx context.Context, task Task) error
}

type Developer struct {
	architector Architector
	coder       Executor
	tester      Executor
	debugger    Executor
	fixer       Executor
}

func NewDeveloper(
	architector Architector,
	coder Executor,
	tester Executor,
	debugger Executor,
	fixer Executor,
) (*Developer, error) {
	return &Developer{
		architector: architector,
		coder:       coder,
		tester:      tester,
		debugger:    debugger,
		fixer:       fixer,
	}, nil
}

func (dev *Developer) Develop(ctx context.Context) error {
	for {
		tasks, err := dev.architector.AnalyzeTasks(ctx)
		if err != nil {
			return fmt.Errorf("architector: analyze tasks: %w", err)
		}

		if len(tasks) == 0 {
			return errors.New("no tasks")
		}

		if allTasksFinished(tasks) {
			return nil
		}

		if hasUnclearTasks(tasks) {
			return fmt.Errorf("these tasks must be clarified: %w", formatUnclearTasksAsError(tasks))
		}

		nextTask, err := dev.architector.NextTask(ctx)
		if err != nil {
			return fmt.Errorf("architector: select next task: %w", err)
		}

		if err := dev.executeTask(ctx, nextTask); err != nil {
			return fmt.Errorf("execute task: %w", err)
		}
	}
}

func (dev *Developer) executeTask(ctx context.Context, task TaskExecute) (err error) {
	var agent Executor

	switch task.Executor {
	case TaskExecutorArchitector:
		agent = dev.architector
	case TaskExecutorCoder:
		agent = dev.coder
	case TaskExecutorFixer:
		agent = dev.fixer
	case TaskExecutorTester:
		agent = dev.tester
	case TaskExecutorDebugger:
		agent = dev.debugger
	default:
		return fmt.Errorf("unexpected task executor %d", task.Executor)
	}

	if err := agent.Exec(ctx, task.Task); err != nil {
		return fmt.Errorf("execute task by agent: %w", err)
	}

	return nil
}

func allTasksFinished(tasks []TaskAnalyze) bool {
	for _, t := range tasks {
		if !t.Done {
			return false
		}
	}

	return true
}

func hasUnclearTasks(tasks []TaskAnalyze) bool {
	for _, t := range tasks {
		if t.ClarificationNeeded {
			return true
		}
	}

	return false
}

func formatUnclearTasksAsError(tasks []TaskAnalyze) error {
	var errs []error

	for _, t := range tasks {
		errs = append(errs, fmt.Errorf("%+v", t.Task))
	}

	return errors.Join(errs...)
}
