//go:build e2e

package e2e

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsers_GetDistinct(t *testing.T) {
	app := setupE2E(t)

	t.Run("Empty Database", func(t *testing.T) {
		resp, err := app.makeRequest(http.MethodGet, "/api/users", nil)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body struct {
			UserIDs []int64 `json:"user_ids"`
		}
		err = json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Empty(t, body.UserIDs)
	})

	t.Run("Populated Database with Duplicates", func(t *testing.T) {
		now := time.Now().UTC()
		app.seedExpense(t, 200, 50.0, "еда", now.Add(-3*time.Hour), "lunch")
		app.seedExpense(t, 100, 30.0, "транспорт", now.Add(-2*time.Hour), "bus")
		app.seedExpense(t, 300, 80.0, "покупки", now.Add(-1*time.Hour), "mall")
		app.seedExpense(t, 100, 20.0, "еда", now, "snack")

		resp, err := app.makeRequest(http.MethodGet, "/api/users", nil)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body struct {
			UserIDs []int64 `json:"user_ids"`
		}
		err = json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)

		expected := []int64{100, 200, 300}
		assert.Equal(t, expected, body.UserIDs)
	})
}

func TestHealth_Check(t *testing.T) {
	app := setupE2E(t)

	resp, err := app.makeRequest(http.MethodGet, "/api/health", nil)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body struct {
		Status string `json:"status"`
		Time   string `json:"time"`
	}
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)

	assert.Equal(t, "ok", body.Status)
	assert.NotEmpty(t, body.Time)
}
