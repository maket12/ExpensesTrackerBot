package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"backend/internal/app/dto"
	"backend/internal/domain/port"
)

type ToolExecutor interface {
	Execute(ctx context.Context, userID int64, tc port.ToolCall) string
}

type AgentToolExecutor struct {
	createUC *CreateExpenseUC
	listUC   *ListExpensesUC
	updateUC *UpdateExpenseUC
	deleteUC *DeleteExpenseUC
}

func NewAgentToolExecutor(
	createUC *CreateExpenseUC,
	listUC *ListExpensesUC,
	updateUC *UpdateExpenseUC,
	deleteUC *DeleteExpenseUC,
) *AgentToolExecutor {
	return &AgentToolExecutor{
		createUC: createUC,
		listUC:   listUC,
		updateUC: updateUC,
		deleteUC: deleteUC,
	}
}

type addExpenseArgs struct {
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	SpentAt  string  `json:"spent_at"`
	Comment  string  `json:"comment"`
}

type listExpensesArgs struct {
	Limit    int     `json:"limit"`
	Category *string `json:"category"`
	From     *string `json:"from"`
	To       *string `json:"to"`
}

type updateExpenseArgs struct {
	ID       string   `json:"id"`
	Amount   *float64 `json:"amount"`
	Category *string  `json:"category"`
	SpentAt  *string  `json:"spent_at"`
	Comment  *string  `json:"comment"`
}

type deleteExpenseArgs struct {
	ID string `json:"id"`
}

func (e *AgentToolExecutor) Execute(ctx context.Context, userID int64, tc port.ToolCall) string {
	switch tc.Name {
	case "add_expense":
		var args addExpenseArgs
		if err := json.Unmarshal([]byte(tc.Arguments), &args); err != nil {
			return fmt.Sprintf(`{"error":"некорректные параметры: %s"}`, err.Error())
		}

		spentAt, err := time.Parse(time.RFC3339, args.SpentAt)
		if err != nil {
			spentAt = time.Now().UTC()
		}

		out, err := e.createUC.Execute(ctx, dto.CreateExpenseInput{
			UserID:   userID,
			Amount:   args.Amount,
			Category: args.Category,
			SpentAt:  spentAt,
			Comment:  args.Comment,
		})
		if err != nil {
			return fmt.Sprintf(`{"error":"ошибка создания траты: %s"}`, err.Error())
		}
		data, _ := json.Marshal(out.Expense)
		return string(data)

	case "list_expenses":
		var args listExpensesArgs
		_ = json.Unmarshal([]byte(tc.Arguments), &args)

		limit := args.Limit
		if limit <= 0 {
			limit = 10
		}

		var fromTime, toTime *time.Time
		if args.From != nil {
			if t, err := time.Parse(time.RFC3339, *args.From); err == nil {
				fromTime = &t
			}
		}
		if args.To != nil {
			if t, err := time.Parse(time.RFC3339, *args.To); err == nil {
				toTime = &t
			}
		}

		out, err := e.listUC.Execute(ctx, dto.ListExpensesInput{
			UserID:   &userID,
			Category: args.Category,
			From:     fromTime,
			To:       toTime,
			Limit:    limit,
		})
		if err != nil {
			return fmt.Sprintf(`{"error":"ошибка получения списка трат: %s"}`, err.Error())
		}

		var total float64
		for _, item := range out.Expenses {
			total += item.Amount
		}

		res := map[string]any{
			"expenses":  out.Expenses,
			"count":     len(out.Expenses),
			"total_sum": total,
		}
		data, _ := json.Marshal(res)
		return string(data)

	case "update_expense":
		var args updateExpenseArgs
		if err := json.Unmarshal([]byte(tc.Arguments), &args); err != nil {
			return fmt.Sprintf(`{"error":"некорректные параметры: %s"}`, err.Error())
		}

		var spentAt *time.Time
		if args.SpentAt != nil {
			t, err := time.Parse(time.RFC3339, *args.SpentAt)
			if err != nil {
				return fmt.Sprintf(`{"error":"некорректный формат spent_at: %s"}`, err.Error())
			}
			spentAt = &t
		}

		out, err := e.updateUC.Execute(ctx, dto.UpdateExpenseInput{
			ID:            args.ID,
			UserID:        userID,
			Amount:        args.Amount,
			Category:      args.Category,
			SpentAt:       spentAt,
			Comment:       args.Comment,
			EnforceUserID: true,
		})
		if err != nil {
			return fmt.Sprintf(`{"error":"ошибка обновления: %s"}`, err.Error())
		}
		data, _ := json.Marshal(out.Expense)
		return string(data)

	case "delete_expense":
		var args deleteExpenseArgs
		if err := json.Unmarshal([]byte(tc.Arguments), &args); err != nil {
			return fmt.Sprintf(`{"error":"некорректные параметры: %s"}`, err.Error())
		}

		out, err := e.deleteUC.Execute(ctx, dto.DeleteExpenseInput{
			ID:            args.ID,
			UserID:        userID,
			EnforceUserID: true,
		})
		if err != nil {
			return fmt.Sprintf(`{"error":"ошибка удаления: %s"}`, err.Error())
		}
		data, _ := json.Marshal(out)
		return string(data)

	default:
		return `{"error":"неизвестный инструмент"}`
	}
}
