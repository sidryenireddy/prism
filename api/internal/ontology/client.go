package ontology

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
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	baseURL := os.Getenv("ONTOLOGY_URL")
	if baseURL == "" {
		baseURL = "https://ontology.rebelinc.ai"
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

// --- Types matching the Ontology API ---

type ObjectType struct {
	ID          string            `json:"id"`
	APIName     string            `json:"apiName"`
	DisplayName string            `json:"displayName"`
	Properties  map[string]Property `json:"properties"`
}

type Property struct {
	APIName     string `json:"apiName"`
	DisplayName string `json:"displayName"`
	DataType    string `json:"dataType"`
}

type ObjectInstance struct {
	ID           string                 `json:"id"`
	ObjectTypeID string                 `json:"objectTypeId"`
	Properties   map[string]interface{} `json:"properties"`
}

type SearchResult struct {
	Objects    []ObjectInstance `json:"objects"`
	TotalCount int64           `json:"totalCount"`
	Page       int             `json:"page"`
	PageSize   int             `json:"pageSize"`
}

type FilterClause struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type SearchQuery struct {
	ObjectTypeID string         `json:"objectTypeId"`
	Query        string         `json:"query"`
	Filters      []FilterClause `json:"filters"`
	PageSize     int            `json:"pageSize"`
	Page         int            `json:"page"`
}

type AggMetric struct {
	Field string `json:"field"`
	Type  string `json:"type"`
}

type AggregationQuery struct {
	ObjectTypeID string         `json:"objectTypeId"`
	GroupBy      []string       `json:"groupBy"`
	Metrics      []AggMetric    `json:"metrics"`
	Filters      map[string]interface{} `json:"filters"`
}

// --- API methods ---

func (c *Client) GetObjectTypes(ctx context.Context, ontologyID string) ([]ObjectType, error) {
	if ontologyID == "" {
		ontologyID = "default"
	}
	url := fmt.Sprintf("%s/api/v1/ontologies/%s/object-types", c.baseURL, ontologyID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching object types: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ontology API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result []ObjectType
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) SearchObjects(ctx context.Context, query SearchQuery) (*SearchResult, error) {
	if query.PageSize == 0 {
		query.PageSize = 100
	}

	body, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v1/search/objects", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("searching objects: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ontology API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) AggregateObjects(ctx context.Context, query AggregationQuery) (json.RawMessage, error) {
	body, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v1/search/aggregate", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("aggregating objects: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ontology API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) GetLinkedObjects(ctx context.Context, objectID, objectTypeID string) (*SearchResult, error) {
	url := fmt.Sprintf("%s/api/v1/links?objectId=%s&objectTypeId=%s", c.baseURL, objectID, objectTypeID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching linked objects: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ontology API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
