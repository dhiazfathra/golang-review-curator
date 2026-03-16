# End-to-End Testing Guide

This guide demonstrates how to test all API endpoints using `curl`. All examples assume the server is running on `http://localhost:8080`.

---

## Prerequisites

```bash
# Start the infrastructure (Postgres + Redis)
make infra-up

# Run database migrations
make migrate

# Build the application
make build

# Run the server
make server-run
```

**Note:** This application does not require authentication for API endpoints. All endpoints are publicly accessible.

---

## 1. Health & Infrastructure

### Metrics Endpoint

```bash
curl -i http://localhost:8080/metrics
```

**Expected:** `200 OK` with Prometheus metrics in text format.

### Selector Health Check

```bash
curl -i http://localhost:8080/api/v1/selectors/health
```

**Expected:** `200 OK` with health status.

---

## 2. Products Module

### Create Product

```bash
curl -i -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Samsung Galaxy S24 Ultra",
    "platform": "shopee",
    "product_url": "https://shopee.co.id/samsung-galaxy-s24-ultra",
    "product_id": "123456789"
  }'
```

**Expected:** `201 Created` with product object including `id`.

**Save the returned `id`** for subsequent requests (e.g., `PRODUCT_ID=...`).

### List Products

```bash
curl -i http://localhost:8080/api/v1/products
```

**Expected:** `200 OK` with array of product objects.

---

## 3. Reviews Module

### List Reviews

```bash
curl -i http://localhost:8080/api/v1/reviews
```

**Expected:** `200 OK` with array of review objects and pagination info.

**Query Parameters:**

| Parameter | Description | Example |
|-----------|-------------|---------|
| `platform` | Filter by platform | `shopee`, `tokopedia`, `blibli` |
| `product_id` | Filter by product ID | `12345` |
| `rating` | Filter by rating (1-5) | `5` |
| `language` | Filter by language code | `id`, `en` |
| `from` | Filter by start date (YYYY-MM-DD) | `2024-01-01` |
| `to` | Filter by end date (YYYY-MM-DD) | `2024-12-31` |
| `limit` | Pagination limit (default: 20) | `50` |
| `offset` | Pagination offset (default: 0) | `0` |

**Example: Filter by platform and rating**

```bash
curl -i "http://localhost:8080/api/v1/reviews?platform=shopee&rating=5&limit=10"
```

### Get Review Summary for Product

```bash
curl -i http://localhost:8080/api/v1/reviews/summary/123456789
```

**Expected:** `200 OK` with summary object containing:

```json
{
  "product_id": "123456789",
  "platform": "shopee",
  "total_count": 150,
  "avg_rating": 4.2,
  "count_by_star": {
    "1": 10,
    "2": 15,
    "3": 25,
    "4": 40,
    "5": 60
  },
  "avg_sentiment": 0.75
}
```

**Query Parameters:**

| Parameter | Description | Example |
|-----------|-------------|---------|
| `platform` | Filter by platform (optional) | `shopee` |

---

## 4. Crawl Jobs Module

### Create Crawl Job

```bash
curl -i -X POST http://localhost:8080/api/v1/crawl/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "platform": "shopee",
    "product_url": "https://shopee.co.id/samsung-galaxy-s24-ultra",
    "product_id": "123456789",
    "max_pages": 10
  }'
```

**Expected:** `201 Created` with job object including `id`, `status`, `enqueued_at`.

**Save the returned `id`** for subsequent requests (e.g., `JOB_ID=...`).

**Note:** Job status can be `pending`, `running`, `done`, or `failed`.

### List Crawl Jobs

```bash
curl -i http://localhost:8080/api/v1/crawl/jobs
```

**Expected:** `200 OK` with array of job objects and pagination info.

**Query Parameters:**

| Parameter | Description | Example |
|-----------|-------------|---------|
| `platform` | Filter by platform | `shopee`, `tokopedia`, `blibli` |
| `status` | Filter by status | `pending`, `running`, `done`, `failed` |
| `limit` | Pagination limit (default: 20) | `50` |
| `offset` | Pagination offset (default: 0) | `0` |

### Get Crawl Job by ID

```bash
curl -i http://localhost:8080/api/v1/crawl/jobs/JOB_ID
```

**Expected:** `200 OK` with job object, or `404 Not Found` if job doesn't exist.

### Retry Failed Crawl Job

```bash
curl -i -X POST http://localhost:8080/api/v1/crawl/jobs/JOB_ID/retry
```

**Expected:** `202 Accepted` with `{"status": "requeued"}`.

**Error:** `400 Bad Request` if job status is not `failed`.

---

## 5. Selectors Module

### List Selectors

```bash
curl -i http://localhost:8080/api/v1/selectors
```

**Expected:** `200 OK` with array of selector configuration objects.

**Response structure:**

```json
{
  "data": [
    {
      "platform": "shopee",
      "field": "review_text",
      "rules": [
        { "type": "css", "value": ".review-text" },
        { "type": "jsonpath", "value": "$.reviews[*].content" }
      ]
    }
  ]
}
```

### Update Selector

```bash
curl -i -X PUT http://localhost:8080/api/v1/selectors/SELECTOR_ID \
  -H "Content-Type: application/json" \
  -d '{
    "platform": "shopee",
    "field": "review_text",
    "rules": [
      { "type": "css", "value": ".new-selector" }
    ]
  }'
```

**Expected:** `202 Accepted` with `{"status": "queued for reload"}`.

**Note:** Selector ID format is typically `{platform}:{field}` (e.g., `shopee:review_text`).

---

## 6. Complete E2E Test Flow

This example demonstrates a full workflow: create product → create crawl job → check job status → list reviews → get review summary.

```bash
# 1. Create a product
PRODUCT_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Product",
    "platform": "shopee",
    "product_url": "https://shopee.co.id/test-product",
    "product_id": "TEST123"
  }')
PRODUCT_ID=$(echo $PRODUCT_RESPONSE | jq -r '.id')
echo "Created product: $PRODUCT_ID"

# 2. Create a crawl job for the product
JOB_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/crawl/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "platform": "shopee",
    "product_url": "https://shopee.co.id/test-product",
    "product_id": "TEST123",
    "max_pages": 5
  }')
JOB_ID=$(echo $JOB_RESPONSE | jq -r '.id')
echo "Created job: $JOB_ID"

# 3. Check job status
curl -s http://localhost:8080/api/v1/crawl/jobs/$JOB_ID | jq

# 4. List all jobs
curl -s http://localhost:8080/api/v1/crawl/jobs | jq

# 5. List reviews (may be empty initially)
curl -s "http://localhost:8080/api/v1/reviews?product_id=TEST123" | jq

# 6. Get review summary
curl -s http://localhost:8080/api/v1/reviews/summary/TEST123 | jq

# 7. List selectors
curl -s http://localhost:8080/api/v1/selectors | jq

# 8. Check selector health
curl -s http://localhost:8080/api/v1/selectors/health | jq
```

---

## Testing Platform Filters

```bash
# Test Shopee platform
curl -s "http://localhost:8080/api/v1/reviews?platform=shopee" | jq '.data | length'

# Test Tokopedia platform
curl -s "http://localhost:8080/api/v1/reviews?platform=tokopedia" | jq '.data | length'

# Test Blibli platform
curl -s "http://localhost:8080/api/v1/reviews?platform=blibli" | jq '.data | length'
```

---

## Testing Rating Filters

```bash
# Get all 5-star reviews
curl -s "http://localhost:8080/api/v1/reviews?rating=5" | jq

# Get reviews with rating >= 4 (by filtering multiple ratings)
curl -s "http://localhost:8080/api/v1/reviews?rating=4" | jq
```

---

## Testing Date Range Filters

```bash
# Get reviews from a specific date range
curl -s "http://localhost:8080/api/v1/reviews?from=2024-01-01&to=2024-12-31" | jq
```

---

## Testing Pagination

```bash
# First page
curl -s "http://localhost:8080/api/v1/reviews?limit=10&offset=0" | jq '.pagination'

# Second page
curl -s "http://localhost:8080/api/v1/reviews?limit=10&offset=10" | jq '.pagination'

# Third page
curl -s "http://localhost:8080/api/v1/reviews?limit=10&offset=20" | jq '.pagination'
```

---

## Error Cases to Test

### 404 Not Found (Invalid Job ID)

```bash
curl -i http://localhost:8080/api/v1/crawl/jobs/invalid-job-id
# Expected: 404 Not Found
```

### 400 Bad Request (Invalid Platform)

```bash
curl -i -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test",
    "platform": "invalid_platform",
    "product_url": "https://example.com",
    "product_id": "123"
  }'
# Expected: 400 Bad Request with validation errors
```

### 400 Bad Request (Retry Non-Failed Job)

```bash
# First create a job
JOB_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/crawl/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "platform": "shopee",
    "product_url": "https://shopee.co.id/test",
    "product_id": "TEST"
  }')
JOB_ID=$(echo $JOB_RESPONSE | jq -r '.id')

# Try to retry a non-failed job (status is "pending")
curl -i -X POST http://localhost:8080/api/v1/crawl/jobs/$JOB_ID/retry
# Expected: 400 Bad Request
```

### 400 Bad Request (Missing Required Fields)

```bash
curl -i -X POST http://localhost:8080/api/v1/crawl/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "platform": "shopee"
  }'
# Expected: 400 Bad Request with validation errors
```

---

## Tips

1. **Use `jq` for JSON parsing**: Install with `brew install jq` (macOS) or `apt install jq` (Linux)
2. **Pretty-print JSON**: Pipe responses through `jq` for readable output
3. **Check response headers**: Use `-i` flag to see HTTP status codes and headers
4. **Extract IDs**: Use `jq -r '.id'` to extract IDs from JSON responses for subsequent requests
5. **Monitor job status**: After creating crawl jobs, poll the job status endpoint to check progress
6. **Monitor logs**: Run `docker-compose logs -f` to see server logs during testing
7. **Reset state**: Use `make infra-down && make infra-up` to restart the infrastructure

---

## Automated Testing with Scripts

Create a test script `test-e2e.sh`:

```bash
#!/bin/bash
set -e

BASE_URL="http://localhost:8080"

echo "==> Testing metrics endpoint"
curl -sf $BASE_URL/metrics > /dev/null
echo "✓ Metrics endpoint"

echo "==> Testing selector health"
curl -sf $BASE_URL/api/v1/selectors/health > /dev/null
echo "✓ Selector health"

echo "==> Testing product creation"
PRODUCT_RESPONSE=$(curl -sf -X POST $BASE_URL/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Product",
    "platform": "shopee",
    "product_url": "https://shopee.co.id/test",
    "product_id": "TEST123"
  }')
PRODUCT_ID=$(echo $PRODUCT_RESPONSE | jq -r '.id')
echo "✓ Product created: $PRODUCT_ID"

echo "==> Testing product listing"
curl -sf $BASE_URL/api/v1/products > /dev/null
echo "✓ Product listing"

echo "==> Testing crawl job creation"
JOB_RESPONSE=$(curl -sf -X POST $BASE_URL/api/v1/crawl/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "platform": "shopee",
    "product_url": "https://shopee.co.id/test",
    "product_id": "TEST123",
    "max_pages": 5
  }')
JOB_ID=$(echo $JOB_RESPONSE | jq -r '.id')
echo "✓ Crawl job created: $JOB_ID"

echo "==> Testing crawl job listing"
curl -sf $BASE_URL/api/v1/crawl/jobs > /dev/null
echo "✓ Crawl job listing"

echo "==> Testing crawl job retrieval"
curl -sf $BASE_URL/api/v1/crawl/jobs/$JOB_ID > /dev/null
echo "✓ Crawl job retrieved"

echo "==> Testing review listing"
curl -sf $BASE_URL/api/v1/reviews > /dev/null
echo "✓ Review listing"

echo "==> Testing review summary"
curl -sf $BASE_URL/api/v1/reviews/summary/TEST123 > /dev/null
echo "✓ Review summary"

echo "==> Testing selectors listing"
curl -sf $BASE_URL/api/v1/selectors > /dev/null
echo "✓ Selectors listing"

echo ""
echo "All E2E tests passed!"
```

Run with:

```bash
chmod +x test-e2e.sh
./test-e2e.sh
```

---

## Supported Platforms

| Platform | Value in API |
|----------|--------------|
| Shopee | `shopee` |
| Tokopedia | `tokopedia` |
| Blibli | `blibli` |

---

## Job Status Values

| Status | Description |
|--------|-------------|
| `pending` | Job is queued and waiting to be processed |
| `running` | Job is currently executing |
| `done` | Job completed successfully |
| `failed` | Job failed (can be retried) |

---

## Next Steps

- **Integration tests**: Write Go tests using `net/http/httptest` for programmatic endpoint testing
- **Load testing**: Use `hey`, `wrk`, or `k6` to test performance under load
- **Monitoring**: Set up dashboards to track API metrics, error rates, and latency
- **Authentication**: Consider adding authentication if the API will be exposed publicly
