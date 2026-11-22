package repositories

import (
	"github.com/bonarizki-dat/Datatables-Gin/datatables"
	"github.com/bonarizki-dat/Datatables-Gin/datatables/dto"
	"github.com/bonarizki/gin-boilerplate/internal/adapters/database"
	"github.com/bonarizki/gin-boilerplate/internal/domain/models"

	"github.com/gin-gonic/gin"
)

func GetDataDatatables(c *gin.Context) (dto.Datatables, error) {
	query := database.GetDB().Model(&models.Example{})
	var example []*models.Example
	result, err := datatables.OfReturn(
		c,
		query,
		&example,
		[]string{"id", "data"},
		map[string]string{ // orderable
			"id":   "id",
			"data": "data",
		},
		datatables.NewOptions().
			WithIndex("DT_RowIndex", false), // global index

	)

	if err != nil {
		return dto.Datatables{}, err
	}

	return result, nil
}
