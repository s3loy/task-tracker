package task

import (
	"errors"
	"fmt"
	"time"
)

type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in-progress"
	StatusDone       TaskStatus = "done"
	StatusOverdue    TaskStatus = "overdue"
)

type Task struct {
	ID          int        `json:"id"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	Archived    bool       `json:"archived,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	StartedAt   *time.Time `json:"startedAt,omitempty"`
}

func (t *Task) Validate() error {
	if t.Description == "" {
		return errors.New("任务描述不能为空")
	}
	if len(t.Description) > 500 {
		return errors.New("任务描述不能超过 500 个字符")
	}
	if !t.Status.IsValid() {
		return fmt.Errorf("无效的任务状态: %s", t.Status)
	}
	return nil
}

func (t *Task) MarkAsTodo() {
	t.Status = StatusTodo
	t.UpdatedAt = time.Now()
	t.CompletedAt = nil
}

func (t *Task) MarkAsInProgress() {
	t.Status = StatusInProgress
	t.UpdatedAt = time.Now()
	t.CompletedAt = nil
	if t.StartedAt == nil {
		now := time.Now()
		t.StartedAt = &now
	}
}

func (t *Task) MarkAsDone() {
	t.Status = StatusDone
	now := time.Now()
	t.UpdatedAt = now
	t.CompletedAt = &now
}

func (t *Task) SetArchived(archived bool) {
	t.Archived = archived
	t.UpdatedAt = time.Now()
}

func (t *Task) UpdateDescription(description string) error {
	if description == "" {
		return errors.New("任务描述不能为空")
	}
	if len(description) > 500 {
		return errors.New("任务描述不能超过 500 个字符")
	}
	t.Description = description
	t.UpdatedAt = time.Now()
	return nil
}

func (t *Task) String() string {
	statusIcon := map[TaskStatus]string{
		StatusTodo:       "[ ]",
		StatusInProgress: "[~]",
		StatusDone:       "[✓]",
		StatusOverdue:    "[⏰]",
	}
	icon := statusIcon[t.Status]
	return fmt.Sprintf("%d. %s %s (创建于: %s)",
		t.ID,
		icon,
		t.Description,
		t.CreatedAt.Format("2006-01-02 15:04"),
	)
}

func (s TaskStatus) IsValid() bool {
	switch s {
	case StatusTodo, StatusInProgress, StatusDone, StatusOverdue:
		return true
	default:
		return false
	}
}

func (s TaskStatus) String() string {
	return string(s)
}

func NewTask(id int, description string) *Task {
	now := time.Now()
	return &Task{
		ID:          id,
		Description: description,
		Status:      StatusTodo,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
