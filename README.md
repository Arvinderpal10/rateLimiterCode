```md
# Rate-Limited API – Testing Guide

> Project Folder Structure

```
```shell
ratelimiter/
├── cmd/
│   └── api/
│       └── main.go                 # Application entrypoint
├── internal/
│   ├── config/
│   │   └── config.go               # Configuration loading
│   ├── limiter/
│   │   ├── bucket.go               # Token bucket implementation
│   │   ├── store.go                # In‑memory store with cleanup
│   │   └── middleware.go           # Request ID & logging
│   └── handlers/
│       ├── request.go              # POST /request handler
│       └── stats.go                # GET /stats handler
├── pkg/
│   └── response/
│       └── json.go                 # JSON error helpers
├── go.mod
└── go.sum
```

---
```md 
## Setup & Running the Server

### Prerequisites
- Go 
- Terminal (bash / zsh)

### Clone Repo :
Clone the repo: https://github.com/Arvinderpal10/rateLimiterCode
cd rateLimiterCode

### Start the Server

```bash
go run ./cmd/api
```

You should see:
```
server listening on port 8080
```

### Custom Port

```bash
PORT=9090 go run ./cmd/api
```

---
```md
## Testing with curl

> All examples assume the server is running at 
http://localhost:8080

---

### 1. Basic Request – Success
```bash
curl -X POST http://localhost:8080/request \
  -H "Content-Type: application/json" \
  -d '{"user_id":"Arvinder","payload":"Hello World"}'
```

Expected response:
```json
{"status":"ok"}
```

---

### 2. Exceed Rate Limit
Send 6 sequential requests – the first 5 succeed, the 6th fails with HTTP 429:

```bash
for i in {1..6}; do
  echo "Request $i:"
  curl -s -X POST http://localhost:8080/request \
    -H "Content-Type: application/json" \
    -d '{"user_id":"Arvinder","payload":"test"}'
  echo
done
```

Expected output (ok for first 5 and last response - error):
```json
{"error":"rate limit exceeded"}
```
---

### 3. View Current Token Statistics

```bash
curl http://localhost:8080/stats
```

Example response:
```json
{"users":{"Arvinder":0.294677736083334}}
```
- The number is the current token balance (max = 5).
- A value below `1.0` means the user cannot make another request yet.

---

### 4. Concurrent Requests (Parallel Test)

Fire 20 requests with 10 in parallel and count HTTP status codes:

```shell
seq 1 20 | xargs -n1 -P10 curl -s -o /dev/null -w "%{http_code}\n" \
  -X POST http://localhost:8080/request \
  -H "Content-Type: application/json" \
  -d '{"user_id":"bob","payload":"x"}'
```

Expected result:  
Exactly five `200` and fifteen `429` status codes.

---

### 5. Check `Retry-After` Header

After hitting the limit, inspect the response headers:

```bash
curl -i -X POST http://localhost:8080/request \
  -H "Content-Type: application/json" \
  -d '{"user_id":"Arvinder","payload":"test"}'
```

Look for:
```
HTTP/1.1 429 Too Many Requests
Retry-After: 60
{"error":"rate limit exceeded"}
```

---

### 6. Error Cases – Invalid Input

Missing `user_id` field:
```bash
curl -X POST http://localhost:8080/request \
  -H "Content-Type: application/json" \
  -d '{"payload":"no user"}'
```
Response (HTTP 400): `{"error":"user_id is required"}`

##### Malformed JSON:
```bash
curl -X POST http://localhost:8080/request \
  -H "Content-Type: application/json" \
  -d 'not json'
```
Response (HTTP 400): `{"error":"invalid JSON"}`

Wrong HTTP method on `/stats`:
```bash
curl -X POST http://localhost:8080/stats
```
Response (HTTP 405): `Method Not Allowed`

---

### 7. Different Users Have Independent Limits

```bash
# Exhaust limit for Arvinder
for i in {1..5}; do
  curl -s -X POST http://localhost:8080/request \
    -H "Content-Type: application/json" \
    -d '{"user_id":"Arvinder","payload":"x"}' > /dev/null
done

# Amit should still be allowed
curl -s -X POST http://localhost:8080/request \
  -H "Content-Type: application/json" \
  -d '{"user_id":"Amit","payload":"y"}'
```

The last command returns `{"status":"ok"}`.

---

## Cleanup

Stop the server with `Ctrl+C`. The in‑memory data is lost on shutdown.

---

## Summary of Endpoints

| Method | Endpoint   | Description                           |
|--------|------------|---------------------------------------|
| POST   | `/request` | Process a payload (rate‑limited)      |
| GET    | `/stats`   | Return current token counts per user  |

## Design Decisions

| Decision                          | Rationale                                                                                           |
| --------------------------------- | --------------------------------------------------------------------------------------------------- |
| Token bucket algorithm            | Provides smooth, continuous rate limiting without fixed window boundary bursts                      |
| Per user buckets                  | Isolates users so one heavy user cannot starve others                                               |
| In memory storage                 | Simple and fast and meets the assignment requirement suitable for single instance deployments       |
| Mutex per bucket                  | Guarantees accurate token consumption under concurrent requests for the same user                   |
| RWMutex for user map              | Allows concurrent reads while safely handling new user creation                                     |
| Background cleanup goroutine      | Prevents unbounded memory growth by removing inactive users                                         |
| Graceful shutdown                 | Ensures in flight requests complete and background routines terminate cleanly                       |
| Request ID middleware             | Improves observability each request can be traced through logs                                      |
| Typed context keys                | Avoids key collisions with other middleware packages                                                |
| Separate lastAccess field         | Cleanup uses actual last request time not refill time ensuring inactive users are evicted correctly |
| Cryptographically random IDs      | Prevents predictable request IDs good for security and tracing                                      |
| Production ready folder structure | Follows Go community standards cmd internal pkg makes the codebase maintainable and testable        |

---

## What would I improve 

| Area           |     Improvement                                                                            |
| -------------- | ------------------------------------------------------------------------------------------ |
| Scalability    | Replace in memory store with Redis for distributed rate limiting                           |
| Tiered Limits  | Support per user limits for example free equals 5 per minute premium equals 100 per minute |
| Observability  | Add Prometheus metrics and Grafana health and ready endpoints                                          |             |
| Resilience     | Add retry mechanism and deadletter queue                                         |
| Testing        | Comprehensive unit and integration tests with race detector                                |                        |
| Deployment     | Dockerfile and Kubernetes manifests                                                        |
---
