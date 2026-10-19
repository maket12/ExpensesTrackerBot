package mapper

import (
	httpdto "backend/internal/adapter/in/http/dto"
	appdto "backend/internal/app/dto"
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

func ToUpdateExpenseInput(id string, req httpdto.UpdateExpenseRequest) appdto.UpdateExpenseInput {
	return appdto.UpdateExpenseInput{
		ID:            id,
		UserID:        req.UserID,
		Amount:        req.Amount,
		Category:      req.Category,
		SpentAt:       req.SpentAt,
		Comment:       req.Comment,
		EnforceUserID: false,
	}
}

func ToExpenseResponse(d appdto.ExpenseDTO) httpdto.ExpenseResponse {
	return httpdto.ExpenseResponse{
		ID:        d.ID,
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
