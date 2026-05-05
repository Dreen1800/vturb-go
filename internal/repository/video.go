package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"vturb-go/internal/model"
)

type VideoRepository struct {
	db *pgxpool.Pool
}

func NewVideoRepository(db *pgxpool.Pool) *VideoRepository {
	return &VideoRepository{db: db}
}

func (r *VideoRepository) Create(ctx context.Context, video *model.Video) error {
	query := `
		INSERT INTO videos (title, embed_token, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query, video.Title, video.EmbedToken, video.Status).
		Scan(&video.ID, &video.CreatedAt, &video.UpdatedAt)
}

func (r *VideoRepository) GetByID(ctx context.Context, id string) (*model.Video, error) {
	query := `
		SELECT id, title, status, embed_token, manifest_url, thumbnail_url, 
		       duration_sec, stream_uid, stream_status, created_at, updated_at
		FROM videos WHERE id = $1
	`
	video := &model.Video{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&video.ID, &video.Title, &video.Status, &video.EmbedToken,
		&video.ManifestURL, &video.ThumbnailURL, &video.DurationSec,
		&video.StreamUID, &video.StreamStatus, &video.CreatedAt, &video.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return video, nil
}

func (r *VideoRepository) GetByEmbedToken(ctx context.Context, token string) (*model.Video, error) {
	query := `
		SELECT id, title, status, embed_token, manifest_url, thumbnail_url, 
		       duration_sec, stream_uid, stream_status, created_at, updated_at
		FROM videos WHERE embed_token = $1
	`
	video := &model.Video{}
	err := r.db.QueryRow(ctx, query, token).Scan(
		&video.ID, &video.Title, &video.Status, &video.EmbedToken,
		&video.ManifestURL, &video.ThumbnailURL, &video.DurationSec,
		&video.StreamUID, &video.StreamStatus, &video.CreatedAt, &video.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return video, nil
}

func (r *VideoRepository) List(ctx context.Context) ([]*model.Video, error) {
	query := `
		SELECT id, title, status, embed_token, manifest_url, thumbnail_url, 
		       duration_sec, stream_uid, stream_status, created_at, updated_at
		FROM videos ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []*model.Video
	for rows.Next() {
		video := &model.Video{}
		err := rows.Scan(
			&video.ID, &video.Title, &video.Status, &video.EmbedToken,
			&video.ManifestURL, &video.ThumbnailURL, &video.DurationSec,
			&video.StreamUID, &video.StreamStatus, &video.CreatedAt, &video.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		videos = append(videos, video)
	}

	return videos, rows.Err()
}

func (r *VideoRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE videos SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, id)
	return err
}

func (r *VideoRepository) UpdateStreamInfo(ctx context.Context, id string, manifestURL, thumbnailURL string, durationSec int, streamStatus string) error {
	query := `
		UPDATE videos 
		SET manifest_url = $1, thumbnail_url = $2, duration_sec = $3, 
		    stream_status = $4, status = $5, updated_at = NOW()
		WHERE id = $6
	`
	_, err := r.db.Exec(ctx, query, manifestURL, thumbnailURL, durationSec, streamStatus, model.StatusReady, id)
	return err
}

func (r *VideoRepository) GetByStreamUID(ctx context.Context, streamUID string) (*model.Video, error) {
	query := `
		SELECT id, title, status, embed_token, manifest_url, thumbnail_url, 
		       duration_sec, stream_uid, stream_status, created_at, updated_at
		FROM videos WHERE stream_uid = $1
	`
	video := &model.Video{}
	err := r.db.QueryRow(ctx, query, streamUID).Scan(
		&video.ID, &video.Title, &video.Status, &video.EmbedToken,
		&video.ManifestURL, &video.ThumbnailURL, &video.DurationSec,
		&video.StreamUID, &video.StreamStatus, &video.CreatedAt, &video.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return video, nil
}

func (r *VideoRepository) UpdateStreamUID(ctx context.Context, id string, streamUID string) error {
	query := `UPDATE videos SET stream_uid = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(ctx, query, streamUID, id)
	return err
}
