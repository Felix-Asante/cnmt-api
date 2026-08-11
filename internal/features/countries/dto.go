package countries

import (
	"time"

	"cnmt/internal/common"
	"cnmt/internal/infra/db"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CountryResponse struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	ISOCode        string    `json:"iso_code"`
	Flag           string    `json:"flag"`
	IsActive       bool      `json:"is_active"`
	CurrencyName   string    `json:"currency_name"`
	CurrencyCode   string    `json:"currency_code"`
	CurrencySymbol string    `json:"currency_symbol"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type PaymentChannelDTO struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type DestCountryResponse struct {
	ID                int64               `json:"id"`
	Name              string              `json:"name"`
	ISOCode           string              `json:"iso_code"`
	Flag              string              `json:"flag"`
	CurrencyName      string              `json:"currency_name"`
	CurrencyCode      string              `json:"currency_code"`
	CurrencySymbol    string              `json:"currency_symbol"`
	MinTransferAmount decimal.Decimal     `json:"min_transfer_amount"`
	MaxTransferAmount decimal.Decimal     `json:"max_transfer_amount"`
	Banks             []PaymentChannelDTO `json:"banks"`
	MobileNetworks    []PaymentChannelDTO `json:"mobile_networks"`
}

func mapCountriesToResponses(countries []db.Country) []CountryResponse {
	responses := make([]CountryResponse, len(countries))
	for i, country := range countries {
		responses[i] = mapCountryToResponse(country)
	}
	return responses
}

func mapCountryToResponse(country db.Country) CountryResponse {
	return CountryResponse{
		ID:             country.ID,
		Name:           country.Name,
		ISOCode:        country.IsoCode,
		Flag:           country.Flag,
		IsActive:       country.IsActive,
		CurrencyName:   country.CurrencyName,
		CurrencyCode:   country.CurrencyCode,
		CurrencySymbol: country.CurrencySymbol,
		CreatedAt:      country.CreatedAt,
		UpdatedAt:      country.UpdatedAt,
	}
}

func mapDestCountriesToResponses(
	rows []db.GetDestCountriesBySrcCountryIDRow,
	channelsByCountry map[int64][]db.GetActivePaymentChannelsByCountryIDsRow,
) ([]DestCountryResponse, error) {
	responses := make([]DestCountryResponse, 0, len(rows))
	for _, row := range rows {
		resp, err := mapDestCountryToResponse(row, channelsByCountry[row.ID])
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}
	return responses, nil
}

func mapDestCountryToResponse(
	row db.GetDestCountriesBySrcCountryIDRow,
	channels []db.GetActivePaymentChannelsByCountryIDsRow,
) (DestCountryResponse, error) {
	minAmount, err := common.PgNumericToDecimal(row.MinTransferAmount)
	if err != nil {
		minAmount = decimal.Zero
	}
	maxAmount, err := common.PgNumericToDecimal(row.MaxTransferAmount)
	if err != nil {
		maxAmount = decimal.Zero
	}

	banks := make([]PaymentChannelDTO, 0)
	networks := make([]PaymentChannelDTO, 0)
	for _, ch := range channels {
		dto := PaymentChannelDTO{ID: ch.ID, Name: ch.Name}
		switch ch.ChannelType {
		case db.ReceivingMethodsBANK:
			banks = append(banks, dto)
		case db.ReceivingMethodsMOBILEMONEY:
			networks = append(networks, dto)
		}
	}

	return DestCountryResponse{
		ID:                row.ID,
		Name:              row.Name,
		ISOCode:           row.IsoCode,
		Flag:              row.Flag,
		CurrencyName:      row.CurrencyName,
		CurrencyCode:      row.CurrencyCode,
		CurrencySymbol:    row.CurrencySymbol,
		MinTransferAmount: minAmount,
		MaxTransferAmount: maxAmount,
		Banks:             banks,
		MobileNetworks:    networks,
	}, nil
}
