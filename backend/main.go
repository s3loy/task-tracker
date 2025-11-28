package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/s3loy/task-tracker/backend/api"
	"github.com/s3loy/task-tracker/backend/cli"
	"github.com/s3loy/task-tracker/backend/store"
)

func main() {
	if len(os.Args) >= 2 && (os.Args[1] == "help" || os.Args[1] == "-h" || os.Args[1] == "--help") {
		cli.Run()
		return
	}
	serverMode := flag.Bool("server", false, "以 API 服务器模式运行")
	port := flag.String("port", "8080", "API 服务器端口")
	flag.Parse()

	dbUser := getEnv("DB_USER", "task_user")
	dbPass := getEnv("DB_PASSWORD", "task_pass")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbName := getEnv("DB_NAME", "task_tracker")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName)

	if err := store.InitStore(connStr); err != nil {
		fmt.Fprintf(os.Stderr, "❌ 初始化存储失败: %v\n", err)
		fmt.Println("提示: 请检查数据库是否启动，以及用户名密码是否正确。")
		os.Exit(1)
	}

	if *serverMode {
		runAPIServer(*port)
	} else {
		cli.Run()
	}
}

func runAPIServer(port string) {
	router := api.SetupRouter()

	addr := ":" + port
	fmt.Printf("🚀 Task Tracker API 服务器正在运行...\n")
	fmt.Printf("📍 地址: http://localhost%s\n", addr)
	fmt.Printf("📖 API 文档: http://localhost%s/api/health\n\n", addr)
	fmt.Println("按 Ctrl+C 停止服务器")

	if err := router.Run(addr); err != nil {
		log.Fatal("启动服务器失败:", err)
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
