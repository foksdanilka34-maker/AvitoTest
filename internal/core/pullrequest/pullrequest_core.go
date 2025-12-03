package pullrequest

import (
	"AvitoTest/internal/models"
	repo "AvitoTest/internal/repository/pullrequest"
	"fmt"
	"log"
	"slices"

	"context"

	"github.com/gofrs/uuid"
)

type pullReq struct {
	rp *repo.PullReq
}

type PullRequest interface {
	CreatePullRequest(ctx context.Context, pulReq *models.PullRequest) (*models.PullRequest, error)
	MergePullRequest(ctx context.Context, pullReqID uuid.UUID) (*models.PullRequestMerge, error)
	ReassignPullRequest(ctx context.Context, reassignReq *models.ReassignPullRequest) (*models.ReassignPullRequestResponse, error)
	GetReview(ctx context.Context, reviewerID uuid.UUID) (*models.GetPullRequestReviewResponse, error)
}

func NewPullReq(rp *repo.PullReq) PullRequest {
	return &pullReq{
		rp: rp,
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

func (p *pullReq) MergePullRequest(ctx context.Context, pullReqID uuid.UUID) (*models.PullRequestMerge, error) {
	return p.rp.MergePullRequest(ctx, pullReqID)
}

func (p *pullReq) ReassignPullRequest(ctx context.Context, reassignReq *models.ReassignPullRequest) (*models.ReassignPullRequestResponse, error) {
	return p.rp.ReassignPullRequest(ctx, reassignReq)
}

func (p *pullReq) GetReview(ctx context.Context, reviewerID uuid.UUID) (*models.GetPullRequestReviewResponse, error) {
	return p.rp.GetReview(ctx, reviewerID)
}
