package tasks

import "context"

// TaskAdd adding two number
type TaskAdd struct {
}

func (t *TaskAdd) Do(ctx context.Context, a, b int) (c int, err error) {
	return 1, nil
}
