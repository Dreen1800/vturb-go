package platform

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type CloudflareClient struct {
	accountID string
	apiToken  string
	baseURL   string
	http      *http.Client
}

func NewCloudflareClient(accountID, apiToken string) *CloudflareClient {
	return &CloudflareClient{
		accountID: accountID,
		apiToken:  apiToken,
		baseURL:   fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/stream", accountID),
		http:      &http.Client{},
	}
}

type CreateUploadResponse struct {
	UploadURL  string `json:"upload_url"`
	StreamUID  string `json:"stream_uid"`
}

func (c *CloudflareClient) CreateTUSUpload(filename string, filesize int64) (*CreateUploadResponse, error) {
	req, err := http.NewRequest("POST", c.baseURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Tus-Resumable", "1.0.0")
	req.Header.Set("Upload-Length", fmt.Sprintf("%d", filesize))
	uploadMetadata := fmt.Sprintf("name %s,filetype %s", b64Encode(filename), b64Encode("video/mp4"))
	log.Printf("📤 Upload-Metadata: %s", uploadMetadata)
	req.Header.Set("Upload-Metadata", uploadMetadata)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cloudflare API error: %s - %s", resp.Status, string(body))
	}

	uploadURL := resp.Header.Get("Location")
	streamUID := resp.Header.Get("Stream-Media-Id")

	if uploadURL == "" {
		return nil, fmt.Errorf("no upload URL returned from Cloudflare")
	}

	return &CreateUploadResponse{
		UploadURL: uploadURL,
		StreamUID: streamUID,
	}, nil
}

type StreamVideoInfo struct {
	UID         string  `json:"uid"`
	Status      string  `json:"status"`
	PlaybackURL string  `json:"playback_url"`
	ThumbnailURL string `json:"thumbnail_url"`
	Duration    float64 `json:"duration"`
}

func (c *CloudflareClient) GetVideoInfo(streamUID string) (*StreamVideoInfo, error) {
	url := fmt.Sprintf("%s/%s", c.baseURL, streamUID)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cloudflare API error: %s - %s", resp.Status, string(body))
	}

	var result struct {
		Success bool `json:"success"`
		Result  struct {
			UID         string  `json:"uid"`
			Status      struct {
				State string `json:"state"`
			} `json:"status"`
			Playback struct {
				HLS string `json:"hls"`
			} `json:"playback"`
			Thumbnail string  `json:"thumbnail"`
			Duration  float64 `json:"duration"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &StreamVideoInfo{
		UID:          result.Result.UID,
		Status:       result.Result.Status.State,
		PlaybackURL:  result.Result.Playback.HLS,
		ThumbnailURL: result.Result.Thumbnail,
		Duration:     result.Result.Duration,
	}, nil
}

func b64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func (c *CloudflareClient) GetAPIToken() string {
	return c.apiToken
}
