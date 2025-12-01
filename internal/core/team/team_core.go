package team

import (
	"context"

	"AvitoTest/internal/models"
	repo "AvitoTest/internal/repository/team"

)

type team struct {
	tRepo *repo.Team
}

type TeamLogic interface {
	TeamAdd(ctx context.Context, teamName string, users []*models.Users) (*models.Teams, error)
	GetTeam(ctx context.Context, teamName string) (*models.Teams, error)
}

func NewTeam(tRepo *repo.Team) TeamLogic {
	return &team{
		tRepo: tRepo,
	}
}

func (t *team) TeamAdd(ctx context.Context, teamName string, users []*models.Users) (*models.Teams, error) {
	return t.tRepo.CreateTeam(ctx, teamName, users)
}

func (t *team) GetTeam(ctx context.Context, teamName string) (*models.Teams, error) {
	return t.tRepo.GetTeam(ctx, teamName)
}