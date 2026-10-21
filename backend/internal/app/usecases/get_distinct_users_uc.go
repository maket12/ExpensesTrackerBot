package usecases

import (
	"context"

	"backend/internal/app/dto"
	"backend/internal/app/mapper"
	"backend/internal/domain/port"
)

type GetDistinctUsersUC struct {
	repo port.ExpenseRepository
}

func NewGetDistinctUsersUC(repo port.ExpenseRepository) *GetDistinctUsersUC {
	return &GetDistinctUsersUC{
		repo: repo,
	}
}

func (uc *GetDistinctUsersUC) Execute(ctx context.Context, _ dto.GetDistinctUsersInput) (dto.GetDistinctUsersOutput, error) {
	userIDs, err := uc.repo.ListUserIDS(ctx)
	if err != nil {
		return dto.GetDistinctUsersOutput{}, err
	}

	return mapper.ToGetDistinctUsersOutput(userIDs), nil
}
