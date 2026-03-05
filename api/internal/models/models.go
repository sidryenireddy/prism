package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CardType represents the type of a card in an analysis
type CardType string

const (
	// Object Set cards
	CardTypeFilterObjectSet   CardType = "filter_object_set"
	CardTypeSearchAround      CardType = "search_around"
	CardTypeSetMathUnion      CardType = "set_math_union"
	CardTypeSetMathIntersect  CardType = "set_math_intersection"
	CardTypeSetMathDifference CardType = "set_math_difference"

	// Visualization cards
	CardTypeBarChart    CardType = "bar_chart"
	CardTypeLineChart   CardType = "line_chart"
	CardTypePieChart    CardType = "pie_chart"
	CardTypeScatterPlot CardType = "scatter_plot"
	CardTypeHeatGrid    CardType = "heat_grid"

	// Table cards
	CardTypeObjectTable CardType = "object_table"
	CardTypePivotTable  CardType = "pivot_table"

	// Numeric cards
	CardTypeCount   CardType = "count"
	CardTypeSum     CardType = "sum"
	CardTypeAverage CardType = "average"
	CardTypeMin     CardType = "min"
	CardTypeMax     CardType = "max"

	// Time Series cards
	CardTypeTimeSeriesChart  CardType = "time_series_chart"
	CardTypeRollingAggregate CardType = "rolling_aggregate"

	// Formula card
	CardTypeFormula CardType = "formula"

	// Overlay chart
	CardTypeOverlayChart CardType = "overlay_chart"

	// Parameter cards
	CardTypeParamObjectSelection CardType = "param_object_selection"
	CardTypeParamDateRange       CardType = "param_date_range"
	CardTypeParamNumeric         CardType = "param_numeric"
	CardTypeParamString          CardType = "param_string"
	CardTypeParamBoolean         CardType = "param_boolean"

	// Action cards
	CardTypeActionButton CardType = "action_button"
)

type Analysis struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Owner       string    `json:"owner" db:"owner"`
	ShareToken  string    `json:"share_token,omitempty" db:"share_token"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Card struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	AnalysisID   uuid.UUID       `json:"analysis_id" db:"analysis_id"`
	CardType     CardType        `json:"card_type" db:"card_type"`
	Label        string          `json:"label" db:"label"`
	Config       json.RawMessage `json:"config" db:"config"`
	PositionX    float64         `json:"position_x" db:"position_x"`
	PositionY    float64         `json:"position_y" db:"position_y"`
	InputCardIDs []uuid.UUID     `json:"input_card_ids" db:"input_card_ids"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
}

type Dashboard struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	AnalysisID uuid.UUID       `json:"analysis_id" db:"analysis_id"`
	Name       string          `json:"name" db:"name"`
	Published  bool            `json:"published" db:"published"`
	Layout     json.RawMessage `json:"layout" db:"layout"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at" db:"updated_at"`
}

type Dataset struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	AnalysisID uuid.UUID       `json:"analysis_id" db:"analysis_id"`
	CardID     uuid.UUID       `json:"card_id" db:"card_id"`
	Name       string          `json:"name" db:"name"`
	Data       json.RawMessage `json:"data" db:"data"`
	RowCount   int             `json:"row_count" db:"row_count"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

// API request/response types

type CreateAnalysisRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Owner       string `json:"owner"`
}

type UpdateAnalysisRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type CreateCardRequest struct {
	CardType     CardType        `json:"card_type"`
	Label        string          `json:"label"`
	Config       json.RawMessage `json:"config"`
	PositionX    float64         `json:"position_x"`
	PositionY    float64         `json:"position_y"`
	InputCardIDs []uuid.UUID     `json:"input_card_ids"`
}

type UpdateCardRequest struct {
	Label        *string          `json:"label,omitempty"`
	Config       *json.RawMessage `json:"config,omitempty"`
	PositionX    *float64         `json:"position_x,omitempty"`
	PositionY    *float64         `json:"position_y,omitempty"`
	InputCardIDs *[]uuid.UUID     `json:"input_card_ids,omitempty"`
}

type CreateDashboardRequest struct {
	AnalysisID uuid.UUID       `json:"analysis_id"`
	Name       string          `json:"name"`
	Layout     json.RawMessage `json:"layout"`
}

type UpdateDashboardRequest struct {
	Name      *string          `json:"name,omitempty"`
	Published *bool            `json:"published,omitempty"`
	Layout    *json.RawMessage `json:"layout,omitempty"`
}

type ExecuteAnalysisResponse struct {
	Results map[string]json.RawMessage `json:"results"` // card_id -> result
}

type CardResult struct {
	CardID   uuid.UUID       `json:"card_id"`
	CardType CardType        `json:"card_type"`
	Data     json.RawMessage `json:"data"`
	Error    string          `json:"error,omitempty"`
}

type SaveDatasetRequest struct {
	AnalysisID string `json:"analysis_id"`
	CardID     string `json:"card_id"`
	Name       string `json:"name"`
}

type ExecuteActionRequest struct {
	CardID string `json:"card_id"`
}
