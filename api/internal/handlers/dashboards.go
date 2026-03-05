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

type DashboardHandler struct {
	db *pgxpool.Pool
}

func NewDashboardHandler(db *pgxpool.Pool) *DashboardHandler {
	return &DashboardHandler{db: db}
}

func (h *DashboardHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(),
		"SELECT id, analysis_id, name, published, layout, created_at, updated_at FROM dashboards ORDER BY updated_at DESC")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var dashboards []models.Dashboard
	for rows.Next() {
		var d models.Dashboard
		if err := rows.Scan(&d.ID, &d.AnalysisID, &d.Name, &d.Published, &d.Layout, &d.CreatedAt, &d.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		dashboards = append(dashboards, d)
	}
	if dashboards == nil {
		dashboards = []models.Dashboard{}
	}

	writeJSON(w, http.StatusOK, dashboards)
}

func (h *DashboardHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid dashboard id")
		return
	}

	var d models.Dashboard
	err = h.db.QueryRow(r.Context(),
		"SELECT id, analysis_id, name, published, layout, created_at, updated_at FROM dashboards WHERE id = $1", id).
		Scan(&d.ID, &d.AnalysisID, &d.Name, &d.Published, &d.Layout, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusNotFound, "dashboard not found")
		return
	}

	writeJSON(w, http.StatusOK, d)
}

func (h *DashboardHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateDashboardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	d := models.Dashboard{
		ID:         uuid.New(),
		AnalysisID: req.AnalysisID,
		Name:       req.Name,
		Published:  false,
		Layout:     req.Layout,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	_, err := h.db.Exec(r.Context(),
		"INSERT INTO dashboards (id, analysis_id, name, published, layout, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		d.ID, d.AnalysisID, d.Name, d.Published, d.Layout, d.CreatedAt, d.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, d)
}

func (h *DashboardHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid dashboard id")
		return
	}

	var req models.UpdateDashboardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var d models.Dashboard
	err = h.db.QueryRow(r.Context(),
		"SELECT id, analysis_id, name, published, layout, created_at, updated_at FROM dashboards WHERE id = $1", id).
		Scan(&d.ID, &d.AnalysisID, &d.Name, &d.Published, &d.Layout, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusNotFound, "dashboard not found")
		return
	}

	if req.Name != nil {
		d.Name = *req.Name
	}
	if req.Published != nil {
		d.Published = *req.Published
	}
	if req.Layout != nil {
		d.Layout = *req.Layout
	}
	d.UpdatedAt = time.Now()

	_, err = h.db.Exec(r.Context(),
		"UPDATE dashboards SET name = $2, published = $3, layout = $4, updated_at = $5 WHERE id = $1",
		d.ID, d.Name, d.Published, d.Layout, d.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, d)
}

func (h *DashboardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid dashboard id")
		return
	}

	_, err = h.db.Exec(r.Context(), "DELETE FROM dashboards WHERE id = $1", id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
