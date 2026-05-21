package dadata

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

const (
	host          = "https://suggestions.dadata.ru/"
	suggestionApi = "suggestions/api/4_1/rs/suggest/address"
	apiKey        = "4c23455eba1a8537c0f936c66eebf9114141696b"
)

type Dadata interface {
	Suggest(ctx context.Context, address, userAgent string) ([]SearchResults, error)
}

type dadata struct {
	cl *http.Client
}

func NewDadata() Dadata {
	transport := &http.Transport{
		TLSHandshakeTimeout: 5 * time.Second,
	}

	return &dadata{
		cl: &http.Client{
			Timeout:   5 * time.Second,
			Transport: transport,
		},
	}
}

type SuggestParams struct {
	Query     string     `json:"query"`
	FromBound Bound      `json:"from_bound"`
	ToBound   Bound      `json:"to_bound"`
	Locations []Location `json:"locations"`
}
type Bound struct {
	Value string `json:"value"`
}
type Location struct {
	City string `json:"city"`
}
type SuggestResponse struct {
	Suggestions []SuggestResult `json:"suggestions"`
}
type SuggestResult struct {
	Value       string      `json:"value"`
	SuggestData SuggestData `json:"data"`
}
type SuggestData struct {
	City           string `json:"city"`
	StreetWithType string `json:"street_with_type"`
	Street         string `json:"street"`
	GeoLat         string `json:"geo_lat"`
	GeoLon         string `json:"geo_lon"`
	Stead          string `json:"stead"`
	House          string `json:"house"`
	SteadType      string `json:"stead_type"`
	HouseType      string `json:"house_type"`
}

func (s *dadata) Suggest(ctx context.Context, address, userAgent string) ([]SearchResults, error) {
	baseURL := host + suggestionApi

	params := SuggestParams{
		Query:     address,
		FromBound: Bound{Value: "country"},
		ToBound:   Bound{Value: "house"},
		Locations: []Location{
			{City: "якутск"},
		},
	}

	payload, err := json.Marshal(params)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal payload")
	}

	reqBody := bytes.NewBuffer(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL, reqBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+apiKey)

	maxRetries := 5
	var resp *http.Response
	var searchResults []SearchResults
	var results SuggestResponse

	for i := 1; i <= maxRetries; i++ {
		resp, err = s.cl.Do(req)
		if err != nil {
			slog.Error("failed to send suggestion request", "error", err, "attempt", i)
			time.Sleep(time.Second * 2)
			continue
		}

		if resp == nil {
			return nil, errors.Wrap(err, "failed to send suggestion request, resp is nil")
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("unexpected status code: %d, body: %s, url: %s", resp.StatusCode, string(body), baseURL)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		resp.Body.Close()

		if err := json.Unmarshal(body, &results); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		break
	}

	seen := make(map[string]struct{})

	for _, result := range results.Suggestions {
		var fullAddress string

		if result.SuggestData.SteadType == "уч" && result.SuggestData.Stead != "" {
			fullAddress = result.SuggestData.Street + " " + result.SuggestData.Stead
		} else if result.SuggestData.HouseType == "д" && result.SuggestData.House != "" {
			fullAddress = result.SuggestData.Street + " " + result.SuggestData.House
		}

		if fullAddress == "" {
			continue
		}
		if _, exists := seen[fullAddress]; exists {
			continue
		}

		seen[fullAddress] = struct{}{}
		searchResults = append(searchResults, SearchResults{
			Lat:     result.SuggestData.GeoLat,
			Lon:     result.SuggestData.GeoLon,
			Address: fullAddress,
		})
	}

	return searchResults, nil
}
