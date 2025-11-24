package controllers

import (
	"net/http"
	"time"
	"uptime-monitor/repository"
	"uptime-monitor/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LogController struct {
	repo *repository.LogRepository
}

func NewLogController(repo *repository.LogRepository) *LogController {
	return &LogController{repo: repo}
}

func (c *LogController) CreateLog(ctx *gin.Context) {
	var log types.Log
	if err := ctx.ShouldBindJSON(&log); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.ID = uuid.New().String()
	log.CreatedAt = time.Now()

	if err := c.repo.CreateLog(&log); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create log"})
		return
	}

	ctx.JSON(http.StatusCreated, log)
}

func (c *LogController) GetLogsByMonitor(ctx *gin.Context) {
	monitorID := ctx.Param("monitorId")
	logs, err := c.repo.GetLogsByMonitorID(monitorID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch logs"})
		return
	}
	ctx.JSON(http.StatusOK, logs)
}
