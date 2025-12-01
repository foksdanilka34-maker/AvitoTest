package team

import (
	"context"
	"fmt"

	"log"

	"AvitoTest/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Team struct {
	pgx *pgxpool.Pool
}

func NewTeam(pgx *pgxpool.Pool) *Team {
	return &Team{
		pgx: pgx,
	}
}

func (t *Team) CreateTeam(ctx context.Context, teamName string, users []*models.Users) (*models.Teams, error) {
	fmt.Println(users)
	queryCreateTeam := `INSERT INTO teams (team_name) VALUES ($1)
						ON CONFLICT (team_name) DO
						UPDATE SET team_name = EXCLUDED.team_name
						RETURNING team_id`
	teamResult := models.Teams{}
	tx, err := t.pgx.Begin(ctx)
	if err != nil {
		log.Println("failed to begin tx", err)
		return nil, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, queryCreateTeam, teamName).Scan(&teamResult.TeamID)
	if err != nil {
		log.Println("failed to tx, insert team error")
		return nil, err
	}

	batch := &pgx.Batch{}
	queryCreateOrUpdateMembers :=
		`INSERT INTO users (user_name, is_active, team_id) VALUES ($1, $2, $3)
		ON CONFLICT (user_name) DO UPDATE SET 
			user_name = EXCLUDED.user_name,
			is_active = EXCLUDED.is_active,
			team_id = EXCLUDED.team_id
		RETURNING user_id, user_name, is_active`

	for _, user := range users {
		batch.Queue(queryCreateOrUpdateMembers, user.UserName, user.IsActive, teamResult.TeamID)
	}

	batchResult := tx.SendBatch(ctx, batch)
	defer batchResult.Close()

	result := make([]*models.Users, 0, len(users))
	for range users {
		user := &models.Users{}
		err = batchResult.QueryRow().Scan(&user.UserID, &user.UserName, &user.IsActive)
		if err != nil {
			batchResult.Close()
			log.Println("failed to batch results", err)
			return nil, fmt.Errorf("failed to batch results")
		}
		fmt.Println(user)
		result = append(result, user)
	}

	batchResult.Close()
	teamResult.Users = result
	if err := tx.Commit(ctx); err != nil {
		log.Println(err)
		return nil, fmt.Errorf("failed to commit transaction")
	}

	return &teamResult, nil
}
