package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sidryenireddy/prism/api/internal/ai"
	"github.com/sidryenireddy/prism/api/internal/models"
)

type AIHandler struct {
	db *pgxpool.Pool
}

func NewAIHandler(db *pgxpool.Pool) *AIHandler {
	return &AIHandler{db: db}
}

func (h *AIHandler) Generate(w http.ResponseWriter, r *http.Request) {
	var req ai.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := ai.Generate(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Create the cards in the database
	analysisID, err := uuid.Parse(req.AnalysisID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid analysis_id")
		return
	}

	var createdCards []models.Card
	createdIDs := make([]uuid.UUID, len(resp.Cards))

	for i, gc := range resp.Cards {
		c := models.Card{
			ID:           uuid.New(),
			AnalysisID:   analysisID,
			CardType:     models.CardType(gc.CardType),
			Label:        gc.Label,
			Config:       gc.Config,
			PositionX:    gc.PositionX,
			PositionY:    gc.PositionY,
			InputCardIDs: []uuid.UUID{},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if gc.InputIndex != nil && *gc.InputIndex >= 0 && *gc.InputIndex < i {
			c.InputCardIDs = []uuid.UUID{createdIDs[*gc.InputIndex]}
		}
		createdIDs[i] = c.ID

		_, err := h.db.Exec(r.Context(),
			"INSERT INTO cards (id, analysis_id, card_type, label, config, position_x, position_y, input_card_ids, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
			c.ID, c.AnalysisID, c.CardType, c.Label, c.Config, c.PositionX, c.PositionY, c.InputCardIDs, c.CreatedAt, c.UpdatedAt)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		createdCards = append(createdCards, c)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"cards": createdCards})
}

func (h *AIHandler) Configure(w http.ResponseWriter, r *http.Request) {
	cardID, err := uuid.Parse(chi.URLParam(r, "cardId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid card id")
		return
	}

	var body struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
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

	resp, err := ai.Configure(r.Context(), ai.ConfigureRequest{Card: c, Prompt: body.Prompt})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if len(resp.Config) > 0 {
		c.Config = resp.Config
	}
	if resp.Label != "" {
		c.Label = resp.Label
	}
	c.UpdatedAt = time.Now()

	_, err = h.db.Exec(r.Context(),
		"UPDATE cards SET label = $2, config = $3, updated_at = $4 WHERE id = $1",
		c.ID, c.Label, c.Config, c.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, c)
}
