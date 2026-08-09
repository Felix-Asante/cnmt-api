package common

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

func ConvertPgNumericToDecimal(pgNum pgtype.Numeric) (decimal.Decimal) {
	
    if !pgNum.Valid {
        return decimal.Zero
    }

    return decimal.NewFromBigInt(pgNum.Int, pgNum.Exp)
}

