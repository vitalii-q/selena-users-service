package handlers

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/vitalii-q/selena-users-service/internal/services"
)

type eligibilityServiceStub struct {
	result services.BookingEligibility
	err    error
}

func (s *eligibilityServiceStub) GetBookingEligibility(uuid.UUID) (services.BookingEligibility, error) {
	return s.result, s.err
}

func TestUserEligibilityHandlerEligibleUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewUserEligibilityHandler(&eligibilityServiceStub{result: services.BookingEligibility{Exists: true, Eligible: true}})
	r := gin.New()
	r.GET("/internal/users/:id/booking-eligibility", handler.GetBookingEligibility)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/internal/users/"+uuid.New().String()+"/booking-eligibility", nil))
	require.Equal(t, 200, w.Code)
	require.JSONEq(t, `{"exists":true,"eligible":true}`, w.Body.String())
}

func TestUserEligibilityHandlerDeletedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewUserEligibilityHandler(&eligibilityServiceStub{result: services.BookingEligibility{Exists: true, Eligible: false}})
	r := gin.New()
	r.GET("/internal/users/:id/booking-eligibility", handler.GetBookingEligibility)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/internal/users/"+uuid.New().String()+"/booking-eligibility", nil))
	require.Equal(t, 200, w.Code)
	require.JSONEq(t, `{"exists":true,"eligible":false}`, w.Body.String())
}

func TestUserEligibilityHandlerInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewUserEligibilityHandler(&eligibilityServiceStub{})
	r := gin.New()
	r.GET("/internal/users/:id/booking-eligibility", handler.GetBookingEligibility)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/internal/users/invalid/booking-eligibility", nil))
	require.Equal(t, 400, w.Code)
}

func TestUserEligibilityHandlerInternalErrorIsSafe(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewUserEligibilityHandler(&eligibilityServiceStub{err: errors.New("database connection details")})
	r := gin.New()
	r.GET("/internal/users/:id/booking-eligibility", handler.GetBookingEligibility)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/internal/users/"+uuid.New().String()+"/booking-eligibility", nil))
	require.Equal(t, 500, w.Code)
	require.NotContains(t, w.Body.String(), "database connection details")
}
