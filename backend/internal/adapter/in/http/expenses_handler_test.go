package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/adapter/in/http/dto"
	"backend/internal/app/usecases"
	"backend/internal/domain/models"
	"backend/internal/domain/port/mocks"
	pkgerrs "backend/pkg/errs"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupTestHandler(t *testing.T) (*ExpensesHandler, *mocks.MockExpenseRepository, *echo.Echo) {
	repo := mocks.NewMockExpenseRepository(t)
	e := echo.New()

	createUC := usecases.NewCreateExpenseUC(repo)
	getUC := usecases.NewGetExpenseUC(repo)
	updateUC := usecases.NewUpdateExpenseUC(repo)
	deleteUC := usecases.NewDeleteExpenseUC(repo)
	listUC := usecases.NewListExpensesUC(repo)
	getUsersUC := usecases.NewGetDistinctUsersUC(repo)

	handler := NewExpensesHandler(nil, createUC, getUC, updateUC, deleteUC, listUC, getUsersUC)
	return handler, repo, e
}

func TestExpensesHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, repo, e := setupTestHandler(t)

		reqBody := dto.CreateExpenseRequest{
			UserID:   100,
			Amount:   50.0,
			Category: "еда",
			SpentAt:  time.Now().UTC().Add(-time.Hour),
			Comment:  "lunch",
		}
		body, _ := json.Marshal(reqBody)

		expense, err := models.NewExpense(reqBody.UserID, reqBody.Amount, reqBody.Category, reqBody.SpentAt, reqBody.Comment)
		require.NoError(t, err)

		repo.On("Create", mock.Anything, mock.AnythingOfType("*models.Expense")).Return(expense, nil).Once()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/expenses", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = handler.Create(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp dto.CreateExpenseResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, expense.ID().String(), resp.Expense.ID)
		assert.Equal(t, int64(100), resp.Expense.UserID)
	})

	t.Run("binding failed", func(t *testing.T) {
		handler, _, e := setupTestHandler(t)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/expenses", bytes.NewReader([]byte("{invalid-json")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.Create(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestExpensesHandler_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, repo, e := setupTestHandler(t)

		expense, err := models.NewExpense(100, 50.0, "еда", time.Now().UTC().Add(-time.Hour), "lunch")
		require.NoError(t, err)

		repo.On("Get", mock.Anything, expense.ID()).Return(expense, nil).Once()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/expenses/"+expense.ID().String(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(expense.ID().String())

		err = handler.Get(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp dto.GetExpenseResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, expense.ID().String(), resp.Expense.ID)
	})

	t.Run("invalid id", func(t *testing.T) {
		handler, _, e := setupTestHandler(t)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/expenses/not-a-uuid", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("not-a-uuid")

		err := handler.Get(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("not found", func(t *testing.T) {
		handler, repo, e := setupTestHandler(t)
		id := uuid.New()

		repo.On("Get", mock.Anything, id).Return(nil, pkgerrs.ErrObjectNotFound).Once()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/expenses/"+id.String(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(id.String())

		err := handler.Get(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestExpensesHandler_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, repo, e := setupTestHandler(t)

		existing, err := models.NewExpense(100, 50.0, "еда", time.Now().UTC().Add(-time.Hour), "lunch")
		require.NoError(t, err)

		newAmount := 75.0
		reqBody := dto.UpdateExpenseRequest{
			Amount: &newAmount,
		}
		body, _ := json.Marshal(reqBody)

		repo.On("Get", mock.Anything, existing.ID()).Return(existing, nil).Once()
		repo.On("Update", mock.Anything, mock.AnythingOfType("*models.Expense")).Return(existing, nil).Once()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/api/expenses/"+existing.ID().String(), bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(existing.ID().String())

		err = handler.Update(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp dto.UpdateExpenseResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, existing.ID().String(), resp.Expense.ID)
	})

	t.Run("invalid id", func(t *testing.T) {
		handler, _, e := setupTestHandler(t)

		req := httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/api/expenses/bad-id", bytes.NewReader([]byte("{}")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("bad-id")

		err := handler.Update(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestExpensesHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, repo, e := setupTestHandler(t)
		id := uuid.New()

		repo.On("Delete", mock.Anything, id).Return(nil).Once()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodDelete, "/api/expenses/"+id.String(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(id.String())

		err := handler.Delete(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("not found", func(t *testing.T) {
		handler, repo, e := setupTestHandler(t)
		id := uuid.New()

		repo.On("Delete", mock.Anything, id).Return(pkgerrs.ErrObjectNotFound).Once()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodDelete, "/api/expenses/"+id.String(), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(id.String())

		err := handler.Delete(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestExpensesHandler_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		handler, repo, e := setupTestHandler(t)

		expense, err := models.NewExpense(100, 50.0, "еда", time.Now().UTC().Add(-time.Hour), "lunch")
		require.NoError(t, err)

		repo.On("List", mock.Anything, mock.Anything).Return([]*models.Expense{expense}, nil).Once()

		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/expenses?user_id=100&category=еда", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err = handler.List(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp dto.ListExpensesResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Len(t, resp.Items, 1)
		assert.Equal(t, expense.ID().String(), resp.Items[0].ID)
	})
}

func TestExpensesHandler_GetDistinctUsers(t *testing.T) {
	handler, repo, e := setupTestHandler(t)

	repo.On("ListUserIDS", mock.Anything).Return([]int64{10, 20, 30}, nil).Once()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.GetDistinctUsers(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp dto.GetDistinctUsersResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, []int64{10, 20, 30}, resp.UserIDs)
}

func TestExpensesHandler_Health(t *testing.T) {
	handler, _, e := setupTestHandler(t)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.Health(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	body, err := io.ReadAll(rec.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "ok")
}
