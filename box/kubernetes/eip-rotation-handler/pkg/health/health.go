package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// Response health check response data
type Response struct {
	Status       string  `json:"status"`
	Code         int     `json:"code"`
	Service      string  `json:"service"`
	ResponseTime float64 `json:"response_time_ms"`
	Timestamp    string  `json:"timestamp"`
}

// Server health check HTTP server
type Server struct {
	server *http.Server
	log    *logrus.Logger
}

// New creates a new health check server
func New(log *logrus.Logger) *Server {
	mux := http.NewServeMux()

	s := &Server{
		server: &http.Server{
			Addr:    ":8080",
			Handler: mux,
		},
		log: log,
	}

	mux.HandleFunc("/healthz", s.healthzHandler)

	return s
}

// healthzHandler handles /healthz endpoint
func (s *Server) healthzHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	responseTimeMs := float64(time.Since(startTime).Nanoseconds()) / 1e6

	response := Response{
		Status:       "healthy",
		Code:         http.StatusOK,
		Service:      "eip-rotation-handler",
		ResponseTime: responseTimeMs,
		Timestamp:    startTime.UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.log.Debugf("Health check request processed - response_time: %.2f ms, status_code: %d", responseTimeMs, http.StatusOK)
}

// Start starts the health check server
func (s *Server) Start(ctx context.Context) {
	go func() {
		s.log.Info("Starting health check server on :8080")
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.WithError(err).Error("Health server failed")
		}
	}()

	go func() {
		<-ctx.Done()
		s.log.Info("Shutting down health check server")
		s.Shutdown()
	}()
}

// Shutdown gracefully shuts down the health check server
func (s *Server) Shutdown() {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.log.WithError(err).Error("Health server shutdown failed")
	}
}
