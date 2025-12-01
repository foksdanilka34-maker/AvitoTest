package team

import (
	"AvitoTest/internal/core/team"
	"AvitoTest/internal/models"
	"context"
	"fmt"
	"log"

	"encoding/json"
	"net/http"
)

type TeamsHandler struct {
	core team.TeamLogic
}

func NewTeamsHandler(c team.TeamLogic) *TeamsHandler {
	return &TeamsHandler{
		core: c,
	}
}

func (h *TeamsHandler) InitRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /team/add", h.AddTeam)

	return mux
}

func (h *TeamsHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	request := &models.CreateTeamRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		message := &models.Response{
			Message: "bad data request",
		}
		json.NewEncoder(w).Encode(message)
		return
	}
	fmt.Println(request.TeamName, request.Members)

	res, err := h.core.TeamAdd(context.Background(), request.TeamName, request.Members)
	if err != nil {
		log.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		message := &models.Response{
			Message: "user already exists",
		}
		json.NewEncoder(w).Encode(message)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := &models.CreateTeamResponse{
		TeamName: res.TeamName,
		Users: res.Users,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

