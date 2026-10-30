package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/vitalii-q/selena-users-service/internal/services/external_services"
)

type UserHotelsHandler struct {
	hotelClient *external_services.HotelServiceClient
}

func NewUserHotelsHandler(hotelClient *external_services.HotelServiceClient) *UserHotelsHandler {
	return &UserHotelsHandler{hotelClient: hotelClient}
}

// GET /users/:userId/hotels
func (h *UserHotelsHandler) GetUserHotelsHandler(c *gin.Context) {
	userId := c.Param("id")
	if _, err := uuid.Parse(userId); err != nil {
		writeUserError(c, http.StatusBadRequest, "INVALID_UUID", "Invalid user ID")
		return
	}

	hotels, err := h.hotelClient.GetHotels()
	if err != nil {
		logrus.WithError(err).Error("failed to get user hotels")
		writeUserError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	hotels = h.hotelClient.ApplyBusinessLogic(userId, hotels)
	c.JSON(http.StatusOK, hotels)
}
