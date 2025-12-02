package team

import (
	"AvitoTest/internal/core/team"
	"AvitoTest/internal/models"
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
	mux.HandleFunc("GET /team/get", h.GetTeam)
	mux.HandleFunc("POST /users/setIsActive", h.SetUserStatus)

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

	res, err := h.core.TeamAdd(r.Context(), request.TeamName, request.Members)
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

func (h *TeamsHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		message := &models.Response{
			Message: "bad data request",
		}
		json.NewEncoder(w).Encode(message)
		return
	}
	res, err := h.core.GetTeam(r.Context(), teamName)
	if err != nil {
		log.Println(err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		message := &models.Response{
			Message: "users not found",
		}
		json.NewEncoder(w).Encode(message)
		return
	}
	response := &models.CreateTeamResponse{
		TeamName: res.TeamName,
		Users: res.Users,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *TeamsHandler) SetUserStatus(w http.ResponseWriter, r *http.Request) {
	req := &models.ChangeUserStatus{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		message := &models.Response{
			Message: "bad data request",
		}
		json.NewEncoder(w).Encode(message)
		return
	}
	user, err := h.core.SetUserStatus(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		message := &models.Response{
			Message: "user not found",
		}
		json.NewEncoder(w).Encode(message)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}