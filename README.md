# omni-5e

REST API and CLI for the D&D 5e System Reference Document. Serves the complete SRD 5.2.1 as structured, queryable JSON.

This is a read-oriented reference API. You can use it to feed data to a VTT, character builder, mobile app, or anything else that needs programmatic access to D&D 5e rules.

## What it covers

Every entity in SRD 5.2.1 is represented:

- Spells (340), with level, school, components, casting time, range, duration
- Monsters (235), with full stat blocks: AC, HP, CR, XP, ability scores, traits, actions
- Classes (12) with subclasses and level tables
- Species (9) with size and speed
- Backgrounds (4) with ability score options and granted feats
- Feats (17) with prerequisites and categories
- Equipment (132): weapons, armor, tools, gear
- Magic items (297) with rarity and attunement
- Conditions (15)
- Glossary terms (139)
- Rule sections (239): character creation, core mechanics, optional systems

## Quick start

### Docker

```bash
git clone https://github.com/azemoning/omni-5e.git
cd omni-5e
docker compose -f deploy/docker-compose.yml up
```

The API runs at `http://localhost:8080`. Swagger UI at `http://localhost:8080/docs`.

### Local

```bash
# Start Postgres
docker compose -f deploy/docker-compose.yml up -d postgres

# Run migrations
go run ./cmd/omni-5e migrate up

# Get the SRD source markdown
git clone --depth 1 https://github.com/downfallx/dnd-5e-srd-markdown.git /tmp/srd-markdown
cp /tmp/srd-markdown/*.md data/raw/5.2.1/

# Import
go run ./cmd/omni-5e import --source data/raw/5.2.1 --version 5.2.1

# Start the server
go run ./cmd/omni-5e serve
```

## API

All resource endpoints accept `?srd_version=X.Y.Z` or the `X-SRD-Version` header to pin a content version. If you omit it, the default version is used.

### Endpoints

**Meta**

| Method | Path | Description |
|---|---|---|
| GET | `/api/v1/meta/srd-versions` | Available SRD versions |
| GET | `/api/v1/license` | CC BY 4.0 attribution |
| GET | `/healthz` | Liveness check |
| GET | `/readyz` | Readiness check |
| GET | `/openapi.json` | OpenAPI 3.1 spec |
| GET | `/docs` | Swagger UI |

**Resources**

| Resource | List | Detail | Filters |
|---|---|---|---|
| Spells | `GET /api/v1/spells` | `GET /api/v1/spells/{slug}` | `class`, `level`, `school`, `ritual`, `concentration`, `q` |
| Monsters | `GET /api/v1/monsters` | `GET /api/v1/monsters/{slug}` | `cr_min`, `cr_max`, `type`, `size`, `category`, `q` |
| Classes | `GET /api/v1/classes` | `GET /api/v1/classes/{slug}` | |
| Subclasses | | `GET /api/v1/classes/{slug}/subclasses` | |
| Level table | | `GET /api/v1/classes/{slug}/levels/{level}` | |
| Species | `GET /api/v1/species` | `GET /api/v1/species/{slug}` | `size` |
| Backgrounds | `GET /api/v1/backgrounds` | `GET /api/v1/backgrounds/{slug}` | |
| Feats | `GET /api/v1/feats` | `GET /api/v1/feats/{slug}` | `category` |
| Equipment | `GET /api/v1/equipment` | `GET /api/v1/equipment/{slug}` | `category` |
| Magic items | `GET /api/v1/magic-items` | `GET /api/v1/magic-items/{slug}` | `rarity`, `attunement` |
| Conditions | `GET /api/v1/conditions` | `GET /api/v1/conditions/{slug}` | |
| Glossary | `GET /api/v1/glossary` | `GET /api/v1/glossary/{slug}` | `category` |
| Rules | `GET /api/v1/rules` | `GET /api/v1/rules/{slug}` | `source_file` |

Note: subclasses and level tables don't have their own list endpoints. They're nested under classes, accessed via `GET /api/v1/classes/{slug}/subclasses` and `GET /api/v1/classes/{slug}/levels/{level}`.

### Response format

List endpoints return:

```json
{
  "data": [ ... ],
  "meta": { "srd_version": "5.2.1", "total": 340, "limit": 50, "offset": 0 },
  "links": { "self": "...", "next": "...", "prev": null }
}
```

Single-resource endpoints put the object directly under `data`.

### Examples

```bash
# Evocation spells, level 3
curl "http://localhost:8080/api/v1/spells?school=evocation&level=3"

# Fireball
curl "http://localhost:8080/api/v1/spells/fireball"

# Monsters CR 5 to 10
curl "http://localhost:8080/api/v1/monsters?cr_min=5&cr_max=10"

# Search spells by keyword
curl "http://localhost:8080/api/v1/spells?q=lightning"

# Pin a specific SRD version
curl "http://localhost:8080/api/v1/spells?srd_version=5.2.1"
```

## CLI

```
omni-5e serve                  # Start HTTP server
omni-5e import                 # --source <dir> --version <semver>
omni-5e export                 # --version <semver> --out <path>
omni-5e validate               # --version <semver>
omni-5e migrate up             # Apply pending migrations
omni-5e migrate down           # Roll back last migration
omni-5e migrate status         # Show migration status
omni-5e db seed                # Import bundled SRD 5.2.1
omni-5e db reset --force       # Drop and recreate schema
omni-5e version                # Build info
omni-5e license                # CC BY 4.0 attribution
```

All commands accept `--config` and read `OMNI5E_*` environment variables.

## Configuration

| Variable | Default | Description |
|---|---|---|
| `OMNI5E_SERVER_HOST` | `0.0.0.0` | Listen address |
| `OMNI5E_SERVER_PORT` | `8080` | Listen port |
| `OMNI5E_DATABASE_HOST` | `localhost` | Postgres host |
| `OMNI5E_DATABASE_PORT` | `5432` | Postgres port |
| `OMNI5E_DATABASE_USER` | `omni5e` | Postgres user |
| `OMNI5E_DATABASE_PASSWORD` | `omni5e` | Postgres password |
| `OMNI5E_DATABASE_NAME` | `omni5e` | Postgres database |
| `OMNI5E_DATABASE_SSLMODE` | `disable` | Postgres SSL mode |
| `OMNI5E_LOG_LEVEL` | `info` | trace, debug, info, warn, error |
| `OMNI5E_LOG_FORMAT` | `json` | json or console |

You can also use a `config.yaml` file. See `config.example.yaml`.

## Architecture

The project follows a ports-and-adapters (hexagonal) layout. Dependencies point inward only:

```
transport (HTTP/CLI)
    |
    v
service (business logic)
    |
    v
repository interfaces (ports) <--- store adapters (Postgres)
    ^
    |
domain (canonical structs)

ingest adapters (per SRD version) ---> domain
```

Each layer only knows the one below it through interfaces. Swapping Postgres for something else means writing a new store adapter. Adding SRD 5.3 means writing a new ingest adapter. Neither change touches the HTTP handlers or service logic.

**Key directories:**

- `internal/domain` -- Plain Go structs. No external dependencies.
- `internal/repository` -- Interfaces only. This is the contract between service and store.
- `internal/store/postgres` -- pgx-based Postgres implementation.
- `internal/service` -- Business logic. Both HTTP handlers and CLI commands call into this.
- `internal/transport/http` -- Fiber v3 handlers with versioned DTOs.
- `internal/ingest/srd521` -- Goldmark-based markdown parsers, one file per entity type.
- `internal/cli` -- Cobra commands wrapping the service layer.
- `migrations` -- Goose SQL migration files.
- `data/raw/5.2.1/` -- Vendored SRD markdown source.

### Adding a new SRD version

1. Create `internal/ingest/srd53x/` implementing the `ContentParser` interface
2. Register it in `init()`
3. Run `omni-5e import --source data/raw/5.3.0 --version 5.3.0`

No changes to domain, repository, service, or transport layers required.

## Data quality

| Entity | Count | Notes |
|---|---|---|
| Spells | 340 | 99.7% have casting time, range, and components |
| Monsters | 235 | All have HP and AC. 98% have CR and XP |
| Classes | 12 | All core classes with subclasses and level tables |
| Species | 9 | All SRD species with size and speed |
| Backgrounds | 4 | All SRD backgrounds with ability scores and feat references |
| Feats | 17 | With category and prerequisites |
| Equipment | 132 | Weapons, armor, tools, gear |
| Magic items | 297 | With rarity and attunement |
| Conditions | 15 | All conditions |
| Glossary terms | 139 | Rules definitions |
| Rule sections | 239 | Character creation, gameplay, toolbox |

The 84 monsters with null traits are not a bug. They legitimately have no Traits section in the SRD (for example, Behir, Young Red Dragon, and Cloud Giant only have Actions).

## Tech stack

| Component | Choice | Why |
|---|---|---|
| Language | Go 1.25+ | Fast compilation, single binary, generics |
| HTTP | Fiber v3 | Low overhead, built on fasthttp |
| CLI | Cobra + Viper | Standard Go CLI pattern |
| Database | PostgreSQL | Relational integrity with JSONB flexibility |
| Migrations | goose | Go-native, plain SQL files |
| Markdown | goldmark | CommonMark-compliant AST parser |
| Logging | zerolog | Zero allocation JSON logging |
| Testing | testify + testcontainers-go | Assertions with real Postgres in CI |
| CI | GitHub Actions | Lint, test, build, vulnerability scan, release |

## Development

```bash
go test ./... -race           # Run tests
golangci-lint run             # Lint
go build -o omni-5e ./cmd/omni-5e   # Build binary
govulncheck ./...             # Vulnerability scan
```

## License

The SRD content is from the D&D 5e System Reference Document 5.2.1 by Wizards of the Coast LLC, licensed under [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/legalcode). Source: https://www.dndbeyond.com/resources/1781-system-reference-document-5-2-1

The omni-5e source code is MIT licensed.
