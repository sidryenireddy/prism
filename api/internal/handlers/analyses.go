package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sidryenireddy/prism/api/internal/models"
)

type AnalysisHandler struct {
	db *pgxpool.Pool
}

func NewAnalysisHandler(db *pgxpool.Pool) *AnalysisHandler {
	return &AnalysisHandler{db: db}
}

func (h *AnalysisHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(),
		"SELECT id, name, description, owner, share_token, created_at, updated_at FROM analyses ORDER BY updated_at DESC")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var analyses []models.Analysis
	for rows.Next() {
		var a models.Analysis
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.Owner, &a.ShareToken, &a.CreatedAt, &a.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		analyses = append(analyses, a)
	}
	if analyses == nil {
		analyses = []models.Analysis{}
	}

	writeJSON(w, http.StatusOK, analyses)
}

func (h *AnalysisHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid analysis id")
		return
	}

	var a models.Analysis
	err = h.db.QueryRow(r.Context(),
		"SELECT id, name, description, owner, share_token, created_at, updated_at FROM analyses WHERE id = $1", id).
		Scan(&a.ID, &a.Name, &a.Description, &a.Owner, &a.ShareToken, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusNotFound, "analysis not found")
		return
	}

	writeJSON(w, http.StatusOK, a)
}

func (h *AnalysisHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	shareToken := uuid.New().String()[:8]

	a := models.Analysis{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Owner:       req.Owner,
		ShareToken:  shareToken,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err := h.db.Exec(r.Context(),
		"INSERT INTO analyses (id, name, description, owner, share_token, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		a.ID, a.Name, a.Description, a.Owner, a.ShareToken, a.CreatedAt, a.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, a)
}

func (h *AnalysisHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid analysis id")
		return
	}

	var req models.UpdateAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var a models.Analysis
	err = h.db.QueryRow(r.Context(),
		"SELECT id, name, description, owner, share_token, created_at, updated_at FROM analyses WHERE id = $1", id).
		Scan(&a.ID, &a.Name, &a.Description, &a.Owner, &a.ShareToken, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusNotFound, "analysis not found")
		return
	}

	if req.Name != nil {
		a.Name = *req.Name
	}
	if req.Description != nil {
		a.Description = *req.Description
	}
	a.UpdatedAt = time.Now()

	_, err = h.db.Exec(r.Context(),
		"UPDATE analyses SET name = $2, description = $3, updated_at = $4 WHERE id = $1",
		a.ID, a.Name, a.Description, a.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, a)
}

func (h *AnalysisHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid analysis id")
		return
	}

	_, err = h.db.Exec(r.Context(), "DELETE FROM analyses WHERE id = $1", id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AnalysisHandler) GetByShareToken(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "missing share token")
		return
	}

	var a models.Analysis
	err := h.db.QueryRow(r.Context(),
		"SELECT id, name, description, owner, share_token, created_at, updated_at FROM analyses WHERE share_token = $1", token).
		Scan(&a.ID, &a.Name, &a.Description, &a.Owner, &a.ShareToken, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusNotFound, "analysis not found")
		return
	}

	writeJSON(w, http.StatusOK, a)
}
