package common

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

const MoneyDecimalPlaces int32 = 2

func PgNumericToDecimal(pgNum pgtype.Numeric) (decimal.Decimal, error) {
	if !pgNum.Valid || pgNum.Int == nil {
		return decimal.Zero, fmt.Errorf("invalid numeric value")
	}
	return decimal.NewFromBigInt(pgNum.Int, pgNum.Exp), nil
}

func ConvertPgNumericToDecimal(pgNum pgtype.Numeric) decimal.Decimal {
	d, err := PgNumericToDecimal(pgNum)
	if err != nil {
		return decimal.Zero
	}
	return d
}

func DecimalToPgNumeric(d decimal.Decimal) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if err := n.Scan(d.StringFixed(MoneyDecimalPlaces)); err != nil {
		return pgtype.Numeric{}, fmt.Errorf("convert decimal to numeric: %w", err)
	}
	return n, nil
}

func RoundMoney(d decimal.Decimal) decimal.Decimal {
	return d.RoundBank(MoneyDecimalPlaces)
}
