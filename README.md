# NexaCard API

NexaCard API is the backend service for the NexaCard ecosystem. It provides public APIs, user/auth APIs, order and payment workflows, and admin APIs.

## Secondary Development Notice

NexaCard is an independently maintained secondary-development branch based on the open-source Dujiao-Next project. NexaCard uses its own branding, release packages, update checks, and documentation under the `NexaCard` GitHub organization. See [NOTICE.md](./NOTICE.md) for details.

## Tech Stack

- Go
- Gin
- GORM
- SQLite / PostgreSQL

## What This Service Does

- Serves REST APIs for user, order, and payment flows
- Handles payment callbacks/webhooks
- Supports product, fulfillment, and configuration management

## Quick Start

```bash
go mod tidy
go run cmd/server/main.go
```

The default health check endpoint is:

- `GET /health`

## Repositories

- API: https://github.com/NexaCard/API
- User: https://github.com/NexaCard/user
- Admin: https://github.com/NexaCard/admin
