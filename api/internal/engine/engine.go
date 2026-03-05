package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/sidryenireddy/prism/api/internal/models"
	"github.com/sidryenireddy/prism/api/internal/ontology"
)

type Engine struct {
	ontologyClient *ontology.Client
}

func New(ontologyClient *ontology.Client) *Engine {
	return &Engine{ontologyClient: ontologyClient}
}

// Execute evaluates all cards in an analysis in topological order
func (e *Engine) Execute(ctx context.Context, cards []models.Card) (map[uuid.UUID]*models.CardResult, error) {
	order, err := topologicalSort(cards)
	if err != nil {
		return nil, fmt.Errorf("sorting card graph: %w", err)
	}

	results := make(map[uuid.UUID]*models.CardResult)

	for _, card := range order {
		inputResults := make(map[uuid.UUID]*models.CardResult)
		for _, inputID := range card.InputCardIDs {
			if r, ok := results[inputID]; ok {
				inputResults[inputID] = r
			}
		}

		result := e.executeCard(ctx, card, inputResults)
		results[card.ID] = result
	}

	return results, nil
}

func (e *Engine) executeCard(ctx context.Context, card models.Card, inputs map[uuid.UUID]*models.CardResult) *models.CardResult {
	result := &models.CardResult{
		CardID:   card.ID,
		CardType: card.CardType,
	}

	var data interface{}
	var err error

	switch card.CardType {
	// Object Set cards
	case models.CardTypeFilterObjectSet:
		data, err = e.executeFilterObjectSet(ctx, card)
	case models.CardTypeSearchAround:
		data, err = e.executeSearchAround(ctx, card, inputs)
	case models.CardTypeSetMathUnion, models.CardTypeSetMathIntersect, models.CardTypeSetMathDifference:
		data, err = e.executeSetMath(card, inputs)

	// Numeric cards
	case models.CardTypeCount, models.CardTypeSum, models.CardTypeAverage, models.CardTypeMin, models.CardTypeMax:
		data, err = e.executeNumeric(ctx, card)

	// Parameter cards
	case models.CardTypeParamObjectSelection, models.CardTypeParamDateRange,
		models.CardTypeParamNumeric, models.CardTypeParamString, models.CardTypeParamBoolean:
		data, err = e.executeParameter(card)

	// Visualization, Table, TimeSeries, Action cards return config + input data for frontend rendering
	default:
		data, err = e.executePassthrough(card, inputs)
	}

	if err != nil {
		result.Error = err.Error()
		result.Data = json.RawMessage(`null`)
	} else {
		raw, _ := json.Marshal(data)
		result.Data = raw
	}

	return result
}

func (e *Engine) executeFilterObjectSet(ctx context.Context, card models.Card) (interface{}, error) {
	var config struct {
		ObjectTypeID string                    `json:"objectTypeId"`
		Filters      []ontology.FilterClause   `json:"filters"`
		Query        string                    `json:"query"`
	}
	if err := json.Unmarshal(card.Config, &config); err != nil {
		return nil, fmt.Errorf("parsing filter config: %w", err)
	}

	query := ontology.SearchQuery{
		ObjectTypeID: config.ObjectTypeID,
		Query:        config.Query,
		Filters:      config.Filters,
		PageSize:     100,
	}

	return e.ontologyClient.SearchObjects(ctx, query)
}

func (e *Engine) executeSearchAround(ctx context.Context, card models.Card, inputs map[uuid.UUID]*models.CardResult) (interface{}, error) {
	// Get the first input's object set and find linked objects
	for _, input := range inputs {
		var sr ontology.SearchResult
		if err := json.Unmarshal(input.Data, &sr); err != nil {
			continue
		}
		if len(sr.Objects) > 0 {
			obj := sr.Objects[0]
			return e.ontologyClient.GetLinkedObjects(ctx, obj.ID, obj.ObjectTypeID)
		}
	}

	return &ontology.SearchResult{Objects: []ontology.ObjectInstance{}}, nil
}

func (e *Engine) executeSetMath(card models.Card, inputs map[uuid.UUID]*models.CardResult) (interface{}, error) {
	var sets []ontology.SearchResult
	for _, input := range inputs {
		var sr ontology.SearchResult
		if err := json.Unmarshal(input.Data, &sr); err == nil {
			sets = append(sets, sr)
		}
	}

	if len(sets) < 2 {
		return &ontology.SearchResult{Objects: []ontology.ObjectInstance{}}, nil
	}

	keySet := func(sr ontology.SearchResult) map[string]ontology.ObjectInstance {
		m := make(map[string]ontology.ObjectInstance)
		for _, o := range sr.Objects {
			m[o.ID] = o
		}
		return m
	}

	a := keySet(sets[0])
	b := keySet(sets[1])
	var result []ontology.ObjectInstance

	switch card.CardType {
	case models.CardTypeSetMathUnion:
		merged := make(map[string]ontology.ObjectInstance)
		for k, v := range a {
			merged[k] = v
		}
		for k, v := range b {
			merged[k] = v
		}
		for _, v := range merged {
			result = append(result, v)
		}
	case models.CardTypeSetMathIntersect:
		for k, v := range a {
			if _, ok := b[k]; ok {
				result = append(result, v)
			}
		}
	case models.CardTypeSetMathDifference:
		for k, v := range a {
			if _, ok := b[k]; !ok {
				result = append(result, v)
			}
		}
	}

	return &ontology.SearchResult{Objects: result, TotalCount: int64(len(result))}, nil
}

func (e *Engine) executeNumeric(ctx context.Context, card models.Card) (interface{}, error) {
	var config struct {
		ObjectTypeID string `json:"objectTypeId"`
		Property     string `json:"property"`
	}
	if err := json.Unmarshal(card.Config, &config); err != nil {
		return nil, fmt.Errorf("parsing numeric config: %w", err)
	}

	// Map card types to Elasticsearch aggregation types
	esAggType := map[models.CardType]string{
		models.CardTypeCount:   "value_count",
		models.CardTypeSum:     "sum",
		models.CardTypeAverage: "avg",
		models.CardTypeMin:     "min",
		models.CardTypeMax:     "max",
	}[card.CardType]

	query := ontology.AggregationQuery{
		ObjectTypeID: config.ObjectTypeID,
		Metrics: []ontology.AggMetric{
			{Field: config.Property, Type: esAggType},
		},
	}

	return e.ontologyClient.AggregateObjects(ctx, query)
}

func (e *Engine) executeParameter(card models.Card) (interface{}, error) {
	return card.Config, nil
}

func (e *Engine) executePassthrough(card models.Card, inputs map[uuid.UUID]*models.CardResult) (interface{}, error) {
	inputData := make(map[string]json.RawMessage)
	for id, r := range inputs {
		inputData[id.String()] = r.Data
	}

	return map[string]interface{}{
		"config": card.Config,
		"inputs": inputData,
	}, nil
}

// topologicalSort orders cards so that dependencies are executed first
func topologicalSort(cards []models.Card) ([]models.Card, error) {
	cardMap := make(map[uuid.UUID]models.Card)
	for _, c := range cards {
		cardMap[c.ID] = c
	}

	visited := make(map[uuid.UUID]bool)
	visiting := make(map[uuid.UUID]bool)
	var order []models.Card

	var visit func(id uuid.UUID) error
	visit = func(id uuid.UUID) error {
		if visited[id] {
			return nil
		}
		if visiting[id] {
			return fmt.Errorf("cycle detected in card graph at card %s", id)
		}

		visiting[id] = true
		card, ok := cardMap[id]
		if !ok {
			return fmt.Errorf("card %s not found", id)
		}

		for _, depID := range card.InputCardIDs {
			if err := visit(depID); err != nil {
				return err
			}
		}

		visiting[id] = false
		visited[id] = true
		order = append(order, card)
		return nil
	}

	// Sort cards by ID for deterministic ordering
	sorted := make([]models.Card, len(cards))
	copy(sorted, cards)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID.String() < sorted[j].ID.String()
	})

	for _, c := range sorted {
		if err := visit(c.ID); err != nil {
			return nil, err
		}
	}

	return order, nil
}
