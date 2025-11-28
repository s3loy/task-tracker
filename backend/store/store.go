package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq" // 导入 Postgres 驱动
	"github.com/s3loy/task-tracker/backend/task"
)

var (
	ErrTaskNotFound = errors.New("任务未找到")
	// db 是我们在包内部持有的数据库连接池对象
	db *sql.DB
)

// InitStore 初始化数据库连接
// 注意：这里我们修改了签名，需要传入连接字符串
func InitStore(connStr string) error {
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("无法连接到 PostgreSQL: %v", err)
	}

	return nil
}

// LoadTasks 加载所有任务
func LoadTasks() ([]task.Task, error) {
	// 按 ID 排序，保证列表顺序稳定
	query := `
		SELECT id, description, status, archived, created_at, updated_at, deadline, completed_at, started_at 
		FROM tasks 
		ORDER BY id ASC
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []task.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *t)
	}
	return tasks, nil
}

// FindByID 根据 ID 查找任务
func FindByID(id int) (*task.Task, error) {
	query := `
		SELECT id, description, status, archived, created_at, updated_at, deadline, completed_at, started_at 
		FROM tasks 
		WHERE id = $1
	`
	row := db.QueryRow(query, id)
	t, err := scanTask(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return t, nil
}

// AddTask 添加任务
func AddTask(description string, deadline *time.Time) (*task.Task, error) {
	newTask := task.NewTask(0, description) // ID 暂填 0，由数据库生成
	if err := newTask.Validate(); err != nil {
		return nil, err
	}

	query := `
		INSERT INTO tasks (description, status, created_at, updated_at, deadline) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id
	`

	// Postgres 会返回生成的 ID
	err := db.QueryRow(query,
		description,
		task.StatusTodo,
		time.Now(),
		time.Now(),
		deadline,
	).Scan(&newTask.ID)

	if err != nil {
		return nil, err
	}

	newTask.Deadline = deadline
	return newTask, nil
}

// UpdateTask 更新任务描述和截止时间
func UpdateTask(id int, description string, deadline *time.Time) error {
	if description == "" {
		return errors.New("任务描述不能为空")
	}

	query := `
		UPDATE tasks 
		SET description = $1, deadline = $2, updated_at = $3 
		WHERE id = $4
	`
	result, err := db.Exec(query, description, deadline, time.Now(), id)
	if err != nil {
		return err
	}
	return checkRowsAffected(result)
}

// DeleteTask 删除任务
func DeleteTask(id int) error {
	query := "DELETE FROM tasks WHERE id = $1"
	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	return checkRowsAffected(result)
}

// MarkTaskStatus 修改状态 (这里整合了原本的三个 Mark 函数逻辑)
func MarkTaskStatus(id int, status task.TaskStatus) error {
	if !status.IsValid() {
		return errors.New("无效的状态")
	}

	now := time.Now()
	var query string
	var args []interface{}

	// 根据不同状态，需要更新不同的时间字段
	switch status {
	case task.StatusDone:
		// 完成：更新 status, updated_at, completed_at
		query = `UPDATE tasks SET status = $1, updated_at = $2, completed_at = $3 WHERE id = $4`
		args = []interface{}{status, now, now, id}
	case task.StatusInProgress:
		// 进行中：更新 status, updated_at，如果 started_at 为空则设置 started_at
		query = `
			UPDATE tasks 
			SET status = $1, updated_at = $2, 
			    started_at = COALESCE(started_at, $3),
				completed_at = NULL 
			WHERE id = $4`
		args = []interface{}{status, now, now, id}
	case task.StatusTodo:
		// 待办：重置 status, updated_at, 清空 completed_at
		query = `UPDATE tasks SET status = $1, updated_at = $2, completed_at = NULL WHERE id = $3`
		args = []interface{}{status, now, id}
	}

	result, err := db.Exec(query, args...)
	if err != nil {
		return err
	}
	return checkRowsAffected(result)
}

// SetTaskArchived 归档/取消归档
func SetTaskArchived(id int, archived bool) error {
	query := `UPDATE tasks SET archived = $1, updated_at = $2 WHERE id = $3`
	result, err := db.Exec(query, archived, time.Now(), id)
	if err != nil {
		return err
	}
	return checkRowsAffected(result)
}

// FilterByStatus 根据状态筛选 (直接在数据库层面筛选，性能更好)
func FilterByStatus(status task.TaskStatus) ([]task.Task, error) {
	query := `
		SELECT id, description, status, archived, created_at, updated_at, deadline, completed_at, started_at 
		FROM tasks 
		WHERE status = $1 
		ORDER BY id ASC
	`
	rows, err := db.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []task.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *t)
	}
	return tasks, nil
}

// ---------------- 辅助函数 ----------------

// scanTask 用于将 SQL 行数据扫描到 Task 结构体中
// 这是一个通用接口，Row 和 Rows 都可以用
type Scannable interface {
	Scan(dest ...interface{}) error
}

func scanTask(row Scannable) (*task.Task, error) {
	var t task.Task
	// 数据库里的 NULL 需要 scan 进 sql.NullTime，或者直接用指针
	// lib/pq 驱动支持直接 scan 到 *time.Time，如果 DB 是 NULL，它会保持 nil
	// 但为了保险和兼容性，最稳妥的方法是用 sql.NullTime 中转，或者依赖驱动特性。
	// 既然你的 task 结构体里 Deadline, CompletedAt 已经是 *time.Time，
	// 我们可以直接传 &t.Deadline。pq 驱动如果看到 NULL，就不会给 t.Deadline 赋值(保持nil)。

	err := row.Scan(
		&t.ID,
		&t.Description,
		&t.Status,
		&t.Archived,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.Deadline,
		&t.CompletedAt,
		&t.StartedAt,
	)
	return &t, err
}

func checkRowsAffected(result sql.Result) error {
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrTaskNotFound
	}
	return nil
}
