# omni-5e

A loosely-coupled, version-agnostic REST API and CLI for the D&D 5e System Reference Document.

omni-5e serves the complete D&D 5e SRD 5.2.1 as clean, structured, queryable JSON. It's designed as a read-oriented reference API — trivially consumable by VTT backends, character builders, mobile apps, or any tool that needs programmatic access to D&D 5e rules data.

## Features

- **Complete SRD coverage** — spells, monsters, classes, subclasses, species, backgrounds, feats, equipment, magic items, conditions, glossary terms, and rule sections
- **Two-axis versioning** — API contract version (`/api/v1`) is independent from content version (`?srd_version=5.2.1`), so new SRD releases don't break API consumers
- **Hexagonal architecture** — ports-and-adapters design: swap the database, search engine, or SRD source without touching HTTP handlers or business logic
- **Full-text search** — Postgres `tsvector` on spells and monsters, no separate search cluster needed
- **Pagination + filtering** — `limit`/`offset` pagination, class/level/school/CR/rarity filters
- **OpenAPI 3.1 spec** — machine-readable API contract at `/openapi.json`, Swagger UI at `/docs`
- **CLI companion** — `import`, `export`, `migrate`, `validate`, `serve` — the CLI and API share the same service layer
- **Docker-native** — `docker compose up` brings up a fully seeded API in one command

## Quick Start

### Docker (recommended)

```bash
git clone https://github.com/azemoning/omni-5e.git
cd omni-5e
docker compose -f deploy/docker-compose.yml up
```

The API is available at `http://localhost:8080`. Swagger UI at `http://localhost:8080/docs`.

### Local development

```bash
# Start Postgres
docker compose -f deploy/docker-compose.yml up -d postgres

# Run migrations
go run ./cmd/omni-5e migrate up

# Import SRD data
# First, clone the source markdown:
git clone --depth 1 https://github.com/downfallx/dnd-5e-srd-markdown.git /tmp/srd-markdown
cp /tmp/srd-markdown/*.md data/raw/5.2.1/

# Import
go run ./cmd/omni-5e import --source data/raw/5.2.1 --version 5.2.1

# Start the server
go run ./cmd/omni-5e serve
```

## API Endpoints

All resource endpoints support `?srd_version=X.Y.Z` or `X-SRD-Version` header to pin a content version. Omit to use the default (latest).

### Meta

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/meta/srd-versions` | List available SRD content versions |
| GET | `/api/v1/license` | CC BY 4.0 attribution |
| GET | `/healthz` | Liveness probe |
| GET | `/readyz` | Readiness probe |
| GET | `/openapi.json` | OpenAPI 3.1 spec |
| GET | `/docs` | Swagger UI |

### Resources

| Resource | List | Detail | Filters |
|---|---|---|---|
| Spells | `GET /api/v1/spells` | `GET /api/v1/spells/{slug}` | `class`, `level`, `school`, `ritual`, `concentration`, `q` |
| Monsters | `GET /api/v1/monsters` | `GET /api/v1/monsters/{slug}` | `cr_min`, `cr_max`, `type`, `size`, `category`, `q` |
| Classes | `GET /api/v1/classes` | `GET /api/v1/classes/{slug}` | — |
| Subclasses | — | `GET /api/v1/classes/{slug}/subclasses` | — |
| Level Table | — | `GET /api/v1/classes/{slug}/levels/{level}` | — |
| Species | `GET /api/v1/species` | `GET /api/v1/species/{slug}` | `size` |
| Backgrounds | `GET /api/v1/backgrounds` | `GET /api/v1/backgrounds/{slug}` | — |
| Feats | `GET /api/v1/feats` | `GET /api/v1/feats/{slug}` | `category` |
| Equipment | `GET /api/v1/equipment` | `GET /api/v1/equipment/{slug}` | `category` |
| Magic Items | `GET /api/v1/magic-items` | `GET /api/v1/magic-items/{slug}` | `rarity`, `attunement` |
| Conditions | `GET /api/v1/conditions` | `GET /api/v1/conditions/{slug}` | — |
| Glossary | `GET /api/v1/glossary` | `GET /api/v1/glossary/{slug}` | `category` |
| Rules | `GET /api/v1/rules` | `GET /api/v1/rules/{slug}` | `source_file` |

### Response format

All list endpoints return a consistent envelope:

```json
{
  "data": [ ... ],
  "meta": { "srd_version": "5.2.1", "total": 340, "limit": 50, "offset": 0 },
  "links": { "self": "...", "next": "...", "prev": null }
}
```

Single-resource endpoints return the object under `data`.

### Example

```bash
# List all evocation spells
curl "http://localhost:8080/api/v1/spells?school=evocation&level=3"

# Get Fireball
curl "http://localhost:8080/api/v1/spells/fireball"

# List monsters CR5-10
curl "http://localhost:8080/api/v1/monsters?cr_min=5&cr_max=10"

# Search spells
curl "http://localhost:8080/api/v1/spells?q=lightning"

# Pin SRD version
curl "http://localhost:8080/api/v1/spells?srd_version=5.2.1"
```

## CLI

```
omni-5e
├── serve                      # Start HTTP server
├── import                     # --source <dir> --version <semver>
├── export                     # --version <semver> --out <path>
├── validate                   # --version <semver>
├── migrate
│   ├── up                     # Apply pending migrations
│   ├── down                   # Roll back last migration
│   └── status                 # Show migration status
├── db
│   ├── seed                   # Import bundled SRD 5.2.1
│   └── reset --force          # Drop + recreate schema
├── version                    # Build info + supported SRD versions
└── license                    # CC BY 4.0 attribution
```

All commands accept `--config` and honor `OMNI5E_*` environment variables.

## Configuration

Configuration via environment variables (prefix `OMNI5E_`), YAML config file, or flags.

| Variable | Default | Description |
|---|---|---|
| `OMNI5E_SERVER_HOST` | `0.0.0.0` | HTTP listen address |
| `OMNI5E_SERVER_PORT` | `8080` | HTTP listen port |
| `OMNI5E_DATABASE_HOST` | `localhost` | Postgres host |
| `OMNI5E_DATABASE_PORT` | `5432` | Postgres port |
| `OMNI5E_DATABASE_USER` | `omni5e` | Postgres user |
| `OMNI5E_DATABASE_PASSWORD` | `omni5e` | Postgres password |
| `OMNI5E_DATABASE_NAME` | `omni5e` | Postgres database |
| `OMNI5E_DATABASE_SSLMODE` | `disable` | Postgres SSL mode |
| `OMNI5E_LOG_LEVEL` | `info` | Log level: trace/debug/info/warn/error |
| `OMNI5E_LOG_FORMAT` | `json` | Log format: json or console |

## Architecture

```
transport (HTTP/CLI) → service (business logic) → repository interfaces (ports)
                                                          ↑
                                              store adapters (Postgres, cache, search)

ingest adapters (per SRD version) → domain (canonical models) → repository interfaces
```

- **Domain** (`internal/domain`) — plain Go structs, zero external dependencies
- **Repository** (`internal/repository`) — interfaces only (ports)
- **Store** (`internal/store/postgres`) — pgx-based Postgres implementation
- **Service** (`internal/service`) — business logic shared by HTTP and CLI
- **Transport** (`internal/transport/http`) — Fiber v3 handlers with versioned DTOs
- **Ingest** (`internal/ingest/srd521`) — goldmark-based markdown parsers per SRD version
- **CLI** (`internal/cli`) — Cobra commands wrapping the service layer

### Adding a new SRD version

1. Create `internal/ingest/srd53x/` implementing the `ContentParser` interface
2. Register it in `init()`
3. Run `omni-5e import --source data/raw/5.3.0 --version 5.3.0`

No changes to domain, repository, service, or transport layers required.

## Data Quality

| Entity | Count | Completeness |
|---|---|---|
| Spells | 340 | 99.7% metadata (casting time, range, components) |
| Monsters | 235 | 100% HP/AC, 98% CR/XP |
| Classes | 12 | All core classes |
| Subclasses | — | Linked to parent classes |
| Species | 9 | All SRD species with size/speed |
| Backgrounds | 4 | All SRD backgrounds with ability scores/feats |
| Feats | 17 | With category/prerequisites |
| Equipment | 132 | Weapons, armor, tools, gear |
| Magic Items | 297 | With rarity/attunement |
| Conditions | 15 | All conditions |
| Glossary Terms | 139 | Rules definitions |
| Rule Sections | 239 | Character creation, gameplay, toolbox |

## Tech Stack

| Component | Choice | Rationale |
|---|---|---|
| Language | Go 1.25+ | Generics, fast compilation, single binary |
| HTTP | Fiber v3 | Low-overhead fasthttp, versioned routing |
| CLI | Cobra + Viper | De facto Go CLI standard |
| Database | PostgreSQL (JSONB + tsvector) | Relational integrity + schema flexibility |
| Query layer | pgx | Type-safe, low-level Postgres driver |
| Migrations | goose | Go-native, reviewable `.sql` files |
| Markdown | goldmark | CommonMark-compliant AST parser |
| Logging | zerolog | Zero-allocation JSON logging |
| Testing | testify + testcontainers-go | Assertions + real Postgres in CI |
| CI/CD | GitHub Actions + GoReleaser | Lint → test → build → vulncheck → release |
| Container | Docker (multi-stage → distroless) | Minimal attack surface |

## Development

```bash
# Run tests
go test ./... -race

# Run linter
golangci-lint run

# Build binary
go build -o omni-5e ./cmd/omni-5e

# Vulnerability scan
govulncheck ./...
```

## License

This project serves content from the D&D 5e System Reference Document 5.2.1 by Wizards of the Coast LLC, licensed under [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/legalcode).

The SRD content is available at https://www.dndbeyond.com/resources/1781-system-reference-document-5-2-1

The omni-5e source code is licensed under MIT.
