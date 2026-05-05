# Configuração de Webhooks - Cloudflare Stream

## O Problema

Sua API roda em `localhost:8080`, mas a Cloudflare precisa enviar webhooks para uma URL **pública** na internet.

## Solução: ngrok (Tunnel)

O ngrok cria uma URL pública temporária que redireciona para seu localhost.

## Como usar

### 1. Inicie a API (se ainda não estiver rodando)

```bash
# Terminal 1
make run-api
```

### 2. Inicie o Tunnel

```bash
# Terminal 2
./scripts/tunnel.sh
```

Ou diretamente:
```bash
ngrok http 8080
```

### 3. Copie a URL pública

O ngrok vai mostrar algo como:
```
Forwarding  https://abc123def.ngrok.io -> http://localhost:8080
```

**Sua URL de webhook será:**
```
https://abc123def.ngrok.io/api/webhooks/cloudflare
```

### 4. Configure no Cloudflare Dashboard

1. Acesse: https://dash.cloudflare.com
2. Vá em **Stream** → **Webhooks**
3. Cole a URL: `https://abc123def.ngrok.io/api/webhooks/cloudflare`
4. Salve

### 5. Teste o Webhook

Faça upload de um vídeo pelo frontend. Quando a Cloudflare terminar de processar, ela enviará o webhook automaticamente!

## Verificando se funcionou

**No terminal da API**, você verá:
```
📩 Webhook received for stream abc123, status: ready
✅ Video abc123 updated to ready status
```

**No frontend**, o status mudará de "Enviando" → "Pronto" (badge verde).

## IMPORTANTE

- O tunnel do ngrok **expira** quando você fecha o terminal
- Cada vez que reiniciar, terá uma **nova URL**
- Precisa **atualizar a URL** no Cloudflare Dashboard
- Em produção, use um domínio fixo (ex: `https://seusite.com/api/webhooks/cloudflare`)

## Fluxo Completo

```
1. Usuário faz upload pelo Frontend
   ↓
2. Frontend envia arquivo para API (localhost:8080)
   ↓
3. API repassa arquivo para Cloudflare Stream
   ↓
4. Cloudflare processa o vídeo (30s - 5min)
   ↓
5. Cloudflare envia webhook → ngrok URL
   ↓
6. ngrok redireciona → localhost:8080/api/webhooks/cloudflare
   ↓
7. API atualiza status do vídeo para "ready"
   ↓
8. Frontend detecta (polling a cada 10s) e mostra "Pronto"
```

## Alternativas ao ngrok

Se preferir, pode usar:
- **Cloudflare Tunnel** (gratuito, mais estável)
- **Localtunnel** (`npx localtunnel --port 8080`)
- **Serveo** (`ssh -R 80:localhost:8080 serveo.net`)
