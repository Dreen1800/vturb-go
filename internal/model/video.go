package model

import "time"

const (
	StatusPending = "pending"
	StatusReady   = "ready"
	StatusFailed  = "failed"
)

type Video struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	EmbedToken   string    `json:"embed_token"`
	ManifestURL  *string   `json:"manifest_url,omitempty"`
	ThumbnailURL *string   `json:"thumbnail_url,omitempty"`
	DurationSec  *int      `json:"duration_sec,omitempty"`
	StreamUID    *string   `json:"stream_uid,omitempty"`
	StreamStatus *string   `json:"stream_status,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
