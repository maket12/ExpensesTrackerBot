package postgres

import (
	"errors"
	"math/big"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func MapToPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func MapToPgText(text string) pgtype.Text {
	return pgtype.Text{
		String: text,
		Valid:  true,
	}
}

func MapPtrToPgText(text *string) pgtype.Text {
	if text == nil {
		return pgtype.Text{}
	}
	return MapToPgText(*text)
}

func MapToPgTimestamptz(time time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:             time,
		InfinityModifier: 0,
		Valid:            true,
	}
}

var (
	ErrNumericIsInvalid       = errors.New("numeric value is invalid")
	ErrNumericValueIsTooLarge = errors.New("numeric value is too large to store in this type")
)

func MapFloat64ToPgNumeric(f float64, precision *int) (pgtype.Numeric, error) {
	precis := 2
	if precision != nil {
		precis = *precision
	}

	var n pgtype.Numeric
	s := strconv.FormatFloat(f, 'f', precis, 64)

	err := n.Scan(s)
	if err != nil {
		return pgtype.Numeric{}, err
	}

	return n, nil
}

func MapPgNumericToFloat64(n pgtype.Numeric) (float64, error) {
	if !n.Valid {
		return 0.0, ErrNumericIsInvalid
	}
	f, _ := n.Float64Value()
	return f.Float64, nil
}

func MapInt64ToPgNumeric(val int64, exp int32) pgtype.Numeric {
	return pgtype.Numeric{
		Int:   big.NewInt(val),
		Exp:   exp,
		Valid: true,
	}
}

func MapPgNumericToInt64(n pgtype.Numeric, exp int32) (int64, error) {
	if !n.Valid {
		return 0, ErrNumericIsInvalid
	}

	acc := new(big.Int).Set(n.Int)

	diff := n.Exp - exp
	if diff > 0 {
		mul := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(diff)), nil)
		acc.Mul(acc, mul)
	} else if diff < 0 {
		div := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-diff)), nil)
		acc.Div(acc, div)
	}

	if !acc.IsInt64() {
		return 0, ErrNumericValueIsTooLarge
	}

	return acc.Int64(), nil
}
