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

type CardHandler struct {
	db     *pgxpool.Pool
	engine *engine.Engine
}

func NewCardHandler(db *pgxpool.Pool, eng *engine.Engine) *CardHandler {
	return &CardHandler{db: db, engine: eng}
}

func (h *CardHandler) List(w http.ResponseWriter, r *http.Request) {
	analysisID, err := uuid.Parse(chi.URLParam(r, "analysisId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid analysis id")
		return
	}

	rows, err := h.db.Query(r.Context(),
		"SELECT id, analysis_id, card_type, label, config, position_x, position_y, input_card_ids, created_at, updated_at FROM cards WHERE analysis_id = $1 ORDER BY created_at",
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
	if cards == nil {
		cards = []models.Card{}
	}

	writeJSON(w, http.StatusOK, cards)
}

func (h *CardHandler) Create(w http.ResponseWriter, r *http.Request) {
	analysisID, err := uuid.Parse(chi.URLParam(r, "analysisId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid analysis id")
		return
	}

	var req models.CreateCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.InputCardIDs == nil {
		req.InputCardIDs = []uuid.UUID{}
	}

	c := models.Card{
		ID:           uuid.New(),
		AnalysisID:   analysisID,
		CardType:     req.CardType,
		Label:        req.Label,
		Config:       req.Config,
		PositionX:    req.PositionX,
		PositionY:    req.PositionY,
		InputCardIDs: req.InputCardIDs,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err = h.db.Exec(r.Context(),
		"INSERT INTO cards (id, analysis_id, card_type, label, config, position_x, position_y, input_card_ids, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
		c.ID, c.AnalysisID, c.CardType, c.Label, c.Config, c.PositionX, c.PositionY, c.InputCardIDs, c.CreatedAt, c.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, c)
}

func (h *CardHandler) Update(w http.ResponseWriter, r *http.Request) {
	cardID, err := uuid.Parse(chi.URLParam(r, "cardId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid card id")
		return
	}

	var req models.UpdateCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var c models.Card
	err = h.db.QueryRow(r.Context(),
		"SELECT id, analysis_id, card_type, label, config, position_x, position_y, input_card_ids, created_at, updated_at FROM cards WHERE id = $1", cardID).
		Scan(&c.ID, &c.AnalysisID, &c.CardType, &c.Label, &c.Config, &c.PositionX, &c.PositionY, &c.InputCardIDs, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusNotFound, "card not found")
		return
	}

	if req.Label != nil {
		c.Label = *req.Label
	}
	if req.Config != nil {
		c.Config = *req.Config
	}
	if req.PositionX != nil {
		c.PositionX = *req.PositionX
	}
	if req.PositionY != nil {
		c.PositionY = *req.PositionY
	}
	if req.InputCardIDs != nil {
		c.InputCardIDs = *req.InputCardIDs
	}
	c.UpdatedAt = time.Now()

	_, err = h.db.Exec(r.Context(),
		"UPDATE cards SET label = $2, config = $3, position_x = $4, position_y = $5, input_card_ids = $6, updated_at = $7 WHERE id = $1",
		c.ID, c.Label, c.Config, c.PositionX, c.PositionY, c.InputCardIDs, c.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, c)
}

func (h *CardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cardID, err := uuid.Parse(chi.URLParam(r, "cardId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid card id")
		return
	}

	_, err = h.db.Exec(r.Context(), "DELETE FROM cards WHERE id = $1", cardID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CardHandler) Execute(w http.ResponseWriter, r *http.Request) {
	analysisID, err := uuid.Parse(chi.URLParam(r, "analysisId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid analysis id")
		return
	}

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

	response := make(map[string]json.RawMessage)
	for id, result := range results {
		raw, _ := json.Marshal(result)
		response[id.String()] = raw
	}

	writeJSON(w, http.StatusOK, models.ExecuteAnalysisResponse{Results: response})
}
