#!/bin/bash

export CLOUDFLARE_API_KEY="your_cloudflare_api_key"
export CLOUDFLARE_EMAIL="your_email@example.com"
export CLOUDFLARE_ZONE_NAME="your_zone_name"

export ARGOCD_URL="your_argocd_url"
export ARGOCD_TOKEN="your_argocd_token"

export ARGOROLLOUTS_URL="your_argorollouts_url"
export ARGOROLLOUTS_TOKEN="your_argorollouts_token"

go run main.go # run tungsten
