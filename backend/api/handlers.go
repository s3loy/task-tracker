package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/s3loy/task-tracker/backend/store"
	"github.com/s3loy/task-tracker/backend/task"

	"github.com/gin-gonic/gin"
)

func GetTasks(c *gin.Context) {
	status := c.Query("status")

	var tasks []task.Task
	var err error

	if status != "" {
		taskStatus := task.TaskStatus(status)
		if !taskStatus.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的状态参数",
			})
			return
		}
		tasks, err = store.FilterByStatus(taskStatus)
	} else {
		tasks, err = store.LoadTasks()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"count": len(tasks),
	})
}

func GetTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的任务 ID",
		})
		return
	}

	task, err := store.FindByID(id)
	if err != nil {
		if err == store.ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "任务未找到",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, task)
}

type CreateTaskRequest struct {
	Description string  `json:"description" binding:"required"`
	Deadline    *string `json:"deadline,omitempty"`
}

func CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}
	var deadline *time.Time
	if req.Deadline != nil && *req.Deadline != "" {
		parsedTime, err := time.Parse(time.RFC3339, *req.Deadline)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的 deadline 格式，应为 ISO 8601 格式",
			})
			return
		}
		if parsedTime.Before(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "UwU你是在逗我吗，过去完成现在的事情嘛",
			})
			return
		}

		deadline = &parsedTime
	}

	newTask, err := store.AddTask(req.Description, deadline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, newTask)
}

type UpdateTaskRequest struct {
	Description string  `json:"description" binding:"required"`
	Deadline    *string `json:"deadline,omitempty"`
}

func UpdateTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的任务 ID",
		})
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	var deadline *time.Time
	if req.Deadline != nil && *req.Deadline != "" {
		parsedTime, err := time.Parse(time.RFC3339, *req.Deadline)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "无效的 deadline 格式，应为 ISO 8601 格式",
			})
			return
		}

		if parsedTime.Before(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "UwU你是在逗我吗，过去完成现在的事情嘛",
			})
			return
		}

		deadline = &parsedTime
	}

	if err := store.UpdateTask(id, req.Description, deadline); err != nil {
		if err == store.ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "任务未找到",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	updatedTask, _ := store.FindByID(id)
	c.JSON(http.StatusOK, updatedTask)
}

func DeleteTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的任务 ID",
		})
		return
	}

	if err := store.DeleteTask(id); err != nil {
		if err == store.ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "任务未找到",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "任务已删除",
	})
}

type UpdateTaskStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func UpdateTaskStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的任务 ID",
		})
		return
	}

	var req UpdateTaskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	status := task.TaskStatus(req.Status)
	if !status.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的状态值",
		})
		return
	}

	if err := store.MarkTaskStatus(id, status); err != nil {
		if err == store.ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "任务未找到",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	updatedTask, _ := store.FindByID(id)
	c.JSON(http.StatusOK, updatedTask)
}

func ArchiveTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的任务 ID",
		})
		return
	}

	if err := store.SetTaskArchived(id, true); err != nil {
		if err == store.ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "任务未找到",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	updatedTask, _ := store.FindByID(id)
	c.JSON(http.StatusOK, updatedTask)
}

func UnarchiveTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的任务 ID",
		})
		return
	}

	if err := store.SetTaskArchived(id, false); err != nil {
		if err == store.ErrTaskNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "任务未找到",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	updatedTask, _ := store.FindByID(id)
	c.JSON(http.StatusOK, updatedTask)
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "task-tracker-api",
	})
}
