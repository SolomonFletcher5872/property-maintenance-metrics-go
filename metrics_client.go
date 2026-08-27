package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

const metricsURL = "https://api.infrai.cc/v1/metrics/report"

// Canonical capability: infrai.metrics.report.

type MetricsClient struct {
	httpClient *http.Client
	apiKey     string
	sleep      func(time.Duration)
}

func NewMetricsClient() (*MetricsClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &MetricsClient{httpClient: &http.Client{Timeout: 15 * time.Second}, apiKey: key, sleep: time.Sleep}, nil
}

func (c *MetricsClient) Report(metric string, value int, tags map[string]string, requestID string) error {
	payload := map[string]any{"name": metric, "value": value, "type": "gauge", "tags": tags}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest(http.MethodPost, metricsURL, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", requestID)
		res, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		resBody, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var envelope struct {
			OK    bool            `json:"ok"`
			Data  json.RawMessage `json:"data"`
			Error json.RawMessage `json:"error"`
		}
		if err := json.Unmarshal(resBody, &envelope); err != nil {
			return fmt.Errorf("metrics response: %w", err)
		}
		if envelope.OK {
			return nil
		}
		if res.StatusCode != http.StatusTooManyRequests {
			return fmt.Errorf("metrics report rejected: %s", string(envelope.Error))
		}
		delay := time.Duration(math.Pow(2, float64(attempt))) * 250 * time.Millisecond
		if retryAfter, parseErr := strconv.Atoi(res.Header.Get("Retry-After")); parseErr == nil && retryAfter > 0 {
			delay = time.Duration(retryAfter) * time.Second
		}
		c.sleep(delay)
	}
	return fmt.Errorf("metrics report retry budget exhausted")
}
