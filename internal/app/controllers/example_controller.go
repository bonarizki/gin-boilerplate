package controllers

import (
	"github.com/bonarizki/gin-datatables/v2/datatables"
	"github.com/bonarizki/gin-boilerplate/internal/app/services"
	"github.com/bonarizki/gin-boilerplate/pkg/utils"
	"github.com/gin-gonic/gin"
)

// ExampleController handles example/datatables-related HTTP requests.
type ExampleController struct {
	service *services.ExampleService
}

// NewExampleController creates a new ExampleController instance.
func NewExampleController(service *services.ExampleService) *ExampleController {
	return &ExampleController{
		service: service,
	}
}

// GetDataDatatables handles GET for DataTables server-side format.
//
// GET /datatables
// Request: DataTables query params (e.g. draw, start, length, order, columns).
// Response: DataTables JSON format (via datatables library).
func (ctrl *ExampleController) GetDataDatatables(c *gin.Context) {
	ctx := c.Request.Context()

	data, err := ctrl.service.GetDataDatatables(ctx, c)
	if err != nil {
		utils.InternalServerError(c, err, "Failed to retrieve data")
		return
	}
	datatables.JSON(c, data)
}
