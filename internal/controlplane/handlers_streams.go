package controlplane

import (
	"encoding/json"
	"net/http"

	"arkis_test/internal/errors"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (s *Server) CreateStream(w http.ResponseWriter, r *http.Request) {
	var req struct {
		StreamName string `json:"stream_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, errors.ErrInvalidInput.Error(), http.StatusBadRequest)
		return
	}

	err := s.streamService.Create(r.Context(), req.StreamName)
	if err != nil {
		if errors.Is(err, errors.ErrAlreadyExists) {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]string{"message": "stream created"})
	if err != nil {
		zap.L().Error("failed to encode response", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) PublishToStream(w http.ResponseWriter, r *http.Request) {
	streamName := chi.URLParam(r, "streamName")
	var req struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, errors.ErrInvalidInput.Error(), http.StatusBadRequest)
		return
	}

	err := s.streamService.SendMessage(r.Context(), streamName, req.Message)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]string{"message": "message sent"})
	if err != nil {
		zap.L().Error("failed to encode response", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) ConsumeFromStream(w http.ResponseWriter, r *http.Request) {
	streamName := chi.URLParam(r, "streamName")

	msg, err := s.streamService.ReceiveMessage(streamName)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Here, we assume `msg` is already a JSON string as the db returns the raw message directly
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(msg))
	if err != nil {
		zap.L().Error("failed to write response", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
