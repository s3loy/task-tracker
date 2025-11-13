package store

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/s3loy/task-tracker/backend/task"
)

var (
	ErrTaskNotFound = errors.New("任务未找到")
	tasks           []task.Task
	mu              sync.RWMutex
	nextID          int
	storePath       = "tasks.json"
)

func InitStore() error {
	mu.Lock()
	defer mu.Unlock()

	if err := loadFromFile(); err != nil {
		if os.IsNotExist(err) {
			tasks = []task.Task{}
			nextID = 1
			return saveToFile()
		}
		return err
	}

	if len(tasks) > 0 {
		maxID := 0
		for _, t := range tasks {
			if t.ID > maxID {
				maxID = t.ID
			}
		}
		nextID = maxID + 1
	} else {
		nextID = 1
	}
	return nil
}

func LoadTasks() ([]task.Task, error) {
	mu.RLock()
	defer mu.RUnlock()

	result := make([]task.Task, len(tasks))
	copy(result, tasks)
	return result, nil
}

func FindByID(id int) (*task.Task, error) {
	mu.RLock()
	defer mu.RUnlock()

	for i := range tasks {
		if tasks[i].ID == id {
			t := tasks[i]
			return &t, nil
		}
	}
	return nil, ErrTaskNotFound
}

func AddTask(description string, deadline *time.Time) (*task.Task, error) {
	mu.Lock()
	defer mu.Unlock()

	newTask := task.NewTask(nextID, description)
	if err := newTask.Validate(); err != nil {
		return nil, err
	}
	if deadline != nil {
		newTask.Deadline = deadline
	}
	tasks = append(tasks, *newTask)
	nextID++
	if err := saveToFile(); err != nil {
		return nil, err
	}
	return newTask, nil
}

func UpdateTask(id int, description string, deadline *time.Time) error {
	mu.Lock()
	defer mu.Unlock()

	for i := range tasks {
		if tasks[i].ID == id {
			if err := tasks[i].UpdateDescription(description); err != nil {
				return err
			}
			tasks[i].Deadline = deadline
			tasks[i].UpdatedAt = time.Now()
			return saveToFile()
		}
	}
	return ErrTaskNotFound
}

func DeleteTask(id int) error {
	mu.Lock()
	defer mu.Unlock()

	for i := range tasks {
		if tasks[i].ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			return saveToFile()
		}
	}
	return ErrTaskNotFound
}

func MarkTaskStatus(id int, status task.TaskStatus) error {
	mu.Lock()
	defer mu.Unlock()

	for i := range tasks {
		if tasks[i].ID == id {
			switch status {
			case task.StatusTodo:
				tasks[i].MarkAsTodo()
			case task.StatusInProgress:
				tasks[i].MarkAsInProgress()
			case task.StatusDone:
				tasks[i].MarkAsDone()
			default:
				return errors.New("无效的状态")
			}
			return saveToFile()
		}
	}
	return ErrTaskNotFound
}

func SetTaskArchived(id int, archived bool) error {
	mu.Lock()
	defer mu.Unlock()

	for i := range tasks {
		if tasks[i].ID == id {
			tasks[i].SetArchived(archived)
			return saveToFile()
		}
	}
	return ErrTaskNotFound
}

func FilterByStatus(status task.TaskStatus) ([]task.Task, error) {
	mu.RLock()
	defer mu.RUnlock()

	var result []task.Task
	for _, t := range tasks {
		if t.Status == status {
			result = append(result, t)
		}
	}
	return result, nil
}

func loadFromFile() error {
	data, err := os.ReadFile(storePath)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		tasks = []task.Task{}
		return nil
	}

	if err := json.Unmarshal(data, &tasks); err != nil {
		return err
	}

	needsSave := false
	for i := range tasks {
		if tasks[i].Status == task.StatusInProgress && tasks[i].StartedAt == nil {
			tasks[i].StartedAt = &tasks[i].CreatedAt
			needsSave = true
		}
	}
	if needsSave {
		return saveToFile()
	}
	return nil
}

func saveToFile() error {
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(storePath, data, 0644)
}
