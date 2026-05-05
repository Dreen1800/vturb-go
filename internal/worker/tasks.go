package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"vturb-go/internal/platform"
	"vturb-go/internal/service"
)

const (
	TypeSyncVideoStatus = "video:sync_status"
)

type SyncVideoStatusPayload struct {
	VideoID   string `json:"video_id"`
	StreamUID string `json:"stream_uid"`
}

func NewSyncVideoStatusTask(videoID, streamUID string) (*asynq.Task, error) {
	payload := SyncVideoStatusPayload{
		VideoID:   videoID,
		StreamUID: streamUID,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeSyncVideoStatus, data), nil
}

func HandleSyncVideoStatusTask(svc *service.VideoService, cf *platform.CloudflareClient) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var payload SyncVideoStatusPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		log.Printf("Processing sync_video_status for video %s, stream %s", payload.VideoID, payload.StreamUID)

		// Get video info from Cloudflare
		info, err := cf.GetVideoInfo(payload.StreamUID)
		if err != nil {
			return fmt.Errorf("failed to get video info: %w", err)
		}

		// Update video in database
		durationSec := int(info.Duration)
		if err := svc.UpdateStreamInfo(ctx, payload.VideoID, info.PlaybackURL, info.ThumbnailURL, durationSec, info.Status); err != nil {
			return fmt.Errorf("failed to update video info: %w", err)
		}

		log.Printf("✅ Video %s synced successfully. Status: %s", payload.VideoID, info.Status)
		return nil
	}
}
