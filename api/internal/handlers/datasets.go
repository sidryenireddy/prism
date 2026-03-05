package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sidryenireddy/prism/api/internal/engine"
	"github.com/sidryenireddy/prism/api/internal/models"
)

type DatasetHandler struct {
	db     *pgxpool.Pool
	engine *engine.Engine
}

func NewDatasetHandler(db *pgxpool.Pool, eng *engine.Engine) *DatasetHandler {
	return &DatasetHandler{db: db, engine: eng}
}

func (h *DatasetHandler) Save(w http.ResponseWriter, r *http.Request) {
	var req models.SaveDatasetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	analysisID, err := uuid.Parse(req.AnalysisID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid analysis_id")
		return
	}
	cardID, err := uuid.Parse(req.CardID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid card_id")
		return
	}

	// Execute the analysis to get results
	rows, err := h.db.Query(r.Context(),
		"SELECT id, analysis_id, card_type, label, config, position_x, position_y, input_card_ids, created_at, updated_at FROM cards WHERE analysis_id = $1",
		analysisID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var c models.Card
		if err := rows.Scan(&c.ID, &c.AnalysisID, &c.CardType, &c.Label, &c.Config, &c.PositionX, &c.PositionY, &c.InputCardIDs, &c.CreatedAt, &c.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		cards = append(cards, c)
	}

	results, err := h.engine.Execute(r.Context(), cards)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	cardResult, ok := results[cardID]
	if !ok || cardResult.Error != "" {
		writeError(w, http.StatusBadRequest, "card has no valid result")
		return
	}

	// Count rows
	rowCount := 0
	var dataWrapper struct {
		Objects    []interface{} `json:"objects"`
		Rows       []interface{} `json:"rows"`
		TotalCount int           `json:"totalCount"`
	}
	if err := json.Unmarshal(cardResult.Data, &dataWrapper); err == nil {
		if len(dataWrapper.Objects) > 0 {
			rowCount = len(dataWrapper.Objects)
		} else if len(dataWrapper.Rows) > 0 {
			rowCount = len(dataWrapper.Rows)
		} else {
			rowCount = dataWrapper.TotalCount
		}
	}

	ds := models.Dataset{
		ID:         uuid.New(),
		AnalysisID: analysisID,
		CardID:     cardID,
		Name:       req.Name,
		Data:       cardResult.Data,
		RowCount:   rowCount,
		CreatedAt:  time.Now(),
	}

	_, err = h.db.Exec(r.Context(),
		"INSERT INTO datasets (id, analysis_id, card_id, name, data, row_count, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		ds.ID, ds.AnalysisID, ds.CardID, ds.Name, ds.Data, ds.RowCount, ds.CreatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, ds)
}

func (h *DatasetHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(),
		"SELECT id, analysis_id, card_id, name, row_count, created_at FROM datasets ORDER BY created_at DESC")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	var datasets []models.Dataset
	for rows.Next() {
		var d models.Dataset
		if err := rows.Scan(&d.ID, &d.AnalysisID, &d.CardID, &d.Name, &d.RowCount, &d.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		d.Data = json.RawMessage(`null`) // Don't return full data in list
		datasets = append(datasets, d)
	}
	if datasets == nil {
		datasets = []models.Dataset{}
	}

	writeJSON(w, http.StatusOK, datasets)
}

func (h *DatasetHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid dataset id")
		return
	}

	var d models.Dataset
	err = h.db.QueryRow(r.Context(),
		"SELECT id, analysis_id, card_id, name, data, row_count, created_at FROM datasets WHERE id = $1", id).
		Scan(&d.ID, &d.AnalysisID, &d.CardID, &d.Name, &d.Data, &d.RowCount, &d.CreatedAt)
	if err != nil {
		writeError(w, http.StatusNotFound, "dataset not found")
		return
	}

	writeJSON(w, http.StatusOK, d)
}

func (h *DatasetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid dataset id")
		return
	}

	_, err = h.db.Exec(r.Context(), "DELETE FROM datasets WHERE id = $1", id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
