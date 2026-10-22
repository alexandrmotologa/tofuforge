package server

import (
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/alexandrmotologa/tofuforge/internal/engine"
	"github.com/alexandrmotologa/tofuforge/internal/model"
	"github.com/alexandrmotologa/tofuforge/internal/parser"
	"github.com/alexandrmotologa/tofuforge/internal/reporter"
)

// Server coordinates the web visualizer and HTTP API.
type Server struct {
	mu         sync.RWMutex
	activePlan *model.AnalyzedPlan
	samples    map[string]*model.AnalyzedPlan
}

// NewServer creates a new Server instance with an initial plan.
func NewServer(initialPlan *model.AnalyzedPlan) *Server {
	s := &Server{
		activePlan: initialPlan,
		samples:    make(map[string]*model.AnalyzedPlan),
	}
	s.loadSampleFixtures()
	return s
}

func (s *Server) loadSampleFixtures() {
	// Attempt loading from fixtures directory if accessible
	paths := map[string]string{
		"aws_ecommerce": filepath.Join("fixtures", "aws_production_ecommerce.json"),
		"gcp_micro":     filepath.Join("fixtures", "gcp_microservices.json"),
		"clean_infra":   filepath.Join("fixtures", "clean_infrastructure.json"),
	}

	for id, p := range paths {
		if data, err := os.ReadFile(p); err == nil {
			if raw, err := parser.ParsePlan(bytes.NewReader(data)); err == nil {
				s.samples[id] = engine.Analyze(raw)
			}
		}
	}
}

// Handler returns the HTTP handler with all registered routes.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// API Routes
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/plan", s.handlePlan)
	mux.HandleFunc("/api/upload", s.handleUpload)
	mux.HandleFunc("/api/samples", s.handleSamples)
	mux.HandleFunc("/api/samples/select", s.handleSelectSample)
	mux.HandleFunc("/api/markdown", s.handleMarkdown)

	// Static assets handler with SPA fallback
	distSub, err := fs.Sub(DistFS, "dist")
	if err == nil {
		fileServer := http.FileServer(http.FS(distSub))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}

			// Try to open file directly
			path := strings.TrimPrefix(r.URL.Path, "/")
			if path == "" {
				path = "index.html"
			}

			f, err := distSub.Open(path)
			if err != nil {
				// Fallback to index.html for SPA client-side routing
				indexFile, err := distSub.Open("index.html")
				if err != nil {
					http.Error(w, "Web Studio assets not found", http.StatusNotFound)
					return
				}
				defer indexFile.Close()
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				io.Copy(w, indexFile)
				return
			}
			f.Close()

			fileServer.ServeHTTP(w, r)
		})
	}

	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "tofuforge",
	})
}

func (s *Server) handlePlan(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	if s.activePlan == nil {
		// Return empty object
		json.NewEncoder(w).Encode(map[string]interface{}{
			"nodes": []string{},
			"edges": []string{},
			"stats": model.SummaryStats{},
		})
		return
	}

	json.NewEncoder(w).Encode(s.activePlan)
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read body
	body, err := io.ReadAll(io.LimitReader(r.Body, 50*1024*1024)) // 50MB limit
	if err != nil {
		http.Error(w, "Failed to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	raw, err := parser.ParsePlan(bytes.NewReader(body))
	if err != nil {
		http.Error(w, "Invalid Terraform/OpenTofu plan JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	analyzed := engine.Analyze(raw)

	s.mu.Lock()
	s.activePlan = analyzed
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analyzed)
}

type SampleMeta struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (s *Server) handleSamples(w http.ResponseWriter, r *http.Request) {
	samples := []SampleMeta{
		{
			ID:          "aws_ecommerce",
			Title:       "AWS E-Commerce (Critical DB Replacement)",
			Description: "Production AWS environment where an RDS database changes identifier, cascading into API and DNS recreation.",
		},
		{
			ID:          "gcp_micro",
			Title:       "GCP Microservices (Open Ingress Security Violation)",
			Description: "Google Cloud architecture with 0.0.0.0/0 firewall ingress to RDP and publicly exposed Cloud SQL.",
		},
		{
			ID:          "clean_infra",
			Title:       "Clean Infrastructure (Safe Plan)",
			Description: "Safe Terraform plan adding secure VPC, subnets, and restricted internal HTTPS security groups.",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(samples)
}

func (s *Server) handleSelectSample(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SampleID string `json:"sample_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	samplePlan, exists := s.samples[req.SampleID]
	if !exists {
		// Re-attempt loading sample if not cached
		var fixturePath string
		switch req.SampleID {
		case "aws_ecommerce":
			fixturePath = filepath.Join("fixtures", "aws_production_ecommerce.json")
		case "gcp_micro":
			fixturePath = filepath.Join("fixtures", "gcp_microservices.json")
		case "clean_infra":
			fixturePath = filepath.Join("fixtures", "clean_infrastructure.json")
		}

		if fixturePath != "" {
			if data, err := os.ReadFile(fixturePath); err == nil {
				if raw, err := parser.ParsePlan(bytes.NewReader(data)); err == nil {
					samplePlan = engine.Analyze(raw)
					s.samples[req.SampleID] = samplePlan
				}
			}
		}
	}

	if samplePlan == nil {
		s.mu.Unlock()
		http.Error(w, "Sample plan not found", http.StatusNotFound)
		return
	}

	s.activePlan = samplePlan
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(samplePlan)
}

func (s *Server) handleMarkdown(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	plan := s.activePlan
	s.mu.RUnlock()

	if plan == nil {
		http.Error(w, "No active plan loaded", http.StatusNotFound)
		return
	}

	md := reporter.GenerateMarkdownReport(plan)

	format := r.URL.Query().Get("format")
	if format == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"markdown": md})
		return
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Write([]byte(md))
}
