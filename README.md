# ReviewCurator

Production-grade product review aggregation for Shopee, Tokopedia, and Blibli (Indonesian e-commerce).

## Prerequisites

- Go 1.22+
- Docker + Docker Compose
- Chromium (installed automatically by go-rod on first run)
- BrightData or Oxylabs residential proxy credentials
- 2Captcha or Anti-Captcha API key

## Quickstart

```bash
# 1. Clone and configure
cp .env.example .env
# Edit .env with your proxy and captcha credentials

# 2. Start infrastructure (Postgres, Redis, OTel, Prometheus, Grafana)
make infra-up

# 3. Run migrations
make migrate

# 4. Start the HTTP API server
make server-run

# 5. Start the background worker (separate terminal)
make worker-run

# 6. Register a product for periodic crawling
curl -X POST http://localhost:8080/api/v1/products \
  -H 'Content-Type: application/json' \
  -d '{"name":"Test Shoe","platform":"shopee","product_url":"https://shopee.co.id/...","product_id":"12345"}'

# 7. Trigger a manual crawl
curl -X POST http://localhost:8080/api/v1/crawl/jobs \
  -H 'Content-Type: application/json' \
  -d '{"platform":"shopee","product_url":"https://shopee.co.id/...","product_id":"12345"}'
```

## Selector Update Runbook

When a platform changes its review section UI:

1. Identify the new CSS/XPath selector in browser DevTools.
2. Update the row in `selector_configs`:
   ```sql
   UPDATE selector_configs
   SET rules = '[{"type":"css","value":"<new-selector>"},{"type":"css","value":"<fallback>"}]'
   WHERE platform = 'shopee' AND field = 'review_text';
   ```
3. Wait up to 60 seconds for `SelectorStore` hot-reload to pick up the change.
4. No deployment required.

## Monitoring

- Grafana: http://localhost:3001 (admin/admin)
- Prometheus: http://localhost:9090
- Asynq dashboard: `ASYNQ_URI=$REDIS_URL asynq dash`

## Adding a New Platform

Run `make gen-adapter PLATFORM=lazada` (M7+ generator) or follow the New Platform Adapter Checklist in the project wiki.

## Environment Variables

See `.env.example` for a fully-documented list of all required variables.
