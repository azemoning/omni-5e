# Contributing to omni-5e

Thanks for your interest. Here's how to get started.

## Setup

1. Fork and clone the repo
2. Start Postgres: `docker compose -f deploy/docker-compose.yml up -d postgres`
3. Run migrations: `go run ./cmd/omni-5e migrate up`
4. Import data: `go run ./cmd/omni-5e import --source data/raw/5.2.1 --version 5.2.1`
5. Run the server: `go run ./cmd/omni-5e serve`

You'll need Go 1.25+ and Docker installed.

## Running tests

```bash
go test ./... -race
```

Integration tests use testcontainers-go to spin up a real Postgres instance. They're slower but catch things unit tests miss.

## Code style

- Follow existing patterns in the codebase
- Don't add comments unless something is genuinely unclear
- Run `golangci-lint run` before committing
- Run `govulncheck ./...` if you're adding or updating dependencies

## Project structure

The architecture is hexagonal (ports-and-adapters). Dependencies point inward:

- `internal/domain` -- Plain structs, no external deps
- `internal/repository` -- Interfaces only
- `internal/store/postgres` -- Postgres implementations of those interfaces
- `internal/service` -- Business logic, calls repository interfaces
- `internal/transport/http` -- Fiber handlers, maps domain to DTOs
- `internal/ingest` -- SRD parsers per content version
- `internal/cli` -- Cobra commands, calls service layer

When adding a new feature, start from the domain and work outward. If you need a new entity field, add it to the domain struct first, then the repository interface, then the store implementation, then the service method, then the HTTP handler and CLI command.

## Adding a new entity type

1. Add the struct to `internal/domain/domain.go`
2. Add the repository interface to `internal/repository/interfaces.go`
3. Add store methods to `internal/store/postgres/`
4. Add service methods to `internal/service/service.go`
5. Add migration SQL to `migrations/`
6. Add HTTP handler and DTO to `internal/transport/http/v1/`
7. Register the route in `internal/transport/http/v1/router.go`

## Adding a new SRD version

1. Create `internal/ingest/srd53x/` implementing the `ContentParser` interface
2. Register it with `ingest.Register()` in an `init()` function
3. Run `omni-5e import --source data/raw/5.3.0 --version 5.3.0`

No changes to domain, repository, service, or transport layers.

## Submitting changes

1. Create a branch from `main`
2. Make your changes
3. Run tests and linter
4. Push and open a PR
5. Describe what you changed and why

Keep PRs focused. One feature or fix per PR is easier to review than a kitchen-sink dump.

## Reporting bugs

Open an issue with:
- What you expected to happen
- What actually happened
- Steps to reproduce
- Go version and OS

If the bug is in the data (wrong spell stats, missing monster, etc.), include the entity name and what's wrong.

## Data sources

The SRD markdown comes from [downfallx/dnd-5e-srd-markdown](https://github.com/downfallx/dnd-5e-srd-markdown), licensed CC BY 4.0 by Wizards of the Coast. If you're reporting data issues, check the source markdown first to confirm whether the problem is in the source or in our parser.

## Code of conduct

Be respectful. Disagree with ideas, not people. Don't be a jerk.
