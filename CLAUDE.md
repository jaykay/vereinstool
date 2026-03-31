# CLAUDE.md – Vereins-Meeting-Tool

## Projekt-Überblick

PWA für Vereinsvorstände zur strukturierten Sitzungsvorbereitung und -dokumentation.
Upvoting priorisiert Themen, automatische Agenda-Sortierung, Beschluss-Dokumentation,
Aufgabenverfolgung.

---

## Tech Stack

| Schicht       | Tech                                      |
|---------------|-------------------------------------------|
| Backend       | Go 1.23+, stdlib `net/http`               |
| Router        | `github.com/go-chi/chi/v5`                |
| Datenbank     | SQLite via `github.com/mattn/go-sqlite3`  |
| DB-Queries    | `sqlc` (typsicher, aus SQL generiert)     |
| Migrationen   | `golang-migrate/migrate`                  |
| Auth          | Sessions (Cookie), bcrypt Passwort-Hash   |
| E-Mail        | SMTP via `github.com/wneessen/go-mail`    |
| Frontend      | SvelteKit + TypeScript                    |
| PWA           | `@vite-pwa/sveltekit`                     |
| Styling       | Tailwind CSS v4                           |
| Deployment    | Kamal 2 auf Hetzner (Docker)              |

Go-Binary bettet SvelteKit-Bundle via `embed.FS` ein – ein Docker-Image, kein separater Frontend-Server.

---

## Projektstruktur

```
/
├── CLAUDE.md
├── TODOS.md                       # Feature-Planung
├── Dockerfile
├── config/deploy.yml              # Kamal Deploy-Konfiguration
├── .kamal/secrets                 # gitignored
│
├── backend/
│   ├── main.go
│   ├── go.mod / go.sum
│   ├── config/config.go           # Env-Vars (godotenv)
│   ├── db/
│   │   ├── migrations/            # SQL up/down
│   │   ├── queries/               # SQL für sqlc
│   │   ├── sqlc.yaml
│   │   └── generated/             # sqlc-Output – NICHT editieren
│   ├── handler/                   # HTTP-Handler + Middleware
│   ├── service/                   # Business-Logik (auth, mailer, agenda)
│   └── static/                    # embed: gebautes SvelteKit
│
└── frontend/
    ├── package.json
    ├── svelte.config.js
    ├── vite.config.ts
    ├── src/
    │   ├── lib/api.ts             # fetch-Wrapper
    │   ├── lib/stores.ts          # Svelte Stores
    │   └── routes/                # SvelteKit Pages
    └── static/                    # PWA-Manifest, Icons
```

---

## Datenbankschema

Siehe `backend/db/migrations/001_init.up.sql` (Source of Truth).

Tabellen: `users`, `sessions`, `password_reset_tokens`, `meetings`, `meeting_attendees`,
`topics`, `votes`, `decisions`, `tasks`.

---

## API-Routen

```
POST   /api/auth/login
POST   /api/auth/logout
POST   /api/auth/forgot-password
POST   /api/auth/reset-password

GET    /api/users                    (admin)
POST   /api/users/invite             (admin)
PATCH  /api/users/:id

GET    /api/meetings
POST   /api/meetings
GET    /api/meetings/:id
PATCH  /api/meetings/:id
POST   /api/meetings/:id/start
POST   /api/meetings/:id/close
GET    /api/meetings/:id/topics

POST   /api/topics
PATCH  /api/topics/:id
DELETE /api/topics/:id
POST   /api/topics/:id/vote
DELETE /api/topics/:id/vote

POST   /api/decisions
GET    /api/decisions
PATCH  /api/decisions/:id

POST   /api/tasks
GET    /api/tasks
PATCH  /api/tasks/:id

GET    /api/health
```

---

## Konventionen & Entscheidungen

- **Kein ORM** – sqlc generiert typsicheres Go aus SQL. Queries in `db/queries/*.sql`.
- **vote_count denormalisiert** – bei Vote-Insert/Delete direkt updaten, kein COUNT(*).
- **Zeiten UTC** in DB, Frontend zeigt lokal via `Intl`.
- **SvelteKit SPA-Modus** (`adapter-static`) – Go serviert Bundle, kein SSR.
- **Kein WebSocket** in Phase 1 – Polling reicht, EventSource später.
- **Fehlerformat**: `{"error": "message"}` + passender HTTP-Status.
- **Session-Cleanup**: tägliche Goroutine löscht abgelaufene Sessions.
- **bcrypt Cost 12**, Session-TTL 30 Tage, Cookie `HttpOnly; Secure; SameSite=Strict`.
- **Seed-Admin**: `SEED_ADMIN_EMAIL` + `SEED_ADMIN_PASSWORD` beim ersten Start.

---

## Lokale Entwicklung

### Voraussetzungen

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Backend

```bash
cd backend
cp ../.env.example .env
go run main.go               # :8080
```

### Frontend

```bash
cd frontend
pnpm install
pnpm dev                     # :5173, proxied /api → :8080
```

### Migrationen & sqlc

```bash
migrate create -ext sql -dir backend/db/migrations -seq beschreibung
migrate -path backend/db/migrations -database "sqlite3://./dev.db" up
cd backend && sqlc generate
```

### E-Mail lokal testen

```bash
docker run -p 1025:1025 -p 8025:8025 mailhog/mailhog
# UI: http://localhost:8025
```

### Env-Vars

Siehe `.env.example` im Projekt-Root.
