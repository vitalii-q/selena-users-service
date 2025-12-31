package helpers

import (
	"time"

	"github.com/vitalii-q/selena-users-service/internal/dto"
	"github.com/vitalii-q/selena-users-service/internal/models"
	"github.com/vitalii-q/selena-users-service/internal/services/external_services"
)

// EnrichUsers заполняет country/city по ID через LocationsClient
func EnrichUsers(
	users []models.User,
	HotelServiceClient *external_services.HotelServiceClient,
) ([]dto.UserResponse, error) {

	result := make([]dto.UserResponse, 0, len(users))

	// Получаем все страны с городами одним запросом (best practice)
	countries, err := HotelServiceClient.GetLocations()
	if err != nil {
		return nil, err
	}

	for _, u := range users {
		var countryName *string
		var cityName *string

		// Ищем страну по country_id
		if u.CountryID != nil {
			for _, c := range countries {
				if c.ID == u.CountryID.String() {

					// [правка] берём адрес строки
					cn := c.Name
					countryName = &cn

					// 3️⃣ Ищем город внутри найденной страны
					if u.CityID != nil {
						for _, city := range c.Cities {
							if city.ID == u.CityID.String() {
								ct := city.Name
								cityName = &ct
								break
							}
						}
					}
					break
				}
			}
		}

		// Форматируем birth
		var birthStr *string
		if u.Birth != nil {
			s := u.Birth.Format("2006-01-02")
			birthStr = &s
		}

		// Собираем DTO
		result = append(result, dto.UserResponse{
			ID:        u.ID,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Email:     u.Email,
			Role:      u.Role,
			Birth:     birthStr,
			Gender:    u.Gender,

			CountryID: u.CountryID,
			Country:   countryName,
			CityID:    u.CityID,
			City:      cityName,

			CreatedAt: u.CreatedAt.Format(time.RFC3339),
			UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
		})
	}

	return result, nil
}
