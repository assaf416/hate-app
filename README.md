# hate-app

A Hebrew (RTL) insurance management app built on the **HATE** stack —
**H**TMX, **A**lpine.js, **T**empl, **E**cho — with Bootstrap 5 for styling
and SQLite3 for storage.

Manage clients, policies, claims, and payments (each client has many
policies, claims, and payments), with per-record file attachments stored as
base64 blobs directly in the database.

## Features

- Full CRUD for **clients**, **policies**, **claims**, and **payments**
- **Attachments** (PDF / Word / Excel / images) stored as base64 in SQLite,
  viewable from client and policy pages, uploadable right after creating a
  policy
- **Global search** by client name or policy number (live HTMX dropdown)
- **Multi-tenant theming** — switch between company color schemes and
  light/dark mode at runtime (persisted in the browser)
- A mock **chat** page between client and representative
- Collapsible sidebar **submenus** for quick "view all" / "create new"
- A Hebrew **Cucumber (godog)** test suite covering every page and form
- A **seed script** for generating sample data

## Stack

| Layer      | Choice                                      |
|------------|----------------------------------------------|
| Backend    | Go + [Echo](https://echo.labstack.com/)       |
| Templates  | [Templ](https://templ.guide/)                 |
| Frontend   | [HTMX](https://htmx.org/) + [Alpine.js](https://alpinejs.dev/) |
| Styling    | Bootstrap 5 (RTL) + CSS variables for theming |
| Database   | SQLite3 (`mattn/go-sqlite3`)                  |
| Tests      | [godog](https://github.com/cucumber/godog) (Cucumber for Go), Hebrew Gherkin |

## Prerequisites

- Go 1.27+
- `CGO_ENABLED=1` (required by `mattn/go-sqlite3`) — a C toolchain (e.g. `gcc`) must be available
- [`templ`](https://templ.guide/) CLI, only if you plan to edit `.templ` files:
  ```bash
  go install github.com/a-h/templ/cmd/templ@latest
  ```

## Getting started

```bash
# Install dependencies
go mod download

# (Only needed if you changed any .templ file)
templ generate

# Run the app
CGO_ENABLED=1 go run .
```

The server listens on **http://localhost:8080**. A SQLite file `insurance.db`
is created automatically in the project root on first run.

### Seeding sample data

```bash
CGO_ENABLED=1 go run ./cmd/seed
```

This wipes and repopulates `insurance.db` with 100 clients, 1000 policies
(plus a sample of claims/payments), and 300 attachments. Pass `-db <path>` to
target a different SQLite file.

## Running the tests

### Cucumber (Hebrew Gherkin) suite

Feature files live under [`features/`](features), with Go step definitions
in [`features/step_definitions/`](features/step_definitions). Each scenario
spins up its own temporary SQLite database and in-process HTTP server, so
scenarios never leak state into each other.

Run the suite with the provided shell script:

```bash
./scripts/run-cucumber.sh
```

or directly with `go test`:

```bash
CGO_ENABLED=1 go test -run TestFeatures -v .
```

### Everything else

```bash
CGO_ENABLED=1 go build ./...
CGO_ENABLED=1 go vet ./...
```

## Project layout

```
db/                   SQLite connection + embedded schema
models/                Data access layer (one file per entity)
handlers/              Echo HTTP handlers
router/                Route registration (shared by main.go and tests)
templates/              Templ view components, one package per entity
static/                 CSS (app + theming) and JS (sidebar sync, theme switcher)
features/               Hebrew Cucumber feature files + step definitions
cmd/seed/               Sample-data generator
scripts/                Dev scripts (Cucumber runner, etc.)
```
