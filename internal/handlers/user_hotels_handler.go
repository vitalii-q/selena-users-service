package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

	hotels, err := h.hotelClient.GetHotels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	hotels = h.hotelClient.ApplyBusinessLogic(userId, hotels)
	c.JSON(http.StatusOK, hotels)
}
