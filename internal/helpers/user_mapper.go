package helpers

import (
	"time"

	"github.com/vitalii-q/selena-users-service/internal/dto"
	"github.com/vitalii-q/selena-users-service/internal/models"
	"github.com/vitalii-q/selena-users-service/internal/services/external_services"
)

// EnrichUsers заполняет country / city по ID через HotelServiceClient
func EnrichUsers(
	users []models.User,
	hotelServiceClient *external_services.HotelServiceClient,
) ([]dto.UserResponse, error) {

	result := make([]dto.UserResponse, 0, len(users))

	// Получаем locations ОДИН раз
	countries, err := hotelServiceClient.GetLocations()
	if err != nil {
		return nil, err
	}

	// Строим map-ы для O(1) lookup
	countryMap := make(map[string]*string)
	cityMap := make(map[string]*string)

	for _, country := range countries {
		countryName := country.Name // [правка] локальная переменная
		countryMap[country.ID] = &countryName

		for _, city := range country.Cities {
			cityName := city.Name // [правка]
			cityMap[city.ID] = &cityName
		}
	}

	// Основной цикл по пользователям
	for _, u := range users {

		var countryName *string
		var cityName *string

		if u.CountryID != nil {
			if name, ok := countryMap[u.CountryID.String()]; ok {
				countryName = name
			}
		}

		if u.CityID != nil {
			if name, ok := cityMap[u.CityID.String()]; ok {
				cityName = name
			}
		}

		var birthStr *string
		if u.Birth != nil {
			s := u.Birth.Format("2006-01-02")
			birthStr = &s
		}

		result = append(result, dto.UserResponse{
			ID:        u.ID,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Email:     u.Email,
			Role:      u.Role,
			Birth:     birthStr,
			Gender:    u.Gender,

			CountryID: u.CountryID,
			Country:   countryName, // null если не найдено

			CityID: u.CityID,
			City:   cityName, // null если не найдено

			CreatedAt: u.CreatedAt.Format(time.RFC3339),
			UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
		})
	}

	return result, nil
}
