# Technical Design Specification: Family Expenses Tracker

## 1. Overview & Objectives
The system is a production-ready family expense tracker featuring two delivery mechanisms sharing a single core domain and PostgreSQL database:
1. **Telegram Bot with AI Agent**: Natural language CRUD interface over personal expenses (no buttons, no slash commands except initial `/start`). Enforces strict user isolation (`user_id` is always derived from Telegram session context).
2. **Web Panel (for Head of Family)**: Full CRUD interface across all family members with filtering (by `user_id`, date period `from`/`to`, and `category`), summary statistics, and interactive modals. No authentication required.
3. **Swagger / OpenAPI**: Live interactive documentation for the REST API available at `/swagger/index.html`.

---

## 2. Architecture & Monorepo Structure

The project strictly follows Hexagonal Architecture and Domain-Driven Design (Rich Domain Model) rules defined in `AGENTS.md`. Dependency direction is strictly inwards:
$$\text{adapter} \longrightarrow \text{app} \longrightarrow \text{domain}$$

### 2.1 Monorepo Layout
```text
expenses-accounter/
├── backend/
│   ├── cmd/
│   │   └── app/
│   │       └── main.go                  # Composition root, config, DI, graceful shutdown
│   ├── internal/
│   │   ├── config/                      # Environment configuration loading
│   │   ├── domain/                      # Core domain layer (zero external framework dependencies)
│   │   │   ├── models/                  # Expense entity, Category & Amount Value Objects
│   │   │   ├── port/                    # ExpenseRepository, LLMClient, ChatMemory interfaces
│   │   │   └── errs/                    # Domain-level validation/invariant errors
│   │   ├── app/                         # Application use cases and AI orchestration
│   │   │   ├── dto/                     # Use-case paired {Name}Input / {Name}Output structs
│   │   │   ├── usecases/                # ExpenseService, AgentOrchestrator
│   │   │   ├── mapper/                  # Pure mapping between domain/models <-> app/dto
│   │   │   └── errs/                    # Application and orchestration errors
│   │   └── adapter/                     # Primary and Secondary adapters
│   │       ├── in/
│   │       │   ├── http/                # Echo REST router, handlers, /dto, /mapper, /errs
│   │       │   └── telegram/            # Telebot v3 long-polling bot, handlers, /dto, /mapper, /errs
│   │       └── out/
│   │           ├── postgres/            # sqlc generated queries, pgx/v5 pool, /dto, /mapper, /errs
│   │           │   ├── queries/         # SQL query files for sqlc
│   │           │   └── sqlc/            # sqlc generated Go code (encapsulated inside adapter)
│   │           ├── llm/                 # go-openai client wrapper, tools, /dto, /mapper, /errs
│   │           └── memory/              # Thread-safe in-memory ring-buffer session store
│   ├── migrations/                      # golang-migrate SQL migration scripts
│   ├── docs/                            # Swaggo generated Swagger files
│   ├── sqlc.yaml                        # sqlc configuration
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile                       # Multi-stage Go build
├── frontend/
│   ├── src/                             # React 18 + TypeScript + Vite + Ant Design
│   │   ├── components/                  # ExpensesTable, ExpenseModal, FiltersBar, SummaryCards
│   │   ├── services/                    # Axios/fetch client for REST API
│   │   ├── types/                       # TypeScript interfaces matching API DTOs
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── nginx.conf                       # Static SPA serving and reverse-proxy for /api & /swagger
│   ├── package.json
│   ├── vite.config.ts
│   └── Dockerfile                       # Multi-stage Vite build to Nginx
├── docs/
│   └── superpowers/
│       └── specs/
│           └── 2026-10-20-family-expenses-tracker-design.md
├── docker-compose.yml
├── .env.example
├── TASK.md
└── AGENTS.md
```

### 2.2 Boundary Isolation & DTO Conventions
- **Domain Models (`internal/domain/models`)**: Rich domain entities and value objects containing business invariants. Free of `json:`, `db:`, or any framework tags.
- **Application DTOs (`internal/app/dto`)**: Named `{Name}Input` and `{Name}Output` (one file per usecase, e.g. `create_expense.go`, `handle_agent_message.go`).
- **Adapter DTOs (`adapter/in/.../dto` & `adapter/out/.../dto`)**: Named `{Name}Request` and `{Name}Response`.
- **Mappers (`/mapper`)**: Explicit mapping functions transforming data between adjacent layers.
- **Errors (`/errs`)**: Layer-specific error packages (`domain/errs`, `app/errs`, `adapter/.../errs`) with error mapping using `errors.Is` / `errors.As`.
- **Clean Code Rule**: Zero code comments unless explicitly requested.

---

## 3. Data Model & Database Persistence

### 3.1 Domain Entity: `Expense`
- `ID`: `uuid.UUID`
- `UserID`: `int64` (Telegram User ID)
- `Amount`: `Amount` Value Object (validates `> 0`, methods `Value() float64`, `String()`)
- `Category`: `Category` Value Object (enum validation against: `еда`, `транспорт`, `жильё`, `развлечения`, `здоровье`, `покупки`, `другое`)
- `SpentAt`: `time.Time` (timestamp of expense)
- `Comment`: `string` (optional note, e.g., "кофе", "такси")
- `CreatedAt`: `time.Time`
- `UpdatedAt`: `time.Time`

### 3.2 SQL Schema (`migrations/000001_create_expenses_table.up.sql`)
```sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    amount NUMERIC(12, 2) NOT NULL CHECK (amount > 0),
    category VARCHAR(50) NOT NULL CHECK (category IN (
        'еда', 'транспорт', 'жильё', 'развлечения', 'здоровье', 'покупки', 'другое'
    )),
    spent_at TIMESTAMPTZ NOT NULL,
    comment TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_expenses_user_id ON expenses(user_id);
CREATE INDEX IF NOT EXISTS idx_expenses_spent_at ON expenses(spent_at);
CREATE INDEX IF NOT EXISTS idx_expenses_category ON expenses(category);
CREATE INDEX IF NOT EXISTS idx_expenses_user_spent_at ON expenses(user_id, spent_at DESC);
```

### 3.3 sqlc Queries (`internal/adapter/out/postgres/queries/expenses.sql`)
- `CreateExpense`: inserts new record and returns row.
- `GetExpenseByID`: returns expense by UUID.
- `UpdateExpense`: updates amount, category, spent_at, comment, and updated_at.
- `DeleteExpense`: deletes row by UUID.
- `ListExpenses`: dynamic or conditional filtering by `user_id`, `category`, `from`, `to`, with `LIMIT` and `OFFSET`.
- `GetLatestExpensesByUser`: returns latest N expenses for a given `user_id` ordered by `spent_at DESC`.
- `GetDistinctUserIDs`: returns distinct list of all `user_id`s in expenses table.

`sqlc` models and query executors are generated inside `internal/adapter/out/postgres/sqlc/` using `pgx/v5`. The postgres adapter mapper converts between sqlc row types and domain models.

---

## 4. Configuration Management

Database credentials use separate environment variables with the `DB_` (and `DBB_`) prefix:
- `DB_HOST` (default: `postgres`)
- `DB_PORT` (default: `5432`)
- `DB_USER` (default: `postgres`)
- `DB_PASSWORD` or `DB_PASS` (default: `postgres`)
- `DB_NAME` (default: `expenses`)
- `DB_SSLMODE` or `DB_SSL_MODE` (default: `disable`)

Additional settings:
- `PORT` (default: `8080`)
- `TELEGRAM_BOT_TOKEN`
- `LLM_BASE_URL` (default: `https://api.openai.com/v1`)
- `LLM_API_KEY`
- `LLM_MODEL` (default: `gpt-4o-mini`)

---

## 5. Application Layer & AI Agent Orchestration

### 5.1 Use Cases (`internal/app/usecases`)
1. **ExpenseUseCase**:
   - `CreateExpense(ctx, dto.CreateExpenseInput) (dto.CreateExpenseOutput, error)`
   - `GetExpense(ctx, dto.GetExpenseInput) (dto.GetExpenseOutput, error)`
   - `UpdateExpense(ctx, dto.UpdateExpenseInput) (dto.UpdateExpenseOutput, error)`
   - `DeleteExpense(ctx, dto.DeleteExpenseInput) (dto.DeleteExpenseOutput, error)`
   - `ListExpenses(ctx, dto.ListExpensesInput) (dto.ListExpensesOutput, error)`
   - `GetDistinctUsers(ctx, dto.GetDistinctUsersInput) (dto.GetDistinctUsersOutput, error)`
2. **AgentUseCase** (`internal/app/usecases/handle_agent_message.go`):
   - `HandleMessage(ctx, dto.HandleMessageInput) (dto.HandleMessageOutput, error)`

### 5.2 AI Agent Orchestration & Tool Calling
- **OpenAI Tool Calling**:
  1. `add_expense`: arguments `amount` (float), `category` (enum), `spent_at` (RFC3339 string), `comment` (string).
  2. `list_expenses`: arguments `limit` (int, default 10), `category` (optional enum), `from` (optional RFC3339 string), `to` (optional RFC3339 string). Always strictly scoped to the Telegram caller's `user_id`.
  3. `update_expense`: arguments `id` (UUID string), `amount` (optional float), `category` (optional enum), `spent_at` (optional RFC3339 string), `comment` (optional string). Validates ownership against caller's `user_id`.
  4. `delete_expense`: arguments `id` (UUID string). Validates ownership against caller's `user_id`.
- **System Prompt Dynamics**:
  - Injected with `time.Now()` formatted as ISO timestamp and human-readable Russian string with day of week (e.g., "Сегодня: Вторник, 20 октября 2026, 03:20 MSK").
  - Clear instructions on resolving relative date references ("вчера" = current day - 1, "в этом месяце" = start of month to now, etc.).
  - Enforced constraint that categories must be strictly one of the 7 valid Russian enum values.
- **Multi-Step Execution Loop**:
  - Supports up to 5 iterations per message to resolve compound actions (e.g. for "последнее было не 350, а 450", agent calls `list_expenses(limit=1)` to find the most recent expense ID, and then calls `update_expense(id=..., amount=450)`).
- **Strict Telegram Isolation**:
  - `user_id` is supplied solely by the Telegram update context (`c.Sender().ID`) and injected into the tool execution handler. The LLM cannot access or mutate any other user's records.
- **Chat Memory (`domain/port.ChatMemory`)**:
  - In-memory thread-safe ring buffer (`sync.RWMutex`) storing the last 10 dialog turns per `user_id`.

---

## 6. Delivery Adapters (Driving / Primary)

### 6.1 HTTP REST API (`internal/adapter/in/http`)
- Implemented with `github.com/labstack/echo/v4`.
- Endpoints:
  - `GET /api/health`
  - `POST /api/expenses`
  - `GET /api/expenses` (query params: `user_id`, `category`, `from`, `to`, `limit`, `offset`)
  - `GET /api/expenses/:id`
  - `PUT /api/expenses/:id`
  - `DELETE /api/expenses/:id`
  - `GET /api/users`
- Swagger UI served at `/swagger/*` with `echo-swagger`.
- Handlers annotated with Swaggo tags generating docs in `backend/docs`.
- Error mapping via `internal/adapter/in/http/errs` into HTTP 400, 403, 404, 500 JSON payloads.

### 6.2 Telegram Bot (`internal/adapter/in/telegram`)
- Implemented with `gopkg.in/telebot.v3` using Long Polling.
- Natural language only:
  - `/start`: Greeting message explaining text-based expense tracking in Russian.
  - `telebot.OnText`: Calls typing indicator `c.Notify(telebot.Typing)`, delegates to `AgentUseCase.HandleMessage`, and replies with the assistant's response.
  - Error responses translated into friendly Russian messages.

---

## 7. Frontend Web Panel (`/frontend`)

- **Tech Stack**: React 18 + TypeScript + Vite + Ant Design (`antd`).
- **Pages / Components**:
  - `Header`: Title and "+ Добавить расход" action button.
  - `SummaryCards`: Total spent in rubles, total count of transactions, and top category for the filtered dataset.
  - `FiltersBar`:
    - Family member (`user_id`): Select populated by `/api/users` with free-form custom input support.
    - Date range: AntD `RangePicker` (`from` / `to`).
    - Category: AntD `Select` (7 categories + "Все категории").
    - "Сбросить фильтры" button.
  - `ExpensesTable`:
    - Columns: Date & Time (`spent_at`), User ID (`user_id`), Category (colored AntD `Tag`), Amount (formatted currency), Comment, Actions (Edit, Delete).
    - Client/server pagination and sorting.
  - `ExpenseModal`:
    - AntD `Form` for creating and editing records: fields for `user_id`, `amount` (> 0), `category`, `spent_at`, and `comment`.
  - `Popconfirm` for safe deletion.
- **Nginx Reverse Proxy**:
  - Serves static assets on port `80` (exposed on host `3000`).
  - Proxies `/api/` and `/swagger/` to `http://backend:8080/`.

---

## 8. Infrastructure & Docker Compose

- `docker-compose.yml`:
  1. `postgres`: `postgres:16-alpine`, port `5432:5432`, healthcheck `pg_isready -U postgres -d expenses`, persistent volume `pgdata`.
  2. `backend`: Multi-stage build (`golang:1.26-alpine` -> `alpine:latest`). Depends on `postgres: condition: service_healthy`. Automatically applies migrations on startup via `golang-migrate`.
  3. `frontend`: Multi-stage build (`node:20-alpine` -> `nginx:alpine`). Depends on `backend`. Exposes port `3000:80`.
- Single command execution: `docker compose up --build`.

---

## 9. Verification & Quality Assurance
- Backend compilation and static analysis:
  - `go build ./...`
  - `go vet ./...`
- Unit tests:
  - Domain invariants (Amount validation, Category validation).
  - Agent multi-step loop and tool call dispatcher.
  - HTTP handlers and mapper tests.
- Conventional Commits enforced across all git commits.
