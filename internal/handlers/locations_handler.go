package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/vitalii-q/selena-users-service/internal/services/external_services"
)

// LocationsHandler handles requests for locations data
type LocationsHandler struct {
	hotelClient *external_services.HotelServiceClient
}

// NewLocationsHandler creates new handler
func NewLocationsHandler(hotelClient *external_services.HotelServiceClient) *LocationsHandler {
	return &LocationsHandler{
		hotelClient: hotelClient,
	}
}

// GetLocationsHandler returns list of countries and cities
// GET /api/v1/locations
func (h *LocationsHandler) GetLocationsHandler(c *gin.Context) {

	locations, err := h.hotelClient.GetLocations()
	if err != nil {
		logrus.WithError(err).Error("failed to get locations")
		writeUserError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
		return
	}

	c.JSON(http.StatusOK, locations)
}
