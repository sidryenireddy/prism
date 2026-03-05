package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sidryenireddy/prism/api/internal/formula"
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
	case models.CardTypeTimeSeriesChart:
		data, err = e.executeTimeSeries(card, inputs)
	case models.CardTypeRollingAggregate:
		data, err = e.executeRollingAggregate(card, inputs)
	case models.CardTypeFormula:
		data, err = e.executeFormula(card, inputs)
	case models.CardTypeOverlayChart:
		data, err = e.executeOverlayChart(card, inputs)
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
		ObjectTypeID string            `json:"objectTypeId"`
		Filters      []mockdata.Filter `json:"filters"`
		Query        string            `json:"query"`
	}
	if err := json.Unmarshal(card.Config, &config); err != nil {
		return nil, fmt.Errorf("parsing filter config: %w", err)
	}

	objects := mockdata.GetObjectsByType(config.ObjectTypeID)
	if objects != nil {
		filtered := mockdata.FilterObjects(objects, config.Filters)
		return map[string]interface{}{
			"objects":    filtered,
			"totalCount": len(filtered),
		}, nil
	}

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
		Field        string `json:"field"`
	}
	if err := json.Unmarshal(card.Config, &config); err != nil {
		return nil, fmt.Errorf("parsing numeric config: %w", err)
	}

	field := config.Property
	if field == "" {
		field = config.Field
	}

	esAggType := map[models.CardType]string{
		models.CardTypeCount:   "value_count",
		models.CardTypeSum:     "sum",
		models.CardTypeAverage: "avg",
		models.CardTypeMin:     "min",
		models.CardTypeMax:     "max",
	}[card.CardType]

	inputObjects := extractObjects(inputs)
	if len(inputObjects) > 0 {
		var mockObjs []mockdata.MockObject
		for _, o := range inputObjects {
			mockObjs = append(mockObjs, mockdata.MockObject{Properties: o})
		}
		value := mockdata.Aggregate(mockObjs, field, esAggType)
		return map[string]interface{}{
			"value": value,
			"label": card.Label,
			"type":  string(card.CardType),
		}, nil
	}

	objects := mockdata.GetObjectsByType(config.ObjectTypeID)
	if objects != nil {
		value := mockdata.Aggregate(objects, field, esAggType)
		return map[string]interface{}{
			"value": value,
			"label": card.Label,
			"type":  string(card.CardType),
		}, nil
	}

	query := ontology.AggregationQuery{
		ObjectTypeID: config.ObjectTypeID,
		Metrics:      []ontology.AggMetric{{Field: field, Type: esAggType}},
	}
	return e.ontologyClient.AggregateObjects(ctx, query)
}

func (e *Engine) executeVisualization(card models.Card, inputs map[uuid.UUID]*models.CardResult) (interface{}, error) {
	var config struct {
		GroupBy    string `json:"groupBy"`
		Metric     string `json:"metric"`
		MetricType string `json:"metricType"`
		ValueField string `json:"valueField"`
	}
	if err := json.Unmarshal(card.Config, &config); err != nil {
		config.GroupBy = "region"
		config.MetricType = "count"
	}
	if config.MetricType == "" {
		config.MetricType = "count"
	}
	if config.ValueField != "" && config.Metric == "" {
		config.Metric = config.ValueField
	}

	inputObjects := extractObjects(inputs)
	if len(inputObjects) == 0 {
		return map[string]interface{}{
			"chartData": []interface{}{},
			"config":    card.Config,
		}, nil
	}

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
		case "avg", "average":
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

// --- Time Series ---

func (e *Engine) executeTimeSeries(card models.Card, inputs map[uuid.UUID]*models.CardResult) (interface{}, error) {
	var config struct {
		TimeField  string   `json:"timeField"`
		ValueField string   `json:"valueField"`
		GroupBy    string   `json:"groupBy"`
		Metric     string   `json:"metric"` // count, sum, avg
		Series     []string `json:"series"` // multiple value fields
	}
	if err := json.Unmarshal(card.Config, &config); err != nil {
		config.TimeField = "order_date"
		config.Metric = "count"
	}
	if config.TimeField == "" {
		config.TimeField = "order_date"
	}
	if config.Metric == "" {
		config.Metric = "count"
	}

	inputObjects := extractObjects(inputs)
	if len(inputObjects) == 0 {
		return map[string]interface{}{"series": []interface{}{}, "timeField": config.TimeField}, nil
	}

	// Sort by time field
	sort.Slice(inputObjects, func(i, j int) bool {
		ti := fmt.Sprintf("%v", inputObjects[i][config.TimeField])
		tj := fmt.Sprintf("%v", inputObjects[j][config.TimeField])
		return ti < tj
	})

	if config.GroupBy != "" {
		// Grouped time series: one line per group value
		groups := map[string][]map[string]interface{}{}
		for _, o := range inputObjects {
			gk := fmt.Sprintf("%v", o[config.GroupBy])
			groups[gk] = append(groups[gk], o)
		}

		var seriesData []map[string]interface{}
		for groupName, objs := range groups {
			points := buildTimePoints(objs, config.TimeField, config.ValueField, config.Metric)
			seriesData = append(seriesData, map[string]interface{}{
				"name":   groupName,
				"points": points,
			})
		}
		sort.Slice(seriesData, func(i, j int) bool {
			return fmt.Sprintf("%v", seriesData[i]["name"]) < fmt.Sprintf("%v", seriesData[j]["name"])
		})

		return map[string]interface{}{
			"series":    seriesData,
			"timeField": config.TimeField,
			"grouped":   true,
		}, nil
	}

	// Multiple series from different value fields
	if len(config.Series) > 1 {
		var seriesData []map[string]interface{}
		for _, vf := range config.Series {
			points := buildTimePoints(inputObjects, config.TimeField, vf, "raw")
			seriesData = append(seriesData, map[string]interface{}{
				"name":   vf,
				"points": points,
			})
		}
		return map[string]interface{}{
			"series":    seriesData,
			"timeField": config.TimeField,
			"grouped":   true,
		}, nil
	}

	// Single series
	points := buildTimePoints(inputObjects, config.TimeField, config.ValueField, config.Metric)
	return map[string]interface{}{
		"series": []map[string]interface{}{
			{"name": config.ValueField, "points": points},
		},
		"timeField": config.TimeField,
		"grouped":   false,
	}, nil
}

func buildTimePoints(objects []map[string]interface{}, timeField, valueField, metric string) []map[string]interface{} {
	// Bucket by month
	buckets := map[string][]float64{}
	bucketOrder := []string{}

	for _, o := range objects {
		ts := fmt.Sprintf("%v", o[timeField])
		t := parseTime(ts)
		key := t.Format("2006-01")

		if _, seen := buckets[key]; !seen {
			bucketOrder = append(bucketOrder, key)
		}

		if metric == "raw" || valueField != "" {
			buckets[key] = append(buckets[key], toFloat(o[valueField]))
		} else {
			buckets[key] = append(buckets[key], 1)
		}
	}

	sort.Strings(bucketOrder)

	var points []map[string]interface{}
	for _, key := range bucketOrder {
		vals := buckets[key]
		var value float64
		switch metric {
		case "count":
			value = float64(len(vals))
		case "sum", "raw":
			for _, v := range vals {
				value += v
			}
		case "avg", "average":
			for _, v := range vals {
				value += v
			}
			value /= float64(len(vals))
		default:
			value = float64(len(vals))
		}
		points = append(points, map[string]interface{}{
			"time":  key,
			"value": math.Round(value*100) / 100,
		})
	}
	return points
}

// --- Rolling Aggregate ---

func (e *Engine) executeRollingAggregate(card models.Card, inputs map[uuid.UUID]*models.CardResult) (interface{}, error) {
	var config struct {
		TimeField  string `json:"timeField"`
		ValueField string `json:"valueField"`
		Window     int    `json:"window"` // number of periods
		Type       string `json:"type"`   // moving_average, moving_sum
	}
	if err := json.Unmarshal(card.Config, &config); err != nil {
		config.Window = 3
		config.Type = "moving_average"
	}
	if config.Window <= 0 {
		config.Window = 3
	}
	if config.TimeField == "" {
		config.TimeField = "order_date"
	}
	if config.Type == "" {
		config.Type = "moving_average"
	}

	inputObjects := extractObjects(inputs)
	if len(inputObjects) == 0 {
		return map[string]interface{}{"series": []interface{}{}, "window": config.Window}, nil
	}

	// Build base time points
	sort.Slice(inputObjects, func(i, j int) bool {
		return fmt.Sprintf("%v", inputObjects[i][config.TimeField]) < fmt.Sprintf("%v", inputObjects[j][config.TimeField])
	})

	basePoints := buildTimePoints(inputObjects, config.TimeField, config.ValueField, "sum")

	// Apply rolling window
	var rollingPoints []map[string]interface{}
	values := make([]float64, len(basePoints))
	for i, p := range basePoints {
		values[i] = toFloat(p["value"])
	}

	for i := range basePoints {
		start := i - config.Window + 1
		if start < 0 {
			start = 0
		}
		windowVals := values[start : i+1]
		var sum float64
		for _, v := range windowVals {
			sum += v
		}
		var val float64
		if config.Type == "moving_average" {
			val = sum / float64(len(windowVals))
		} else {
			val = sum
		}
		rollingPoints = append(rollingPoints, map[string]interface{}{
			"time":  basePoints[i]["time"],
			"value": math.Round(val*100) / 100,
		})
	}

	return map[string]interface{}{
		"series": []map[string]interface{}{
			{"name": "raw", "points": basePoints},
			{"name": config.Type, "points": rollingPoints},
		},
		"window":    config.Window,
		"timeField": config.TimeField,
		"grouped":   true,
	}, nil
}

// --- Formula ---

func (e *Engine) executeFormula(card models.Card, inputs map[uuid.UUID]*models.CardResult) (interface{}, error) {
	var config struct {
		Expression string `json:"expression"`
		Mode       string `json:"mode"` // "aggregate" or "per_row"
	}
	if err := json.Unmarshal(card.Config, &config); err != nil {
		return nil, fmt.Errorf("parsing formula config: %w", err)
	}
	if config.Expression == "" {
		return map[string]interface{}{"value": 0, "label": "No formula"}, nil
	}

	node, err := formula.Parse(config.Expression)
	if err != nil {
		return nil, fmt.Errorf("parsing formula: %w", err)
	}

	inputObjects := extractObjects(inputs)

	if config.Mode == "per_row" {
		// Evaluate formula for each row
		results := formula.EvaluatePerRow(node, inputObjects)
		var rows []map[string]interface{}
		for i, obj := range inputObjects {
			row := map[string]interface{}{}
			for k, v := range obj {
				row[k] = v
			}
			row["_formula_result"] = results[i].AsFloat()
			rows = append(rows, row)
		}
		cols := []string{}
		if len(inputObjects) > 0 {
			for k := range inputObjects[0] {
				if k != "objectTypeId" && k != "objectTypeID" {
					cols = append(cols, k)
				}
			}
		}
		sort.Strings(cols)
		cols = append(cols, "_formula_result")
		return map[string]interface{}{
			"rows":       rows,
			"columns":    cols,
			"totalCount": len(rows),
			"mode":       "per_row",
		}, nil
	}

	// Aggregate mode (default)
	result := formula.Evaluate(node, inputObjects)
	return map[string]interface{}{
		"value":      result.AsFloat(),
		"label":      card.Label,
		"expression": config.Expression,
		"type":       "formula",
	}, nil
}

// --- Overlay Chart ---

func (e *Engine) executeOverlayChart(card models.Card, inputs map[uuid.UUID]*models.CardResult) (interface{}, error) {
	// Overlay combines chart data from multiple input charts
	var layers []map[string]interface{}

	for inputID, input := range inputs {
		if input.Error != "" || len(input.Data) == 0 {
			continue
		}

		var inputData map[string]interface{}
		if err := json.Unmarshal(input.Data, &inputData); err != nil {
			continue
		}

		layer := map[string]interface{}{
			"cardId": inputID.String(),
		}

		// Check for chart data
		if chartData, ok := inputData["chartData"]; ok {
			layer["chartData"] = chartData
			layer["type"] = "bar" // default
		}
		// Check for time series data
		if series, ok := inputData["series"]; ok {
			layer["series"] = series
			layer["type"] = "line"
		}

		// Override type from config if provided
		var cfg struct {
			LayerTypes map[string]string `json:"layerTypes"`
		}
		if err := json.Unmarshal(card.Config, &cfg); err == nil && cfg.LayerTypes != nil {
			if lt, ok := cfg.LayerTypes[inputID.String()]; ok {
				layer["type"] = lt
			}
		}

		layers = append(layers, layer)
	}

	return map[string]interface{}{
		"layers": layers,
		"config": card.Config,
	}, nil
}

// --- Action ---

func (e *Engine) executeAction(ctx context.Context, card models.Card, inputs map[uuid.UUID]*models.CardResult) (interface{}, error) {
	var config struct {
		ActionTypeID      string            `json:"actionTypeId"`
		ParameterMappings map[string]string `json:"parameterMappings"`
		Execute           bool              `json:"execute"`
	}
	if err := json.Unmarshal(card.Config, &config); err != nil {
		return nil, fmt.Errorf("parsing action config: %w", err)
	}

	// Build parameters from inputs
	params := map[string]interface{}{}
	inputObjects := extractObjects(inputs)
	for paramName, fieldRef := range config.ParameterMappings {
		if len(inputObjects) > 0 {
			parts := strings.SplitN(fieldRef, ".", 2)
			if len(parts) == 2 {
				params[paramName] = inputObjects[0][parts[1]]
			} else {
				params[paramName] = inputObjects[0][fieldRef]
			}
		}
	}

	result := map[string]interface{}{
		"actionTypeId": config.ActionTypeID,
		"parameters":   params,
		"status":       "ready",
	}

	if config.Execute {
		// Call ontology action API
		err := e.ontologyClient.ExecuteAction(ctx, config.ActionTypeID, params)
		if err != nil {
			result["status"] = "error"
			result["error"] = err.Error()
		} else {
			result["status"] = "executed"
		}
	}

	return result, nil
}

func (e *Engine) executeParameter(card models.Card) (interface{}, error) {
	var config map[string]interface{}
	if err := json.Unmarshal(card.Config, &config); err != nil {
		return card.Config, nil
	}
	value := config["value"]
	if value == nil {
		value = config["defaultValue"]
	}
	return map[string]interface{}{
		"value":  value,
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

func parseTime(s string) time.Time {
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
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
