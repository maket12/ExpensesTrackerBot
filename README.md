# Family Expenses Tracker (AI Bot & Web Panel)

Сервис учёта семейных расходов, предоставляющий два независимых интерфейса к единому ядру и базе данных:
1. **Telegram-бот с ИИ-агентом**: управление расходами на естественном языке (текстовые сообщения без кнопок и команд). Реализована строгая изоляция: каждый член семьи имеет доступ исключительно к своим тратам.
2. **Веб-панель для главы семьи**: полный CRUD-интерфейс над расходами всех членов семьи с фильтрацией (по пользователю, категориям, интервалам дат) и аналитикой. Авторизация не требуется.

---

## 1. Архитектура и принципы проектирования

Проект спроектирован по принципам **Domain-Driven Design (DDD)**, **Hexagonal Architecture (Ports and Adapters)** и **Clean Architecture**:
- Строгое направление зависимостей: `Adapter -> App -> Domain`. Внутренние слои ничего не знают о внешних транспортах и драйверах.
- **Rich Domain Model**: бизнес-правила и инварианты инкапсулированы внутри доменных сущностей и Value Objects.

```
                              ┌──────────────────────────────────┐
                              │       Driving Adapters           │
                              │  (HTTP Echo REST / Telegram Bot) │
                              └─────────────────┬────────────────┘
                                                │
                                                ▼
┌────────────────────────────────────────────────────────────────────────────────┐
│ Application Layer (internal/app)                                               │
│  - Use Cases (Create, Get, Update, Delete, List, GetDistinctUsers, AgentLoop) │
│  - Use Case DTOs ({Name}Input / {Name}Output)                                  │
│  - Mappers & Application Errors                                                │
└───────────────────────────────────────┬────────────────────────────────────────┘
                                        │
                                        ▼
┌────────────────────────────────────────────────────────────────────────────────┐
│ Domain Core (internal/domain)                                                  │
│  - Models & Value Objects (Expense, Category) — Rich Domain Model              │
│  - Domain Invariants & Validations                                             │
│  - Ports / Interfaces (ExpenseRepository, AIAssistant, ChatMemory)             │
└───────────────────────────────────────▲────────────────────────────────────────┘
                                        │
                              ┌─────────┴────────────────────────┐
                              │       Driven Adapters            │
                              │  (Postgres pgx, OpenAI LLM, Mem) │
                              └──────────────────────────────────┘
```

### Изоляция слоёв и соглашения об именовании

- **`domain/models`**: Богатые доменные модели без тегов сериализации (`json`, `db`). Инварианты проверяются фабриками (`NewExpense`, `RestoreExpense`) и методами мутации (`Update`).
- **`app/dto`**: Строгое именование `{Name}Input` и `{Name}Output` (один файл на юзкейс: `create_expense.go`, `list_expenses.go` и т.д.).
- **`adapter/in/.../dto` и `adapter/out/.../dto`**: Транспортные структуры вида `{Name}Request` и `{Name}Response`.
- **`mapper`**: Каждый слой имеет собственный пакет мапперов для чистой трансляции структур между смежными слоями.
- **`errs`**: Иерархия ошибок от доменных (`ErrSpentAtAfterCreatedAt`) и прикладных (`ErrInvalidInput`, `ErrExpenseNotFound`, `ErrForbiddenAccess`) до централизованного маппинга в HTTP-коды и дружелюбные ответы.

---

## 2. Модель данных (`Expense`)

| Поле | Тип | Описание и инварианты |
|---|---|---|
| `id` | UUID | Первичный ключ (генерируется при создании) |
| `user_id` | `int64` | Telegram User ID владельца (`> 0`) |
| `amount` | `numeric(12,2)` | Сумма расхода (`1.0 <= amount <= 1 000 000.0`) |
| `category` | Enum | Строго ограниченный список: `еда`, `транспорт`, `жильё`, `развлечения`, `здоровье`, `покупки`, `другое` |
| `spent_at` | `timestamptz` | Дата и время совершения расхода (не может быть в будущем и старше 30 дней) |
| `comment` | `text` | Необязательный комментарий (от 2 до 256 символов) |
| `created_at` | `timestamptz` | Время добавления записи |
| `updated_at` | `timestamptz` | Время последнего изменения |

---

## 3. ИИ-агент (Telegram Bot)

- **OpenAI Tool/Function Calling**: агент не генерирует сырой SQL, а вызывает строго типизированные функции (`add_expense`, `list_expenses`, `update_expense`, `delete_expense`).
- **Динамический системный промпт**: перед каждым запросом внедряется текущее точное время (`time.Now().UTC()`) и день недели, что позволяет точно вычислять относительные даты («вчера», «в прошлый вторник», «за эту неделю»).
- **Многошаговый цикл рассуждений (Agent Loop)**: агент поддерживает цепочки вызовов инструментов (например, для запроса «последнее было не 350, а 450» агент сначала запрашивает последние траты, находит ID нужной записи и вызывает `update_expense`).
- **Безопасность и изоляция**: `user_id` никогда не принимается из аргументов модели, а всегда принудительно инжектируется из контекста Telegram-сессии. Член семьи физически не может просмотреть или изменить траты другого пользователя.
- **Контекстная память**: история последних реплик сохраняется в in-memory хранилище (`ChatMemory`), поддерживая естественный диалог с уточнениями.
- **Чистый вывод**: строго запрещены Markdown-таблицы (не поддерживаемые клиентом Telegram) и системные UUID. Траты выводятся компактным списком с HTML-выделением.

### Примеры диалога с ботом

| Сообщение пользователя | Поведение ИИ-агента | Формат ответа в Telegram |
|---|---|---|
| *«кофе 350»* | Вызывает функцию `add_expense` с суммой `350` и категорией `еда` | «Записал: кофе <b>350 ₽</b> (еда)» |
| *«вчера такси 1200»* | Разрешает дату «вчера» относительно системного времени, вызывает `add_expense` | «Записал: такси <b>1 200 ₽</b> (транспорт) за вчера» |
| *«какие траты за неделю?»* | Запрашивает `list_expenses` за последние 7 дней | «Вот ваши траты за неделю:<br>• <b>1 200 ₽</b> — транспорт: такси (22.10 в 18:30)<br>• <b>350 ₽</b> — еда: кофе (23.10 в 09:15)<br><br><b>Итого за неделю: 1 550 ₽</b>» |
| *«последнее было не 350, а 450»* | В цикле рассуждений: сначала `list_expenses(limit=5)` для поиска ID последней записи, затем `update_expense` | «Обновил последнюю запись: сумма изменена на <b>450 ₽</b>» |
| *«удали такси»* | Находит нужный расход по контексту и удаляет через `delete_expense` | «Удалил трату: такси на <b>1 200 ₽</b>» |

---

## 4. Стек технологий

### Бэкенд
- **Go 1.26.2**
- **HTTP Router**: `github.com/labstack/echo/v4` + Swagger (`echo-swagger`)
- **База данных**: PostgreSQL 16 + `github.com/jackc/pgx/v5`
- **Генерация запросов**: `sqlc`
- **Миграции**: `golang-migrate` (запускаются отдельным легковесным контейнером в compose)
- **Telegram SDK**: `gopkg.in/telebot.v3` (Long Polling)
- **LLM SDK**: `github.com/sashabaranov/go-openai` (совместим с OpenAI, OpenRouter, Groq, Ollama, vLLM)
- **Логирование**: встроенный структурированный `log/slog`
- **Graceful Shutdown**: `golang.org/x/sync/errgroup` + `signal.NotifyContext`

### Фронтенд
- **React 18 + TypeScript 5 + Vite**
- **UI-библиотека**: Ant Design (`antd`, `@ant-design/icons`)
- **HTTP-клиент**: `axios`
- **Тестирование**: `vitest`

### Инфраструктура
- **Docker & Docker Compose** (multi-stage сборки, healthchecks, порядок старта через `condition`)
- **CI/CD**: GitHub Actions (проверка линтерами, сборка и прогон unit/integration/e2e тестов)

---

## 5. Структура репозитория

```
├── .github/workflows/
│   └── ci.yml                 # CI/CD пайплайн (бэкенд + фронтенд)
├── backend/
│   ├── cmd/app/
│   │   ├── config/            # Загрузка и валидация конфигурации через env
│   │   └── main.go            # Composition root, DI, graceful shutdown
│   ├── docs/                  # Swagger/OpenAPI спецификации
│   ├── internal/
│   │   ├── domain/
│   │   │   ├── models/        # Rich Domain Models (Expense, Category)
│   │   │   └── port/          # Интерфейсы репозиториев, LLM и памяти
│   │   ├── app/
│   │   │   ├── dto/           # {Name}Input и {Name}Output для use-cases
│   │   │   ├── usecases/      # Бизнес-логика и оркестрация ИИ-агента
│   │   │   ├── mapper/        # Мапперы domain <-> app/dto
│   │   │   └── errs/          # Ошибки уровня приложения
│   │   └── adapter/
│   │       ├── in/
│   │       │   ├── http/      # REST API (хендлеры, роутер, DTO, error mapper)
│   │       │   └── telegram/  # Telegram Bot (обработчики сообщений, DTO)
│   │       └── out/
│   │           ├── postgres/  # Репозиторий расходов на базе pgx/sqlc
│   │           ├── llm/       # Клиент OpenAI Tool Calling и промпты
│   │           └── memory/    # In-memory хранилище истории диалога
│   ├── migrations/            # SQL-миграции схемы БД + embed.go
│   ├── pkg/
│   │   ├── errs/              # Унифицированные структуры ошибок и HTTP-маппинг
│   │   └── postgres/          # Клиент подключения и Testcontainers
│   └── tests/
│       └── e2e/               # Сквозные E2E-тесты HTTP API
├── frontend/
│   ├── src/
│   │   ├── components/        # UI-компоненты (модалка, таблица, фильтры, карточки)
│   │   ├── services/          # API-клиент (axios)
│   │   └── types/             # Типы расходов и фильтров
│   ├── package.json           # Зависимости и скрипты
│   └── vite.config.ts         # Конфигурация Vite
├── docker-compose.yml         # Оркестрация сервисов
└── .env.example               # Шаблон переменных окружения
```

---

## 6. Тестирование

Проект покрыт пирамидой тестов: от изолированных модульных до интеграционных и e2e.

### 1. Модульные тесты (Unit Tests)
Покрывают бизнес-логику домена, валидацию инвариантов модели, преобразования мапперов, юзкейсы с моками и хендлеры:
- `backend/internal/domain/models/expense_test.go`
- `backend/internal/app/usecases/*_test.go`
- `backend/internal/adapter/in/http/expenses_handler_test.go`
- `backend/internal/adapter/in/http/mapper/*_test.go`
- `frontend/src/services/api.test.ts`, `frontend/src/types/expense.test.ts`

Запуск:
```bash
cd backend
go test -v -count=1 ./...
```

### 2. Интеграционные тесты (Integration Tests)
Тестируют реальное взаимодействие репозитория `ExpensesRepository` с PostgreSQL в изолированном контейнере Docker через `testcontainers-go`:
- Выполняют автоматический накат миграций схемы.
- Очищают таблицы между тестами (`TruncateTables`).
- Проверяют все SQL-запросы: вставку, выборку по ID, частичное обновление, удаление, сложные фильтры и выборку уникальных ID.

Запуск:
```bash
cd backend
go test -v -tags=integration ./internal/adapter/out/postgres/...
```

### 3. Сквозные тесты (E2E Tests)
Поднимают реальный HTTP-сервер `httptest.NewServer` с подключением к тестовому контейнеру БД и проверяют работу полного контура REST API через реальные HTTP-запросы:
- `backend/tests/e2e/expenses_test.go`: полное тестирование CRUD над расходами, валидация некорректных параметров (400), обработка несуществующих записей (404), фильтрация и пагинация.
- `backend/tests/e2e/users_test.go`: получение уникальных ID членов семьи и healthcheck.

Запуск:
```bash
cd backend
go test -v -tags=e2e ./tests/e2e/...
```

---

## 7. Инструкция по запуску

### 1. Предварительные требования
- Docker и Docker Compose (V2)
- Telegram Bot Token (получить у [@BotFather](https://t.me/BotFather))
- Ключ доступа к LLM (OpenAI API или любой совместимый провайдер, например, OpenRouter)

### 2. Конфигурация окружения
Скопируйте шаблон переменных окружения:
```bash
cp .env.example .env
```

Заполните переменные в `.env`:
```env
# База данных
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=expenses

# Telegram Bot
TELEGRAM_BOT_TOKEN=123456789:ABCdefGHIjklMNOpqrsTUVwxyz

# Настройки LLM (пример для OpenRouter или OpenAI)
LLM_BASE_URL=https://openrouter.ai/api/v1
LLM_API_KEY=sk-or-v1-xxxxxxxxxxxxxxxxxxxx
LLM_MODEL=openrouter/free
# Либо для OpenAI:
# LLM_BASE_URL=https://api.openai.com/v1
# LLM_API_KEY=sk-proj-xxxxxxxxxxxxxxxxxxxx
# LLM_MODEL=gpt-4o-mini

# Бэкенд
HTTP_PORT=8080
LOG_LEVEL=INFO
ENVIRONMENT=production
```

### 3. Запуск всего приложения
Запуск осуществляется одной командой:
```bash
docker compose up --build
```

Compose автоматически:
1. Запустит PostgreSQL 16 и дождется его готовности (`healthcheck`).
2. Запустит контейнер `migrations`, накатит схему БД и корректно завершится (`service_completed_successfully`).
3. Соберёт и запустит бэкенд на Go (`http://localhost:8080`).
4. Соберёт и запустит SPA веб-панели под Nginx (`http://localhost:3000`).

### 4. Точки доступа
- **Веб-панель**: [http://localhost:3000](http://localhost:3000)
- **REST API бэкенда**: [http://localhost:8080/api](http://localhost:8080/api)
- **Интерактивная документация Swagger**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
- **Healthcheck**: [http://localhost:8080/api/health](http://localhost:8080/api/health)
- **Telegram Bot**: готов принимать сообщения в Telegram сразу после старта бэкенда.

### 5. Спецификация REST API

| Метод | Эндпоинт | Назначение |
|---|---|---|
| `POST` | `/api/expenses` | Создание новой траты (возвращает 201 Created) |
| `GET` | `/api/expenses` | Список трат с фильтрами (`user_id`, `category`, `from`, `to`, `limit`, `offset`) |
| `GET` | `/api/expenses/{id}` | Получение траты по UUID (возвращает 404, если не найдена) |
| `PUT` | `/api/expenses/{id}` | Частичное / полное обновление полей траты |
| `DELETE` | `/api/expenses/{id}` | Удаление траты (204 No Content при успехе, 404 если не найдена) |
| `GET` | `/api/users` | Список уникальных Telegram ID членов семьи для выпадающих списков UI |
| `GET` | `/api/health` | Проверка статуса сервиса |
| `GET` | `/swagger/*` | Интерактивный интерфейс Swagger UI с возможностью отправки запросов |

---

## 8. Локальный запуск и разработка (без Docker)

Если вы хотите запустить компоненты локально для отладки:

1. **Запуск PostgreSQL и накат миграций**:
   ```bash
   docker run --name local-pg -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=expenses -p 5432:5432 -d postgres:16-alpine
   migrate -path backend/migrations -database "postgres://postgres:postgres@localhost:5432/expenses?sslmode=disable" up
   ```
2. **Запуск бэкенда**:
   ```bash
   cd backend
   go run cmd/app/main.go
   ```
3. **Запуск веб-панели**:
   ```bash
   cd frontend
   pnpm install
   pnpm dev
   ```

---

## 9. Контроль качества и CI/CD

В проекте настроен автоматический CI-пайплайн на базе GitHub Actions (`.github/workflows/ci.yml`), который при каждом пуше и Pull Request проверяет:
- **Backend Linting**: строгая проверка через `golangci-lint` (включая `errcheck`, `govet`, `staticcheck`, `unused`, `gofmt` и др.).
- **Backend Tests**: 100% прогон модульных, интеграционных и сквозных (E2E) тестов в изолированном PostgreSQL-контейнере через Testcontainers.
- **Frontend Quality**: статическая типизация TypeScript (`tsc --noEmit`), сборка продакшен-бандла Vite и выполнение тестов Vitest.
