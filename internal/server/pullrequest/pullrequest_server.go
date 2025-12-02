package pullrequest

import (
	"AvitoTest/internal/core/pullrequest"
	"AvitoTest/internal/models"
	"fmt"

	"log"

	"encoding/json"
	"net/http"
)

type PullRequestHandler struct {
	core pullrequest.PullRequest
}

func NewTeamsHandler(c pullrequest.PullRequest) *PullRequestHandler {
	return &PullRequestHandler{
		core: c,
	}
}

func (h *PullRequestHandler) ReqisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /pullRequest/create", h.CreatePullRequest)
}

func (h *PullRequestHandler) CreatePullRequest(w http.ResponseWriter, r *http.Request) {
	pulRequest := &models.PullRequest{}
	fmt.Println(pulRequest)
	if err := json.NewDecoder(r.Body).Decode(pulRequest); err != nil {
		log.Println(err)
		http.Error(w, "error decoding json", http.StatusBadRequest)
		return
	}

	result, err := h.core.CreatePullRequest(r.Context(), pulRequest)
	if err != nil {
		log.Println(err)
		http.Error(w, "error", http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}