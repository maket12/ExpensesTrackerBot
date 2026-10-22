//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/stretchr/testify/require"

	adapterhttp "backend/internal/adapter/in/http"
	adapterpostgres "backend/internal/adapter/out/postgres"
	"backend/internal/app/usecases"
	"backend/internal/domain/models"
	"backend/migrations"
	pkgpostgres "backend/pkg/postgres"
)

const migrationVersion = 1

var validCategories = []string{
	"еда",
	"транспорт",
	"жильё",
	"развлечения",
	"здоровье",
	"покупки",
	"другое",
}

type testApp struct {
	server       *httptest.Server
	client       *http.Client
	pg           *pkgpostgres.TestContainer
	dbClient     *pkgpostgres.Client
	expensesRepo *adapterpostgres.ExpensesRepository
}

var (
	appInstance *testApp
	once        sync.Once
)

func setupE2E(t *testing.T) *testApp {
	once.Do(func() {
		ctx := context.Background()

		pg, err := pkgpostgres.StartTestContainer(ctx)
		require.NoError(t, err)

		err = pg.MigrateUp(migrations.FS, migrationVersion)
		require.NoError(t, err)

		pgClient, err := pkgpostgres.NewClient(ctx, pg.Config)
		require.NoError(t, err)

		expensesRepo := adapterpostgres.NewExpensesRepository(pgClient)

		createExpenseUC := usecases.NewCreateExpenseUC(expensesRepo)
		getExpenseUC := usecases.NewGetExpenseUC(expensesRepo)
		updateExpenseUC := usecases.NewUpdateExpenseUC(expensesRepo)
		deleteExpenseUC := usecases.NewDeleteExpenseUC(expensesRepo)
		listExpensesUC := usecases.NewListExpensesUC(expensesRepo)
		getDistinctUsersUC := usecases.NewGetDistinctUsersUC(expensesRepo)

		logger := slog.New(slog.NewTextHandler(io.Discard, nil))

		expensesHandler := adapterhttp.NewExpensesHandler(
			logger,
			createExpenseUC,
			getExpenseUC,
			updateExpenseUC,
			deleteExpenseUC,
			listExpensesUC,
			getDistinctUsersUC,
		)

		router := adapterhttp.NewRouter(expensesHandler).InitRoutes()
		ts := httptest.NewServer(router)

		appInstance = &testApp{
			server:       ts,
			client:       ts.Client(),
			pg:           pg,
			dbClient:     pgClient,
			expensesRepo: expensesRepo,
		}
	})

	appInstance.cleanData(t, context.Background())
	return appInstance
}

func (a *testApp) cleanData(t *testing.T, ctx context.Context) {
	err := a.pg.TruncateTables(ctx, "expenses")
	require.NoError(t, err)
}

func (a *testApp) makeRequest(method, path string, body interface{}) (*http.Response, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, a.server.URL+path, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return a.client.Do(req)
}

func (a *testApp) randomCategory() string {
	return gofakeit.RandomString(validCategories)
}

func (a *testApp) randomExpensePayload() map[string]interface{} {
	return map[string]interface{}{
		"user_id":  int64(gofakeit.Number(100, 10000)),
		"amount":   math.Round(gofakeit.Float64Range(10.0, 500.0)*100) / 100,
		"category": a.randomCategory(),
		"spent_at": time.Now().UTC().Add(-time.Duration(gofakeit.Number(1, 48)) * time.Hour).Format(time.RFC3339),
		"comment":  gofakeit.Sentence(3),
	}
}

func (a *testApp) seedExpense(t *testing.T, userID int64, amount float64, category string, spentAt time.Time, comment string) *models.Expense {
	expense, err := models.NewExpense(userID, amount, category, spentAt, comment)
	require.NoError(t, err)

	created, err := a.expensesRepo.Create(context.Background(), expense)
	require.NoError(t, err)
	return created
}
