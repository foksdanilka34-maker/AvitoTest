package pullrequest

import (
	
	"AvitoTest/internal/models"
	repo "AvitoTest/internal/repository/pullrequest"

	"context"
)

type pullReq struct {
	rp *repo.PullReq
}

type PullRequest interface {
	CreatePullRequest(ctx context.Context, pulReq *models.PullRequest) (*models.PullRequest, error)
}

func NewPullReq(rp *repo.PullReq) PullRequest {
	return &pullReq{
		rp : rp,
	}
}

func (p *pullReq) CreatePullRequest(ctx context.Context, pulReq *models.PullRequest) (*models.PullRequest, error) {
	return p.rp.CreatePullRequest(ctx, pulReq)
}