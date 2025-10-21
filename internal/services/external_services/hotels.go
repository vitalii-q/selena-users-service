package external_services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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
	Client  *http.Client
}

// NewHotelServiceClient — конструктор клиента
func NewHotelServiceClient(baseURL string) *HotelServiceClient {
	return &HotelServiceClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetHotels — получение списка отелей
func (c *HotelServiceClient) GetHotels() ([]Hotel, error) {
	resp, err := c.Client.Get(c.BaseURL + "/hotels") // [правка] используем http.Client с таймаутом
	if err != nil {
		return nil, fmt.Errorf("failed to get hotels: %w", err) // [правка] добавляем контекст ошибки
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hotel service returned status: %d", resp.StatusCode) // [правка]
	}

	var hotels []Hotel
	if err := json.NewDecoder(resp.Body).Decode(&hotels); err != nil {
		return nil, fmt.Errorf("failed to decode hotels response: %w", err) // [правка]
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