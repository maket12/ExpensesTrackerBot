package http

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"backend/internal/adapter/in/http/dto"
	"backend/internal/adapter/in/http/errs"
	"backend/internal/adapter/in/http/mapper"
	appdto "backend/internal/app/dto"
	"backend/internal/app/usecases"
	pkgerrs "backend/pkg/errs"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	expenseUseCase *usecases.ExpenseUseCase
}

func NewHandler(expenseUseCase *usecases.ExpenseUseCase) *Handler {
	return &Handler{
		expenseUseCase: expenseUseCase,
	}
}

// @Summary Create a new expense
// @Description Create a new expense record for any family member
// @Tags expenses
// @Accept json
// @Produce json
// @Param request body dto.CreateExpenseRequest true "Expense request"
// @Success 201 {object} dto.ExpenseResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/expenses [post]
func (h *Handler) CreateExpense(c echo.Context) error {
	ctx := c.Request().Context()
	var req dto.CreateExpenseRequest
	if err := c.Bind(&req); err != nil {
		slog.WarnContext(ctx, "HTTP create expense: invalid body", "error", err)
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: pkgerrs.ErrInvalidJSON.Error()})
	}

	out, err := h.expenseUseCase.CreateExpense(ctx, mapper.ToCreateExpenseInput(req))
	if err != nil {
		httpErr := errs.FromDomainOrAppError(err)
		slog.WarnContext(ctx, "HTTP create expense failed", "user_id", req.UserID, "error", err, "code", httpErr.Code)
		return c.JSON(httpErr.Code, dto.ErrorResponse{Error: httpErr.Message})
	}

	slog.InfoContext(ctx, "HTTP create expense success", "expense_id", out.Expense.ID, "user_id", out.Expense.UserID, "amount", out.Expense.Amount)
	return c.JSON(http.StatusCreated, mapper.ToExpenseResponse(out.Expense))
}

// @Summary Get expense by ID
// @Description Retrieve a single expense by UUID
// @Tags expenses
// @Produce json
// @Param id path string true "Expense UUID"
// @Success 200 {object} dto.ExpenseResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/expenses/{id} [get]
func (h *Handler) GetExpense(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	out, err := h.expenseUseCase.GetExpense(ctx, appdto.GetExpenseInput{ID: id})
	if err != nil {
		httpErr := errs.FromDomainOrAppError(err)
		slog.WarnContext(ctx, "HTTP get expense failed", "id", id, "error", err, "code", httpErr.Code)
		return c.JSON(httpErr.Code, dto.ErrorResponse{Error: httpErr.Message})
	}

	slog.DebugContext(ctx, "HTTP get expense success", "id", id)
	return c.JSON(http.StatusOK, mapper.ToExpenseResponse(out.Expense))
}

// @Summary Update an existing expense
// @Description Update expense details by UUID
// @Tags expenses
// @Accept json
// @Produce json
// @Param id path string true "Expense UUID"
// @Param request body dto.UpdateExpenseRequest true "Update request"
// @Success 200 {object} dto.ExpenseResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/expenses/{id} [put]
func (h *Handler) UpdateExpense(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")
	var req dto.UpdateExpenseRequest
	if err := c.Bind(&req); err != nil {
		slog.WarnContext(ctx, "HTTP update expense: invalid body", "id", id, "error", err)
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: pkgerrs.ErrInvalidJSON.Error()})
	}

	out, err := h.expenseUseCase.UpdateExpense(ctx, mapper.ToUpdateExpenseInput(id, req))
	if err != nil {
		httpErr := errs.FromDomainOrAppError(err)
		slog.WarnContext(ctx, "HTTP update expense failed", "id", id, "error", err, "code", httpErr.Code)
		return c.JSON(httpErr.Code, dto.ErrorResponse{Error: httpErr.Message})
	}

	slog.InfoContext(ctx, "HTTP update expense success", "id", id, "amount", out.Expense.Amount, "category", out.Expense.Category)
	return c.JSON(http.StatusOK, mapper.ToExpenseResponse(out.Expense))
}

// @Summary Delete an expense
// @Description Delete an expense by UUID
// @Tags expenses
// @Produce json
// @Param id path string true "Expense UUID"
// @Success 204 "No Content"
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/expenses/{id} [delete]
func (h *Handler) DeleteExpense(c echo.Context) error {
	ctx := c.Request().Context()
	id := c.Param("id")

	_, err := h.expenseUseCase.DeleteExpense(ctx, appdto.DeleteExpenseInput{
		ID:            id,
		EnforceUserID: false,
	})
	if err != nil {
		httpErr := errs.FromDomainOrAppError(err)
		slog.WarnContext(ctx, "HTTP delete expense failed", "id", id, "error", err, "code", httpErr.Code)
		return c.JSON(httpErr.Code, dto.ErrorResponse{Error: httpErr.Message})
	}

	slog.InfoContext(ctx, "HTTP delete expense success", "id", id)
	return c.NoContent(http.StatusNoContent)
}

// @Summary List expenses
// @Description Retrieve a list of expenses with optional filtering
// @Tags expenses
// @Produce json
// @Param user_id query int false "Telegram User ID"
// @Param category query string false "Category"
// @Param from query string false "From timestamp (RFC3339)"
// @Param to query string false "To timestamp (RFC3339)"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} dto.ListExpensesResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/expenses [get]
func (h *Handler) ListExpenses(c echo.Context) error {
	ctx := c.Request().Context()
	var input appdto.ListExpensesInput

	if userIDStr := c.QueryParam("user_id"); userIDStr != "" {
		if uid, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			input.UserID = &uid
		}
	}

	if cat := c.QueryParam("category"); cat != "" {
		input.Category = &cat
	}

	if fromStr := c.QueryParam("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			input.From = &t
		}
	}

	if toStr := c.QueryParam("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			input.To = &t
		}
	}

	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			input.Limit = l
		}
	}

	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			input.Offset = o
		}
	}

	out, err := h.expenseUseCase.ListExpenses(ctx, input)
	if err != nil {
		httpErr := errs.FromDomainOrAppError(err)
		slog.WarnContext(ctx, "HTTP list expenses failed", "error", err, "code", httpErr.Code)
		return c.JSON(httpErr.Code, dto.ErrorResponse{Error: httpErr.Message})
	}

	slog.DebugContext(ctx, "HTTP list expenses success", "count", len(out.Expenses))
	return c.JSON(http.StatusOK, dto.ListExpensesResponse{
		Items: mapper.ToExpenseResponses(out.Expenses),
	})
}

// @Summary Get distinct user IDs
// @Description List all distinct family member user IDs in database
// @Tags users
// @Produce json
// @Success 200 {object} dto.GetDistinctUsersResponse
// @Router /api/users [get]
func (h *Handler) GetDistinctUsers(c echo.Context) error {
	ctx := c.Request().Context()
	out, err := h.expenseUseCase.GetDistinctUsers(ctx, appdto.GetDistinctUsersInput{})
	if err != nil {
		httpErr := errs.FromDomainOrAppError(err)
		slog.WarnContext(ctx, "HTTP get distinct users failed", "error", err, "code", httpErr.Code)
		return c.JSON(httpErr.Code, dto.ErrorResponse{Error: httpErr.Message})
	}

	slog.DebugContext(ctx, "HTTP get distinct users success", "count", len(out.UserIDs))
	return c.JSON(http.StatusOK, dto.GetDistinctUsersResponse{
		UserIDs: out.UserIDs,
	})
}

// @Summary Health check
// @Description Check service availability
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/health [get]
func (h *Handler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}
