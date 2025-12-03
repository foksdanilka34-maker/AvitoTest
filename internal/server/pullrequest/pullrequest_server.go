package pullrequest

import (
	"AvitoTest/internal/core/pullrequest"
	"AvitoTest/internal/models"

	"log"

	"encoding/json"
	"net/http"

	"github.com/gofrs/uuid"
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
	mux.HandleFunc("POST /pullRequest/merge", h.MergePullRequest)
	mux.HandleFunc("POST /pullRequest/reassign", h.ReassignPullRequest)

	mux.HandleFunc("GET /users/getReview", h.GetReviewersRequests)
	mux.HandleFunc("GET /users/stats", h.GetStats)
}

func (h *PullRequestHandler) CreatePullRequest(w http.ResponseWriter, r *http.Request) {
	pulRequest := &models.PullRequest{}
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

func (h *PullRequestHandler) MergePullRequest(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Query().Get("pull_request_id")
	if param == "" {
		return
	}

	result, err := h.core.MergePullRequest(r.Context(), uuid.FromStringOrNil(param))
	if err != nil {
		log.Println("ERROR", err)
		http.Error(w, "failed merge request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (h *PullRequestHandler) ReassignPullRequest(w http.ResponseWriter, r *http.Request) {
	reassignRequest := &models.ReassignPullRequest{}
	if err := json.NewDecoder(r.Body).Decode(reassignRequest); err != nil {
		http.Error(w, "failed to decod query", http.StatusBadRequest)
		return
	}

	result, err := h.core.ReassignPullRequest(r.Context(), reassignRequest)
	if err != nil {
		log.Println("REASSIGN ERROR ", err)
		http.Error(w, "faile to reassign user", http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (h *PullRequestHandler) GetReviewersRequests(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Query().Get("reviewer_id")
	if param == "" {
		return
	}

	result, err := h.core.GetReview(r.Context(), uuid.FromStringOrNil(param))
	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (h *PullRequestHandler) GetStats (w http.ResponseWriter, r *http.Request) {
	result, err := h.core.GetStats(r.Context())
	if err != nil {
		http.Error(w, "bd", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}