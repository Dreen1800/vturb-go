package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/hibiken/asynq"
	"vturb-go/internal/service"
	"vturb-go/internal/worker"
)

type CloudflareWebhookPayload struct {
	UID        string `json:"uid"`
	Status     struct {
		State string `json:"state"`
	} `json:"status"`
	Playback   struct {
		HLS string `json:"hls"`
	} `json:"playback"`
	Thumbnail  string  `json:"thumbnail"`
	Duration   float64 `json:"duration"`
}

func CloudflareWebhookHandler(svc *service.VideoService, client *asynq.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload CloudflareWebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		log.Printf("📩 Webhook received for stream %s, status: %s", payload.UID, payload.Status.State)

		// Find video by stream_uid
		video, err := svc.GetVideoByStreamUID(r.Context(), payload.UID)
		if err != nil {
			log.Printf("Video not found for stream %s, error: %v", payload.UID, err)
		}

		// If status is ready, update immediately
		if payload.Status.State == "ready" && video != nil {
			durationSec := int(payload.Duration)
			if err := svc.UpdateStreamInfo(r.Context(), video.ID, payload.Playback.HLS, payload.Thumbnail, durationSec, payload.Status.State); err != nil {
				log.Printf("Failed to update video info: %v", err)
			} else {
				log.Printf("✅ Video %s updated to ready status", video.ID)
			}
		}

		// Queue sync job for background processing
		if video != nil {
			task, err := worker.NewSyncVideoStatusTask(video.ID, payload.UID)
			if err != nil {
				log.Printf("Failed to create sync task: %v", err)
			} else {
				if _, err := client.Enqueue(task); err != nil {
					log.Printf("Failed to enqueue sync task: %v", err)
				}
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}
