package models

import (
	"fmt"
	"math"

	"backend/internal/domain/errs"
)

type Amount struct {
	value float64
}

func NewAmount(val float64) (Amount, error) {
	rounded := math.Round(val*100) / 100
	if rounded <= 0 {
		return Amount{}, errs.ErrInvalidAmount
	}
	return Amount{value: rounded}, nil
}

func (a Amount) Value() float64 {
	return a.value
}

func (a Amount) String() string {
	return fmt.Sprintf("%.2f", a.value)
}
