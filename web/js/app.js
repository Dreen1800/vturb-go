const API_URL = '';

console.log('🚀 App.js carregado! Versão 2.0');
console.log('📍 URL atual:', window.location.href);

// Capturar erros globais
window.onerror = function(msg, url, line) {
    console.error('❌ Erro global:', msg, 'em', url, 'linha', line);
    return false;
};

class VTurbApp {
    constructor() {
        console.log('🏗️ Criando VTurbApp...');
        this.currentPage = 'dashboard';
        this.videos = [];
        this.selectedFile = null;
        this.currentVideo = null;
        this.uploadInProgress = false;
        
        // Sempre aguardar DOM estar pronto
        if (document.readyState === 'loading') {
            console.log('⏳ Aguardando DOM...');
            document.addEventListener('DOMContentLoaded', () => this.init());
        } else {
            console.log('✅ DOM já pronto');
            this.init();
        }
    }

    init() {
        console.log('🔧 Inicializando app...');
        try {
            this.setupNavigation();
            this.setupUpload();
            this.loadVideos();
            
            // Auto-refresh a cada 10 segundos
            setInterval(() => this.loadVideos(), 10000);
            console.log('✅ App inicializado com sucesso!');
        } catch (error) {
            console.error('❌ Erro fatal na inicialização:', error);
            alert('Erro ao inicializar app. Veja o console (F12).');
        }
    }

    // Navigation
    setupNavigation() {
        console.log('🔗 Configurando navegação...');
        const navLinks = document.querySelectorAll('.nav-links li');
        console.log('Links encontrados:', navLinks.length);
        
        navLinks.forEach(link => {
            link.addEventListener('click', (e) => {
                console.log('👆 Clique em:', e.currentTarget.dataset.page);
                const page = e.currentTarget.dataset.page;
                this.navigate(page);
            });
        });
    }

    navigate(page) {
        this.currentPage = page;
        
        // Update nav
        document.querySelectorAll('.nav-links li').forEach(li => {
            li.classList.toggle('active', li.dataset.page === page);
        });
        
        // Update pages
        document.querySelectorAll('.page').forEach(p => {
            p.classList.toggle('active', p.id === page);
        });
        
        // Update title
        const titles = {
            dashboard: 'Dashboard',
            gallery: 'Galeria de Vídeos',
            upload: 'Upload de Vídeo'
        };
        document.getElementById('page-title').textContent = titles[page];
        
        // Refresh data
        if (page === 'dashboard' || page === 'gallery') {
            this.loadVideos();
        }
    }

    // API Calls
    async api(endpoint, options = {}) {
        try {
            const response = await fetch(`${API_URL}/api${endpoint}`, {
                headers: {
                    'Content-Type': 'application/json',
                },
                ...options
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}`);
            }
            
            return await response.json();
        } catch (error) {
            this.showToast(`Erro: ${error.message}`, 'error');
            throw error;
        }
    }

    async loadVideos() {
        console.log('📥 Carregando vídeos...');
        try {
            this.videos = await this.api('/videos');
            console.log('✅ Vídeos carregados:', this.videos.length);
            this.updateDashboard();
            this.renderGallery();
        } catch (error) {
            console.error('❌ Falha ao carregar vídeos:', error);
            // Mostrar estado vazio em vez de travar
            this.videos = [];
            this.updateDashboard();
            this.renderGallery();
        }
    }

    // Dashboard
    updateDashboard() {
        const total = this.videos.length;
        const ready = this.videos.filter(v => v.status === 'ready').length;
        const pending = this.videos.filter(v => v.status === 'pending' || v.status === 'uploading').length;
        
        document.getElementById('total-videos').textContent = total;
        document.getElementById('ready-videos').textContent = ready;
        document.getElementById('pending-videos').textContent = pending;
        
        // Recent videos (last 6)
        const recent = this.videos.slice(0, 6);
        this.renderVideoGrid('recent-videos-list', recent);
    }

    // Gallery
    renderGallery() {
        this.renderVideoGrid('gallery-grid', this.videos);
    }

    renderVideoGrid(containerId, videos) {
        const container = document.getElementById(containerId);
        
        if (videos.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-video"></i>
                    <p>Nenhum vídeo encontrado</p>
                    <button class="btn btn-primary" onclick="app.navigate('upload')">
                        Fazer Upload
                    </button>
                </div>
            `;
            return;
        }
        
        container.innerHTML = videos.map(video => this.createVideoCard(video)).join('');
    }

    createVideoCard(video) {
        const statusColors = {
            ready: 'success',
            pending: 'warning',
            uploading: 'info',
            failed: 'error'
        };
        
        const statusLabels = {
            ready: 'Pronto',
            pending: 'Pendente',
            uploading: 'Enviando',
            failed: 'Falhou'
        };
        
        const thumbnail = video.thumbnail_url || '/web/img/placeholder.svg';
        const duration = video.duration_sec > 0 ? this.formatDuration(video.duration_sec) : '';
        
        return `
            <div class="video-card" data-id="${video.id}">
                <div class="video-thumbnail" onclick="app.openPlayer('${video.id}')">
                    <img src="${thumbnail}" alt="${video.title}" onerror="this.src='/web/img/placeholder.svg'">
                    ${duration ? `<span class="video-duration">${duration}</span>` : ''}
                    <div class="video-overlay">
                        <i class="fas fa-play"></i>
                    </div>
                </div>
                <div class="video-info">
                    <h4 onclick="app.openPlayer('${video.id}')">${video.title}</h4>
                    <div class="video-meta">
                        <span class="status-badge ${statusColors[video.status] || 'warning'}">
                            ${statusLabels[video.status] || video.status}
                        </span>
                        <span class="video-date">${this.formatDate(video.created_at)}</span>
                    </div>
                    ${video.status === 'ready' ? `
                        <div class="video-actions">
                            <button class="btn-icon" onclick="app.openPlayer('${video.id}')" title="Assistir">
                                <i class="fas fa-play"></i>
                            </button>
                            <button class="btn-icon" onclick="app.copyEmbedCode('${video.embed_token}')" title="Copiar Embed">
                                <i class="fas fa-code"></i>
                            </button>
                        </div>
                    ` : ''}
                </div>
            </div>
        `;
    }

    // Upload
    setupUpload() {
        console.log('📤 Configurando upload...');
        const dropZone = document.getElementById('drop-zone');
        const fileInput = document.getElementById('file-input');
        
        if (!dropZone) {
            console.error('❌ drop-zone não encontrado!');
            return;
        }
        if (!fileInput) {
            console.error('❌ file-input não encontrado!');
            return;
        }
        
        dropZone.addEventListener('dragover', (e) => {
            e.preventDefault();
            dropZone.classList.add('drag-over');
        });
        
        dropZone.addEventListener('dragleave', () => {
            dropZone.classList.remove('drag-over');
        });
        
        dropZone.addEventListener('drop', (e) => {
            e.preventDefault();
            dropZone.classList.remove('drag-over');
            
            const files = e.dataTransfer.files;
            if (files.length > 0) {
                this.selectFile(files[0]);
            }
        });
        
        fileInput.addEventListener('change', (e) => {
            if (e.target.files.length > 0) {
                this.selectFile(e.target.files[0]);
            }
        });
    }

    selectFile(file) {
        if (!file.type.startsWith('video/')) {
            this.showToast('Por favor, selecione um arquivo de vídeo', 'error');
            return;
        }
        
        this.selectedFile = file;
        
        // Show form
        document.getElementById('drop-zone').style.display = 'none';
        document.getElementById('upload-form').style.display = 'block';
        
        // Fill info
        document.getElementById('file-name').textContent = file.name;
        document.getElementById('file-size').textContent = this.formatFileSize(file.size);
        document.getElementById('video-title').value = file.name.replace(/\.[^/.]+$/, '');
    }

    resetUpload() {
        this.selectedFile = null;
        document.getElementById('drop-zone').style.display = 'flex';
        document.getElementById('upload-form').style.display = 'none';
        document.getElementById('file-input').value = '';
        document.querySelector('.progress-container').style.display = 'none';
    }

    async startUpload() {
        console.log('🚀 Iniciando upload...');
        
        if (!this.selectedFile) {
            console.error('❌ Nenhum arquivo selecionado');
            this.showToast('Selecione um arquivo primeiro', 'error');
            return;
        }
        
        if (this.uploadInProgress) {
            console.warn('⚠️ Upload já em andamento');
            return;
        }
        
        const title = document.getElementById('video-title').value || this.selectedFile.name;
        this.uploadInProgress = true;
        
        const btnUpload = document.getElementById('btn-upload');
        btnUpload.disabled = true;
        btnUpload.innerHTML = '<i class="fas fa-spinner fa-spin"></i> Enviando...';
        
        try {
            // 1. Create video
            console.log('📤 Etapa 1: Criando vídeo na API...');
            this.showToast('Criando vídeo...', 'info');
            const video = await this.api('/videos', {
                method: 'POST',
                body: JSON.stringify({ title })
            });
            console.log('✅ Vídeo criado:', video);
            
            // 2. Upload via proxy do backend (resolve CORS)
            console.log('📤 Etapa 2: Enviando arquivo via proxy...');
            this.showToast('Enviando vídeo (isso pode levar alguns minutos)...', 'info');
            
            const progressContainer = document.querySelector('.progress-container');
            const progressFill = document.getElementById('progress-fill');
            const progressText = document.getElementById('progress-text');
            progressContainer.style.display = 'flex';
            
            await this.uploadViaProxy(video.id, this.selectedFile, (percentage) => {
                progressFill.style.width = percentage + '%';
                progressText.textContent = percentage.toFixed(1) + '%';
                console.log(`📊 Progresso: ${percentage.toFixed(1)}%`);
            });
            
            // 3. Finalize
            console.log('📤 Etapa 3: Finalizando upload...');
            this.showToast('Finalizando...', 'info');
            await this.api(`/videos/${video.id}/finalize`, {
                method: 'PATCH'
            });
            
            console.log('✅ Upload completo!');
            this.showToast('Upload concluído com sucesso!', 'success');
            this.resetUpload();
            this.navigate('gallery');
            this.loadVideos();
            
        } catch (error) {
            console.error('❌ Upload failed:', error);
            this.showToast('Falha no upload: ' + error.message, 'error');
        } finally {
            this.uploadInProgress = false;
            btnUpload.disabled = false;
            btnUpload.innerHTML = '<i class="fas fa-upload"></i> Fazer Upload';
        }
    }

    uploadViaProxy(videoId, file, onProgress) {
        return new Promise((resolve, reject) => {
            const xhr = new XMLHttpRequest();
            
            xhr.upload.addEventListener('progress', (e) => {
                if (e.lengthComputable) {
                    const percentage = (e.loaded / e.total) * 100;
                    onProgress(percentage);
                }
            });
            
            xhr.addEventListener('load', () => {
                if (xhr.status >= 200 && xhr.status < 300) {
                    console.log('✅ Upload proxy completo!');
                    resolve(JSON.parse(xhr.responseText));
                } else {
                    console.error('❌ Erro no upload:', xhr.status, xhr.responseText);
                    reject(new Error(`Upload failed: ${xhr.status}`));
                }
            });
            
            xhr.addEventListener('error', () => {
                console.error('❌ Erro de rede no upload');
                reject(new Error('Network error'));
            });
            
            xhr.open('POST', `${API_URL}/api/videos/${videoId}/proxy-upload`);
            xhr.setRequestHeader('X-Filename', file.name);
            xhr.setRequestHeader('Content-Type', 'application/octet-stream');
            xhr.send(file);
        });
    }

    // Player
    openPlayer(videoId) {
        const video = this.videos.find(v => v.id === videoId);
        if (!video) return;
        
        this.currentVideo = video;
        
        document.getElementById('player-title').textContent = video.title;
        document.getElementById('player-status').textContent = video.status;
        document.getElementById('player-duration').textContent = 
            video.duration_sec > 0 ? this.formatDuration(video.duration_sec) : '';
        
        // Embed code
        const embedCode = `<script src="https://cdn.jsdelivr.net/npm/hls.js@latest"></script>
<div>
  <script src="${window.location.origin}/embed.js?token=${video.embed_token}"></script>
</div>`;
        document.getElementById('embed-code').value = embedCode;
        
        // Setup player
        const player = document.getElementById('video-player');
        player.poster = video.thumbnail_url || '';
        
        if (video.manifest_url && Hls.isSupported()) {
            const hls = new Hls();
            hls.loadSource(video.manifest_url);
            hls.attachMedia(player);
        } else if (video.manifest_url) {
            player.src = video.manifest_url;
        }
        
        document.getElementById('player-modal').style.display = 'flex';
    }

    closePlayer() {
        const player = document.getElementById('video-player');
        player.pause();
        player.src = '';
        document.getElementById('player-modal').style.display = 'none';
    }

    copyEmbed() {
        const textarea = document.getElementById('embed-code');
        textarea.select();
        document.execCommand('copy');
        this.showToast('Código embed copiado!', 'success');
    }

    copyEmbedCode(token) {
        const code = `<script src="https://cdn.jsdelivr.net/npm/hls.js@latest"></script>
<div>
  <script src="${window.location.origin}/embed.js?token=${token}"></script>
</div>`;
        
        navigator.clipboard.writeText(code).then(() => {
            this.showToast('Código embed copiado!', 'success');
        });
    }

    // Utilities
    formatDuration(seconds) {
        const mins = Math.floor(seconds / 60);
        const secs = Math.floor(seconds % 60);
        return `${mins}:${secs.toString().padStart(2, '0')}`;
    }

    formatFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleDateString('pt-BR', {
            day: '2-digit',
            month: '2-digit',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    }

    showToast(message, type = 'info') {
        const toast = document.getElementById('toast');
        toast.textContent = message;
        toast.className = `toast ${type} show`;
        
        setTimeout(() => {
            toast.classList.remove('show');
        }, 3000);
    }
}

// Initialize app
const app = new VTurbApp();
