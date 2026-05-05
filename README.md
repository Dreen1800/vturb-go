# VTurb Go

MVP de plataforma de vídeo com upload via Cloudflare Stream e player embed.

## Stack

- **Go 1.26** + Chi Router
- **PostgreSQL 16** (pgx/v5)
- **Redis** (go-redis/v9)
- **Asynq** (jobs em background)
- **Cloudflare Stream** (upload TUS + HLS playback)

## Estrutura

```
vturb-go/
├── cmd/
│   ├── api/          # Servidor HTTP
│   └── worker/       # Processador de jobs
├── internal/
│   ├── config/       # Variáveis de ambiente
│   ├── handler/      # HTTP handlers
│   ├── service/      # Lógica de negócio
│   ├── repository/   # Acesso ao Postgres
│   ├── model/        # Structs
│   ├── platform/     # Client Cloudflare Stream
│   └── worker/       # Definição dos jobs Asynq
├── migrations/
├── docker-compose.yml
├── Dockerfile
└── Makefile
```

## Quick Start

### 1. Configurar variáveis de ambiente

```bash
cp .env.example .env
# Edite .env com suas credenciais Cloudflare
```

> **Nota sobre portas**: Se você já tem Postgres/Redis rodando localmente, o Docker usará portas alternativas:
> - Postgres: `5433` (ao invés de 5432)
> - Redis: `6380` (ao invés de 6379)

### 2. Subir infraestrutura (Postgres + Redis)

```bash
make docker-up
```

### 3. Rodar migrations

```bash
make migrate
```

### 4. Iniciar serviços (em terminais separados)

```bash
# Terminal 1 - API HTTP
make run-api

# Terminal 2 - Worker de jobs
make run-worker
```

Ou tudo via Docker (sem Go local):
```bash
docker-compose up -d api worker
```

## API Endpoints

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| POST | `/api/videos` | Cria vídeo |
| GET | `/api/videos` | Lista vídeos |
| GET | `/api/videos/:id` | Detalhes do vídeo |
| POST | `/api/videos/:id/upload` | Gera URL TUS para upload |
| PATCH | `/api/videos/:id/finalize` | Finaliza upload |
| POST | `/api/webhooks/cloudflare` | Webhook da Cloudflare |
| POST | `/api/embed/resolve` | Resolve token → manifest |
| GET | `/embed.js?token=xxx` | Script do player |
| GET | `/health` | Healthcheck |

## Fluxo de Upload

1. `POST /api/videos` → Cria vídeo, retorna ID + token
2. `POST /api/videos/:id/upload` → Gera endpoint TUS no Cloudflare
3. Cliente faz upload resumável via TUS direto para Cloudflare
4. Cloudflare processa e envia webhook
5. Webhook atualiza status + dispara job de sync
6. `GET /embed.js?token=xxx` → Carrega player com HLS

## Variáveis de Ambiente

```env
PORT=8080
ENV=development
# Use porta 5433 se Postgres estiver no Docker, 5432 se local
DATABASE_URL=postgres://vturb:vturb@localhost:5433/vturb
# Use porta 6380 se Redis estiver no Docker, 6379 se local
REDIS_URL=redis://localhost:6380/0
CLOUDFLARE_STREAM_ACCOUNT_ID=xxx
CLOUDFLARE_STREAM_API_TOKEN=xxx
EMBED_API_URL=http://localhost:8080
```

## Docker

```bash
# Tudo de uma vez
docker-compose up -d

# Logs
docker-compose logs -f api
docker-compose logs -f worker
```

## Embed no HTML

```html
<script src="https://cdn.jsdelivr.net/npm/hls.js@latest"></script>
<div data-vturb-token="vt_xxx">
  <script src="http://localhost:8080/embed.js?token=vt_xxx"></script>
</div>
```

## License

MIT
