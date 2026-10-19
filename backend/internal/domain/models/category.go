package models

import (
	"strings"

	"backend/internal/domain/errs"
)

type Category string

const (
	CategoryFood           Category = "еда"
	CategoryTransport      Category = "транспорт"
	CategoryHousing        Category = "жильё"
	CategoryEntertainment  Category = "развлечения"
	CategoryHealth         Category = "здоровье"
	CategoryShopping       Category = "покупки"
	CategoryOther          Category = "другое"
)

var validCategories = map[Category]struct{}{
	CategoryFood:          {},
	CategoryTransport:     {},
	CategoryHousing:       {},
	CategoryEntertainment: {},
	CategoryHealth:        {},
	CategoryShopping:      {},
	CategoryOther:         {},
}

func NewCategory(val string) (Category, error) {
	cat := Category(strings.ToLower(strings.TrimSpace(val)))
	if _, ok := validCategories[cat]; !ok {
		return "", errs.ErrInvalidCategory
	}
	return cat, nil
}

func (c Category) String() string {
	return string(c)
}

func (c Category) IsValid() bool {
	_, ok := validCategories[c]
	return ok
}

func AllCategories() []Category {
	return []Category{
		CategoryFood,
		CategoryTransport,
		CategoryHousing,
		CategoryEntertainment,
		CategoryHealth,
		CategoryShopping,
		CategoryOther,
	}
}
