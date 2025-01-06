package justfiles

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/WinPooh32/go-coder/pkg/llm"
	"github.com/WinPooh32/go-coder/pkg/tasktracker"
	"gopkg.in/yaml.v3"
)

const (
	scoreThreshold = 0.01
	searchLimit    = 10
)

type taskData struct {
	ID          string    `yaml:"id"`
	Title       string    `yaml:"title"`
	Description string    `yaml:"description"`
	Vector      []float32 `yaml:"vector,flow"`

	done bool `yaml:"-"`
}

type TaskTracker struct {
	dir   string
	embed llm.Embedder
}

func NewTaskTracker(dir string, embed llm.Embedder) (*TaskTracker, error) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("make tasks directory: %w", err)
	}

	return &TaskTracker{
		dir:   dir,
		embed: embed,
	}, nil
}

func (t *TaskTracker) Set(ctx context.Context, id string, task tasktracker.Task) error {
	if task.Done {
		if err := os.MkdirAll(filepath.Join(t.dir, "done"), os.ModePerm); err != nil {
			return fmt.Errorf("make done folder: %w", err)
		}
	}

	tsk, err := t.get(id)
	if err != nil && !errors.Is(err, tasktracker.ErrNotFound) {
		return fmt.Errorf("get task: %w", err)
	}

	if !tsk.done && task.Done != tsk.done {
		filename := filepath.Join(t.dir, formatBasename(id))

		if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove file %q: %w", filename, err)
		}
	} else if tsk.done && task.Done != tsk.done {
		filename := filepath.Join(t.dir, "done", formatBasename(id))

		if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove file %q: %w", filename, err)
		}
	}

	vec, err := t.embed.Embed(ctx, formatMdText(task))
	if err != nil {
		return fmt.Errorf("get task embedding: %w", err)
	}

	newTask := taskData{
		ID:          id,
		Title:       task.Title,
		Description: task.Description,
		Vector:      vec,
		done:        task.Done,
	}

	if err := t.writeTaskToFile(id, newTask); err != nil {
		return err
	}

	return nil
}

func (t *TaskTracker) writeTaskToFile(id string, newTask taskData) error {
	var filename string

	if newTask.done {
		filename = filepath.Join(t.dir, "done", formatBasename(id))
	} else {
		filename = filepath.Join(t.dir, formatBasename(id))
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create task file %q: %w", filename, err)
	}
	defer f.Close()

	b, err := yaml.Marshal(&newTask)
	if err != nil {
		return fmt.Errorf("marshal task yaml: %w", err)
	}

	if _, err := f.Write(b); err != nil {
		return fmt.Errorf("write task to file %q: %w", filename, err)
	}

	return nil
}

func (t *TaskTracker) Get(_ context.Context, id string) (task tasktracker.Task, err error) {
	tsk, err := t.get(id)
	if err != nil {
		return task, fmt.Errorf("get task: %w", err)
	}

	return tasktracker.Task{
		Title:       tsk.Title,
		Description: tsk.Description,
		Done:        tsk.done,
	}, nil
}

func (t *TaskTracker) Search(ctx context.Context, query string) ([]tasktracker.SearchResult, error) {
	vec, err := t.embed.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get query embedding: %w", err)
	}

	tsks, err := t.getAll()
	if err != nil {
		return nil, fmt.Errorf("get all tasks: %w", err)
	}

	results, err := t.rankSearchResults(vec, tsks, scoreThreshold, searchLimit)
	if err != nil {
		return nil, fmt.Errorf("rank search results: %w", err)
	}

	return results, nil
}

func (t *TaskTracker) get(id string) (tsk taskData, err error) {
	if len(id) == 0 {
		return tsk, errors.New("empty task id")
	}

	name := formatBasename(id)

	done := false

	f, err := os.Open(filepath.Join(t.dir, name))
	if os.IsNotExist(err) {
		if f, err = os.Open(filepath.Join(t.dir, "done", name)); err != nil {
			if os.IsNotExist(err) {
				return tsk, tasktracker.ErrNotFound
			}

			return tsk, fmt.Errorf("open task file: %w", err)
		}

		done = true
	} else if err != nil {
		return tsk, fmt.Errorf("open task file: %w", err)
	}

	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return tsk, fmt.Errorf("read task file: %w", err)
	}

	if err := yaml.Unmarshal(b, &tsk); err != nil {
		return tsk, fmt.Errorf("unmarshal task yaml: %w", err)
	}

	tsk.done = done

	var errs []error

	if len(tsk.Title) == 0 {
		errs = append(errs, errors.New("empty task title"))
	}

	if len(tsk.Description) == 0 {
		errs = append(errs, errors.New("empty task description"))
	}

	return tsk, errors.Join(errs...)
}

func (t *TaskTracker) getAll() (tsks []taskData, err error) {
	err = filepath.WalkDir(t.dir, func(_ string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(info.Name())
		id := strings.TrimSuffix(info.Name(), ext)

		tsk, err := t.get(id)
		if err != nil {
			return fmt.Errorf("get task by id %q: %w", id, err)
		}

		tsks = append(tsks, tsk)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk dir %q: %w", t.dir, err)
	}

	return tsks, nil
}

func (t *TaskTracker) rankSearchResults(
	q []float32, tsks []taskData, threshold float32, limit int,
) ([]tasktracker.SearchResult, error) {
	results := make([]tasktracker.SearchResult, len(tsks))

	var maxDst float32

	for i, t := range tsks {
		dst, err := distance(q, t.Vector)
		if err != nil {
			return nil, fmt.Errorf("calc vectors distance %q: %w", t.ID, err)
		}

		if dst > maxDst {
			maxDst = dst
		}

		results[i] = tasktracker.SearchResult{
			Task: tasktracker.Task{
				Title:       t.Title,
				Description: t.Description,
				Done:        t.done,
			},
			Score: dst,
		}
	}

	// Normalize and invert.

	var filteredResults []tasktracker.SearchResult

	if maxDst > 0 {
		for _, res := range results {
			res.Score = 1 - (res.Score / maxDst)

			if res.Score > threshold {
				filteredResults = append(filteredResults, res)
			}
		}
	}

	slices.SortFunc(filteredResults, func(a, b tasktracker.SearchResult) int {
		return cmp.Compare(b.Score, a.Score) // DESC order
	})

	if len(filteredResults) > limit {
		filteredResults = slices.Clip(filteredResults[:limit])
	}

	return filteredResults, nil
}

func (t *TaskTracker) Del(_ context.Context, id string) error {
	if id == "" {
		return errors.New("empty id")
	}

	p := filepath.Join(t.dir, formatBasename(id))
	if err := os.Remove(p); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("remove file %q: %w", p, err)
		}
	}

	p = filepath.Join(t.dir, "done", formatBasename(id))
	if err := os.Remove(p); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("remove file %q: %w", p, err)
		}
	}

	return nil
}

// distance calculates the Euclidean distance between two vectors.
func distance(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, errors.New("vectors must be of the same length")
	}

	var sum float32

	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}

	return float32(math.Sqrt(float64(sum))), nil
}

func formatBasename(id string) string {
	return id + ".yaml"
}

func formatMdText(task tasktracker.Task) string {
	return fmt.Sprintf("# %s\n\n%s", task.Title, task.Description)
}
