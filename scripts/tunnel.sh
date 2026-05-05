#!/bin/bash

echo "🚀 VTurb Go - Tunnel para Cloudflare Webhooks"
echo ""
echo "Este script cria um tunnel público para receber webhooks da Cloudflare"
echo ""

# Verificar se API está rodando
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ API não está rodando em localhost:8080"
    echo "   Inicie primeiro: make run-api"
    exit 1
fi

echo "✅ API detectada em localhost:8080"
echo "🌐 Iniciando tunnel ngrok..."
echo ""
echo "Quando aparecer a URL pública, configure no Cloudflare Stream:"
echo "   Dashboard → Stream → Webhooks"
echo ""
echo "URL do webhook: https://SEU-ID.ngrok.io/api/webhooks/cloudflare"
echo ""

# Iniciar ngrok
ngrok http 8080
