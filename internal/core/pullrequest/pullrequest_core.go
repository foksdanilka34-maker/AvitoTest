package pullrequest

import (
	"AvitoTest/internal/models"
	repo "AvitoTest/internal/repository/pullrequest"
	"fmt"
	"log"
	"slices"

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
	if pulReq.PullReqName == "" {
		return nil, fmt.Errorf("empty")
	}

	if slices.Contains(pulReq.AssignedReviewes, pulReq.AuthorID) {
		log.Println("author cannot be reviewer")
		return nil, fmt.Errorf("author cannot be reviewer")
	}

	if len(pulReq.AssignedReviewes) == 0 {
		log.Println("pull Request short called")
		return p.rp.CreatePullRequestShort(ctx, pulReq)
	} else {
		log.Println("pull Request called")
		return p.rp.CreatePullRequest(ctx, pulReq)
	}
}