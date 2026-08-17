# Project Overview & Architecture Guide

## Product Mission
An automated job search platform that scans job boards, parses postings, ranks matches against user profiles, tailors resumes to target descriptions, and generates custom cover letters using LLMs.

---

## Tech Stack
- **Frontend:** React 19, TypeScript, Vite, Tailwind CSS, Lucide React, shadcn/ui, TanStack Query, Zustand, Vitest, Ky.
- **Backend:** Go (1.23+), Gin, SQLite (modernc.org/sqlite), GORM, `tygo` (type generator), standard library structured logging (`slog`).
- **Repo Structure:** Monorepo with `/web` and `/server`.

---

## Directory Conventions

### Backend (`/server`)
Follow Standard Go Project Layout:
```
server/
├── cmd/
│   └── api/             # main.go entrypoint
├── internal/
│   ├── api/             # Gin HTTP handlers, routes, middlewares
│   ├── model/           # Core domain structs and DTOs (Source of Truth for Tygo)
│   ├── repository/      # GORM database queries and persistence
│   ├── service/         # Business logic (job scanning, AI tailoring, matching)
│   ├── ai/              # LLM client, prompt templates, output parsers
│   └── scraper/         # Job board ingestors / API connectors
├── config/              # Environment config loading
├── migrations/          # Database migration files
├── data/                # Local SQLite DB storage (e.g. app.db)
├── tygo.yaml            # Tygo configuration for TS type generation
└── go.mod
```

### Frontend (`/web`)
Feature-based folder structure:
```
web/
├── src/
│   ├── assets/
│   ├── components/ui/   # shadcn/ui primitives (Button, Modal, Input, Dialog)
│   ├── features/
│   │   ├── jobs/        # Job list, match cards, filters
│   │   ├── resumes/     # Resume builder, tailoring preview, diff viewer
│   │   └── letters/     # Cover letter editor and exports
│   ├── hooks/           # Shared reusable hooks
│   ├── lib/             # API client (Ky wrapper), query client
│   ├── types/
│   │   └── api.gen.ts   # Auto-generated DTOs from Tygo (DO NOT EDIT MANUALLY)
│   ├── App.tsx
│   └── main.tsx
├── package.json
├── tsconfig.json        # Configured with '@/*' path aliases
├── vite.config.ts
└── tailwind.config.js
```

---

## Type Sync Pipeline (`tygo`)

1. **Single Source of Truth:** All API DTOs and payload interfaces are defined as Go structs in `/server/internal/model/*.go`.
2. **Auto-Generation:** Run `make types` (or `cd server && tygo generate`) to compile Go structs into `/web/src/types/api.gen.ts`.
3. **Guardrail:** Never manually modify `/web/src/types/api.gen.ts`. Always update the Go model and re-generate.

---

## Coding Rules & Agent Guardrails

1. **Strict Type Contracts:** All frontend API calls must consume generated types from `@/types/api.gen`.
2. **Deterministic Architecture:** 
   - Backend logic belongs in `internal/service`, never directly inside Gin handlers.
   - Database operations stay strictly inside `internal/repository` using GORM.
3. **Frontend Rules:**
   - Use modular components under `features/<name>`.
   - Use `shadcn/ui` components for reusable primitive UI elements.
   - Use TanStack Query for all server-state mutations and queries.
   - Use `ky` for HTTP requests configured inside `src/lib/api.ts`.
4. **Error Handling & Logs:**
   - Go: Handle errors explicitly with clear wrap context (`fmt.Errorf("fetching job %s: %w", id, err)`).
   - Go: Use `log/slog` for structured logging.
5. **Security & Secrets:**
   - Never hardcode API keys, database credentials, or prompt tokens. Always load from `.env`.