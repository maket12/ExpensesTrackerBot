# Project Context: Family Expenses Tracker (AI Bot & Web Panel)
You are a Senior Fullstack Engineer and Backend Architect. Your task is to build a production-ready system according to the technical specification.

## 1. Domain & Business Rules
The system is a family expense tracker with two interfaces sharing the same backend and database:
1. **Telegram Bot with AI Agent**: Users manage their personal expenses via natural language ONLY (no buttons, no slash commands except initial `/start`).
   - **Strict Isolation**: A user in Telegram can ONLY create, view, update, or delete their own expenses (`user_id` must always be enforced from the Telegram session context, never trusted from LLM arguments).
2. **Web Panel (for Head of Family)**: Full CRUD interface for expenses across ALL family members with filters (`user_id`, date period `from`/`to`, `category`). No authentication required.

### Data Model (`Expense`)
- `id`: UUID (Primary Key)
- `user_id`: `int64` (Telegram User ID of the owner)
- `amount`: `numeric(12,2)`, must be `> 0`
- `category`: Enum strictly limited to: `еда`, `транспорт`, `жильё`, `развлечения`, `здоровье`, `покупки`, `другое`
- `spent_at`: `timestamptz` (when the expense occurred, parsed from natural language or selected in UI)
- `comment`: `text` (optional description, e.g., "кофе", "такси")
- `created_at`: `timestamptz`
- `updated_at`: `timestamptz`

## 2. Tech Stack & Libraries
- **Backend (Go 1.26.2)**:
  - Architecture: Monolith with two delivery adapters (HTTP REST API + Telegram Bot Long Polling) sharing the same UseCase layer.
  - Database: PostgreSQL 16 + `jackc/pgx/v5`.
  - Migrations: `golang-migrate`.
  - HTTP Router: `github.com/labstack/echo/v4.
  - Telegram SDK: `gopkg.in/telebot.v3` (or `go-telegram/bot`).
  - LLM Client: `sashabaranov/go-openai` configured via env vars (`LLM_BASE_URL`, `LLM_API_KEY`, `LLM_MODEL`) using OpenAI Tool/Function Calling.
- **Frontend**:
  - React + TypeScript + Vite + Tailwind CSS / Ant Design (communicates with backend via REST API).
- **Infrastructure**:
  - Everything must start with a single `docker compose up --build` command.

## 3. Monorepo & Hexagonal Directory Structure
- `/backend` — Go module root
  - `/cmd/app/main.go` — composition root, config initialization, DI, and graceful shutdown (`errgroup` / `signal.NotifyContext`).
  - `/internal`
    - `/domain` — core domain layer with zero external dependencies:
      - `/models` — Rich Domain Models (`Expense`), Value Objects (`Category`, `Amount`, etc.), and domain-level validation/invariants.
      - `/port` — repository and external service interfaces operating on domain models.
      - `/errs` (or `errors.go`) — domain-specific errors.
    - `/app` — application layer:
      - `/dto` — one file per use case (`<usecase_name>.go`) containing paired `{Name}Input` and `{Name}Output` structs.
      - `/usecases` — application business logic implementations and AI agent orchestration.
      - `/mapper` — pure mapping functions between `domain/models` and `app/dto`.
      - `/errs` — application-layer errors.
    - `/adapter`
      - `/in` (Driving / Primary Adapters)
        - `/http` — REST API adapter containing handlers/router, `/dto` (`{Name}Request` and `{Name}Response`), `/mapper` (converts HTTP DTOs <-> `app/dto`), and `/errs` (HTTP error mapping).
        - `/telegram` — Telegram bot adapter containing handlers, `/dto` (`{Name}Request` and `{Name}Response`), `/mapper` (converts Telegram updates <-> `app/dto`), and `/errs`.
      - `/out` (Driven / Secondary Adapters)
        - `/postgres` — PostgreSQL repository implementation, `/dto` (`{Name}Request` / `{Name}Response` or DB row models), `/mapper` (converts DB structs <-> `domain/models`), and `/errs`.
        - `/llm` — OpenAI SDK client wrapper, `/dto` (`{Name}Request` and `{Name}Response` for LLM/tool payloads), `/mapper` (converts domain/port structures <-> `go-openai` types), and `/errs`.
        - `/memory` — conversation history storage implementation with its own `/dto`, `/mapper`, and `/errs`.
  - `/migrations` — SQL migration files.
- `/frontend` — Web panel SPA source code and `Dockerfile`.
- `/docker-compose.yml` & `.env.example` — root orchestration and configuration template.

## 4. Layer Isolation: DTO Naming, Mappers & Error Handling Rules
Strict boundary isolation must be maintained across `domain`, `app`, and `adapter` layers:

### DTO Naming & File Conventions
- **No Leaking Types**: Never pass `adapter` structs into `app` or `domain`. Never expose `domain/models` directly in `adapter/in` responses without going through `app/dto` and `adapter/in/.../dto`.
- **`domain/models`**: Rich domain entities and value objects with business methods and invariants. No serialization tags (`json`, `db`).
- **`app/dto`**:
  - Must strictly follow the `{Name}Input` and `{Name}Output` naming convention, where `{Name}` is the use case name (e.g., `CreateExpenseInput` / `CreateExpenseOutput`, `ListExpensesInput` / `ListExpensesOutput`, `HandleMessageInput` / `HandleMessageOutput`).
  - Organize by one file per use case (e.g., `internal/app/dto/create_expense.go`, `internal/app/dto/list_expenses.go`).
- **`adapter/in/.../dto` & `adapter/out/.../dto`**:
  - Must strictly follow the `{Name}Request` and `{Name}Response` naming convention (e.g., `CreateExpenseRequest` / `CreateExpenseResponse`).
  - Can be organized as one file per operation (`<name>.go`) or grouped cleanly within the `/dto` package of that specific adapter.
- **Explicit Mappers (`/mapper`)**:
  - Every layer (`app`, `adapter/in/*`, `adapter/out/*`) must use its dedicated `/mapper` package to transform structs between adjacent layers (`adapter/in/dto` <-> `app/dto` <-> `domain/models` <-> `adapter/out/dto`).

### Per-Layer Errors (`/errs`)
- **Domain Errors (`internal/domain/errs`)**: Represent business invariant violations inside rich models and value objects (e.g., `ErrInvalidAmount`, `ErrInvalidCategory`).
- **Application Errors (`internal/app/errs`)**: Represent use-case and orchestration failures (e.g., `ErrExpenseNotFound`, `ErrForbiddenAccess`, `ErrAgentMaxIterations`).
- **Adapter Errors (`internal/adapter/.../errs`)**:
  - Outbound adapters (`postgres`, `llm`, `memory`) catch low-level driver/SDK errors (e.g., `pgx.ErrNoRows`, OpenAI API errors) and translate them into domain/application errors or wrapped adapter errors.
  - Inbound adapters (`http`, `telegram`) inspect `app/errs` and `domain/errs` via `errors.Is` / `errors.As` and map them to transport-specific responses (HTTP status codes + JSON error responses for Web; friendly Russian text for Telegram).

## 5. AI Agent Implementation Guidelines
- Use **Tool Calling (Function Calling)** for all CRUD operations (`add_expense`, `list_expenses`, `update_expense`, `delete_expense`).
- Always inject the current timestamp (`time.Now()`) and day of the week into the system prompt so the LLM accurately resolves relative dates ("вчера", "в этом месяце", "за неделю").
- Support a multi-step tool execution loop inside `app/usecases` (e.g., to handle "последнее было не 350, а 450", the agent calls `list_expenses` to find the latest record ID, then calls `update_expense`).
- Maintain short conversation history per `user_id` via `domain/port` memory interface so follow-up commands work naturally.

## 6. General Engineering Principles
- Follow DDD (Rich Domain Model), Clean Architecture, and Hexagonal Architecture principles. Dependency direction: `adapter -> app -> domain`.
- Keep code production-ready, strongly typed, and clean.
- Always propagate `context.Context` as the first argument in Go functions across `app` and `adapter` layers.
- Do NOT output comments in code unless explicitly requested by the user.

## 7. Verification & Commands
After making changes in `/backend`, always verify compilation and formatting:
- `go build ./...`
- `go vet ./...`
- `golangci-lint run` (if configured)

## 8. Git Commit Standards
You must commit all modifications using the Conventional Commits specification.
- Format: `<prefix>: <short imperative description in English without a trailing period>`
- Allowed Prefixes: `feat:`, `fix:`, `refactor:`, `docs:`, `chore:`
- Example: `feat: add expense repository postgres implementation`