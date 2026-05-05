package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"vturb-go/internal/platform"
	"vturb-go/internal/service"
)

type ProxyUploadHandler struct {
	svc *service.VideoService
	cf  *platform.CloudflareClient
}

func NewProxyUploadHandler(svc *service.VideoService, cf *platform.CloudflareClient) *ProxyUploadHandler {
	return &ProxyUploadHandler{svc: svc, cf: cf}
}

func (h *ProxyUploadHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	log.Printf("📥 Recebendo upload proxy para vídeo %s", id)

	// Verificar vídeo existe
	video, err := h.svc.GetVideo(r.Context(), id)
	if err != nil {
		http.Error(w, "video not found", http.StatusNotFound)
		return
	}

	// Obter nome e tamanho do arquivo do header
	filename := r.Header.Get("X-Filename")
	if filename == "" {
		filename = "video.mp4"
	}

	fileSize := r.ContentLength
	if fileSize <= 0 {
		http.Error(w, "content length required", http.StatusBadRequest)
		return
	}

	log.Printf("📁 Arquivo: %s (%d bytes)", filename, fileSize)

	// 1. Criar URL de upload na Cloudflare
	uploadInfo, err := h.cf.CreateTUSUpload(filename, fileSize)
	if err != nil {
		log.Printf("❌ Erro ao criar upload: %v", err)
		http.Error(w, fmt.Sprintf("failed to create upload: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("🔗 URL TUS criada: %s", uploadInfo.UploadURL)

	// 2. Salvar stream_uid
	if err := h.svc.UpdateStreamUID(r.Context(), video.ID, uploadInfo.StreamUID); err != nil {
		log.Printf("⚠️ Erro ao salvar stream_uid: %v", err)
	}

	// 3. Ler o arquivo inteiro em memória e enviar para Cloudflare
	log.Printf("📤 Iniciando upload para Cloudflare...")
	log.Printf("📊 Tamanho do arquivo a enviar: %d bytes", fileSize)
	
	// Ler todo o body
	fileData, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Erro ao ler body: %v", err)
		http.Error(w, "failed to read request body", http.StatusInternalServerError)
		return
	}
	
	log.Printf("📊 Bytes lidos do body: %d", len(fileData))
	
	if len(fileData) == 0 {
		log.Printf("❌ Body está vazio!")
		http.Error(w, "request body is empty", http.StatusBadRequest)
		return
	}
	
	// Obter token da configuração
	apiToken := h.cf.GetAPIToken()
	
	// Criar request PATCH para a Cloudflare (TUS protocol) com os dados lidos
	log.Printf("🌐 URL do PATCH: %s", uploadInfo.UploadURL)
	log.Printf("📊 Tamanho do body a enviar: %d bytes", len(fileData))
	
	cfReq, err := http.NewRequest("PATCH", uploadInfo.UploadURL, bytes.NewReader(fileData))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copiar headers necessários + Autorização
	cfReq.Header.Set("Authorization", "Bearer "+apiToken)
	cfReq.Header.Set("Tus-Resumable", "1.0.0")
	cfReq.Header.Set("Upload-Offset", "0")
	cfReq.Header.Set("Content-Type", "application/offset+octet-stream")
	cfReq.ContentLength = int64(len(fileData))

	client := &http.Client{}
	resp, err := client.Do(cfReq)
	if err != nil {
		log.Printf("❌ Erro no upload: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Ler body da resposta para debug
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyStr := string(bodyBytes)
	
	log.Printf("📥 Resposta Cloudflare: %d - Body: %s", resp.StatusCode, bodyStr)

	// 4. Retornar resultado
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    resp.StatusCode >= 200 && resp.StatusCode < 300,
		"stream_uid": uploadInfo.StreamUID,
		"status":     "uploaded",
		"cf_status":  resp.StatusCode,
		"cf_body":    bodyStr,
	})
}
