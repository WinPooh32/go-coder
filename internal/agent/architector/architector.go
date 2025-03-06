package architector

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/WinPooh32/go-coder/internal/developer"
	"github.com/WinPooh32/go-coder/internal/project"
	"github.com/WinPooh32/go-coder/pkg/llm"
	"github.com/WinPooh32/go-coder/pkg/prompt"
	"github.com/WinPooh32/go-coder/pkg/tasktracker"
)

type LLMFormatter interface {
	WithJSONShema(schema json.RawMessage) (llm.MessageGenerator, error)
}

type TaskAnalysisGenerators struct {
	Chat      llm.MessageGenerator
	Formatter LLMFormatter

	// For internal use.
	withAnalyzeTaskFormat llm.MessageGenerator
}

type LLMs struct {
	TaskAnalysisGenerators
}

type Architector struct {
	project project.Config
	tracker tasktracker.Tracker

	llms    LLMs
	prompts map[string]prompt.Prompt

	analyzedTasks []developer.TaskAnalyze
}

func New(projcfg project.Config, tracker tasktracker.Tracker, llms LLMs) (*Architector, error) {
	prompts, err := prompt.Load(analyzePrompts, "_assets/*.tpl")
	if err != nil {
		return nil, fmt.Errorf("load prompt templates: %w", err)
	}

	llms.TaskAnalysisGenerators.withAnalyzeTaskFormat, err = llms.Formatter.WithJSONShema(analyzeTaskSchema)
	if err != nil {
		return nil, fmt.Errorf("make generator with analyze task format: %w", err)
	}

	return &Architector{
		project:       projcfg,
		tracker:       tracker,
		prompts:       prompts,
		llms:          llms,
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
