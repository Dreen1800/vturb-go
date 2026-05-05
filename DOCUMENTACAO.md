# VTurb Go - Documentação Completa

## Visão Geral

VTurb Go é uma plataforma de upload e streaming de vídeos construída com Go, PostgreSQL, Redis e Cloudflare Stream. Permite upload de vídeos via drag-and-drop, processamento automático na nuvem e embed de vídeos em qualquer site.

**URL de Produção:** https://vturb-go-production.up.railway.app

---

## Tecnologias Utilizadas

- **Backend:** Go 1.26 + Chi Router
- **Banco de Dados:** PostgreSQL (Railway)
- **Fila de Processamento:** Redis + Asynq
- **Storage/Streaming:** Cloudflare Stream (TUS Protocol)
- **Frontend:** Vanilla JavaScript SPA
- **Deploy:** Railway (Docker)
- **Player:** HLS.js para streaming adaptativo

---

## Estrutura do Projeto

```
vturb-go/
├── cmd/
│   ├── api/           # Servidor HTTP principal
│   └── worker/        # Worker de processamento
├── internal/
│   ├── config/        # Configurações
│   ├── handler/       # HTTP handlers
│   ├── model/         # Modelos de dados
│   ├── platform/      # Clientes externos (Cloudflare)
│   ├── repository/    # Acesso a dados
│   ├── service/       # Lógica de negócio
│   └── worker/        # Jobs em background
├── migrations/        # Scripts SQL
├── web/               # Frontend SPA
│   ├── css/
│   ├── js/
│   └── index.html
├── Dockerfile
├── Procfile
└── docker-compose.yml
```

---

## Configuração Local

### Pré-requisitos
- Go 1.26+
- Docker (para PostgreSQL e Redis)
- Cloudflare Stream API Token

### 1. Clonar e Configurar

```bash
git clone https://github.com/Dreen1800/vturb-go.git
cd vturb-go
cp .env.example .env
```

### 2. Configurar Variáveis (.env)

```env
PORT=8080
ENV=development
DATABASE_URL=postgres://vturb:vturb@localhost:5433/vturb
REDIS_URL=redis://localhost:6380/0
CLOUDFLARE_STREAM_ACCOUNT_ID=sua_account_id
CLOUDFLARE_STREAM_API_TOKEN=seu_token
CLOUDFLARE_STREAM_WEBHOOK_SECRET=seu_secret
EMBED_API_URL=http://localhost:8080
```

### 3. Iniciar Bancos de Dados

```bash
docker-compose up -d
```

Isso inicia:
- PostgreSQL na porta 5433
- Redis na porta 6380

### 4. Executar Migrations

As migrations rodam automaticamente na primeira execução.

### 5. Iniciar Servidor

```bash
# API Server
go run ./cmd/api

# Worker (em outro terminal)
go run ./cmd/worker
```

Acesse: http://localhost:8080

---

## Deploy na Railway

### 1. Login e Configuração

```bash
# Login (abre navegador)
railway login

# Criar projeto
railway init --name vturb-go

# Linkar projeto
railway link
```

### 2. Configurar Bancos de Dados

No dashboard da Railway ou via CLI:
- Adicionar PostgreSQL
- Adicionar Redis

### 3. Configurar Variáveis

```bash
railway variables --set "CLOUDFLARE_STREAM_ACCOUNT_ID=xxx"
railway variables --set "CLOUDFLARE_STREAM_API_TOKEN=xxx"
railway variables --set "CLOUDFLARE_STREAM_WEBHOOK_SECRET=xxx"
railway variables --set "EMBED_API_URL=https://vturb-go-production.up.railway.app"
```

### 4. Deploy

```bash
railway up
```

---

## API Endpoints

### Videos
- `POST /api/videos` - Criar vídeo
- `GET /api/videos` - Listar vídeos
- `GET /api/videos/:id` - Obter vídeo
- `POST /api/videos/:id/upload` - Gerar URL de upload
- `POST /api/videos/:id/proxy-upload` - Upload via proxy
- `PATCH /api/videos/:id/finalize` - Finalizar upload

### Webhooks
- `POST /api/webhooks/cloudflare` - Webhook do Cloudflare Stream

### Embed
- `GET /embed.js?token=xxx` - Script de embed
- `POST /api/embed/resolve` - Resolver token para URL

### Health
- `GET /health` - Status da aplicação

---

## Fluxo de Upload

1. **Criar Vídeo:**
   ```bash
   curl -X POST /api/videos \
     -d '{"title":"Meu Video","filename":"video.mp4","size_bytes":1000000}'
   ```
   Retorna: `id`, `embed_token`

2. **Upload:**
   ```bash
   curl -X POST /api/videos/{id}/proxy-upload \
     -H "X-Filename: video.mp4" \
     --data-binary @video.mp4
   ```

3. **Processamento:**
   - Cloudflare processa o vídeo
   - Webhook notifica quando "ready"
   - Status atualiza automaticamente

4. **Obter URL:**
   ```bash
   curl -X POST /api/embed/resolve \
     -d '{"token":"vt_xxx"}'
   ```
   Retorna: `manifest_url`, `thumbnail_url`

---

## Embed em Sites

### Opção 1: Script Automático
```html
<script src="https://cdn.jsdelivr.net/npm/hls.js@latest"></script>
<div>
  <script src="https://vturb-go-production.up.railway.app/embed.js?token=SEU_TOKEN"></script>
</div>
```

### Opção 2: Player Manual
```html
<video controls>
  <source src="https://customer-c4cwl9vbm74g1nh9.cloudflarestream.com/STREAM_UID/manifest/video.m3u8" 
          type="application/vnd.apple.mpegurl">
</video>
<script src="https://cdn.jsdelivr.net/npm/hls.js@latest"></script>
```

---

## Testes Realizados

### Testes de Upload
- ✅ Criação de vídeo via API
- ✅ Upload proxy (retorna 204)
- ✅ Upload com metadata base64
- ✅ Processamento no Cloudflare
- ✅ Webhook de atualização de status

### Testes de Playback
- ✅ Geração de HLS manifest
- ✅ Thumbnail disponível
- ✅ Player com HLS.js
- ✅ Embed.js funcional
- ✅ CORS configurado

### Testes de Deploy
- ✅ Deploy na Railway
- ✅ PostgreSQL configurado
- ✅ Redis configurado
- ✅ Migrations automáticas
- ✅ Variáveis de ambiente
- ✅ Frontend servindo arquivos estáticos

---

## Variáveis de Ambiente

| Variável | Descrição | Obrigatório |
|----------|-----------|-------------|
| `PORT` | Porta do servidor | Sim |
| `DATABASE_URL` | URL do PostgreSQL | Sim |
| `REDIS_URL` | URL do Redis | Sim |
| `CLOUDFLARE_STREAM_ACCOUNT_ID` | ID da conta Cloudflare | Sim |
| `CLOUDFLARE_STREAM_API_TOKEN` | Token da API Cloudflare | Sim |
| `CLOUDFLARE_STREAM_WEBHOOK_SECRET` | Secret do webhook | Não |
| `EMBED_API_URL` | URL base para embeds | Sim |

---

## Troubleshooting

### Erro 404 no Cloudflare
Verifique se `CLOUDFLARE_STREAM_ACCOUNT_ID` está correto.

### Erro 500 no Upload
Verifique se `CLOUDFLARE_STREAM_API_TOKEN` tem permissões de upload.

### Banco de Dados não conecta
Verifique se `DATABASE_URL` está configurado na Railway.

### Frontend não carrega
Verifique se a pasta `web/` está copiada no Dockerfile.

---

## Links Úteis

- **Aplicação:** https://vturb-go-production.up.railway.app
- **GitHub:** https://github.com/Dreen1800/vturb-go
- **Railway Dashboard:** https://railway.com/project/111f4a56-9e77-4010-b964-f11a55e15092

---

## Autor

Desenvolvido como MVP de plataforma de vídeo VTurb.
