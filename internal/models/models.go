package models

import (
	"github.com/gofrs/uuid"
)


type Users struct {
	UserID 		uuid.UUID  `json:"user_id,omitempty"`
	UserName 	string	   `json:"user_name"`
	IsActive 	bool	   `json:"is_active"`
	TeamID 		uuid.UUID  `json:"team_id,omitempty"`
}

type Teams struct {
	TeamID 		uuid.UUID  `json:"team_id"`
	TeamName 	string	   `json:"team_name"`
	Users 		[]*Users    `json:"users"`
}

type CreateTeamRequest struct {
	TeamName string 	`json:"team_name"`
	Members	 []*Users   `json:"users"`
}

type Response struct {
	Message  string `json:"message"`
}

type CreateTeamResponse struct {
    TeamName string   `json:"team_name"`
    Users    []*Users `json:"users"`
}