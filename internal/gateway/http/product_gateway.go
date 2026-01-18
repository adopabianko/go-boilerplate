package httpgateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go-boilerplate/internal/entity"
)

type ProductGateway interface {
	GetProducts(ctx context.Context, limit, skip int) ([]entity.Product, int64, error)
}

type productGateway struct {
	baseURL string
	client  *http.Client
}

func NewProductGateway(baseURL string) ProductGateway {
	return &productGateway{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type dummyJSONResponse struct {
	Products []entity.Product `json:"products"`
	Total    int64            `json:"total"`
	Skip     int              `json:"skip"`
	Limit    int              `json:"limit"`
}

func (g *productGateway) GetProducts(ctx context.Context, limit, skip int) ([]entity.Product, int64, error) {
	url := fmt.Sprintf("%s?limit=%d&skip=%d", g.baseURL, limit, skip)

	var lastErr error
	maxRetries := 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Simple backoff: 200ms, 400ms, etc.
			time.Sleep(time.Duration(attempt*200) * time.Millisecond)
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, 0, err
		}

		resp, err := g.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("failed to fetch products, status: %d", resp.StatusCode)
			if resp.StatusCode >= 500 {
				// Retry on Server Errors
				continue
			}
			// Don't retry on client errors (4xx) unless strict req
			return nil, 0, lastErr
		}

		var data dummyJSONResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, 0, err
		}

		return data.Products, data.Total, nil
	}

	return nil, 0, fmt.Errorf("max retries reached: %w", lastErr)
}
