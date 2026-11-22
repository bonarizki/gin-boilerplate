package models

import "time"

// Example is a sample domain row used by the DataTables demo endpoint.
type Example struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Data      string    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName is Database TableName of this model
func (e *Example) TableName() string {
	return "examples"
}
