package otlptracegrpc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

type CollectorResp struct {
	Collectors struct {
		Traces  []string `json:"traces"`
		Metrics []string `json:"metrics"`
	} `json:"collectors"`
}

func (c *client) loadCollectors(ctx context.Context) (*CollectorResp, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.configService, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bts, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var collector CollectorResp
	if err = json.Unmarshal(bts, &resp); err != nil {
		return nil, err
	}

	return &collector, nil
}
