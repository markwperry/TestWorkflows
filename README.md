# TestWorkflows

A simple Go API built with [Gin](https://github.com/gin-gonic/gin), used as a test harness for GitHub Actions workflow development.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/helloworld` | Returns `{"message": "Hello, World!"}` |
| GET | `/health` | Returns `{"status": "ok"}` |

## Development

### Prerequisites

- Go 1.25+

### Run locally

```bash
go run .
# Server starts on :8080 (override with PORT env var)
```

### Run tests

```bash
go test ./...
```

### Build

```bash
go build -o server .
```

### Docker

```bash
docker build -t testworkflows .
docker run -p 8080:8080 testworkflows
```

## Branching Strategy

This project uses a simplified branching model with `main` and `release` branches:

```
feature branches → main → release
```

- **PRs to `main`**: Run unit/integration tests and linting (fast feedback)
- **PRs to `release`**: Run full E2E tests, build, and deploy
- **Hotfixes**: Branch from `release`, PR with `fix(hotfix):` prefix, auto-backported to `main`

See `.github/workflows/` for the full CI/CD configuration.
