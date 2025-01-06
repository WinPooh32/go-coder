package tasktracker

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("task not found")

type Tracker interface {
	Set(ctx context.Context, id string, task Task) error
	Get(ctx context.Context, id string) (Task, error)
	Del(ctx context.Context, id string) error
	Search(ctx context.Context, query string) ([]SearchResult, error)
}

type Task struct {
	Title       string
	Description string
	Done        bool
}

type SearchResult struct {
	Task
	Score float32
}