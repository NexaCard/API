# NexaCard API

Backend service for NexaCard. It provides public storefront APIs, user authentication, order and payment workflows, admin APIs, upload serving, workers, and release update checks.

## Secondary Development Notice

NexaCard is an independently maintained secondary-development branch based on the open-source Dujiao-Next project. NexaCard uses its own branding, release packages, update checks, documentation, and repositories under the `NexaCard` GitHub organization. See [NOTICE.md](./NOTICE.md).

## Default Ports

- API: `5175`
- Health check: `GET /health`
- Admin dev proxy target: `http://localhost:5175`
- User dev proxy target: `http://localhost:5175`

## Requirements

- Go `1.26.3`
- Redis
- SQLite by default, PostgreSQL optional

## Local Development

```bash
go mod download
go run ./cmd/server -mode api
```

Use `config.yml.example` as the production template. The real `config.yml`, database files, uploads, and logs are intentionally ignored by git.

## Production Build

```bash
go build -trimpath -tags release -ldflags="-s -w -X github.com/NexaCard/API/internal/version.Version=v1.0.0" -o nexacard-api ./cmd/server
```

Recommended runtime:

```bash
./nexacard-api -mode all
```

`-mode all` starts API and worker services together. Use `-mode api` or `-mode worker` only when splitting services.

## Production Checklist

- Set `server.mode: release`
- Set `server.host: 127.0.0.1` when using a reverse proxy
- Keep `server.port: 5175`
- Replace `app.secret_key`, JWT secrets, and default admin credentials
- Set explicit `cors.allowed_origins` for storefront and admin domains
- Use a non-default `web.admin_path` if serving embedded fullstack builds

## Release Updates

The admin update check uses NexaCard releases only:

- https://github.com/NexaCard/API/releases

It does not point back to upstream Dujiao release channels.

## Related Repositories

- API: https://github.com/NexaCard/API
- User: https://github.com/NexaCard/user
- Admin: https://github.com/NexaCard/admin
- Docs: https://github.com/NexaCard/docs
