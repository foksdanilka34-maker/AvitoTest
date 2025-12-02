package pullrequest

import (
	"AvitoTest/internal/models"
	"context"
	"fmt"
	"log"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PullReq struct {
	pgx *pgxpool.Pool
}

func NewPullReq(pgx *pgxpool.Pool) *PullReq {
	return &PullReq{
		pgx: pgx,
	}
}

func (p *PullReq) CreatePullRequest(ctx context.Context, pulReq *models.PullRequest) (*models.PullRequest, error) {
	queryPullRequest := `INSERT INTO pull_requests (request_name, creator_id)
							VALUES ($1, (SELECT user_id FROM users WHERE user_id = ($2) AND is_active = TRUE))
							RETURNING request_id, request_name, creator_id, status, created_at`

	tx, err := p.pgx.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error begin tx %v", err)
	}
	defer tx.Rollback(ctx)

	pulReqResult := &models.PullRequest{}
	err = tx.QueryRow(ctx, queryPullRequest, pulReq.PullReqName, pulReq.AuthorID).Scan(
		&pulReqResult.PullReqID,
		&pulReqResult.PullReqName,
		&pulReq.AuthorID,
		&pulReq.Status,
		&pulReq.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	queryPullRequestReviewers := `INSERT INTO pr_reviewers (reviewer_id) VALUES (
								SELECT user_id FROM users WHERE user_id IN ($1) AND is_active = TRUE)
								RETURNING request_id, reviewer_id)`
	
	rows, err := tx.Query(ctx, queryPullRequestReviewers, pulReq.AssignedReviewes)
	if err != nil {
		return nil, fmt.Errorf("error reviewers tx %v", err)
	}

	foundUsers := 0
	for rows.Next() {
		var requestID *uuid.UUID
		var reviewer_id *uuid.UUID

		err = rows.Scan(&requestID, &reviewer_id)
		if err != nil {
			return nil, fmt.Errorf("error scanning reviewers %v", err)
		}
		pulReqResult.AssignedReviewes[foundUsers] = reviewer_id
		foundUsers++
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		return nil, fmt.Errorf("failed to commit transaction")
	}

	return pulReqResult, nil
}