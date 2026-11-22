package seeders

import (
	"context"
	"fmt"

	"github.com/bonarizki/gin-boilerplate/internal/domain/models"
	"github.com/bonarizki/gin-boilerplate/pkg/logger"
	"gorm.io/gorm"
)

// demoExamples seed rows for the example CRUD/datatables endpoints.
var demoExamples = []string{
	"Example row 1",
	"Example row 2",
}

func seedDemoExamples(ctx context.Context, db *gorm.DB) error {
	db = db.WithContext(ctx)
	var count int64
	if err := db.Model(&models.Example{}).Count(&count).Error; err != nil {
		return fmt.Errorf("count examples: %w", err)
	}
	if count > 0 {
		return nil
	}

	rows := make([]models.Example, 0, len(demoExamples))
	for _, data := range demoExamples {
		rows = append(rows, models.Example{Data: data})
	}
	if err := db.Create(&rows).Error; err != nil {
		return fmt.Errorf("create examples: %w", err)
	}

	logger.Infof("Seeded %d demo example rows", len(rows))
	return nil
}
