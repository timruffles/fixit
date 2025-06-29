# Fixit

Community software for fixing things - a web application built with Go, PostgreSQL, and Tailwind CSS.

## Architecture

- **cmd/** - Application binaries/entrypoints
- **web/** - HTTP handlers, templates, and web-facing code
- **engine/** - Business logic, repositories, and database operations

The application uses:
- [Ent ORM](https://entgo.io/) with PostgreSQL
- [Authboss v3](https://github.com/volatiletech/authboss) for authentication
- [Tailwind CSS](https://tailwindcss.com/) for styling
- [Cobra](https://github.com/spf13/cobra) for CLI commands

## Prerequisites

Install the following dependencies:

- **Go 1.24+**: [Download Go](https://golang.org/dl/)
- **Docker & Docker Compose**: [Install Docker](https://docs.docker.com/get-docker/)
- **Air** (for hot reload): `go install github.com/cosmtrek/air@latest`
- **Flyctl** (for deployment): `brew install flyctl`

## Getting Started

1. **Clone and setup**:
   ```bash
   git clone <repository-url>
   cd fixit
   ```

2. **Start the database**:
   ```bash
   docker-compose up -d
   ```

3. **Generate Ent code**:
   ```bash
   go generate ./engine/ent
   ```

4. **Install dependencies**:
   ```bash
   go mod download
   ```

5. **Run with hot reload**:
   ```bash
   air
   ```
   
   Or run directly:
   ```bash
   go run cmd/server/main.go
   ```

## Development

### Database
- PostgreSQL runs in Docker via `docker-compose.yml`
- Primary keys use UUIDv7 (application-generated)
- Integration tests use separate `fixit_test` database

### Testing
```bash
# Run integration tests
go test ./web/integration -v

# Run all tests
go test ./...
```

### Code Generation
```bash
# Generate Ent schema code
go generate ./engine/ent
```

