package mapper

import (
	httpdto "backend/internal/adapter/in/http/dto"
	appdto "backend/internal/app/dto"
	pkgerrs "backend/pkg/errs"

	"github.com/google/uuid"
)

func ToCreateExpenseInput(req httpdto.CreateExpenseRequest) appdto.CreateExpenseInput {
	return appdto.CreateExpenseInput{
		UserID:   req.UserID,
		Amount:   req.Amount,
		Category: req.Category,
		SpentAt:  req.SpentAt,
		Comment:  req.Comment,
	}
}

func ToGetExpenseInput(idStr string) (appdto.GetExpenseInput, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return appdto.GetExpenseInput{}, pkgerrs.NewValueInvalidError("id")
	}
	return appdto.GetExpenseInput{
		ID: id,
	}, nil
}

func ToUpdateExpenseInput(idStr string, req httpdto.UpdateExpenseRequest) (appdto.UpdateExpenseInput, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return appdto.UpdateExpenseInput{}, pkgerrs.NewValueInvalidError("id")
	}
	return appdto.UpdateExpenseInput{
		ID:            id,
		UserID:        req.UserID,
		Amount:        req.Amount,
		Category:      req.Category,
		SpentAt:       req.SpentAt,
		Comment:       req.Comment,
		EnforceUserID: false,
	}, nil
}

func ToDeleteExpenseInput(idStr string) (appdto.DeleteExpenseInput, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return appdto.DeleteExpenseInput{}, pkgerrs.NewValueInvalidError("id")
	}
	return appdto.DeleteExpenseInput{
		ID:            id,
		EnforceUserID: false,
	}, nil
}

func ToExpenseResponse(d appdto.ExpenseDTO) httpdto.ExpenseResponse {
	return httpdto.ExpenseResponse{
		ID:        d.ID.String(),
		UserID:    d.UserID,
		Amount:    d.Amount,
		Category:  d.Category,
		SpentAt:   d.SpentAt,
		Comment:   d.Comment,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func ToExpenseResponses(dtos []appdto.ExpenseDTO) []httpdto.ExpenseResponse {
	result := make([]httpdto.ExpenseResponse, len(dtos))
	for i, d := range dtos {
		result[i] = ToExpenseResponse(d)
	}
	return result
}
