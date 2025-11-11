package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/s3loy/task-tracker/backend/store"
	"github.com/s3loy/task-tracker/backend/task"
)

func Run() {
	if err := store.InitStore(); err != nil {
		fmt.Fprintf(os.Stderr, "初始化存储失败: %v\n", err)
		os.Exit(1)
	}
	args := os.Args

	if len(args) < 2 || args[1] == "-h" || args[1] == "--help" {
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
	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n", command)
		fmt.Println("使用 -h 或 --help 查看帮助信息")
		os.Exit(1)
	}
}

func handleAdd(args []string) {
	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "错误: 缺少任务描述")
		fmt.Println("用法: task-cli add \"<description>\"")
		os.Exit(1)
	}

	description := args[2]
	newTask, err := store.AddTask(description)
	if err != nil {
		fmt.Fprintf(os.Stderr, "添加任务失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ 任务添加成功 (ID: %d)\n", newTask.ID)
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
			title = "待办任务"
		case "in-progress":
			status = task.StatusInProgress
			title = "进行中的任务"
		case "done":
			status = task.StatusDone
			title = "已完成的任务"
		case "inbox":
			status = task.StatusInbox
			title = "收集箱任务"
		default:
			fmt.Fprintf(os.Stderr, "错误: 无效的状态 '%s'\n", statusStr)
			fmt.Println("可用状态: todo, in-progress, done, inbox")
			os.Exit(1)
		}

		tasks, err = store.FilterByStatus(status)
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载任务失败: %v\n", err)
			os.Exit(1)
		}
	} else {
		title = "所有任务"
		tasks, err = store.LoadTasks()
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载任务失败: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("=== " + title + " ===")
	if len(tasks) == 0 {
		fmt.Println("暂无任务")
		return
	}

	for _, t := range tasks {
		fmt.Println(t.String())
	}
	fmt.Printf("\n共 %d 个任务\n", len(tasks))
}

func handleUpdate(args []string) {
	if len(args) < 4 {
		fmt.Fprintln(os.Stderr, "错误: 缺少参数")
		fmt.Println("用法: task-cli update <id> \"<new description>\"")
		os.Exit(1)
	}

	id, err := strconv.Atoi(args[2]) // string conversion  ASCII to Integer
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 无效的任务 ID '%s'\n", args[2])
		os.Exit(1)
	}

	description := args[3]
	if err := store.UpdateTask(id, description); err != nil {
		fmt.Fprintf(os.Stderr, "更新任务失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ 任务 %d 更新成功\n", id)
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

func printHelp() {
	fmt.Println("┌─────────────────────────────────────────┐")
	fmt.Println("│   Task Tracker - 命令行任务管理工具     │")
	fmt.Println("└─────────────────────────────────────────┘")
	fmt.Println("也许还在开发中？")
	fmt.Println("\n用法:")
	fmt.Println("  task-cli <command> [arguments]")
	fmt.Println("\n可用命令:")
	fmt.Println("  add \"<description>\"       添加一个新任务")
	fmt.Println("  list [status]             列出任务")
	fmt.Println("                            可选状态: todo, in-progress, done, inbox")
	fmt.Println("  update <id> \"<desc>\"      更新任务描述")
	fmt.Println("  delete <id>               删除任务")
	fmt.Println("  mark-done <id>            标记任务为完成")
	fmt.Println("  mark-in-progress <id>     标记任务为进行中")
	fmt.Println("  mark-todo <id>            标记任务为待办")
}
