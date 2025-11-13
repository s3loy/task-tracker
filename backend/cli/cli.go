package cli

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/s3loy/task-tracker/backend/store"
	"github.com/s3loy/task-tracker/backend/task"
)

func Run() {
	if err := store.InitStore(); err != nil {
		fmt.Fprintf(os.Stderr, "初始化存储失败: %v\n", err)
		os.Exit(1)
	}
	args := os.Args

	if len(args) < 2 || args[1] == "-h" || args[1] == "--help" || args[1] == "help" {
		printHelp()
		return
	}

	command := args[1]

	switch command {
	case "add":
		handleAdd(args)
	case "list":
		handleList(args)
	case "update":
		handleUpdate(args)
	case "delete":
		handleDelete(args)
	case "mark-done":
		handleMarkDone(args)
	case "mark-in-progress":
		handleMarkInProgress(args)
	case "mark-todo":
		handleMarkTodo(args)
	case "archive":
		handleArchive(args)
	case "unarchive":
		handleUnarchive(args)
	case "show":
		handleShow(args)
	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n", command)
		fmt.Println("使用 -h 或 --help 查看帮助信息")
		os.Exit(1)
	}
}

func handleAdd(args []string) {
	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "错误: 缺少任务描述")
		fmt.Println("用法: task-cli add \"<description>\" [deadline]")
		fmt.Println("示例: task-cli add \"完成项目文档\" \"2025-11-15T18:00:00+08:00\"")
		os.Exit(1)
	}

	description := args[2]
	var deadline *time.Time

	// 如果提供了deadline参数
	if len(args) >= 4 {
		parsedTime, err := time.Parse(time.RFC3339, args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 无效的 deadline 格式 '%s'\n", args[3])
			fmt.Println("格式应为: 2025-11-15T18:00:00+08:00")
			os.Exit(1)
		}
		if parsedTime.Before(time.Now()) {
			fmt.Fprintln(os.Stderr, "错误: deadline 不能是过去的时间")
			os.Exit(1)
		}
		deadline = &parsedTime
	}

	newTask, err := store.AddTask(description, deadline)
	if err != nil {
		fmt.Fprintf(os.Stderr, "添加任务失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ 任务添加成功 (ID: %d)\n", newTask.ID)
	if deadline != nil {
		fmt.Printf("  截止时间: %s\n", deadline.Format("2006-01-02 15:04"))
	}
}

func handleList(args []string) {
	var tasks []task.Task
	var err error
	var title string

	if len(args) >= 3 {
		statusStr := args[2]
		var status task.TaskStatus

		switch statusStr {
		case "todo":
			status = task.StatusTodo
			title = "📝 待办任务"
		case "in-progress":
			status = task.StatusInProgress
			title = "🔄 进行中的任务"
		case "done":
			status = task.StatusDone
			title = "✅ 已完成的任务"
		case "overdue":
			status = task.StatusOverdue
			title = "⏰ 已过期的任务"
		default:
			fmt.Fprintf(os.Stderr, "错误: 无效的状态 '%s'\n", statusStr)
			fmt.Println("可用状态: todo, in-progress, done, overdue")
			os.Exit(1)
		}

		tasks, err = store.FilterByStatus(status)
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载任务失败: %v\n", err)
			os.Exit(1)
		}
	} else {
		title = "📋 所有任务"
		tasks, err = store.LoadTasks()
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载任务失败: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("\n=== " + title + " ===")
	if len(tasks) == 0 {
		fmt.Println("暂无任务")
		return
	}

	for _, t := range tasks {
		displayTask(&t, false)
	}
	fmt.Printf("\n共 %d 个任务\n\n", len(tasks))
}

func handleUpdate(args []string) {
	if len(args) < 4 {
		fmt.Fprintln(os.Stderr, "错误: 缺少参数")
		fmt.Println("用法: task-cli update <id> \"<new description>\" [deadline]")
		fmt.Println("示例: task-cli update 1 \"更新后的描述\" \"2025-11-15T18:00:00+08:00\"")
		os.Exit(1)
	}

	id, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无效的任务 ID '%s'\n", args[2])
		os.Exit(1)
	}

	description := args[3]
	var deadline *time.Time

	// 如果提供了deadline参数
	if len(args) >= 5 {
		parsedTime, err := time.Parse(time.RFC3339, args[4])
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: 无效的 deadline 格式 '%s'\n", args[4])
			fmt.Println("格式应为: 2025-11-15T18:00:00+08:00")
			os.Exit(1)
		}
		if parsedTime.Before(time.Now()) {
			fmt.Fprintln(os.Stderr, "错误: deadline 不能是过去的时间")
			os.Exit(1)
		}
		deadline = &parsedTime
	}

	if err := store.UpdateTask(id, description, deadline); err != nil {
		fmt.Fprintf(os.Stderr, "更新任务失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ 任务 %d 更新成功\n", id)
	if deadline != nil {
		fmt.Printf("  新的截止时间: %s\n", deadline.Format("2006-01-02 15:04"))
	}
}

func handleDelete(args []string) {
	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "错误: 缺少任务 ID")
		fmt.Println("用法: task-cli delete <id>")
		os.Exit(1)
	}

	id, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无效的任务 ID '%s'\n", args[2])
		os.Exit(1)
	}

	if err := store.DeleteTask(id); err != nil {
		fmt.Fprintf(os.Stderr, "删除任务失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ 任务 %d 已删除\n", id)
}

func handleMarkDone(args []string) {
	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "错误: 缺少任务 ID")
		fmt.Println("用法: task-cli mark-done <id>")
		os.Exit(1)
	}

	id, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无效的任务 ID '%s'\n", args[2])
		os.Exit(1)
	}

	if err := store.MarkTaskStatus(id, task.StatusDone); err != nil {
		fmt.Fprintf(os.Stderr, "标记任务失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ 任务 %d 已标记为完成\n", id)
}

func handleMarkInProgress(args []string) {
	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "错误: 缺少任务 ID")
		fmt.Println("用法: task-cli mark-in-progress <id>")
		os.Exit(1)
	}

	id, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无效的任务 ID '%s'\n", args[2])
		os.Exit(1)
	}

	if err := store.MarkTaskStatus(id, task.StatusInProgress); err != nil {
		fmt.Fprintf(os.Stderr, "标记任务失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ 任务 %d 已标记为进行中\n", id)
}

func handleMarkTodo(args []string) {
	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "错误: 缺少任务 ID")
		fmt.Println("用法: task-cli mark-todo <id>")
		os.Exit(1)
	}

	id, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无效的任务 ID '%s'\n", args[2])
		os.Exit(1)
	}

	if err := store.MarkTaskStatus(id, task.StatusTodo); err != nil {
		fmt.Fprintf(os.Stderr, "标记任务失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ 任务 %d 已标记为待办\n", id)
}

func handleArchive(args []string) {
	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "错误: 缺少任务 ID")
		fmt.Println("用法: task-cli archive <id>")
		os.Exit(1)
	}

	id, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无效的任务 ID '%s'\n", args[2])
		os.Exit(1)
	}

	if err := store.SetTaskArchived(id, true); err != nil {
		fmt.Fprintf(os.Stderr, "归档任务失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ 任务 %d 已归档\n", id)
}

func handleUnarchive(args []string) {
	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "错误: 缺少任务 ID")
		fmt.Println("用法: task-cli unarchive <id>")
		os.Exit(1)
	}

	id, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无效的任务 ID '%s'\n", args[2])
		os.Exit(1)
	}

	if err := store.SetTaskArchived(id, false); err != nil {
		fmt.Fprintf(os.Stderr, "取消归档失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ 任务 %d 已取消归档\n", id)
}

func handleShow(args []string) {
	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "错误: 缺少任务 ID")
		fmt.Println("用法: task-cli show <id>")
		os.Exit(1)
	}

	id, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无效的任务 ID '%s'\n", args[2])
		os.Exit(1)
	}

	task, err := store.FindByID(id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "查找任务失败: %v\n", err)
		os.Exit(1)
	}

	displayTask(task, true)
}

func displayTask(t *task.Task, detailed bool) {
	statusIcon := map[task.TaskStatus]string{
		task.StatusTodo:       "[ ]",
		task.StatusInProgress: "[~]",
		task.StatusDone:       "[✓]",
		task.StatusOverdue:    "[⏰]",
	}

	icon := statusIcon[t.Status]

	if detailed {
		fmt.Println("\n" + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("ID:          %d\n", t.ID)
		fmt.Printf("描述:        %s\n", t.Description)
		fmt.Printf("状态:        %s %s\n", icon, t.Status)
		fmt.Printf("创建时间:    %s\n", t.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("更新时间:    %s\n", t.UpdatedAt.Format("2006-01-02 15:04:05"))

		if t.StartedAt != nil {
			fmt.Printf("开始时间:    %s\n", t.StartedAt.Format("2006-01-02 15:04:05"))
			elapsed := time.Since(*t.StartedAt)
			fmt.Printf("已进行:      %s\n", formatDuration(elapsed))
		}

		if t.Deadline != nil {
			fmt.Printf("截止时间:    %s\n", t.Deadline.Format("2006-01-02 15:04:05"))
			if t.Status != task.StatusDone && t.Status != task.StatusOverdue {
				remaining := time.Until(*t.Deadline)
				if remaining > 0 {
					fmt.Printf("剩余时间:    %s\n", formatDuration(remaining))
				}
			}
		}

		if t.CompletedAt != nil {
			fmt.Printf("完成时间:    %s\n", t.CompletedAt.Format("2006-01-02 15:04:05"))
		}

		if t.Archived {
			fmt.Printf("归档状态:    📦 已归档\n")
		}
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	} else {
		archivedMark := ""
		if t.Archived {
			archivedMark = " 📦"
		}

		fmt.Printf("%d. %s %s%s", t.ID, icon, t.Description, archivedMark)

		if t.Deadline != nil && t.Status != task.StatusDone {
			remaining := time.Until(*t.Deadline)
			if remaining > 0 {
				fmt.Printf(" ⏳ 剩余 %s", formatDuration(remaining))
			}
		}

		if t.StartedAt != nil && t.Status == task.StatusInProgress {
			elapsed := time.Since(*t.StartedAt)
			fmt.Printf(" ⏱️  已进行 %s", formatDuration(elapsed))
		}

		fmt.Println()
	}
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}

	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%d天 %d小时 %d分钟", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%d小时 %d分钟", hours, minutes)
	} else {
		return fmt.Sprintf("%d分钟", minutes)
	}
}

func printHelp() {
	fmt.Println("┌─────────────────────────────────────────┐")
	fmt.Println("│   Task Tracker - 命令行任务管理工具     │")
	fmt.Println("└─────────────────────────────────────────┘")
	fmt.Println("\n用法:")
	fmt.Println("  task-cli <command> [arguments]")

	fmt.Println("\n📋 任务管理命令:")
	fmt.Println("  add \"<描述>\" [deadline]      添加新任务")
	fmt.Println("                               示例: task-cli add \"写文档\" \"2025-11-15T18:00:00+08:00\"")
	fmt.Println("  list [status]                列出任务")
	fmt.Println("                               可选状态: todo, in-progress, done, overdue")
	fmt.Println("  show <id>                    显示任务详情")
	fmt.Println("  update <id> \"<描述>\" [deadline]  更新任务")
	fmt.Println("  delete <id>                  删除任务")

	fmt.Println("\n🔄 状态管理命令:")
	fmt.Println("  mark-todo <id>               标记为待办")
	fmt.Println("  mark-in-progress <id>        标记为进行中")
	fmt.Println("  mark-done <id>               标记为完成")

	fmt.Println("\n📦 归档管理命令:")
	fmt.Println("  archive <id>                 归档任务")
	fmt.Println("  unarchive <id>               取消归档")

	fmt.Println("\n💡 提示:")
	fmt.Println("  - deadline 格式: 2025-11-15T18:00:00+08:00 (ISO 8601)")
	fmt.Println("  - 使用 -h 或 --help 显示此帮助信息")
	fmt.Println()
}
