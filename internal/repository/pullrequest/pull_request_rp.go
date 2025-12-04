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
	_ "github.com/jackc/pgx/v5/pgtype"
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
	defer rows.Close()

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

func (p *PullReq) MergePullRequest(ctx context.Context, pullRequestID uuid.UUID) (*models.PullRequestMerge, error) {
	queryMergeRequest := `UPDATE pull_requests SET status = 'MERGED',
							merged_at = COALESCE(merged_at, now())
							WHERE request_id = $1
							RETURNING request_id, request_name, creator_id, 
							status, merged_at
						`
	tx, err := p.pgx.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to beign transaction %v", err)
	}
	defer tx.Rollback(ctx)

	mergeRequest := &models.PullRequestMerge{}
	err = tx.QueryRow(ctx, queryMergeRequest, pullRequestID).Scan(
		&mergeRequest.PullReqID,
		&mergeRequest.PullReqName,
		&mergeRequest.AuthorID,
		&mergeRequest.Status,
		&mergeRequest.MergedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error durung merging request %v", err)
	}

	queryMergeRequestSelectReviewers := `SELECT reviewer_id FROM pr_reviewers 
										WHERE request_id = ($1)`

	rows, err := tx.Query(ctx, queryMergeRequestSelectReviewers, pullRequestID)
	if err != nil {
		return nil, fmt.Errorf("query failed %v", err)
	}
	for rows.Next() {
		var reviewerId uuid.UUID
		err = rows.Scan(&reviewerId)
		if err != nil {
			return nil, fmt.Errorf("failed to scan %v", err)
		}
		mergeRequest.AssignedReviewes = append(mergeRequest.AssignedReviewes, reviewerId)
	}
	defer rows.Close()

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit %v", err)
	}

	return mergeRequest, nil
}

func (p *PullReq) ReassignPullRequest(ctx context.Context, reassignReq *models.ReassignPullRequest) (*models.ReassignPullRequestResponse, error) {
	querySelectStatus := `SELECT creator_id, status FROM pull_requests WHERE request_id = $1`

	tx, err := p.pgx.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	var prAuthor uuid.UUID
	var status string
	err = tx.QueryRow(ctx, querySelectStatus, reassignReq.PullRequestID).Scan(&prAuthor, &status)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull_request")
	}

	if status == models.MERGED {
		return nil, fmt.Errorf("request is already merged")
	}

	queryReassignUser := `SELECT u.user_id
						FROM users u
						JOIN pull_requests pr ON pr.creator_id = (
							SELECT creator_id
							FROM pull_requests
							WHERE request_id = ($1)
						)
						WHERE u.team_id = (
							SELECT team_id
							FROM users
							WHERE user_id = pr.creator_id
						)
						AND u.user_id != pr.creator_id 
						AND u.user_id != ($2)
						AND u.is_active = TRUE
						AND u.user_id NOT IN (
						SELECT reviewer_id
						FROM pr_reviewers
						WHERE request_id = ($1)
						)
						ORDER BY random()
						LIMIT 1;`

	var newReviewerID uuid.UUID
	err = tx.QueryRow(ctx, queryReassignUser, reassignReq.PullRequestID, reassignReq.OldUserID).Scan(&newReviewerID)

	if err != nil {
		return nil, fmt.Errorf("error choosing reassigned users %v", err)
	}

	queryUpdateUser := `UPDATE pr_reviewers
						SET reviewer_id = $1,
						assigned_at = NOW()
						WHERE request_id = $2 AND reviewer_id = $3`

	tag, err := tx.Exec(ctx, queryUpdateUser, newReviewerID, reassignReq.PullRequestID, reassignReq.OldUserID)

	if err != nil {
		return nil, fmt.Errorf("error updating reviewer %v", err)
	}

	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("no rows were updated")
	}

	querySelectAssignedReviewers := `SELECT reviewer_id FROM pr_reviewers
									WHERE request_id = $1`

	rows, err := tx.Query(ctx, querySelectAssignedReviewers, reassignReq.PullRequestID)
	if err != nil {
		return nil, fmt.Errorf("error selecting users %v", err)
	}
	defer rows.Close()

	assignedReviewers := []uuid.UUID{}
	for rows.Next() {
		var reviewerID uuid.UUID
		err = rows.Scan(&reviewerID)
		if err != nil {
			return nil, fmt.Errorf("error scanning user %v", err)
		}
		assignedReviewers = append(assignedReviewers, reviewerID)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error commiting reassign %v", err)
	}

	reassignResponse := &models.ReassignPullRequestResponse{
		PullRequestID:    reassignReq.PullRequestID,
		ReplacedByID:     newReviewerID,
		AuthorID:         prAuthor,
		Status:           status,
		AssignedReviewes: assignedReviewers,
	}

	return reassignResponse, nil
}

func (p *PullReq) GetReview(ctx context.Context, reviewerID uuid.UUID) (*models.GetPullRequestReviewResponse, error) {
	getPullRequestReview := `SELECT request_id, creator_id, request_name, status 
							FROM pull_requests
							WHERE request_id IN (
								SELECT request_id FROM pr_reviewers WHERE
								reviewer_id = ($1))`

	rows, err := p.pgx.Query(ctx, getPullRequestReview, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user pr %v", err)
	}

	reviewerPR := &models.GetPullRequestReviewResponse{}
	for rows.Next() {
		userReviews := &models.GetPullRequestReview{}
		err = rows.Scan(
			&userReviews.PullReqID,
			&userReviews.AuthorID,
			&userReviews.PullReqName,
			&userReviews.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pr's %v", err)
		}
		reviewerPR.PullRequests = append(reviewerPR.PullRequests, userReviews)
	}

	return reviewerPR, nil
}

func (p *PullReq) GetStats(ctx context.Context) (*models.GetStatsResponse, error) {
	queryStats := `SELECT 
					u.user_name,
					COUNT(CASE WHEN pr.status = 'OPEN' THEN 1 END) AS open_tasks,
					COUNT(CASE WHEN pr.status = 'MERGED' THEN 1 END) AS merged_tasks,
					COUNT(*) AS total_tasks
					FROM pr_reviewers r
					JOIN pull_requests pr ON pr.request_id = r.request_id
					JOIN users u ON r.reviewer_id = u.user_id
					WHERE pr.status IN ('OPEN', 'MERGED')  
					GROUP BY u.user_name`

	rows, err := p.pgx.Query(ctx, queryStats)
	if err != nil {
		return nil, fmt.Errorf("error quuery stat scan %v", err)
	}

	response := &models.GetStatsResponse{}
	for rows.Next() {
		stats := &models.Stats{}
		err = rows.Scan(
			&stats.Name,
			&stats.OpenTasks,
			&stats.MergedTasks,
			&stats.TotalTask,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning stats %v", err)
		}
		stats.MergedTasksRate = float64(stats.MergedTasks) / float64(stats.TotalTask) * 100
		response.Stats = append(response.Stats, stats)
	}
	return response, nil
}

func (p *PullReq) DeactivateMembers(ctx context.Context, membersID *models.DeactivateMembers) (*models.TotalDeactivate, error) {
	tx, err := p.pgx.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction %v", err)
	}
	defer tx.Rollback(ctx)

	queryDeactivateAndSelect := `
        WITH deactivated AS (
            UPDATE users SET is_active = false
            WHERE user_id = ANY($1)
            AND team_id = (SELECT team_id FROM teams WHERE team_name = $2)
            AND is_active = TRUE
            RETURNING user_id, user_name
        )
        SELECT d.user_id, d.user_name, pr.request_id, pr.creator_id
        FROM deactivated d
        JOIN pr_reviewers prr ON prr.reviewer_id = d.user_id
        JOIN pull_requests pr ON pr.request_id = prr.request_id
        WHERE pr.status = 'OPEN'`

	rows, err := tx.Query(ctx, queryDeactivateAndSelect, membersID.Members, membersID.TeamName)
	if err != nil {
		return nil, fmt.Errorf("error deactivate and select %v", err)
	}

	type affectedPR struct {
		userID    uuid.UUID
		userName  string
		requestID uuid.UUID
		creatorID uuid.UUID
	}

	affectedPRs := []affectedPR{}
	for rows.Next() {
		var apr affectedPR
		err = rows.Scan(&apr.userID, &apr.userName, &apr.requestID, &apr.creatorID)
		if err != nil {
			return nil, fmt.Errorf("error scanning affected pr %v", err)
		}
		affectedPRs = append(affectedPRs, apr)
	}
	rows.Close()

	if len(affectedPRs) == 0 {
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("error commiting deactivate %v", err)
		}
		return &models.TotalDeactivate{
			TeamName:     membersID.TeamName,
			Replacements: []*models.ReassignResult{},
		}, nil
	}

	querySelectNewReviewer := `
        SELECT u.user_id
        FROM users u
        WHERE u.team_id = (SELECT team_id FROM users WHERE user_id = $1)
        AND u.user_id != $1
        AND u.user_id != $2
        AND u.is_active = TRUE
        AND u.user_id NOT IN (
            SELECT reviewer_id FROM pr_reviewers WHERE request_id = $3
        )
        ORDER BY random()
        LIMIT 1`

	selectBatch := &pgx.Batch{}
	for _, apr := range affectedPRs {
		selectBatch.Queue(querySelectNewReviewer, apr.creatorID, apr.userID, apr.requestID)
	}

	selectResults := tx.SendBatch(ctx, selectBatch)

	newReviewers := make([]uuid.UUID, len(affectedPRs))
	for i := range affectedPRs {
		var newReviewerID uuid.UUID
		err = selectResults.QueryRow().Scan(&newReviewerID)
		if err != nil {
			selectResults.Close()
			return nil, fmt.Errorf("error selecting new reviewer for PR %v: %v", affectedPRs[i].requestID, err)
		}
		newReviewers[i] = newReviewerID
	}
	selectResults.Close()

	queryUpdateReviewer := `
        UPDATE pr_reviewers
        SET reviewer_id = $1, assigned_at = NOW()
        WHERE request_id = $2 AND reviewer_id = $3`

	updateBatch := &pgx.Batch{}
	for i, apr := range affectedPRs {
		updateBatch.Queue(queryUpdateReviewer, newReviewers[i], apr.requestID, apr.userID)
	}

	updateResults := tx.SendBatch(ctx, updateBatch)

	for i := range affectedPRs {
		tag, err := updateResults.Exec()
		if err != nil {
			updateResults.Close()
			return nil, fmt.Errorf("error updating reviewer %v", err)
		}
		if tag.RowsAffected() == 0 {
			updateResults.Close()
			return nil, fmt.Errorf("no rows updated for PR %v", affectedPRs[i].requestID)
		}
	}
	updateResults.Close()

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("error commiting deactivate %v", err)
	}

	resDeactiv := &models.TotalDeactivate{
		TeamName:     membersID.TeamName,
		Replacements: make([]*models.ReassignResult, len(affectedPRs)),
	}

	for i, apr := range affectedPRs {
		resDeactiv.Replacements[i] = &models.ReassignResult{
			OldUser:   apr.userID,
			NewUser:   newReviewers[i],
			RequestID: apr.requestID,
		}
	}

	return resDeactiv, nil
}
