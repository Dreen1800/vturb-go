package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"vturb-go/internal/config"
	"vturb-go/internal/service"
)

func ServeEmbedJS(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			token = r.URL.Query().Get("t")
		}

		script := fmt.Sprintf(`(function() {
			'use strict';
			
			var token = '%s';
			var apiUrl = '%s';
			
			if (!token) {
				console.error('VTurb: No token provided');
				return;
			}
			
			function init() {
				fetch(apiUrl + '/api/embed/resolve', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ token: token })
				})
				.then(function(res) { return res.json(); })
				.then(function(data) {
					if (data.status !== 'ready' || !data.manifest_url) {
						console.log('VTurb: Video not ready yet');
						return;
					}
					createPlayer(data.manifest_url, data.thumbnail_url);
				})
				.catch(function(err) {
					console.error('VTurb: Failed to resolve embed', err);
				});
			}
			
			function createPlayer(manifestUrl, thumbnailUrl) {
				// Find container - try currentScript first, fallback to parent div
				var container = null;
				if (document.currentScript && document.currentScript.parentElement) {
					container = document.currentScript.parentElement;
				} else {
					// Fallback: find script by src and get its parent
					var scripts = document.getElementsByTagName('script');
					for (var i = 0; i < scripts.length; i++) {
						if (scripts[i].src && scripts[i].src.indexOf('embed.js') !== -1) {
							container = scripts[i].parentElement;
							break;
						}
					}
				}
				
				if (!container) {
					console.error('VTurb: Could not find container element');
					return;
				}
				
				var video = document.createElement('video');
				video.controls = true;
				video.style.width = '100%%';
				video.style.maxWidth = '100%%';
				video.style.display = 'block';
				
				if (thumbnailUrl) {
					video.poster = thumbnailUrl;
				}
				
				// Check HLS support
				if (video.canPlayType('application/vnd.apple.mpegurl')) {
					video.src = manifestUrl;
				} else if (window.Hls) {
					var hls = new window.Hls();
					hls.loadSource(manifestUrl);
					hls.attachMedia(video);
				} else {
					console.error('VTurb: HLS.js not loaded and native HLS not supported');
					return;
				}
				
				container.appendChild(video);
				console.log('VTurb: Player created successfully');
			}
			
			if (document.readyState === 'loading') {
				document.addEventListener('DOMContentLoaded', init);
			} else {
				init();
			}
		})();`, token, cfg.Embed.APIURL)

		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write([]byte(script))
	}
}

type ResolveEmbedRequest struct {
	Token string `json:"token"`
}

type ResolveEmbedResponse struct {
	ManifestURL  string  `json:"manifest_url"`
	ThumbnailURL *string `json:"thumbnail_url,omitempty"`
	DurationSec  *int    `json:"duration_sec,omitempty"`
	Status       string  `json:"status"`
}

func ResolveEmbedHandler(svc *service.VideoService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ResolveEmbedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Token == "" {
			http.Error(w, "token is required", http.StatusBadRequest)
			return
		}

		video, err := svc.GetVideoByToken(r.Context(), req.Token)
		if err != nil {
			http.Error(w, "video not found", http.StatusNotFound)
			return
		}

		resp := ResolveEmbedResponse{
			Status: video.Status,
		}

		if video.ManifestURL != nil {
			resp.ManifestURL = *video.ManifestURL
		}
		if video.ThumbnailURL != nil {
			resp.ThumbnailURL = video.ThumbnailURL
		}
		if video.DurationSec != nil {
			resp.DurationSec = video.DurationSec
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
