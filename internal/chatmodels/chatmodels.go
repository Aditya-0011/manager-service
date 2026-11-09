package chatmodels

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	modelsURL   = "https://ai-gateway.vercel.sh/v1/models"
	maxBodySize = 16 << 20
)

var client = &http.Client{Timeout: 5 * time.Second}

type Model struct {
	ID   string
	Name string
}

type catalog struct {
	Data []struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Type    string `json:"type"`
		Pricing *struct {
			Input            string `json:"input"`
			Output           string `json:"output"`
			VariesByProvider bool   `json:"varies_by_provider"`
		} `json:"pricing"`
	} `json:"data"`
}

func ListAvailable(ctx context.Context) ([]Model, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, modelsURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("model list returned status %d", res.StatusCode)
	}

	var parsed catalog
	if err := json.NewDecoder(io.LimitReader(res.Body, maxBodySize)).Decode(&parsed); err != nil {
		return nil, err
	}

	models := make([]Model, 0, 8)
	for _, entry := range parsed.Data {
		if entry.Type != "language" || entry.Pricing == nil || entry.Pricing.VariesByProvider {
			continue
		}
		if isZero(entry.Pricing.Input) && isZero(entry.Pricing.Output) {
			models = append(models, Model{ID: entry.ID, Name: entry.Name})
		}
	}
	return models, nil
}

func isZero(value string) bool {
	parsed, err := strconv.ParseFloat(value, 64)
	return err == nil && parsed == 0
}
