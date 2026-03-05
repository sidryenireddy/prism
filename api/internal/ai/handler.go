package ai

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/sidryenireddy/prism/api/internal/mockdata"
	"github.com/sidryenireddy/prism/api/internal/models"
)

type GenerateRequest struct {
	AnalysisID string        `json:"analysis_id"`
	Prompt     string        `json:"prompt"`
	Cards      []models.Card `json:"cards"`
}

type GenerateResponse struct {
	Cards []GeneratedCard `json:"cards"`
}

type GeneratedCard struct {
	CardType     string          `json:"card_type"`
	Label        string          `json:"label"`
	Config       json.RawMessage `json:"config"`
	PositionX    float64         `json:"position_x"`
	PositionY    float64         `json:"position_y"`
	InputIndex   *int            `json:"input_index,omitempty"`
}

type ConfigureRequest struct {
	Card   models.Card `json:"card"`
	Prompt string      `json:"prompt"`
}

type ConfigureResponse struct {
	Config json.RawMessage `json:"config"`
	Label  string          `json:"label"`
}

func Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return generateFallback(req)
	}

	objectTypes := make([]map[string]interface{}, len(mockdata.ObjectTypes))
	for i, ot := range mockdata.ObjectTypes {
		objectTypes[i] = map[string]interface{}{
			"id":         ot.ID,
			"name":       ot.Name,
			"properties": ot.Properties,
		}
	}

	systemPrompt := `You are an analytics assistant for Prism. Generate card configurations for an analysis canvas.

Available object types:
` + mustJSON(objectTypes) + `

Available card types:
- filter_object_set: {objectTypeId, filters: [{field, operator, value}], query}
- search_around: {linkType} (follows links from input object set)
- set_math_union/set_math_intersection/set_math_difference: no config, takes 2 inputs
- bar_chart/line_chart/pie_chart/scatter_plot/heat_grid: {groupBy, metric, metricType: "count"|"sum"|"avg", colors}
- object_table: {columns: ["prop1","prop2"], pageSize}
- pivot_table: {rowField, columnField, valueField, aggregation}
- count/sum/average/min/max: {objectTypeId, property}
- param_object_selection/param_date_range/param_numeric/param_string/param_boolean: {label, defaultValue}
- action_button: {actionTypeId, parameterMappings}

Return JSON array of cards. Each card has: card_type, label, config, position_x, position_y, input_index (index of card in this array that feeds into this one, or null for root cards).
Position cards in a top-down flow: root cards at y=50, dependent cards at y=250, y=450, etc. Space horizontally at x=100, x=400, x=700.`

	body := map[string]interface{}{
		"model":      "claude-sonnet-4-20250514",
		"max_tokens": 2048,
		"system":     systemPrompt,
		"messages": []map[string]interface{}{
			{"role": "user", "content": fmt.Sprintf("Existing cards: %s\n\nUser request: %s", mustJSON(req.Cards), req.Prompt)},
		},
	}

	bodyBytes, _ := json.Marshal(body)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{
		Timeout:   60 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return generateFallback(req)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return generateFallback(req)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var anthropicResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil || len(anthropicResp.Content) == 0 {
		return generateFallback(req)
	}

	text := anthropicResp.Content[0].Text
	// Extract JSON array from response
	start := -1
	for i, c := range text {
		if c == '[' {
			start = i
			break
		}
	}
	if start == -1 {
		return generateFallback(req)
	}
	depth := 0
	end := -1
	for i := start; i < len(text); i++ {
		if text[i] == '[' {
			depth++
		} else if text[i] == ']' {
			depth--
			if depth == 0 {
				end = i + 1
				break
			}
		}
	}
	if end == -1 {
		return generateFallback(req)
	}

	var cards []GeneratedCard
	if err := json.Unmarshal([]byte(text[start:end]), &cards); err != nil {
		return generateFallback(req)
	}

	return &GenerateResponse{Cards: cards}, nil
}

func Configure(ctx context.Context, req ConfigureRequest) (*ConfigureResponse, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return configureFallback(req)
	}

	systemPrompt := `You are an analytics card configuration assistant. Update the card config based on the user's request.
Return JSON with "config" (the updated config object) and "label" (updated label string).
Card type: ` + string(req.Card.CardType) + `
Current config: ` + string(req.Card.Config) + `
Current label: ` + req.Card.Label

	body := map[string]interface{}{
		"model":      "claude-sonnet-4-20250514",
		"max_tokens": 1024,
		"system":     systemPrompt,
		"messages": []map[string]interface{}{
			{"role": "user", "content": req.Prompt},
		},
	}

	bodyBytes, _ := json.Marshal(body)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{
		Timeout:   60 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return configureFallback(req)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return configureFallback(req)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var anthropicResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(respBody, &anthropicResp); err != nil || len(anthropicResp.Content) == 0 {
		return configureFallback(req)
	}

	text := anthropicResp.Content[0].Text
	start := -1
	for i, c := range text {
		if c == '{' {
			start = i
			break
		}
	}
	if start == -1 {
		return configureFallback(req)
	}
	depth := 0
	end := -1
	for i := start; i < len(text); i++ {
		if text[i] == '{' {
			depth++
		} else if text[i] == '}' {
			depth--
			if depth == 0 {
				end = i + 1
				break
			}
		}
	}
	if end == -1 {
		return configureFallback(req)
	}

	var result ConfigureResponse
	if err := json.Unmarshal([]byte(text[start:end]), &result); err != nil {
		return configureFallback(req)
	}
	return &result, nil
}

func generateFallback(req GenerateRequest) (*GenerateResponse, error) {
	// Smart fallback: parse the prompt for common patterns
	cards := []GeneratedCard{
		{
			CardType:  "filter_object_set",
			Label:     "All Customers",
			Config:    json.RawMessage(`{"objectTypeId":"ot-customers","filters":[],"query":""}`),
			PositionX: 100,
			PositionY: 50,
		},
		{
			CardType:   "bar_chart",
			Label:      "Customer Analysis",
			Config:     json.RawMessage(`{"groupBy":"region","metric":"lifetime_value","metricType":"avg"}`),
			PositionX:  100,
			PositionY:  250,
			InputIndex: intPtr(0),
		},
	}

	prompt := req.Prompt
	if containsAny(prompt, []string{"order", "revenue", "sales"}) {
		cards = []GeneratedCard{
			{
				CardType:  "filter_object_set",
				Label:     "All Orders",
				Config:    json.RawMessage(`{"objectTypeId":"ot-orders","filters":[],"query":""}`),
				PositionX: 100,
				PositionY: 50,
			},
			{
				CardType:   "bar_chart",
				Label:      "Revenue by Region",
				Config:     json.RawMessage(`{"groupBy":"region","metric":"amount","metricType":"sum"}`),
				PositionX:  100,
				PositionY:  250,
				InputIndex: intPtr(0),
			},
		}
		if containsAny(prompt, []string{"count", "how many", "total"}) {
			cards = append(cards, GeneratedCard{
				CardType:   "count",
				Label:      "Total Orders",
				Config:     json.RawMessage(`{"objectTypeId":"ot-orders","property":"id"}`),
				PositionX:  400,
				PositionY:  250,
				InputIndex: intPtr(0),
			})
		}
	}
	if containsAny(prompt, []string{"product", "inventory", "stock"}) {
		cards = []GeneratedCard{
			{
				CardType:  "filter_object_set",
				Label:     "All Products",
				Config:    json.RawMessage(`{"objectTypeId":"ot-products","filters":[],"query":""}`),
				PositionX: 100,
				PositionY: 50,
			},
			{
				CardType:   "bar_chart",
				Label:      "Products by Category",
				Config:     json.RawMessage(`{"groupBy":"category","metric":"price","metricType":"avg"}`),
				PositionX:  100,
				PositionY:  250,
				InputIndex: intPtr(0),
			},
		}
	}
	if containsAny(prompt, []string{"line chart", "trend", "over time", "by month"}) {
		if len(cards) >= 2 {
			cards[1].CardType = "line_chart"
		}
	}
	if containsAny(prompt, []string{"pie", "breakdown", "distribution"}) {
		if len(cards) >= 2 {
			cards[1].CardType = "pie_chart"
		}
	}
	if containsAny(prompt, []string{"table", "list", "show me all"}) {
		cards = append(cards, GeneratedCard{
			CardType:   "object_table",
			Label:      "Data Table",
			Config:     json.RawMessage(`{"columns":[],"pageSize":20}`),
			PositionX:  400,
			PositionY:  250,
			InputIndex: intPtr(0),
		})
	}

	return &GenerateResponse{Cards: cards}, nil
}

func configureFallback(req ConfigureRequest) (*ConfigureResponse, error) {
	return &ConfigureResponse{
		Config: req.Card.Config,
		Label:  req.Card.Label,
	}, nil
}

func intPtr(i int) *int { return &i }

func mustJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func containsAny(s string, subs []string) bool {
	lower := toLower(s)
	for _, sub := range subs {
		if containsStr(lower, toLower(sub)) {
			return true
		}
	}
	return false
}

func containsStr(s, sub string) bool {
	if len(sub) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}
