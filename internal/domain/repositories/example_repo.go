package repositories

import (
	"github.com/bonarizki/gin-boilerplate/internal/domain/models"
	"github.com/bonarizki/gin-datatables/v2/datatables"
	"github.com/bonarizki/gin-datatables/v2/datatables/dto"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ExampleRepository serves the examples table in DataTables format.
type ExampleRepository interface {
	// Datatables takes *gin.Context because the gin-datatables library parses
	// the DataTables query parameters from it; the query uses its request context.
	Datatables(c *gin.Context) (dto.Datatables, error)
}

type exampleRepo struct {
	db *gorm.DB
}

// NewExampleRepository returns an ExampleRepository backed by db.
func NewExampleRepository(db *gorm.DB) ExampleRepository {
	return &exampleRepo{db: db}
}

func (r *exampleRepo) Datatables(c *gin.Context) (dto.Datatables, error) {
	query := r.db.WithContext(c.Request.Context()).Model(&models.Example{})
	var examples []*models.Example
	return datatables.OfReturn(
		c,
		query,
		&examples,
		[]string{"id", "data"},
		map[string]string{
			"id":   "id",
			"data": "data",
		},
		datatables.NewOptions().WithIndex("DT_RowIndex", false),
	)
}
