package controlplane

import (
	"context"
	"net/http"
	"time"

	"arkis_test/internal/application"
	"arkis_test/internal/errors"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type Server struct {
	httpServer    *http.Server
	streamService *application.StreamService
}

func NewServer(streamService *application.StreamService) *Server {
	s := &Server{
		streamService: streamService,
	}

	s.setupRouter()
	return s
}

func (s *Server) setupRouter() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", s.healthCheck)

	r.Post("/streams", s.CreateStream)
	r.Post("/streams/{streamName}/messages", s.PublishToStream)
	r.Get("/streams/{streamName}/messages", s.ConsumeFromStream)

	s.httpServer = &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) Start() {
	zap.L().Info("starting server on", zap.String("address", s.httpServer.Addr))
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		zap.L().Fatal("could not listen on", zap.String("address", s.httpServer.Addr), zap.Error(err))
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	zap.L().Info("cleaning up server resources")
	s.streamService.Shutdown()
	zap.L().Info("shutting down server")
	return s.httpServer.Shutdown(ctx)
}
