package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"vturb-go/internal/model"
	"vturb-go/internal/repository"
)

type VideoService struct {
	repo *repository.VideoRepository
}

func NewVideoService(repo *repository.VideoRepository) *VideoService {
	return &VideoService{repo: repo}
}

func generateEmbedToken() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return "vt_" + hex.EncodeToString(bytes), nil
}

func (s *VideoService) CreateVideo(ctx context.Context, title string) (*model.Video, error) {
	token, err := generateEmbedToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate embed token: %w", err)
	}

	video := &model.Video{
		Title:      title,
		Status:     model.StatusPending,
		EmbedToken: token,
	}

	if err := s.repo.Create(ctx, video); err != nil {
		return nil, fmt.Errorf("failed to create video: %w", err)
	}

	return video, nil
}

func (s *VideoService) GetVideo(ctx context.Context, id string) (*model.Video, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *VideoService) GetVideoByToken(ctx context.Context, token string) (*model.Video, error) {
	return s.repo.GetByEmbedToken(ctx, token)
}

func (s *VideoService) GetVideoByStreamUID(ctx context.Context, streamUID string) (*model.Video, error) {
	return s.repo.GetByStreamUID(ctx, streamUID)
}

func (s *VideoService) ListVideos(ctx context.Context) ([]*model.Video, error) {
	return s.repo.List(ctx)
}

func (s *VideoService) UpdateStreamUID(ctx context.Context, id string, streamUID string) error {
	return s.repo.UpdateStreamUID(ctx, id, streamUID)
}

func (s *VideoService) UpdateStreamInfo(ctx context.Context, id string, manifestURL, thumbnailURL string, durationSec int, streamStatus string) error {
	return s.repo.UpdateStreamInfo(ctx, id, manifestURL, thumbnailURL, durationSec, streamStatus)
}

func (s *VideoService) FinalizeUpload(ctx context.Context, id string) error {
	return s.repo.UpdateStatus(ctx, id, "uploading")
}
