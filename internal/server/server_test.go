package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexandrmotologa/tofuforge/internal/model"
)

func TestServerEndpoints(t *testing.T) {
	initialPlan := &model.AnalyzedPlan{
		FormatVersion: "1.2",
		Stats: model.SummaryStats{
			TotalResources: 2,
			CreateCount:    2,
		},
	}

	srv := NewServer(initialPlan)
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	// 1. Test /api/health
	resp, err := http.Get(ts.URL + "/api/health")
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	// 2. Test /api/plan
	resp, err = http.Get(ts.URL + "/api/plan")
	if err != nil {
		t.Fatalf("get plan failed: %v", err)
	}
	var planResp model.AnalyzedPlan
	if err := json.NewDecoder(resp.Body).Decode(&planResp); err != nil {
		t.Fatalf("failed to decode plan response: %v", err)
	}
	if planResp.Stats.TotalResources != 2 {
		t.Errorf("expected 2 total resources, got %d", planResp.Stats.TotalResources)
	}

	// 3. Test /api/upload
	validPlanJSON := `{
		"format_version": "1.2",
		"resource_changes": [
			{
				"address": "aws_s3_bucket.test",
				"type": "aws_s3_bucket",
				"change": {
					"actions": ["create"],
					"after": { "bucket": "test-bkt" }
				}
			}
		]
	}`

	resp, err = http.Post(ts.URL+"/api/upload", "application/json", bytes.NewBufferString(validPlanJSON))
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("upload status expected 200, got %d", resp.StatusCode)
	}

	// 4. Test /api/markdown
	resp, err = http.Get(ts.URL + "/api/markdown")
	if err != nil {
		t.Fatalf("markdown failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for markdown, got %d", resp.StatusCode)
	}
}
