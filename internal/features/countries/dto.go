package countries

import (
	"time"

	"cnmt/internal/common"
	"cnmt/internal/infra/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

type SourceCountryDTO struct {
	ID             int64               `json:"id"`
	Name           string              `json:"name"`
	ISOCode        string              `json:"iso_code"`
	Flag           string              `json:"flag"`
	CurrencyName   string              `json:"currency_name"`
	CurrencyCode   string              `json:"currency_code"`
	CurrencySymbol string              `json:"currency_symbol"`
	Banks          []PaymentChannelDTO `json:"banks"`
	MobileNetworks []PaymentChannelDTO `json:"mobile_networks"`
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
	DefaultExchangeRate decimal.Decimal     `json:"default_exchange_rate"`
	FeeType           db.FeeType          `json:"fee_type"`
	Fee               decimal.Decimal     `json:"fee"`
	Banks             []PaymentChannelDTO `json:"banks"`
	MobileNetworks    []PaymentChannelDTO `json:"mobile_networks"`
}


type CreatePaymentChannelRequest struct {
	Name string `json:"name" validate:"required,min=3,max=255"`
	ChannelType db.ReceivingMethods `json:"channel_type" validate:"required,oneof=BANK MOBILE_MONEY"`
}
type CreateCountryRequest struct {
	Name string `json:"name" validate:"required,min=3,max=255"`
	ISOCode string `json:"iso_code" validate:"required,min=2,max=3"`
	Flag string `json:"flag" validate:"required,min=1,max=16"`
	CurrencyName string `json:"currency_name" validate:"required"`
	CurrencyCode string `json:"currency_code" validate:"required,min=3,max=3"`
	CurrencySymbol string `json:"currency_symbol" validate:"required,min=1,max=3"`
	PaymentChannels []*CreatePaymentChannelRequest `json:"payment_channels" validate:"required,omitempty"`
}

type CreateRouteRequest struct {
	SourceCountryID   int64           `json:"source_country_id" validate:"required,gt=0"`
	DestCountryID     int64           `json:"dest_country_id" validate:"required,gt=0"`
	ExchangeRate      decimal.Decimal `json:"exchange_rate" validate:"required"`
	FeeType           db.FeeType      `json:"fee_type" validate:"required,oneof=fixed percentage"`
	Fee               decimal.Decimal `json:"fee" validate:"required"`
	MinTransferAmount decimal.Decimal `json:"min_transfer_amount" validate:"required"`
	MaxTransferAmount decimal.Decimal `json:"max_transfer_amount" validate:"required"`
}

type UpdateRouteRequest struct {
	ExchangeRate      decimal.Decimal `json:"exchange_rate" validate:"required"`
	FeeType           db.FeeType      `json:"fee_type" validate:"required,oneof=fixed percentage"`
	Fee               decimal.Decimal `json:"fee" validate:"required"`
	MinTransferAmount decimal.Decimal `json:"min_transfer_amount" validate:"required"`
	MaxTransferAmount decimal.Decimal `json:"max_transfer_amount" validate:"required"`
}

type RouteResponse struct {
	ID                   uuid.UUID       `json:"id"`
	SourceCountryID      int64           `json:"source_country_id"`
	DestinationCountryID int64           `json:"destination_country_id"`
	IsActive             bool            `json:"is_active"`
	DefaultExchangeRate  decimal.Decimal `json:"default_exchange_rate"`
	Fee                  decimal.Decimal `json:"fee"`
	FeeType              db.FeeType      `json:"fee_type"`
	MinTransferAmount    decimal.Decimal `json:"min_transfer_amount"`
	MaxTransferAmount    decimal.Decimal `json:"max_transfer_amount"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
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

func MapSourceCountries(
	rows []db.GetAllSourceCountriesRow,
	channelsByCountry map[int64][]db.GetActivePaymentChannelsByCountryIDsRow,
) []SourceCountryDTO {
	out := make([]SourceCountryDTO, len(rows))
	for i, row := range rows {
		banks, networks := groupPaymentChannels(channelsByCountry[row.ID])
		out[i] = SourceCountryDTO{
			ID:             row.ID,
			Name:           row.Name,
			ISOCode:        row.IsoCode,
			Flag:           row.Flag,
			CurrencyName:   row.CurrencyName,
			CurrencyCode:   row.CurrencyCode,
			CurrencySymbol: row.CurrencySymbol,
			Banks:          banks,
			MobileNetworks: networks,
		}
	}
	return out
}

func mapDestCountriesToResponses(
	rows []db.GetDestCountriesBySrcCountryIDRow,
	channelsByCountry map[int64][]db.GetActivePaymentChannelsByCountryIDsRow,
) ([]DestCountryResponse, error) {
	responses := make([]DestCountryResponse, 0, len(rows))
	for _, row := range rows {
		resp, err := MapDestCountryFromSrcRow(row, channelsByCountry[row.ID])
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}
	return responses, nil
}

func MapDestCountryFromSrcRow(
	row db.GetDestCountriesBySrcCountryIDRow,
	channels []db.GetActivePaymentChannelsByCountryIDsRow,
) (DestCountryResponse, error) {
	return mapDestCountry(
		row.ID,
		row.Name,
		row.IsoCode,
		row.Flag,
		row.CurrencyName,
		row.CurrencyCode,
		row.CurrencySymbol,
		row.MinTransferAmount,
		row.MaxTransferAmount,
		row.DefaultExchangeRate,
		row.FeeType,
		row.Fee,
		channels,
	)
}

func MapDestCountryFromRouteRow(
	row db.GetAllActiveRouteDestinationsRow,
	channels []db.GetActivePaymentChannelsByCountryIDsRow,
) (DestCountryResponse, error) {
	return mapDestCountry(
		row.ID,
		row.Name,
		row.IsoCode,
		row.Flag,
		row.CurrencyName,
		row.CurrencyCode,
		row.CurrencySymbol,
		row.MinTransferAmount,
		row.MaxTransferAmount,
		row.DefaultExchangeRate,
		row.FeeType,
		row.Fee,
		channels,
	)
}

func mapDestCountry(
	id int64,
	name, isoCode, flag, currencyName, currencyCode, currencySymbol string,
	minAmountRaw, maxAmountRaw pgtype.Numeric,
	defaultExchangeRateRaw pgtype.Numeric,
	feeTypeRaw db.FeeType,
	feeRaw pgtype.Numeric,
	channels []db.GetActivePaymentChannelsByCountryIDsRow,
) (DestCountryResponse, error) {
	minAmount, err := common.PgNumericToDecimal(minAmountRaw)
	if err != nil {
		minAmount = decimal.Zero
	}
	maxAmount, err := common.PgNumericToDecimal(maxAmountRaw)
	if err != nil {
		maxAmount = decimal.Zero
	}

	banks, networks := groupPaymentChannels(channels)

	defaultExchangeRate, err := common.PgNumericToDecimal(defaultExchangeRateRaw)
	if err != nil {
		defaultExchangeRate = decimal.Zero
	}
	fee, err := common.PgNumericToDecimal(feeRaw)
	if err != nil {
		fee = decimal.Zero
	}
	return DestCountryResponse{
		ID:                id,
		Name:              name,
		ISOCode:           isoCode,
		Flag:              flag,
		CurrencyName:      currencyName,
		CurrencyCode:      currencyCode,
		CurrencySymbol:    currencySymbol,
		MinTransferAmount: minAmount,
		MaxTransferAmount: maxAmount,
		DefaultExchangeRate: defaultExchangeRate,
		FeeType:           feeTypeRaw,
		Fee:               fee,
		Banks:             banks,
		MobileNetworks:    networks,
	}, nil
}

func mapRouteToResponse(route db.Route) (RouteResponse, error) {
	var rate, fee, minAmount, maxAmount decimal.Decimal
	var err error
	
	if route.DefaultExchangeRate.Valid {
		rate, err = common.PgNumericToDecimal(route.DefaultExchangeRate)
		if err != nil {
			return RouteResponse{}, err
		}
	}
	if route.Fee.Valid {
		fee, err = common.PgNumericToDecimal(route.Fee)
		if err != nil {
			return RouteResponse{}, err
		}
	}
	if route.MinTransferAmount.Valid {
		minAmount, err = common.PgNumericToDecimal(route.MinTransferAmount)
		if err != nil {
			return RouteResponse{}, err
		}
	}
	if route.MaxTransferAmount.Valid {
		maxAmount, err = common.PgNumericToDecimal(route.MaxTransferAmount)
		if err != nil {
			return RouteResponse{}, err
		}
	}

	return RouteResponse{
		ID:                   route.ID,
		SourceCountryID:      route.SourceCountryID,
		DestinationCountryID: route.DestinationCountryID,
		IsActive:             route.IsActive,
		DefaultExchangeRate:  rate,
		Fee:                  fee,
		FeeType:              route.FeeType,
		MinTransferAmount:    minAmount,
		MaxTransferAmount:    maxAmount,
		CreatedAt:            route.CreatedAt,
		UpdatedAt:            route.UpdatedAt,
	}, nil
}

func groupPaymentChannels(channels []db.GetActivePaymentChannelsByCountryIDsRow) (banks, networks []PaymentChannelDTO) {
	banks = make([]PaymentChannelDTO, 0)
	networks = make([]PaymentChannelDTO, 0)
	for _, ch := range channels {
		dto := PaymentChannelDTO{ID: ch.ID, Name: ch.Name}
		switch ch.ChannelType {
		case db.ReceivingMethodsBANK:
			banks = append(banks, dto)
		case db.ReceivingMethodsMOBILEMONEY:
			networks = append(networks, dto)
		}
	}
	return banks, networks
}
