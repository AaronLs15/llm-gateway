# llm-gateway

An HTTP service that sits between client apps and multiple LLM providers behind one
common interface. Routing, fallback, rate limiting, cost tracking, observability.

Built as a Go learning project — **standard library only**, no external dependencies.

## Status

Built incrementally, one milestone at a time. Current: **M1 — hello server**.

Roadmap:

| Milestone | What |
|-----------|------|
| M1 | HTTP server + `/healthz` + graceful shutdown |
| M2 | Domain types + `Provider` interface + mock provider |
| M3 | `POST /v1/chat` endpoint |
| M4 | Config from env |
| M5 | Middleware (request-id, logging, recover) |
| M6 | Multi-provider registry + routing |
| M7 | Fallback chain |
| M8 | Auth + per-client rate limiting |
| M9 | Real Anthropic / OpenAI providers |
| M10 | Observability + cost tracking |

## Run

```sh
go run ./cmd/gateway
# in another shell:
curl localhost:8080/healthz   # -> ok
```

Stop with Ctrl-C; the server drains in-flight requests before exiting.

## Layout

```
cmd/gateway/   entrypoint (wiring + startup)
internal/      private packages (config, server, chat, provider, routing, middleware)
```

Requires Go 1.22+ (uses method-aware `net/http` routing).
