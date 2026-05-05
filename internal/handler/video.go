package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"vturb-go/internal/service"
)

type CreateVideoRequest struct {
	Title string `json:"title"`
}

type CreateVideoResponse struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	EmbedToken string `json:"embed_token"`
}

func CreateVideoHandler(svc *service.VideoService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateVideoRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Title == "" {
			http.Error(w, "title is required", http.StatusBadRequest)
			return
		}

		video, err := svc.CreateVideo(r.Context(), req.Title)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(video)
	}
}

func ListVideosHandler(svc *service.VideoService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		videos, err := svc.ListVideos(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(videos)
	}
}

func GetVideoHandler(svc *service.VideoService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		video, err := svc.GetVideo(r.Context(), id)
		if err != nil {
			http.Error(w, "video not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(video)
	}
}
