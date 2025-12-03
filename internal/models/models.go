package models

import (
	"github.com/gofrs/uuid"
	"time"
)

const (
	TEAM_EXISTS = "TEAM_EXISTS"
	PR_EXISTS
	PR_MERGED
	MERGED = "MERGED"
	NOT_ASSIGNED
	NO_CANDIDATE
	NOT_FOUND
)

type Users struct {
	TeamID   uuid.UUID `json:"-"`
	UserID   uuid.UUID `json:"user_id"`
	UserName string    `json:"user_name"`
	IsActive bool      `json:"is_active"`
}

type Teams struct {
	TeamID   uuid.UUID `json:"team_id"`
	TeamName string    `json:"team_name"`
	Users    []*Users  `json:"users"`
}

type CreateTeamRequest struct {
	TeamName string   `json:"team_name"`
	Members  []*Users `json:"users"`
}

type Response struct {
	Message string `json:"message"`
}

type CreateTeamResponse struct {
	TeamName string   `json:"team_name"`
	Users    []*Users `json:"users"`
}

type UserSetResponse struct {
	UserID   uuid.UUID `json:"user_id"`
	TeamName string    `json:"team_name"`
	UserName string    `json:"user_name"`
	IsActive bool      `json:"is_active"`
}

type ChangeUserStatus struct {
	UserID   uuid.UUID `json:"user_id"`
	IsActive bool      `json:"is_active"`
}

type PullRequest struct {
	PullReqID        uuid.UUID   `json:"pull_request_id,omitempty"`
	AuthorID         uuid.UUID   `json:"author_id"`
	PullReqName      string      `json:"pull_request_name"`
	Status           string      `json:"status,omitempty"`
	AssignedReviewes []uuid.UUID `json:"assigned_reviewers,omitempty"`
	CreatedAt        *time.Time  `json:"-"`
}

type PullRequestMerge struct {
	PullReqID        uuid.UUID   `json:"pull_request_id,omitempty"`
	AuthorID         uuid.UUID   `json:"author_id"`
	PullReqName      string      `json:"pull_request_name"`
	Status           string      `json:"status,omitempty"`
	AssignedReviewes []uuid.UUID `json:"assigned_reviewers,omitempty"`
	MergedAt         *time.Time  `json:"merged_at,omitempty"`
}

type ReassignPullRequest struct {
	PullRequestID uuid.UUID `json:"pull_request_id"`
	OldUserID     uuid.UUID `json:"old_user_id"`
}

type ReassignPullRequestResponse struct {
	PullRequestID    uuid.UUID   `json:"pull_request_id"`
	ReplacedByID     uuid.UUID   `json:"replaced_id"`
	AuthorID         uuid.UUID   `json:"author_id"`
	Status           string      `json:"status,omitempty"`
	AssignedReviewes []uuid.UUID `json:"assigned_reviewers,omitempty"`
}

type GetPullRequestReview struct {
	PullReqID   uuid.UUID `json:"pull_request_id,omitempty"`
	AuthorID    uuid.UUID `json:"author_id"`
	PullReqName string    `json:"pull_request_name"`
	Status      string    `json:"status,omitempty"`
}

type GetPullRequestReviewResponse struct {
	PullRequests []*GetPullRequestReview `json:"pull_requests"`
}

type Stats struct {
	Name string `json:"name"`
	OpenTasks int `json:"open_tasks"`
	MergedTasks int `json:"merged_tasks"`
	TotalTask int `json:"total_tasks"`
	MergedTasksRate float64 `json:"merged_rate"`
}

type GetStatsResponse struct {
	Stats []*Stats `json:"stats_response"`
}
