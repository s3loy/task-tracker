package store

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/s3loy/task-tracker/backend/task"
)

const dbFile = "tasks.json"

var (
	ErrTaskNotFound = errors.New("任务未找到")
	ErrInvalidID    = errors.New("无效的任务 ID")
)

func LoadTasks() ([]task.Task, error) {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return []task.Task{}, nil
	}

	data, err := os.ReadFile(dbFile)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return []task.Task{}, nil
	}

	var tasks []task.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

func SaveTasks(tasks []task.Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(dbFile, data, 0644); err != nil {
		return err
	}

	return nil
}

func GetNextID() (int, error) {
	tasks, err := LoadTasks()
	if err != nil {
		return 0, err
	}

	if len(tasks) == 0 {
		return 1, nil
	}

	maxID := 0
	for _, t := range tasks {
		if t.ID > maxID {
			maxID = t.ID
		}
	}

	return maxID + 1, nil
}

func AddTask(description string) (*task.Task, error) {
	tasks, err := LoadTasks()
	if err != nil {
		return nil, err
	}

	id, err := GetNextID()
	if err != nil {
		return nil, err
	}

	newTask := task.NewTask(id, description)

	if err := newTask.Validate(); err != nil {
		return nil, err
	}

	tasks = append(tasks, *newTask)

	if err := SaveTasks(tasks); err != nil {
		return nil, err
	}

	return newTask, nil
}

func FindByID(id int) (*task.Task, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}

	tasks, err := LoadTasks()
	if err != nil {
		return nil, err
	}

	for i := range tasks {
		if tasks[i].ID == id {
			return &tasks[i], nil
		}
	}

	return nil, ErrTaskNotFound
}

func UpdateTask(id int, description string) error {
	tasks, err := LoadTasks()
	if err != nil {
		return err
	}

	found := false
	for i := range tasks {
		if tasks[i].ID == id {
			if err := tasks[i].UpdateDescription(description); err != nil {
				return err
			}
			found = true
			break
		}
	}

	if !found {
		return ErrTaskNotFound
	}

	return SaveTasks(tasks)
}

func DeleteTask(id int) error {
	tasks, err := LoadTasks()
	if err != nil {
		return err
	}

	found := false
	newTasks := make([]task.Task, 0, len(tasks))
	for _, t := range tasks {
		if t.ID == id {
			found = true
			continue
		}
		newTasks = append(newTasks, t)
	}

	if !found {
		return ErrTaskNotFound
	}

	return SaveTasks(newTasks)
}

func MarkTaskStatus(id int, status task.TaskStatus) error {
	tasks, err := LoadTasks()
	if err != nil {
		return err
	}

	found := false
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
			found = true
			break
		}
	}

	if !found {
		return ErrTaskNotFound
	}

	return SaveTasks(tasks)
}

func FilterByStatus(status task.TaskStatus) ([]task.Task, error) {
	tasks, err := LoadTasks()
	if err != nil {
		return nil, err
	}

	filtered := make([]task.Task, 0)
	for _, t := range tasks {
		if t.Status == status {
			filtered = append(filtered, t)
		}
	}

	return filtered, nil
}

func InitStore() error {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return SaveTasks([]task.Task{})
	}
	return nil
}
