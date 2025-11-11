# Task Tracker CLI 

```
backend/
├── main.go              # 程序入口
├── go.mod               # Go 模块定义
├── task-cli             # 编译后的可执行文件
├── tasks.json           # 任务数据存储文件（自动生成）
├── cli/
│   └── cli.go          # CLI 命令解析和处理
├── store/
│   └── file_store.go   # 数据持久化层（JSON 文件操作）
└── task/
    └── task.go         # 任务数据模型和业务逻辑
```
