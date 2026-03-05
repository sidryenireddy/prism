package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/sidryenireddy/prism/api/internal/mockdata"
	"github.com/sidryenireddy/prism/api/internal/models"
	"github.com/sidryenireddy/prism/api/internal/ontology"
)

type Engine struct {
	ontologyClient *ontology.Client
}

func New(ontologyClient *ontology.Client) *Engine {
	return &Engine{ontologyClient: ontologyClient}
}

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
	case models.CardTypeFilterObjectSet:
		data, err = e.executeFilterObjectSet(ctx, card)
	case models.CardTypeSearchAround:
		data, err = e.executeSearchAround(ctx, card, inputs)
	case models.CardTypeSetMathUnion, models.CardTypeSetMathIntersect, models.CardTypeSetMathDifference:
		data, err = e.executeSetMath(card, inputs)
	case models.CardTypeCount, models.CardTypeSum, models.CardTypeAverage, models.CardTypeMin, models.CardTypeMax:
		data, err = e.executeNumeric(ctx, card, inputs)
	case models.CardTypeParamObjectSelection, models.CardTypeParamDateRange,
		models.CardTypeParamNumeric, models.CardTypeParamString, models.CardTypeParamBoolean:
		data, err = e.executeParameter(card)
	case models.CardTypeActionButton:
		data, err = e.executeAction(ctx, card, inputs)
	case models.CardTypeBarChart, models.CardTypeLineChart, models.CardTypePieChart,
		models.CardTypeScatterPlot, models.CardTypeHeatGrid:
		data, err = e.executeVisualization(card, inputs)
	case models.CardTypeObjectTable:
		data, err = e.executeObjectTable(card, inputs)
	case models.CardTypePivotTable:
		data, err = e.executePivotTable(card, inputs)
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
		ObjectTypeID string             `json:"objectTypeId"`
		Filters      []mockdata.Filter  `json:"filters"`
		Query        string             `json:"query"`
	}
	if err := json.Unmarshal(card.Config, &config); err != nil {
		return nil, fmt.Errorf("parsing filter config: %w", err)
	}

	// Try mock data first
	objects := mockdata.GetObjectsByType(config.ObjectTypeID)
	if objects != nil {
		filtered := mockdata.FilterObjects(objects, config.Filters)
		return map[string]interface{}{
			"objects":    filtered,
			"totalCount": len(filtered),
		}, nil
	}

	// Fallback to ontology API
	query := ontology.SearchQuery{
		ObjectTypeID: config.ObjectTypeID,
		Query:        config.Query,
		PageSize:     100,
	}
	for _, f := range config.Filters {
		query.Filters = append(query.Filters, ontology.FilterClause{
			Field: f.Field, Operator: f.Operator, Value: f.Value,
		})
	}
	return e.ontologyClient.SearchObjects(ctx, query)
}

func (e *Engine) executeSearchAround(ctx context.Context, card models.Card, inputs map[uuid.UUID]*models.CardResult) (interface{}, error) {
	inputObjects := extractObjects(inputs)
	if len(inputObjects) > 0 {
		obj := inputObjects[0]
		id, _ := obj["id"].(string)
		if id == "" {
			if fid, ok := obj["id"].(float64); ok {
				id = fmt.Sprintf("%v", int(fid))
			}
		}
		otID, _ := obj["objectTypeId"].(string)
		if otID == "" {
			otID, _ = obj["objectTypeID"].(string)
		}

		// Try mock data
		mockID := id
		if otID == "ot-customers" {
			mockID = fmt.Sprintf("c-%s", id)
		} else if otID == "ot-orders" {
			mockID = fmt.Sprintf("o-%s", id)
		} else if otID == "ot-products" {
			mockID = fmt.Sprintf("p-%s", id)
		}
		linked := mockdata.GetLinkedObjects(mockID, otID)
		if linked != nil {
			return map[string]interface{}{
				"objects":    linked,
				"totalCount": len(linked),
			}, nil
		}

		return e.ontologyClient.GetLinkedObjects(ctx, id, otID)
	}
	return map[string]interface{}{"objects": []interface{}{}, "totalCount": 0}, nil
}

func (e *Engine) executeSetMath(card models.Card, inputs map[uuid.UUID]*models.CardResult) (interface{}, error) {
	var allSets [][]map[string]interface{}
	for _, input := range inputs {
		objs := extractObjectsFromResult(input)
		if len(objs) > 0 {
			allSets = append(allSets, objs)
		}
	}
	if len(allSets) < 2 {
		return map[string]interface{}{"objects": []interface{}{}, "totalCount": 0}, nil
	}

	keyOf := func(o map[string]interface{}) string {
		if id, ok := o["id"]; ok {
			return fmt.Sprintf("%v", id)
		}
		return fmt.Sprintf("%v", o)
	}

	aMap := map[string]map[string]interface{}{}
	for _, o := range allSets[0] {
		aMap[keyOf(o)] = o
	}
	bMap := map[string]map[string]interface{}{}
	for _, o := range allSets[1] {
		bMap[keyOf(o)] = o
	}

	var result []map[string]interface{}
	switch card.CardType {
	case models.CardTypeSetMathUnion:
		merged := map[string]map[string]interface{}{}
		for k, v := range aMap {
			merged[k] = v
		}
		for k, v := range bMap {
			merged[k] = v
		}
		for _, v := range merged {
			result = append(result, v)
		}
	case models.CardTypeSetMathIntersect:
		for k, v := range aMap {
			if _, ok := bMap[k]; ok {
				result = append(result, v)
			}
		}
	case models.CardTypeSetMathDifference:
		for k, v := range aMap {
			if _, ok := bMap[k]; !ok {
				result = append(result, v)
			}
		}
	}

	return map[string]interface{}{"objects": result, "totalCount": len(result)}, nil
}

func (e *Engine) executeNumeric(ctx context.Context, card models.Card, inputs map[uuid.UUID]*models.CardResult) (interface{}, error) {
	var config struct {
		ObjectTypeID string `json:"objectTypeId"`
		Property     string `json:"property"`
	}
	if err := json.Unmarshal(card.Config, &config); err != nil {
		return nil, fmt.Errorf("parsing numeric config: %w", err)
	}

	esAggType := map[models.CardType]string{
		models.CardTypeCount:   "value_count",
		models.CardTypeSum:     "sum",
		models.CardTypeAverage: "avg",
		models.CardTypeMin:     "min",
		models.CardTypeMax:     "max",
	}[card.CardType]

	// Try from input objects first
	inputObjects := extractObjects(inputs)
	if len(inputObjects) > 0 {
		var mockObjs []mockdata.MockObject
		for _, o := range inputObjects {
			mockObjs = append(mockObjs, mockdata.MockObject{Properties: o})
		}
		value := mockdata.Aggregate(mockObjs, config.Property, esAggType)
		return map[string]interface{}{
			"value": value,
			"label": card.Label,
			"type":  string(card.CardType),
		}, nil
	}

	// Try mock data
	objects := mockdata.GetObjectsByType(config.ObjectTypeID)
	if objects != nil {
		value := mockdata.Aggregate(objects, config.Property, esAggType)
		return map[string]interface{}{
			"value": value,
			"label": card.Label,
			"type":  string(card.CardType),
		}, nil
	}

	// Fallback to ontology
	query := ontology.AggregationQuery{
		ObjectTypeID: config.ObjectTypeID,
		Metrics:      []ontology.AggMetric{{Field: config.Property, Type: esAggType}},
	}
	return e.ontologyClient.AggregateObjects(ctx, query)
}

func (e *Engine) executeVisualization(card models.Card, inputs map[uuid.UUID]*models.CardResult) (interface{}, error) {
	var config struct {
		GroupBy    string `json:"groupBy"`
		Metric    string `json:"metric"`
		MetricType string `json:"metricType"`
	}
	if err := json.Unmarshal(card.Config, &config); err != nil {
		config.GroupBy = "region"
		config.MetricType = "count"
	}
	if config.MetricType == "" {
		config.MetricType = "count"
	}

	inputObjects := extractObjects(inputs)
	if len(inputObjects) == 0 {
		return map[string]interface{}{
			"chartData": []interface{}{},
			"config":    card.Config,
		}, nil
	}

	// Group objects
	groups := map[string][]map[string]interface{}{}
	for _, o := range inputObjects {
		key := fmt.Sprintf("%v", o[config.GroupBy])
		groups[key] = append(groups[key], o)
	}

	var chartData []map[string]interface{}
	for name, objs := range groups {
		entry := map[string]interface{}{"name": name}
		switch config.MetricType {
		case "count":
			entry["value"] = len(objs)
		case "sum":
			var s float64
			for _, o := range objs {
				s += toFloat(o[config.Metric])
			}
			entry["value"] = s
		case "avg":
			var s float64
			for _, o := range objs {
				s += toFloat(o[config.Metric])
			}
			entry["value"] = s / float64(len(objs))
		default:
			entry["value"] = len(objs)
		}
		chartData = append(chartData, entry)
	}

	// Sort by name for consistent ordering
	sort.Slice(chartData, func(i, j int) bool {
		return fmt.Sprintf("%v", chartData[i]["name"]) < fmt.Sprintf("%v", chartData[j]["name"])
	})

	return map[string]interface{}{
		"chartData": chartData,
		"config":    card.Config,
	}, nil
}

func (e *Engine) executeObjectTable(card models.Card, inputs map[uuid.UUID]*models.CardResult) (interface{}, error) {
	var config struct {
		Columns  []string `json:"columns"`
		PageSize int      `json:"pageSize"`
	}
	if err := json.Unmarshal(card.Config, &config); err != nil {
		config.PageSize = 20
	}
	if config.PageSize == 0 {
		config.PageSize = 20
	}

	inputObjects := extractObjects(inputs)

	// Auto-detect columns if not specified
	if len(config.Columns) == 0 && len(inputObjects) > 0 {
		for k := range inputObjects[0] {
			if k != "objectTypeId" && k != "objectTypeID" {
				config.Columns = append(config.Columns, k)
			}
		}
		sort.Strings(config.Columns)
	}

	rows := inputObjects
	if len(rows) > config.PageSize {
		rows = rows[:config.PageSize]
	}

	return map[string]interface{}{
		"rows":       rows,
		"columns":    config.Columns,
		"totalCount": len(inputObjects),
		"pageSize":   config.PageSize,
	}, nil
}

func (e *Engine) executePivotTable(card models.Card, inputs map[uuid.UUID]*models.CardResult) (interface{}, error) {
	var config struct {
		RowField    string `json:"rowField"`
		ColumnField string `json:"columnField"`
		ValueField  string `json:"valueField"`
		Aggregation string `json:"aggregation"`
	}
	if err := json.Unmarshal(card.Config, &config); err != nil {
		return nil, fmt.Errorf("parsing pivot config: %w", err)
	}
	if config.Aggregation == "" {
		config.Aggregation = "count"
	}

	inputObjects := extractObjects(inputs)

	// Build pivot
	rowKeys := map[string]bool{}
	colKeys := map[string]bool{}
	cells := map[string]map[string][]float64{}

	for _, o := range inputObjects {
		rk := fmt.Sprintf("%v", o[config.RowField])
		ck := fmt.Sprintf("%v", o[config.ColumnField])
		rowKeys[rk] = true
		colKeys[ck] = true
		if cells[rk] == nil {
			cells[rk] = map[string][]float64{}
		}
		cells[rk][ck] = append(cells[rk][ck], toFloat(o[config.ValueField]))
	}

	// Aggregate
	pivotData := map[string]map[string]float64{}
	for rk, cols := range cells {
		pivotData[rk] = map[string]float64{}
		for ck, vals := range cols {
			switch config.Aggregation {
			case "count":
				pivotData[rk][ck] = float64(len(vals))
			case "sum":
				var s float64
				for _, v := range vals {
					s += v
				}
				pivotData[rk][ck] = s
			case "avg":
				var s float64
				for _, v := range vals {
					s += v
				}
				pivotData[rk][ck] = s / float64(len(vals))
			}
		}
	}

	var rows []string
	for k := range rowKeys {
		rows = append(rows, k)
	}
	sort.Strings(rows)
	var cols []string
	for k := range colKeys {
		cols = append(cols, k)
	}
	sort.Strings(cols)

	return map[string]interface{}{
		"pivotData":  pivotData,
		"rowKeys":    rows,
		"columnKeys": cols,
		"config":     config,
	}, nil
}

func (e *Engine) executeAction(ctx context.Context, card models.Card, inputs map[uuid.UUID]*models.CardResult) (interface{}, error) {
	var config struct {
		ActionTypeID      string            `json:"actionTypeId"`
		ParameterMappings map[string]string `json:"parameterMappings"`
	}
	if err := json.Unmarshal(card.Config, &config); err != nil {
		return nil, fmt.Errorf("parsing action config: %w", err)
	}

	// Build parameters from inputs
	params := map[string]interface{}{}
	for paramName, inputRef := range config.ParameterMappings {
		for _, input := range inputs {
			params[paramName] = input.Data
			_ = inputRef
			break
		}
	}

	return map[string]interface{}{
		"actionTypeId": config.ActionTypeID,
		"parameters":   params,
		"status":       "ready",
	}, nil
}

func (e *Engine) executeParameter(card models.Card) (interface{}, error) {
	var config map[string]interface{}
	if err := json.Unmarshal(card.Config, &config); err != nil {
		return card.Config, nil
	}
	return map[string]interface{}{
		"value":  config["defaultValue"],
		"label":  config["label"],
		"config": config,
	}, nil
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

// --- helpers ---

func extractObjects(inputs map[uuid.UUID]*models.CardResult) []map[string]interface{} {
	for _, input := range inputs {
		return extractObjectsFromResult(input)
	}
	return nil
}

func extractObjectsFromResult(input *models.CardResult) []map[string]interface{} {
	if input == nil || len(input.Data) == 0 {
		return nil
	}
	var wrapper struct {
		Objects []map[string]interface{} `json:"objects"`
	}
	if err := json.Unmarshal(input.Data, &wrapper); err == nil && len(wrapper.Objects) > 0 {
		// Flatten: if objects have "properties" sub-object, merge up
		var flat []map[string]interface{}
		for _, o := range wrapper.Objects {
			if props, ok := o["properties"].(map[string]interface{}); ok {
				merged := map[string]interface{}{}
				for k, v := range o {
					if k != "properties" {
						merged[k] = v
					}
				}
				for k, v := range props {
					merged[k] = v
				}
				flat = append(flat, merged)
			} else {
				flat = append(flat, o)
			}
		}
		return flat
	}
	return nil
}

func toFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case json.Number:
		f, _ := n.Float64()
		return f
	}
	return 0
}

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
