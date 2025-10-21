package external_services

import (
	"encoding/json"
	"net/http"
)

// Hotel — структура данных от HotelService
type Hotel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Address     string `json:"address,omitempty"`
	City        string `json:"city"`
	Country     string `json:"country,omitempty"`
}

// HotelServiceClient — клиент для работы с HotelService
type HotelServiceClient struct {
	BaseURL string
}

// NewHotelServiceClient — конструктор клиента
func NewHotelServiceClient(baseURL string) *HotelServiceClient {
	return &HotelServiceClient{BaseURL: baseURL}
}

// GetHotels — получение списка отелей
func (c *HotelServiceClient) GetHotels() ([]Hotel, error) {
	resp, err := http.Get(c.BaseURL + "/hotels")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var hotels []Hotel
	if err := json.NewDecoder(resp.Body).Decode(&hotels); err != nil {
		return nil, err
	}
	return hotels, nil
}

// ApplyBusinessLogic — фильтр отелей для пользователя
func (c *HotelServiceClient) ApplyBusinessLogic(userId string, hotels []Hotel) []Hotel {
	var filtered []Hotel
	for _, h := range hotels {
		if h.City == "Augsburg" { // пример фильтра
			filtered = append(filtered, h)
		}
	}
	return filtered
}
