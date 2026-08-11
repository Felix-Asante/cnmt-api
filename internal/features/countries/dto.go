package countries

import (
	"time"

	"cnmt/internal/infra/db"
)

type CountryResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	ISOCode   string `json:"iso_code"`
	Flag      string `json:"flag"`
	IsActive  bool   `json:"is_active"`
	CurrencyName string `json:"currency_name"`
	CurrencyCode string `json:"currency_code"`
	CurrencySymbol string `json:"currency_symbol"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func mapCountriesToResponses(countries []db.Country) []CountryResponse {
	responses := make([]CountryResponse, len(countries))
	for i, country := range countries {
		responses[i] = CountryResponse{
			ID: country.ID,
			Name: country.Name,
			ISOCode: country.IsoCode,
			Flag: country.Flag,
			IsActive: country.IsActive,
			CurrencyName: country.CurrencyName,
			CurrencyCode: country.CurrencyCode,
			CurrencySymbol: country.CurrencySymbol,
			CreatedAt: country.CreatedAt,
			UpdatedAt: country.UpdatedAt,
		}
	}
	return responses
}

func mapCountryToResponse(country db.Country) CountryResponse {
	return CountryResponse{
		ID: country.ID,
		Name: country.Name,
		ISOCode: country.IsoCode,
		Flag: country.Flag,
		IsActive: country.IsActive,
		CurrencyName: country.CurrencyName,
		CurrencyCode: country.CurrencyCode,
		CurrencySymbol: country.CurrencySymbol,
	}
}