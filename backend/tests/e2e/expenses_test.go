//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpense_CRUD(t *testing.T) {
	app := setupE2E(t)

	var createdID string
	payload := app.randomExpensePayload()

	t.Run("Create Expense Success", func(t *testing.T) {
		resp, err := app.makeRequest(http.MethodPost, "/api/expenses", payload)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		defer func() { _ = resp.Body.Close() }()

		var body struct {
			Expense struct {
				ID        string    `json:"id"`
				UserID    int64     `json:"user_id"`
				Amount    float64   `json:"amount"`
				Category  string    `json:"category"`
				SpentAt   time.Time `json:"spent_at"`
				Comment   string    `json:"comment"`
				CreatedAt time.Time `json:"created_at"`
				UpdatedAt time.Time `json:"updated_at"`
			} `json:"expense"`
		}
		err = json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)

		createdID = body.Expense.ID
		assert.NotEmpty(t, createdID)
		_, err = uuid.Parse(createdID)
		assert.NoError(t, err)

		assert.Equal(t, payload["user_id"], body.Expense.UserID)
		assert.InDelta(t, payload["amount"].(float64), body.Expense.Amount, 0.001)
		assert.Equal(t, payload["category"], body.Expense.Category)
		assert.Equal(t, payload["comment"], body.Expense.Comment)
		assert.False(t, body.Expense.CreatedAt.IsZero())
	})

	t.Run("Create Expense Bad Request - Invalid JSON", func(t *testing.T) {
		resp, err := app.makeRequest(http.MethodPost, "/api/expenses", "{invalid-json")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Create Expense Bad Request - Invalid Category", func(t *testing.T) {
		invalidPayload := app.randomExpensePayload()
		invalidPayload["category"] = "unsupported_cat"

		resp, err := app.makeRequest(http.MethodPost, "/api/expenses", invalidPayload)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Create Expense Bad Request - Zero Amount", func(t *testing.T) {
		invalidPayload := app.randomExpensePayload()
		invalidPayload["amount"] = 0

		resp, err := app.makeRequest(http.MethodPost, "/api/expenses", invalidPayload)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Create Expense Bad Request - Future SpentAt", func(t *testing.T) {
		invalidPayload := app.randomExpensePayload()
		invalidPayload["spent_at"] = time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

		resp, err := app.makeRequest(http.MethodPost, "/api/expenses", invalidPayload)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Get Expense Success", func(t *testing.T) {
		resp, err := app.makeRequest(http.MethodGet, "/api/expenses/"+createdID, nil)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body struct {
			Expense struct {
				ID       string  `json:"id"`
				UserID   int64   `json:"user_id"`
				Amount   float64 `json:"amount"`
				Category string  `json:"category"`
				Comment  string  `json:"comment"`
			} `json:"expense"`
		}
		err = json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)

		assert.Equal(t, createdID, body.Expense.ID)
		assert.Equal(t, payload["user_id"], body.Expense.UserID)
		assert.InDelta(t, payload["amount"].(float64), body.Expense.Amount, 0.001)
		assert.Equal(t, payload["category"], body.Expense.Category)
		assert.Equal(t, payload["comment"], body.Expense.Comment)
	})

	t.Run("Get Expense Not Found", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		resp, err := app.makeRequest(http.MethodGet, "/api/expenses/"+nonExistentID, nil)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Get Expense Bad Request - Invalid UUID", func(t *testing.T) {
		resp, err := app.makeRequest(http.MethodGet, "/api/expenses/not-a-valid-uuid", nil)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Update Expense Success", func(t *testing.T) {
		newAmount := 777.77
		newCategory := "здоровье"
		newComment := gofakeit.Sentence(2)
		updatePayload := map[string]interface{}{
			"amount":   newAmount,
			"category": newCategory,
			"comment":  newComment,
		}

		resp, err := app.makeRequest(http.MethodPut, "/api/expenses/"+createdID, updatePayload)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body struct {
			Expense struct {
				ID       string  `json:"id"`
				Amount   float64 `json:"amount"`
				Category string  `json:"category"`
				Comment  string  `json:"comment"`
			} `json:"expense"`
		}
		err = json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)

		assert.Equal(t, createdID, body.Expense.ID)
		assert.InDelta(t, newAmount, body.Expense.Amount, 0.001)
		assert.Equal(t, newCategory, body.Expense.Category)
		assert.Equal(t, newComment, body.Expense.Comment)

		getResp, err := app.makeRequest(http.MethodGet, "/api/expenses/"+createdID, nil)
		require.NoError(t, err)
		defer func() { _ = getResp.Body.Close() }()

		assert.Equal(t, http.StatusOK, getResp.StatusCode)
	})

	t.Run("Update Expense Not Found", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		updatePayload := map[string]interface{}{
			"amount": 100.0,
		}

		resp, err := app.makeRequest(http.MethodPut, "/api/expenses/"+nonExistentID, updatePayload)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Delete Expense Success", func(t *testing.T) {
		resp, err := app.makeRequest(http.MethodDelete, "/api/expenses/"+createdID, nil)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		getResp, err := app.makeRequest(http.MethodGet, "/api/expenses/"+createdID, nil)
		require.NoError(t, err)
		defer func() { _ = getResp.Body.Close() }()

		assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
	})

	t.Run("Delete Expense Not Found", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		resp, err := app.makeRequest(http.MethodDelete, "/api/expenses/"+nonExistentID, nil)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestExpense_List(t *testing.T) {
	app := setupE2E(t)

	now := time.Now().UTC()
	user1 := int64(101)
	user2 := int64(102)

	e1 := app.seedExpense(t, user1, 100.0, "еда", now.Add(-3*time.Hour), "groceries")
	e2 := app.seedExpense(t, user1, 50.0, "транспорт", now.Add(-2*time.Hour), "taxi")
	e3 := app.seedExpense(t, user2, 200.0, "еда", now.Add(-1*time.Hour), "cafe")
	e4 := app.seedExpense(t, user2, 300.0, "покупки", now.Add(-30*time.Minute), "mall")

	t.Run("List All Expenses", func(t *testing.T) {
		resp, err := app.makeRequest(http.MethodGet, "/api/expenses", nil)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
		}
		err = json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)

		assert.Len(t, body.Items, 4)
	})

	t.Run("Filter by User ID", func(t *testing.T) {
		path := fmt.Sprintf("/api/expenses?user_id=%d", user1)
		resp, err := app.makeRequest(http.MethodGet, path, nil)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body struct {
			Items []struct {
				ID     string `json:"id"`
				UserID int64  `json:"user_id"`
			} `json:"items"`
		}
		err = json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)

		assert.Len(t, body.Items, 2)
		for _, it := range body.Items {
			assert.Equal(t, user1, it.UserID)
		}
	})

	t.Run("Filter by Category", func(t *testing.T) {
		resp, err := app.makeRequest(http.MethodGet, "/api/expenses?category=еда", nil)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body struct {
			Items []struct {
				ID       string `json:"id"`
				Category string `json:"category"`
			} `json:"items"`
		}
		err = json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)

		assert.Len(t, body.Items, 2)
		for _, it := range body.Items {
			assert.Equal(t, "еда", it.Category)
		}
	})

	t.Run("Filter by User ID and Category", func(t *testing.T) {
		path := fmt.Sprintf("/api/expenses?user_id=%d&category=еда", user1)
		resp, err := app.makeRequest(http.MethodGet, path, nil)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
		}
		err = json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)

		require.Len(t, body.Items, 1)
		assert.Equal(t, e1.ID().String(), body.Items[0].ID)
	})

	t.Run("Filter by Date Range", func(t *testing.T) {
		from := now.Add(-2*time.Hour - 30*time.Minute).Format(time.RFC3339)
		to := now.Add(-45 * time.Minute).Format(time.RFC3339)

		path := fmt.Sprintf("/api/expenses?from=%s&to=%s", from, to)
		resp, err := app.makeRequest(http.MethodGet, path, nil)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
		}
		err = json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)

		assert.Len(t, body.Items, 2)
		ids := []string{body.Items[0].ID, body.Items[1].ID}
		assert.Contains(t, ids, e2.ID().String())
		assert.Contains(t, ids, e3.ID().String())
	})

	t.Run("Pagination with Limit and Offset", func(t *testing.T) {
		resp, err := app.makeRequest(http.MethodGet, "/api/expenses?limit=2&offset=1", nil)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
		}
		err = json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)

		assert.Len(t, body.Items, 2)
		assert.Equal(t, e3.ID().String(), body.Items[0].ID)
		assert.Equal(t, e2.ID().String(), body.Items[1].ID)
	})

	t.Run("Invalid Date Filter Format", func(t *testing.T) {
		resp, err := app.makeRequest(http.MethodGet, "/api/expenses?from=not-a-date", nil)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	_ = e4
}
