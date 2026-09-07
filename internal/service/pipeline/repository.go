package pipeline

import (
	"context"
	"sort"
	"sync"
)

// Repository is the persistence boundary required by Runner. CreateIfAbsent
// makes idempotent submission atomic for a storage implementation.
type Repository interface {
	CreateIfAbsent(ctx context.Context, task Task) (stored Task, created bool, err error)
	Get(ctx context.Context, taskID string) (Task, error)
	Update(ctx context.Context, task Task) error
	ListIncomplete(ctx context.Context) ([]Task, error)
}

// MemoryRepository is an in-process implementation suitable for tests and the
// MVP bootstrap. It returns defensive copies at every boundary.
type MemoryRepository struct {
	mu               sync.RWMutex
	tasks            map[string]Task
	byIdempotencyKey map[string]string
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		tasks:            make(map[string]Task),
		byIdempotencyKey: make(map[string]string),
	}
}

func (r *MemoryRepository) CreateIfAbsent(_ context.Context, task Task) (Task, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if taskID, exists := r.byIdempotencyKey[task.IdempotencyKey]; exists {
		return cloneTask(r.tasks[taskID]), false, nil
	}
	r.tasks[task.ID] = cloneTask(task)
	r.byIdempotencyKey[task.IdempotencyKey] = task.ID
	return cloneTask(task), true, nil
}

func (r *MemoryRepository) Get(_ context.Context, taskID string) (Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	task, exists := r.tasks[taskID]
	if !exists {
		return Task{}, ErrTaskNotFound
	}
	return cloneTask(task), nil
}

func (r *MemoryRepository) Update(_ context.Context, task Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tasks[task.ID]; !exists {
		return ErrTaskNotFound
	}
	r.tasks[task.ID] = cloneTask(task)
	return nil
}

func (r *MemoryRepository) ListIncomplete(_ context.Context) ([]Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tasks := make([]Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		if !task.Status.Terminal() {
			tasks = append(tasks, cloneTask(task))
		}
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
	})
	return tasks, nil
}
