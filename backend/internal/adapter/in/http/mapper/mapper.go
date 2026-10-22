package mapper

import (
	"time"

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

func ToCreateExpenseResponse(d appdto.ExpenseDTO) httpdto.CreateExpenseResponse {
	return httpdto.CreateExpenseResponse{
		Expense: ToExpense(d),
	}
}

func ToGetExpenseInput(req httpdto.GetExpenseRequest) (appdto.GetExpenseInput, error) {
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return appdto.GetExpenseInput{}, pkgerrs.NewValueInvalidError("id")
	}
	return appdto.GetExpenseInput{
		ID: id,
	}, nil
}

func ToGetExpenseResponse(d appdto.ExpenseDTO) httpdto.GetExpenseResponse {
	return httpdto.GetExpenseResponse{
		Expense: ToExpense(d),
	}
}

func ToUpdateExpenseInput(req httpdto.UpdateExpenseRequest) (appdto.UpdateExpenseInput, error) {
	id, err := uuid.Parse(req.ID)
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

func ToUpdateExpenseResponse(d appdto.ExpenseDTO) httpdto.UpdateExpenseResponse {
	return httpdto.UpdateExpenseResponse{
		Expense: ToExpense(d),
	}
}

func ToDeleteExpenseInput(req httpdto.DeleteExpenseRequest) (appdto.DeleteExpenseInput, error) {
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return appdto.DeleteExpenseInput{}, pkgerrs.NewValueInvalidError("id")
	}
	return appdto.DeleteExpenseInput{
		ID:            id,
		EnforceUserID: false,
	}, nil
}

func ToListExpensesInput(req httpdto.ListExpensesRequest) (appdto.ListExpensesInput, error) {
	var input appdto.ListExpensesInput

	input.UserID = req.UserID
	input.Category = req.Category

	if req.From != nil && *req.From != "" {
		t, err := time.Parse(time.RFC3339, *req.From)
		if err != nil {
			return appdto.ListExpensesInput{}, pkgerrs.NewValueInvalidErrorWithReason("from", err)
		}
		input.From = &t
	}

	if req.To != nil && *req.To != "" {
		t, err := time.Parse(time.RFC3339, *req.To)
		if err != nil {
			return appdto.ListExpensesInput{}, pkgerrs.NewValueInvalidErrorWithReason("to", err)
		}
		input.To = &t
	}

	if req.Limit != nil {
		input.Limit = *req.Limit
	}
	if req.Offset != nil {
		input.Offset = *req.Offset
	}

	return input, nil
}

func ToListExpensesResponse(dtos []appdto.ExpenseDTO) httpdto.ListExpensesResponse {
	expenses := ToExpenseResponses(dtos)
	return httpdto.ListExpensesResponse{
		Items:    expenses,
		Expenses: expenses,
	}
}

func ToGetDistinctUsersInput(_ httpdto.GetDistinctUsersRequest) appdto.GetDistinctUsersInput {
	return appdto.GetDistinctUsersInput{}
}

func ToGetDistinctUsersResponse(userIDs []int64) httpdto.GetDistinctUsersResponse {
	if userIDs == nil {
		userIDs = []int64{}
	}
	return httpdto.GetDistinctUsersResponse{
		UserIDs: userIDs,
	}
}

func ToExpense(d appdto.ExpenseDTO) httpdto.Expense {
	return httpdto.Expense{
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

func ToExpenseResponse(d appdto.ExpenseDTO) httpdto.Expense {
	return ToExpense(d)
}

func ToExpenseResponses(dtos []appdto.ExpenseDTO) []httpdto.Expense {
	result := make([]httpdto.Expense, len(dtos))
	for i, d := range dtos {
		result[i] = ToExpense(d)
	}
	return result
}
