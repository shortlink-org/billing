package dto

import "google.golang.org/genproto/googleapis/type/money"

// MoneyDTO represents a monetary amount.
type MoneyDTO struct {
	CurrencyCode string `json:"currency_code" validate:"required,len=3"`
	Units        int64  `json:"units"`
	Nanos        int32  `json:"nanos" validate:"gte=-999999999,lte=999999999"`
}

// ToMoney converts MoneyDTO to domain money type.
func (dto *MoneyDTO) ToMoney() *money.Money {
	if dto == nil {
		return nil
	}
	return &money.Money{
		CurrencyCode: dto.CurrencyCode,
		Units:        dto.Units,
		Nanos:        dto.Nanos,
	}
}

// FromMoney converts domain money type to MoneyDTO.
func FromMoney(m *money.Money) *MoneyDTO {
	if m == nil {
		return nil
	}
	return &MoneyDTO{
		CurrencyCode: m.GetCurrencyCode(),
		Units:        m.GetUnits(),
		Nanos:        m.GetNanos(),
	}
}