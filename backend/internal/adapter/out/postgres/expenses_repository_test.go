//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"backend/internal/adapter/out/postgres"
	"backend/internal/domain/models"
	"backend/internal/domain/port"
	pkgerrs "backend/pkg/errs"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type ExpensesRepoSuite struct {
	BaseRepoSuite
	repo        *postgres.ExpensesRepository
	testExpense *models.Expense
}

func TestExpensesRepoSuite(t *testing.T) {
	suite.Run(t, new(ExpensesRepoSuite))
}

func (s *ExpensesRepoSuite) SetupSuite() {
	s.SetupBase(1)
	s.repo = postgres.NewExpensesRepository(s.dbClient)
	s.ctx = context.Background()
}

func (s *ExpensesRepoSuite) SetupTest() {
	err := s.pgContainer.TruncateTables(s.ctx, "expenses")
	s.Require().NoError(err)

	expense, err := models.NewExpense(
		100,
		150.50,
		"еда",
		time.Now().UTC().Add(-time.Hour),
		"lunch",
	)
	s.Require().NoError(err)
	s.testExpense = expense
}

func (s *ExpensesRepoSuite) TestCreateAndGet() {
	created, err := s.repo.Create(s.ctx, s.testExpense)
	s.Require().NoError(err)
	s.Require().NotNil(created)
	s.Require().Equal(s.testExpense.ID(), created.ID())
	s.Require().Equal(s.testExpense.UserID(), created.UserID())
	s.Require().InDelta(s.testExpense.Amount(), created.Amount(), 0.001)
	s.Require().Equal(s.testExpense.Category(), created.Category())
	s.Require().WithinDuration(s.testExpense.SpentAt(), created.SpentAt(), time.Second)
	s.Require().Equal(s.testExpense.Comment(), created.Comment())

	got, err := s.repo.Get(s.ctx, s.testExpense.ID())
	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Require().Equal(s.testExpense.ID(), got.ID())
	s.Require().Equal(s.testExpense.UserID(), got.UserID())
	s.Require().InDelta(s.testExpense.Amount(), got.Amount(), 0.001)
	s.Require().Equal(s.testExpense.Category(), got.Category())
	s.Require().WithinDuration(s.testExpense.SpentAt(), got.SpentAt(), time.Second)
	s.Require().Equal(s.testExpense.Comment(), got.Comment())
}

func (s *ExpensesRepoSuite) TestGet_NotFound() {
	got, err := s.repo.Get(s.ctx, uuid.New())
	s.Require().Error(err)
	var notFoundErr *pkgerrs.ObjectNotFoundError
	s.Require().ErrorAs(err, &notFoundErr)
	s.Require().Nil(got)
}

func (s *ExpensesRepoSuite) TestUpdate_Success() {
	_, err := s.repo.Create(s.ctx, s.testExpense)
	s.Require().NoError(err)

	newAmount := 299.99
	newCategory := "транспорт"
	newSpentAt := time.Now().UTC().Add(-30 * time.Minute)
	newComment := "taxi ride"

	err = s.testExpense.Update(&newAmount, &newCategory, &newSpentAt, &newComment)
	s.Require().NoError(err)

	updated, err := s.repo.Update(s.ctx, s.testExpense)
	s.Require().NoError(err)
	s.Require().NotNil(updated)

	got, err := s.repo.Get(s.ctx, s.testExpense.ID())
	s.Require().NoError(err)
	s.Require().NotNil(got)
	s.Require().InDelta(newAmount, got.Amount(), 0.001)
	s.Require().Equal(models.Category(newCategory), got.Category())
	s.Require().WithinDuration(newSpentAt, got.SpentAt(), time.Second)
	s.Require().Equal(newComment, got.Comment())
}

func (s *ExpensesRepoSuite) TestUpdate_NotFound() {
	unpersisted, err := models.NewExpense(
		100,
		50.0,
		"еда",
		time.Now().UTC().Add(-time.Hour),
		"unpersisted",
	)
	s.Require().NoError(err)

	updated, err := s.repo.Update(s.ctx, unpersisted)
	s.Require().Error(err)
	var notFoundErr *pkgerrs.ObjectNotFoundError
	s.Require().ErrorAs(err, &notFoundErr)
	s.Require().Nil(updated)
}

func (s *ExpensesRepoSuite) TestDelete_Success() {
	_, err := s.repo.Create(s.ctx, s.testExpense)
	s.Require().NoError(err)

	err = s.repo.Delete(s.ctx, s.testExpense.ID())
	s.Require().NoError(err)

	got, err := s.repo.Get(s.ctx, s.testExpense.ID())
	s.Require().Error(err)
	var notFoundErr *pkgerrs.ObjectNotFoundError
	s.Require().ErrorAs(err, &notFoundErr)
	s.Require().Nil(got)
}

func (s *ExpensesRepoSuite) TestDelete_NotFound() {
	err := s.repo.Delete(s.ctx, uuid.New())
	s.Require().Error(err)
	var notFoundErr *pkgerrs.ObjectNotFoundError
	s.Require().ErrorAs(err, &notFoundErr)
}

func (s *ExpensesRepoSuite) TestList_Filtered() {
	now := time.Now().UTC()

	e1, err := models.NewExpense(100, 10.0, "еда", now.Add(-3*time.Hour), "groceries")
	s.Require().NoError(err)
	_, err = s.repo.Create(s.ctx, e1)
	s.Require().NoError(err)

	e2, err := models.NewExpense(100, 20.0, "транспорт", now.Add(-2*time.Hour), "bus")
	s.Require().NoError(err)
	_, err = s.repo.Create(s.ctx, e2)
	s.Require().NoError(err)

	e3, err := models.NewExpense(200, 30.0, "еда", now.Add(-1*time.Hour), "dinner")
	s.Require().NoError(err)
	_, err = s.repo.Create(s.ctx, e3)
	s.Require().NoError(err)

	uid := int64(100)
	byUser, err := s.repo.List(s.ctx, port.ExpenseFilter{UserID: &uid, Limit: 10})
	s.Require().NoError(err)
	s.Require().Len(byUser, 2)

	cat := models.CategoryFood
	byCategory, err := s.repo.List(s.ctx, port.ExpenseFilter{Category: &cat, Limit: 10})
	s.Require().NoError(err)
	s.Require().Len(byCategory, 2)

	from := now.Add(-2*time.Hour - 30*time.Minute)
	to := now
	byDate, err := s.repo.List(s.ctx, port.ExpenseFilter{From: &from, To: &to, Limit: 10})
	s.Require().NoError(err)
	s.Require().Len(byDate, 2)

	paginated, err := s.repo.List(s.ctx, port.ExpenseFilter{Limit: 1, Offset: 1})
	s.Require().NoError(err)
	s.Require().Len(paginated, 1)
	s.Require().Equal(e2.ID(), paginated[0].ID())
}

func (s *ExpensesRepoSuite) TestListLatest() {
	now := time.Now().UTC()

	e1, err := models.NewExpense(100, 10.0, "еда", now.Add(-3*time.Hour), "oldest")
	s.Require().NoError(err)
	_, err = s.repo.Create(s.ctx, e1)
	s.Require().NoError(err)

	e2, err := models.NewExpense(100, 20.0, "транспорт", now.Add(-2*time.Hour), "middle")
	s.Require().NoError(err)
	_, err = s.repo.Create(s.ctx, e2)
	s.Require().NoError(err)

	e3, err := models.NewExpense(100, 30.0, "покупки", now.Add(-1*time.Hour), "latest")
	s.Require().NoError(err)
	_, err = s.repo.Create(s.ctx, e3)
	s.Require().NoError(err)

	latest, err := s.repo.ListLatest(s.ctx, 100, 2)
	s.Require().NoError(err)
	s.Require().Len(latest, 2)
	s.Require().Equal(e3.ID(), latest[0].ID())
	s.Require().Equal(e2.ID(), latest[1].ID())
}

func (s *ExpensesRepoSuite) TestListUserIDS() {
	now := time.Now().UTC()

	e1, err := models.NewExpense(100, 10.0, "еда", now.Add(-2*time.Hour), "user 100")
	s.Require().NoError(err)
	_, err = s.repo.Create(s.ctx, e1)
	s.Require().NoError(err)

	e2, err := models.NewExpense(200, 20.0, "транспорт", now.Add(-1*time.Hour), "user 200")
	s.Require().NoError(err)
	_, err = s.repo.Create(s.ctx, e2)
	s.Require().NoError(err)

	e3, err := models.NewExpense(100, 30.0, "покупки", now, "user 100 again")
	s.Require().NoError(err)
	_, err = s.repo.Create(s.ctx, e3)
	s.Require().NoError(err)

	userIDs, err := s.repo.ListUserIDS(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal([]int64{100, 200}, userIDs)
}
