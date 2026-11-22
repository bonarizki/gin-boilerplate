package controllers

import (
	"github.com/bonarizki/gin-boilerplate/internal/app/services"
	"github.com/bonarizki/gin-boilerplate/pkg/utils"
	"github.com/gin-gonic/gin"
)

// HealthController handles health check and metrics endpoints.
type HealthController struct {
	service *services.HealthService
}

// NewHealthController creates a new HealthController instance.
func NewHealthController(service *services.HealthService) *HealthController {
	return &HealthController{
		service: service,
	}
}

// Health performs application health check.
//
// GET /health
// Returns health status of the application and its dependencies.
func (ctrl *HealthController) Health(c *gin.Context) {
	ctx := c.Request.Context()

	response := ctrl.service.CheckHealth(ctx)

	if response.Status == "unhealthy" {
		utils.ServiceUnavailableWithData(c, response, "Service is unhealthy")
		return
	}

	utils.Ok(c, response, "Service is healthy")
}

// Metrics returns application metrics.
//
// GET /metrics
// Returns basic request counters and uptime statistics.
func (ctrl *HealthController) Metrics(c *gin.Context) {
	ctx := c.Request.Context()

	response := ctrl.service.GetMetrics(ctx)
	utils.Ok(c, response, "Metrics retrieved successfully")
}
