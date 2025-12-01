package team

import (
	"context"

	repo "AvitoTest/internal/repository/team"
	"AvitoTest/internal/models"
)

type team struct {
	tRepo *repo.Team
}

type TeamLogic interface {
	TeamAdd(ctx context.Context, teamName string, users []*models.Users) (*models.Teams, error)
}

func NewTeam(tRepo *repo.Team) TeamLogic {
	return &team{
		tRepo: tRepo,
	}
}

func (t *team) TeamAdd(ctx context.Context, teamName string, users []*models.Users) (*models.Teams, error) {
	return t.tRepo.CreateTeam(ctx, teamName, users)
}