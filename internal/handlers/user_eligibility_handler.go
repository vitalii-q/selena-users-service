package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/vitalii-q/selena-users-service/internal/services"
)

type UserEligibilityHandler struct {
	service services.BookingEligibilityService
}

func NewUserEligibilityHandler(service services.BookingEligibilityService) *UserEligibilityHandler {
	return &UserEligibilityHandler{service: service}
}

func (h *UserEligibilityHandler) GetBookingEligibility(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		writeUserError(c, http.StatusBadRequest, "INVALID_UUID", "Invalid user ID")
		return
	}
	eligibility, err := h.service.GetBookingEligibility(id)
	if err != nil {
		logrus.WithError(err).Error("failed to check user booking eligibility")
		writeUserError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}
	c.JSON(http.StatusOK, gin.H{"exists": eligibility.Exists, "eligible": eligibility.Eligible})
}
