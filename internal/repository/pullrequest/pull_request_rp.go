package pullrequest

import (
	"AvitoTest/internal/models"
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

	queryPullRequestReviewers := `
			INSERT INTO pr_reviewers (request_id, reviewer_id)
			SELECT $3, u.user_id
			FROM users u
			JOIN users creator ON creator.user_id = $4
			WHERE u.user_id IN ($1, $2)
			AND u.is_active = TRUE
			AND u.team_id = creator.team_id
			RETURNING request_id, reviewer_id
	`

	rows, err := tx.Query(ctx, queryPullRequestReviewers,
		pulReq.AssignedReviewes[0],
		pulReq.AssignedReviewes[1],
		pulReqResult.PullReqID,
		pulReq.AuthorID,

	)
	if err != nil {
		return nil, fmt.Errorf("error reviewers tx %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var requestID uuid.UUID
		var reviewerID uuid.UUID

		err = rows.Scan(&requestID, &reviewerID)
		if err != nil {
			return nil, fmt.Errorf("error scanning reviewers %v", err)
		}
		pulReqResult.AssignedReviewes = append(pulReqResult.AssignedReviewes, reviewerID)
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		return nil, fmt.Errorf("failed to commit transaction")
	}

	return pulReqResult, nil
}

func (p *PullReq) CreatePullRequestShort(ctx context.Context, pulReq *models.PullRequest) (*models.PullRequest, error) {
	queryPullRequestInsertAuthor := `INSERT INTO pull_requests (request_name, creator_id)
									VALUES ($1, (SELECT user_id FROM users WHERE user_id = ($2) AND is_active = TRUE))
									RETURNING request_id, request_name, creator_id, status, created_at`
	tx, err := p.pgx.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("tx error %v", err)
	}
	defer tx.Rollback(ctx)

	pulReqResult := &models.PullRequest{}
	err = tx.QueryRow(ctx, queryPullRequestInsertAuthor, pulReq.PullReqName, pulReq.AuthorID).Scan(
		&pulReqResult.PullReqID,
		&pulReqResult.PullReqName,
		&pulReqResult.AuthorID,
		&pulReqResult.Status,
		&pulReqResult.CreatedAt)
	
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return nil, fmt.Errorf("a pull request with this name is already open")
			}
		}
		return nil, fmt.Errorf("insert prS error %v", err)
	}

	queryPullRequestChooseReviewers := `INSERT INTO pr_reviewers(request_id, reviewer_id)
										SELECT $1, u.user_id FROM users u 
										JOIN users creator ON creator.user_id = $2
										WHERE u.team_id = creator.team_id
										AND u.is_active = TRUE AND u.user_id != creator.user_id
										ORDER BY random()
										LIMIT 2
										RETURNING reviewer_id`

	rows, err := tx.Query(ctx, queryPullRequestChooseReviewers, pulReqResult.PullReqID, pulReq.AuthorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Println("no data were returned")
		}
		return nil, fmt.Errorf("error choosing reviewers %v", err)
	}

	for rows.Next() {
		var reviewer_id uuid.UUID
		err = rows.Scan(&reviewer_id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				log.Println("empty scanning chosen reviewers")
				continue
			} else {
				return nil, fmt.Errorf("hoho %v", err)
			}
		}
		log.Println("get reviewer id - ", reviewer_id)
		pulReqResult.AssignedReviewes = append(pulReqResult.AssignedReviewes, reviewer_id)
	}

	if len(pulReqResult.AssignedReviewes) == 0 {
		return nil, fmt.Errorf("no reviewers were assigned to pr")
	}
	
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error while comittin %v", err)
	}

	return pulReqResult, nil
}

func (p *PullReq) MergePullRequest(ctx context.Context, pullRequestID uuid.UUID) (error, error){
	queryMergeRequest := `UPDATE pull_requests SET status = 'MERGED',
							merget_at = COALESCE(merged_at, now())
							WHERE request_id = $1
							RETURNING request_id, request_name, creator_id, 
							status, created_at, merged_at
						`
	tx, err := p.pgx.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to beign transaction %v", err)
	}

	mergeRequest := &models.PullRequest{}
	err = tx.QueryRow(ctx, queryMergeRequest, pullRequestID).Scan(
		mergeRequest.PullReqID,
		mergeRequest.PullReqName,
		mergeRequest.AuthorID,
		mergeRequest.Status,
		mergeRequest.CreatedAt,
		mergeRequestM
	)
}