# Go Backend Guidelines & Architecture

## Role & Scope
You are an expert Go backend engineer. All work in `/server` must adhere to idiomatic Go 1.23+, Gin routing, GORM persistence with SQLite, clean architecture, and standard structured logging.

---

## Architectural Boundaries

1. **Layer Hierarchy (Strict One-Way Dependency):**
   `cmd/api/` ──► `internal/api/` ──► `internal/service/` ──► `internal/repository/` ──► SQLite (GORM)
   - **`internal/model/`**: Pure domain structs and DTOs. Annotated with `json:"..."` and `gorm:"..."` tags. This is the single source of truth for frontend types.
   - **`internal/repository/`**: GORM database operations, transaction handling, and query filters. Returns domain models or domain errors.
   - **`internal/service/`**: Core business logic, LLM prompt orchestration, job matching scoring, and scraper execution.
   - **`internal/api/`**: Gin HTTP handlers, route groups, middleware, request validation, and HTTP response serialization. No direct GORM or business logic here.
   - **`internal/ai/`**: LLM API clients, structured JSON output parsers, and prompt builders.

---

## SQLite & GORM Configuration Rules

- Use `glebarez/sqlite`.
- Enable SQLite WAL (Write-Ahead Logging) mode and busy timeout on startup for concurrent reads/writes:
  ```go
  db, err := gorm.Open(sqlite.Open("data/app.db?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"), &gorm.Config{})
  ```
- Store the SQLite database file in `server/data/app.db` and ensure `server/data/` is added to `.gitignore`.

---

## Type Sharing & DTO Rules (`tygo`)

- Every request/response struct in `internal/model` must have explicit `json:"camelCase"` tags.
- Run `tygo generate` whenever structs in `internal/model` are modified or added.
- The `tygo.yaml` config must target `../../web/src/types/api.gen.ts`.

---

## Coding Standards & Idioms

- **Gin Handlers:** Bind JSON with `c.ShouldBindJSON(&req)` and return standard error envelopes.
- **Context Handling:** Pass `c.Request.Context()` down to service and repository calls. Use timeouts for AI calls and web scraping.
- **GORM Best Practices:** Avoid global DB instances; inject `*gorm.DB` through repository constructors.
- **Error Handling:** Always wrap errors with context: `fmt.Errorf("service: fetching job %s: %w", id, err)`.
- **Logging:** Use standard library `log/slog` with structured attributes:
  ```go
  slog.Info("job parsed", "job_id", job.ID, "company", job.Company)
  ```