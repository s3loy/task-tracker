package api

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := r.Group("/api")
	{
		api.GET("/health", HealthCheck)

		tasks := api.Group("/tasks")
		{
			tasks.GET("", GetTasks)
			tasks.GET("/:id", GetTask)
			tasks.POST("", CreateTask)
			tasks.PUT("/:id", UpdateTask)
			tasks.DELETE("/:id", DeleteTask)
			tasks.PATCH("/:id/status", UpdateTaskStatus)
			tasks.PATCH("/:id/archive", ArchiveTask)
			tasks.PATCH("/:id/unarchive", UnarchiveTask)
		}
	}

	return r
}
