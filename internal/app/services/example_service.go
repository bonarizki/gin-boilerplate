package services

import (
	"context"

	"github.com/bonarizki/gin-boilerplate/internal/domain/repositories"
	"github.com/bonarizki/gin-datatables/v2/datatables/dto"
	"github.com/gin-gonic/gin"
)

// ExampleService handles example/datatables-related business logic.
type ExampleService struct {
	repo repositories.ExampleRepository
}

// NewExampleService creates a new ExampleService instance.
func NewExampleService(repo repositories.ExampleRepository) *ExampleService {
	return &ExampleService{repo: repo}
}

// GetDataDatatables returns data in DataTables format for server-side processing.
//
// Expects DataTables request query params from gin.Context; returns datatables result or error.
func (s *ExampleService) GetDataDatatables(ctx context.Context, c *gin.Context) (dto.Datatables, error) {
	data, err := s.repo.Datatables(c)
	return data, err
}
