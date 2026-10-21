package mapper

import (
	"strings"

	"backend/internal/app/dto"
	"backend/internal/domain/models"
	"backend/internal/domain/port"
	pkgerrs "backend/pkg/errs"
)

func ToExpenseDTO(e *models.Expense) dto.ExpenseDTO {
	if e == nil {
		return dto.ExpenseDTO{}
	}
	return dto.ExpenseDTO{
		ID:        e.ID().String(),
		UserID:    e.UserID(),
		Amount:    e.Amount(),
		Category:  e.Category().String(),
		SpentAt:   e.SpentAt(),
		Comment:   e.Comment(),
		CreatedAt: e.CreatedAt(),
		UpdatedAt: e.UpdatedAt(),
	}
}

func ToExpenseDTOList(expenses []*models.Expense) []dto.ExpenseDTO {
	result := make([]dto.ExpenseDTO, len(expenses))
	for i, e := range expenses {
		result[i] = ToExpenseDTO(e)
	}
	return result
}

func ToCreateExpenseOutput(expense *models.Expense) dto.CreateExpenseOutput {
	return dto.CreateExpenseOutput{
		Expense: ToExpenseDTO(expense),
	}
}

func ToGetExpenseOutput(expense *models.Expense) dto.GetExpenseOutput {
	return dto.GetExpenseOutput{
		Expense: ToExpenseDTO(expense),
	}
}

func ToUpdateExpenseOutput(expense *models.Expense) dto.UpdateExpenseOutput {
	return dto.UpdateExpenseOutput{
		Expense: ToExpenseDTO(expense),
	}
}

func ToDeleteExpenseOutput() dto.DeleteExpenseOutput {
	return dto.DeleteExpenseOutput{
		Success: true,
	}
}

func ToListExpensesOutput(expenses []*models.Expense) dto.ListExpensesOutput {
	return dto.ListExpensesOutput{
		Expenses: ToExpenseDTOList(expenses),
	}
}

func ToGetDistinctUsersOutput(userIDs []int64) dto.GetDistinctUsersOutput {
	return dto.GetDistinctUsersOutput{
		UserIDs: userIDs,
	}
}

func ToHandleAgentMessageOutput(reply string) dto.HandleAgentMessageOutput {
	return dto.HandleAgentMessageOutput{
		ResponseText: reply,
	}
}

func ToDomainFilter(input dto.ListExpensesInput) (port.ExpenseFilter, error) {
	var cat *models.Category
	if input.Category != nil && *input.Category != "" && *input.Category != "все" {
		c := models.Category(strings.TrimSpace(*input.Category))
		if !c.IsValid() {
			return port.ExpenseFilter{}, pkgerrs.NewValueInvalidError("category")
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

	return port.ExpenseFilter{
		UserID:   input.UserID,
		Category: cat,
		From:     input.From,
		To:       input.To,
		Limit:    limit,
		Offset:   offset,
	}, nil
}
