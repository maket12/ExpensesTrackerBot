package mapper

import (
	"backend/internal/app/dto"
	"backend/internal/domain/models"
)

func ToExpenseDTO(e *models.Expense) dto.ExpenseDTO {
	if e == nil {
		return dto.ExpenseDTO{}
	}
	return dto.ExpenseDTO{
		ID:        e.ID.String(),
		UserID:    e.UserID,
		Amount:    e.Amount.Value(),
		Category:  e.Category.String(),
		SpentAt:   e.SpentAt,
		Comment:   e.Comment,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func ToExpenseDTOList(expenses []*models.Expense) []dto.ExpenseDTO {
	result := make([]dto.ExpenseDTO, len(expenses))
	for i, e := range expenses {
		result[i] = ToExpenseDTO(e)
	}
	return result
}

func ToDomainFilter(input dto.ListExpensesInput) (models.ExpenseFilter, error) {
	var cat *models.Category
	if input.Category != nil && *input.Category != "" && *input.Category != "все" {
		c, err := models.NewCategory(*input.Category)
		if err != nil {
			return models.ExpenseFilter{}, err
		}
		cat = &c
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 50
	}

	offset := input.Offset
	if offset < 0 {
		offset = 0
	}

	return models.ExpenseFilter{
		UserID:   input.UserID,
		Category: cat,
		From:     input.From,
		To:       input.To,
		Limit:    limit,
		Offset:   offset,
	}, nil
}
