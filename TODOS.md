# TODOS – Vereins-Meeting-Tool

Featureplanung und Implementierungsfortschritt.

---

## Phase 1: Grundgerüst

- [ ] **Projektstruktur anlegen**
  - [ ] Go-Modul initialisieren (`backend/`)
  - [ ] SvelteKit init (`frontend/`)
  - [ ] Tailwind CSS v4 einrichten
  - [ ] Vite Proxy `/api → :8080` konfigurieren
  - [ ] `.env.example` anlegen
  - [ ] `.gitignore` (DB-Dateien, `.env`, `node_modules`, `build/`, `generated/`)

- [ ] **Dockerfile + Docker Compose**
  - [ ] Multi-Stage Dockerfile (Frontend build → Go build → Alpine runtime)
  - [ ] `docker-compose.yml` für lokale Entwicklung (optional, MailHog)
  - [ ] Lokaler Build-Test

- [ ] **Datenbank**
  - [ ] Migration `001_init.up.sql` + `001_init.down.sql` schreiben
  - [ ] `sqlc.yaml` konfigurieren
  - [ ] SQL-Queries schreiben (`users.sql`, `meetings.sql`, `topics.sql`, `decisions.sql`, `tasks.sql`)
  - [ ] `sqlc generate` ausführen und prüfen

---

## Phase 2: Auth

- [ ] **Backend Auth**
  - [ ] `config/config.go` – Env-Vars laden (godotenv)
  - [ ] `service/auth.go` – bcrypt (Cost 12), Session-Erstellung (UUID, 30d TTL)
  - [ ] `handler/middleware.go` – `AuthRequired` Middleware, CORS
  - [ ] `handler/auth.go` – Login, Logout
  - [ ] Seed-Admin beim ersten Start (`SEED_ADMIN_EMAIL` / `SEED_ADMIN_PASSWORD`)
  - [ ] Session-Cleanup Goroutine (täglich abgelaufene Sessions löschen)

- [ ] **Passwort-Reset**
  - [ ] `POST /api/auth/forgot-password` – Token generieren (32 Byte hex, 1h gültig)
  - [ ] `POST /api/auth/reset-password` – Token validieren, Passwort setzen
  - [ ] Reset-E-Mail senden

- [ ] **User-Verwaltung (Admin)**
  - [ ] `GET /api/users` – alle User auflisten
  - [ ] `POST /api/users/invite` – User anlegen, temporäres PW per Mail senden
  - [ ] `PATCH /api/users/:id` – User bearbeiten (Name, Rolle, aktiv/inaktiv)

- [ ] **Frontend Auth**
  - [ ] Login-Seite (`/login`)
  - [ ] Auth-Guard / geschützte Routen
  - [ ] `lib/api.ts` – fetch-Wrapper mit Cookie-Auth
  - [ ] `lib/stores.ts` – User-Store
  - [ ] Passwort-vergessen Seite
  - [ ] Passwort-Reset Seite

---

## Phase 3: Meetings CRUD

- [ ] **Backend**
  - [ ] `GET /api/meetings` – alle Sitzungen (gefiltert nach Status)
  - [ ] `POST /api/meetings` – neue Sitzung erstellen
  - [ ] `GET /api/meetings/:id` – Sitzungsdetail
  - [ ] `PATCH /api/meetings/:id` – Sitzung bearbeiten (Titel, Datum, Ort)
  - [ ] `POST /api/meetings/:id/start` – Status → `active`
  - [ ] `POST /api/meetings/:id/close` – Status → `closed`
  - [ ] Teilnehmerverwaltung (`meeting_attendees`)

- [ ] **Frontend**
  - [ ] Dashboard (`/dashboard`) – Übersicht aktiver/geplanter Sitzungen
  - [ ] Neue Sitzung erstellen (`/meetings/new`)
  - [ ] Sitzungsdetail (`/meetings/[id]`) – Live-Agenda, Teilnehmer

---

## Phase 4: Topics + Voting

- [ ] **Backend**
  - [ ] `POST /api/topics` – Thema einreichen (Titel, Beschreibung, Kategorie, geschätzte Dauer)
  - [ ] `PATCH /api/topics/:id` – Thema bearbeiten
  - [ ] `DELETE /api/topics/:id` – Thema löschen
  - [ ] `GET /api/meetings/:id/topics` – Themen sortiert nach `vote_count DESC`
  - [ ] `POST /api/topics/:id/vote` – Upvote abgeben (1 pro User)
  - [ ] `DELETE /api/topics/:id/vote` – Upvote zurücknehmen
  - [ ] `service/agenda.go` – Sortierlogik (vote_count, ggf. position Override)
  - [ ] `vote_count` bei Vote-Insert/Delete denormalisiert updaten

- [ ] **Frontend**
  - [ ] Thema einreichen (`/topics/new`)
  - [ ] Themenliste in Sitzungsansicht mit Upvote-Buttons
  - [ ] Sortierte Agenda-Ansicht
  - [ ] Kategorie-Filter (Finanzen, Satzung, Veranstaltungen, Sonstiges)

---

## Phase 5: Decisions + Tasks

- [ ] **Decisions (Backend)**
  - [ ] `POST /api/decisions` – Beschluss erfassen (Text, Ja/Nein/Enthaltung)
  - [ ] `GET /api/decisions` – Beschluss-Register (filterbar nach Meeting, Datum)
  - [ ] `PATCH /api/decisions/:id` – Beschluss korrigieren

- [ ] **Tasks (Backend)**
  - [ ] `POST /api/tasks` – Aufgabe erstellen (Titel, Beschreibung, Zuständiger, Fälligkeit)
  - [ ] `GET /api/tasks` – Aufgaben filtern (assigned_to, status)
  - [ ] `PATCH /api/tasks/:id` – Status ändern (open → done/cancelled)

- [ ] **Frontend**
  - [ ] Beschlüsse während Sitzung erfassen (in Sitzungsansicht)
  - [ ] Beschluss-Register (`/decisions`) – durchsuchbar, filterbar
  - [ ] Aufgaben erstellen (in Sitzungsansicht)
  - [ ] Meine Aufgaben (`/tasks`) – offene Aufgaben mit Fälligkeiten

---

## Phase 6: E-Mail-Service

- [ ] **`service/mailer.go`** – SMTP-Client (go-mail oder net/smtp)
- [ ] E-Mail-Templates (HTML + Plaintext):
  - [ ] Einladung (Willkommen + Login-Link + temporäres PW)
  - [ ] Passwort-Reset (Reset-Link, 1h gültig)
  - [ ] Neues Thema eingereicht (Titel, Beschreibung, Link zum Upvoten)
  - [ ] Sitzung startet (finale Agenda sortiert nach Votes)
  - [ ] Sitzung abgeschlossen (Protokoll mit Beschlüssen + offenen Aufgaben)
  - [ ] Aufgabe zugewiesen (Aufgabe, Fälligkeit, Link)

---

## Phase 7: Frontend Polish + PWA

- [ ] **Layout & Navigation**
  - [ ] Responsive Layout mit Sidebar/Bottom-Nav
  - [ ] Breadcrumbs / Zurück-Navigation
  - [ ] Loading-States, Error-States
  - [ ] Toast-Benachrichtigungen

- [ ] **PWA**
  - [ ] `manifest.json` (Name, Icons, Theme-Color)
  - [ ] Service Worker via `@vite-pwa/sveltekit`
  - [ ] Offline-Hinweis
  - [ ] App-Icons generieren (192px, 512px)

---

## Phase 8: Deployment

- [ ] **Hetzner-Server** provisionieren (CX22)
- [ ] **Kamal 2** einrichten
  - [ ] `config/deploy.yml` mit Secrets
  - [ ] `.kamal/secrets` anlegen
  - [ ] Docker Registry (Docker Hub oder GHCR)
  - [ ] TLS via Kamal Proxy (automatisch)
- [ ] DNS: `vorstand.flugplatz-uelzen.de` → Server-IP
- [ ] Erster Deploy + Smoke-Test
- [ ] Backup-Strategie für SQLite-Datei

---

## Phase 9: Nice-to-have (später)

- [ ] **Protokoll-Export** – Sitzungsprotokoll als PDF/DOCX (`GET /api/meetings/:id/export`)
- [ ] **Live-Updates** – EventSource/SSE statt Polling
- [ ] **Wiederkehrende Themen** – `is_recurring` automatisch in nächste Sitzung übernehmen
- [ ] **Benachrichtigungen** – Push-Notifications via Service Worker
- [ ] **Audit-Log** – Wer hat wann was geändert
- [ ] **Dunkelmodus**
- [ ] **Dateianhänge** an Topics/Decisions
