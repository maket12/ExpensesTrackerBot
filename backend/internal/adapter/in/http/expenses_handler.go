package http

import (
	"log/slog"
	"net/http"
	"time"

	"backend/internal/adapter/in/http/dto"
	"backend/internal/adapter/in/http/mapper"
	"backend/internal/app/usecases"
	pkgerrs "backend/pkg/errs"

	"github.com/labstack/echo/v4"
)

type ExpensesHandler struct {
	log                *slog.Logger
	createExpenseUC    *usecases.CreateExpenseUC
	getExpenseUC       *usecases.GetExpenseUC
	updateExpenseUC    *usecases.UpdateExpenseUC
	deleteExpenseUC    *usecases.DeleteExpenseUC
	listExpensesUC     *usecases.ListExpensesUC
	getDistinctUsersUC *usecases.GetDistinctUsersUC
}

func NewExpensesHandler(
	log *slog.Logger,
	createExpenseUC *usecases.CreateExpenseUC,
	getExpenseUC *usecases.GetExpenseUC,
	updateExpenseUC *usecases.UpdateExpenseUC,
	deleteExpenseUC *usecases.DeleteExpenseUC,
	listExpensesUC *usecases.ListExpensesUC,
	getDistinctUsersUC *usecases.GetDistinctUsersUC,
) *ExpensesHandler {
	return &ExpensesHandler{
		log:                log,
		createExpenseUC:    createExpenseUC,
		getExpenseUC:       getExpenseUC,
		updateExpenseUC:    updateExpenseUC,
		deleteExpenseUC:    deleteExpenseUC,
		listExpensesUC:     listExpensesUC,
		getDistinctUsersUC: getDistinctUsersUC,
	}
}

// Create handles creation of a new expense.
// @Summary Create a new expense
// @Description Create a new expense record for any family member
// @Tags expenses
// @Accept json
// @Produce json
// @Param request body dto.CreateExpenseRequest true "Expense request"
// @Success 201 {object} dto.CreateExpenseResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/expenses [post]
func (h *ExpensesHandler) Create(c echo.Context) error {
	var req dto.CreateExpenseRequest
	if err := c.Bind(&req); err != nil {
		return h.returnErr(c, "binding failed", pkgerrs.ErrInvalidJSON)
	}

	out, err := h.createExpenseUC.Execute(c.Request().Context(), mapper.ToCreateExpenseInput(req))
	if err != nil {
		return h.returnErr(c, "failed to create expense", err)
	}

	slog.InfoContext(c.Request().Context(), "successfully created expense",
		slog.String("expense_id", out.Expense.ID.String()),
		slog.Int64("user_id", out.Expense.UserID),
		slog.Float64("amount", out.Expense.Amount),
	)

	return c.JSON(http.StatusCreated, mapper.ToCreateExpenseResponse(out.Expense))
}

// Get retrieves an expense by ID.
// @Summary Get expense by ID
// @Description Retrieve a single expense by UUID
// @Tags expenses
// @Produce json
// @Param id path string true "Expense UUID"
// @Success 200 {object} dto.GetExpenseResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/expenses/{id} [get]
func (h *ExpensesHandler) Get(c echo.Context) error {
	var req dto.GetExpenseRequest
	if err := c.Bind(&req); err != nil {
		return h.returnErr(c, "binding failed", pkgerrs.ErrInvalidIdentifier)
	}

	input, err := mapper.ToGetExpenseInput(req)
	if err != nil {
		return h.returnErr(c, "invalid id", err)
	}

	out, err := h.getExpenseUC.Execute(c.Request().Context(), input)
	if err != nil {
		return h.returnErr(c, "failed to get expense", err)
	}

	slog.InfoContext(c.Request().Context(), "successfully retrieved expense",
		slog.String("expense_id", out.Expense.ID.String()),
		slog.Int64("user_id", out.Expense.UserID),
	)

	return c.JSON(http.StatusOK, mapper.ToGetExpenseResponse(out.Expense))
}

// Update updates an existing expense.
// @Summary Update an existing expense
// @Description Update expense details by UUID
// @Tags expenses
// @Accept json
// @Produce json
// @Param id path string true "Expense UUID"
// @Param request body dto.UpdateExpenseRequest true "Update request"
// @Success 200 {object} dto.UpdateExpenseResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/expenses/{id} [put]
func (h *ExpensesHandler) Update(c echo.Context) error {
	var req dto.UpdateExpenseRequest
	if err := c.Bind(&req); err != nil {
		return h.returnErr(c, "binding failed", pkgerrs.ErrInvalidJSON)
	}

	input, err := mapper.ToUpdateExpenseInput(req)
	if err != nil {
		return h.returnErr(c, "invalid update input", err)
	}

	out, err := h.updateExpenseUC.Execute(c.Request().Context(), input)
	if err != nil {
		return h.returnErr(c, "failed to update expense", err)
	}

	slog.InfoContext(c.Request().Context(), "successfully updated expense",
		slog.String("expense_id", out.Expense.ID.String()),
		slog.Int64("user_id", out.Expense.UserID),
		slog.Float64("amount", out.Expense.Amount),
	)

	return c.JSON(http.StatusOK, mapper.ToUpdateExpenseResponse(out.Expense))
}

// Delete deletes an expense by ID.
// @Summary Delete an expense
// @Description Delete an expense by UUID
// @Tags expenses
// @Produce json
// @Param id path string true "Expense UUID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/expenses/{id} [delete]
func (h *ExpensesHandler) Delete(c echo.Context) error {
	var req dto.DeleteExpenseRequest
	if err := c.Bind(&req); err != nil {
		return h.returnErr(c, "binding failed", pkgerrs.ErrInvalidIdentifier)
	}

	input, err := mapper.ToDeleteExpenseInput(req)
	if err != nil {
		return h.returnErr(c, "invalid id", err)
	}

	_, err = h.deleteExpenseUC.Execute(c.Request().Context(), input)
	if err != nil {
		return h.returnErr(c, "failed to delete expense", err)
	}

	slog.InfoContext(c.Request().Context(), "successfully deleted expense",
		slog.String("expense_id", req.ID),
	)

	return c.NoContent(http.StatusNoContent)
}

// List retrieves expenses with optional filters.
// @Summary List expenses
// @Description Retrieve a list of expenses with optional filtering
// @Tags expenses
// @Produce json
// @Param user_id query int false "Telegram User ID"
// @Param category query string false "Category" Enums(еда, транспорт, жильё, развлечения, здоровье, покупки, другое, все)
// @Param from query string false "From timestamp (RFC3339)"
// @Param to query string false "To timestamp (RFC3339)"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} dto.ListExpensesResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/expenses [get]
func (h *ExpensesHandler) List(c echo.Context) error {
	var req dto.ListExpensesRequest
	if err := c.Bind(&req); err != nil {
		return h.returnErr(c, "binding failed", pkgerrs.ErrInvalidJSON)
	}

	input, err := mapper.ToListExpensesInput(req)
	if err != nil {
		return h.returnErr(c, "invalid list filter", err)
	}

	out, err := h.listExpensesUC.Execute(c.Request().Context(), input)
	if err != nil {
		return h.returnErr(c, "failed to list expenses", err)
	}

	slog.InfoContext(c.Request().Context(), "successfully listed expenses",
		slog.Int("count", len(out.Expenses)),
	)

	return c.JSON(http.StatusOK, mapper.ToListExpensesResponse(out.Expenses))
}

// GetDistinctUsers retrieves all distinct user IDs.
// @Summary Get distinct user IDs
// @Description List all distinct family member user IDs in database
// @Tags users
// @Produce json
// @Success 200 {object} dto.GetDistinctUsersResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/users [get]
func (h *ExpensesHandler) GetDistinctUsers(c echo.Context) error {
	out, err := h.getDistinctUsersUC.Execute(c.Request().Context())
	if err != nil {
		return h.returnErr(c, "failed to get distinct users", err)
	}

	slog.InfoContext(c.Request().Context(), "successfully retrieved distinct users",
		slog.Int("count", len(out.UserIDs)),
	)

	return c.JSON(http.StatusOK, mapper.ToGetDistinctUsersResponse(out.UserIDs))
}

// Health returns service status.
// @Summary Health check
// @Description Check service availability
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/health [get]
func (h *ExpensesHandler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *ExpensesHandler) returnErr(c echo.Context, msg string, err error) error {
	if err == nil {
		return c.NoContent(http.StatusBadRequest)
	}

	outErr := mapper.HttpError(err)

	if h.log != nil {
		h.log.LogAttrs(c.Request().Context(), outErr.Level, msg,
			slog.Int("code", outErr.Code),
			slog.String("public_msg", outErr.Message),
			slog.Any("cause", outErr.Reason),
		)
	} else {
		slog.LogAttrs(c.Request().Context(), outErr.Level, msg,
			slog.Int("code", outErr.Code),
			slog.String("public_msg", outErr.Message),
			slog.Any("cause", outErr.Reason),
		)
	}

	return c.JSON(outErr.Code, mapper.MapErrorToResponse(outErr.Message))
}
