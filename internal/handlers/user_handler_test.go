package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"selena-users-service/internal/models"
	"selena-users-service/internal/services"
)

func TestGetUserHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	mockService := new(MockUserService)
	handler := NewUserHandler(mockService)
	router.GET("/users/:id", handler.GetUserHandler)

	userID := uuid.New()
	expectedUser := &models.User{ID: userID, FirstName: "Alex", LastName: "Green"}

	mockService.On("GetUserByID", userID).Return(expectedUser, nil)

	req, _ := http.NewRequest("GET", "/users/"+userID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responseUser models.User
	json.Unmarshal(w.Body.Bytes(), &responseUser)

	assert.Equal(t, expectedUser.FirstName, responseUser.FirstName)
	mockService.AssertExpectations(t)
}
