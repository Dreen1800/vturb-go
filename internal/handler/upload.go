package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"vturb-go/internal/platform"
	"vturb-go/internal/service"
)

type UploadURLRequest struct {
	Filename string `json:"filename"`
	Filesize int64  `json:"filesize"`
}

type UploadURLResponse struct {
	UploadURL string `json:"upload_url"`
	StreamUID string `json:"stream_uid"`
}

func GenerateUploadURLHandler(svc *service.VideoService, cf *platform.CloudflareClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		var req UploadURLRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Filename == "" || req.Filesize == 0 {
			http.Error(w, "filename and filesize are required", http.StatusBadRequest)
			return
		}

		// Verify video exists
		video, err := svc.GetVideo(r.Context(), id)
		if err != nil {
			http.Error(w, "video not found", http.StatusNotFound)
			return
		}

		// Create TUS upload with Cloudflare
		resp, err := cf.CreateTUSUpload(req.Filename, req.Filesize)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Save stream_uid
		if err := svc.UpdateStreamUID(r.Context(), video.ID, resp.StreamUID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func FinalizeUploadHandler(svc *service.VideoService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		// Verify video exists
		_, err := svc.GetVideo(r.Context(), id)
		if err != nil {
			http.Error(w, "video not found", http.StatusNotFound)
			return
		}

		// Update status to uploading
		if err := svc.FinalizeUpload(r.Context(), id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "uploading",
			"message": "Upload finalized. Video will be processed by Cloudflare.",
		})
	}
}
