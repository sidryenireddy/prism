# Prism

Analytics and dashboard platform built on ontology objects. Users create analyses by adding cards to a canvas, connecting them together to form an analysis graph. Cards take inputs and produce outputs of specific types (object sets, numbers, time series, charts, tables).

## Architecture

- **api/** - Go backend: analysis/card/dashboard CRUD, card execution engine, object set operations, aggregations
- **frontend/** - Next.js + TypeScript: canvas-based analysis editor, card configuration, live results, dashboard builder
- **spec/** - OpenAPI specification
- **docker-compose.yml** - Full stack: API, frontend, PostgreSQL

## Card Types

- **Object Set**: filter, search around, set math (union/intersection/difference)
- **Visualization**: bar chart, line chart, pie chart, scatter plot, heat grid, map
- **Table**: object table, pivot table, transform table (group by/join/filter)
- **Numeric**: count, sum, average, min, max
- **Time Series**: time series chart, rolling aggregate, formula plot
- **Parameter**: object selection, date range, numeric, string, boolean
- **Action**: action button (triggers ontology actions)

## Quick Start

```bash
# Start all services
docker-compose up --build

# Or run individually
make api       # Go API on :8080
make frontend  # Next.js on :3000
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ONTOLOGY_URL` | `https://ontology.rebelinc.ai` | Ontology API base URL |
| `DATABASE_URL` | `postgres://prism:prism@localhost:5432/prism?sslmode=disable` | PostgreSQL connection |
| `PORT` | `8080` | API server port |

## Development

```bash
make dev        # Run API + frontend in development mode
make migrate    # Run database migrations
make test       # Run all tests
make lint       # Lint all code
make generate   # Generate code from OpenAPI spec
```
