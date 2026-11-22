package services_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/bonarizki/gin-boilerplate/internal/app/services"
	"github.com/bonarizki/gin-datatables/v2/datatables/dto"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type stubExampleRepo struct {
	result dto.Datatables
	err    error
}

func (s stubExampleRepo) Datatables(*gin.Context) (dto.Datatables, error) {
	return s.result, s.err
}

func TestExampleService_GetDataDatatables(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/datatables", nil)

	want := dto.Datatables{Draw: 3, RecordsTotal: 10}
	got, err := services.NewExampleService(stubExampleRepo{result: want}).GetDataDatatables(context.Background(), c)
	assert.NoError(t, err)
	assert.Equal(t, want, got)

	boom := errors.New("boom")
	_, err = services.NewExampleService(stubExampleRepo{err: boom}).GetDataDatatables(context.Background(), c)
	assert.ErrorIs(t, err, boom)
}
