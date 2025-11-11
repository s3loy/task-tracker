.DEFAULT_GOAL := help
BACKEND_DIR := ./backend
BINARY_NAME := task-cli
build:
	@echo "编译项目 ..."
	@go build -o $(BINARY_NAME) -v $(BACKEND_DIR)
run: build
	@echo "运行程序..."
	@./$(BINARY_NAME) $(ARGS)
clean:
	@echo "清理二进制文件..."
	@rm -f $(BINARY_NAME)
demo: build
	@echo "📺 运行演示..."
	@./$(BINARY_NAME) add "学习 Go 语言"
	@./$(BINARY_NAME) add "完成项目文档"
	@./$(BINARY_NAME) add "编写单元测试"
	@./$(BINARY_NAME) mark-in-progress 1
	@./$(BINARY_NAME) mark-done 2
	@echo ""
	@./$(BINARY_NAME) list
	@make clean
help:
	@echo "Task Tracker - Makefile"
	@echo ""
	@echo "用法: make [命令]"
	@echo ""
	@echo "可用命令:"
	@echo "  build         编译 backend 目录中的 Go 应用程序"
	@echo "  run           编译并运行 (使用 ARGS=... 传递参数)"
	@echo "  clean         清理编译生成的文件"
	@echo "  demo          运行一个预设的演示流程"
	@echo "  help          显示此帮助信息"