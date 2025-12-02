package pullrequest

import (
	"AvitoTest/internal/models"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PullReq struct {
	pgx *pgxpool.Pool
}

func NewTeam(pgx *pgxpool.Pool) *PullReq {
	return &PullReq{
		pgx: pgx,
	}
}

func (p *PullReq) CreatePullRequestShort(ctx context.Context, pulReq models.PullRequest) {
	queryPullRequestShort := `INSERT INTO pull_request (request_name, creator_id, status)
							VALUES ($1, (SELECT user_id FROM users WHERE user_id IN($2) AND is_active = TRUE), $3)
							`
}