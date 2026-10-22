# HTTP Handler & Error Mapper Refactoring Plan

Refactor the HTTP delivery adapter to adhere to strict DTO Request/Response boundaries, standardized error handling via `returnErr` / `mapper.HttpError`, and consistent structured slog logging.

## Proposed Changes

### 1. DTO Layer (`internal/adapter/in/http/dto/http.go`)
- Define dedicated Request and Response structs for each endpoint:
  - `CreateExpenseRequest` & `CreateExpenseResponse { Expense Expense }`
  - `GetExpenseRequest { ID string `param:"id"` }` & `GetExpenseResponse { Expense Expense }`
  - `UpdateExpenseRequest { ID string `param:"id"`, ... }` & `UpdateExpenseResponse { Expense Expense }`
  - `DeleteExpenseRequest { ID string `param:"id"` }` & `DeleteExpenseResponse {}`
  - `ListExpensesRequest` with `query` tags & `ListExpensesResponse { Items []Expense, Expenses []Expense }`
  - `GetDistinctUsersRequest {}` & `GetDistinctUsersResponse { UserIDs []int64 }`
- Define `Expense` representation struct and backward-compatible alias `ExpenseResponse = Expense`.
- Retain `ErrorResponse { Error string }`.

### 2. Error Mapper (`internal/adapter/in/http/mapper/error_mapper.go`)
- Clean out foreign `avito-kitchen` imports.
- Import `backend/internal/app/errs`, `backend/internal/domain/models`, and `backend/pkg/errs`.
- Implement `HttpError(err error) *pkgerrs.OutErr` handling:
  - `apperrs.WrappedError` inspection:
    - `apperrs.ErrInvalidInput` -> 400 Bad Request
    - `apperrs.ErrForbiddenAccess` -> 403 Forbidden
    - `apperrs.ErrExpenseNotFound` -> 404 Not Found
    - DB errors (`ErrCreateExpenseDB`, `ErrGetExpenseDB`, etc.) -> 500 Internal Server Error
  - Direct errors:
    - `pkgerrs.ErrInvalidJSON`, `pkgerrs.ErrInvalidIdentifier`, `pkgerrs.ErrValueIsInvalid`, `pkgerrs.ErrValueIsRequired`, `models.ErrSpentAtAfterCreatedAt`, `apperrs.ErrInvalidInput` -> 400 Bad Request
    - `apperrs.ErrExpenseNotFound`, `pkgerrs.ErrObjectNotFound` -> 404 Not Found
    - `apperrs.ErrForbiddenAccess` -> 403 Forbidden
    - `pkgerrs.ErrUnauthenticated` -> 401 Unauthorized
    - `pkgerrs.ErrObjectAlreadyExists` -> 409 Conflict
    - DB and service errors -> 500 Internal Server Error
    - Default fallback -> 500 Internal Server Error
- Implement `MapErrorToResponse(msg string) dto.ErrorResponse`.

### 3. HTTP Mapper (`internal/adapter/in/http/mapper/mapper.go`)
- Update / add mapping functions:
  - `ToCreateExpenseInput(req httpdto.CreateExpenseRequest) appdto.CreateExpenseInput`
  - `ToCreateExpenseResponse(d appdto.ExpenseDTO) httpdto.CreateExpenseResponse`
  - `ToGetExpenseInput(req httpdto.GetExpenseRequest) (appdto.GetExpenseInput, error)`
  - `ToGetExpenseResponse(d appdto.ExpenseDTO) httpdto.GetExpenseResponse`
  - `ToUpdateExpenseInput(req httpdto.UpdateExpenseRequest) (appdto.UpdateExpenseInput, error)`
  - `ToUpdateExpenseResponse(d appdto.ExpenseDTO) httpdto.UpdateExpenseResponse`
  - `ToDeleteExpenseInput(req httpdto.DeleteExpenseRequest) (appdto.DeleteExpenseInput, error)`
  - `ToListExpensesInput(req httpdto.ListExpensesRequest) (appdto.ListExpensesInput, error)`
  - `ToListExpensesResponse(dtos []appdto.ExpenseDTO) httpdto.ListExpensesResponse`
  - `ToGetDistinctUsersInput(req httpdto.GetDistinctUsersRequest) appdto.GetDistinctUsersInput`
  - `ToGetDistinctUsersResponse(userIDs []int64) httpdto.GetDistinctUsersResponse`
  - `ToExpense(d appdto.ExpenseDTO) httpdto.Expense`
  - `ToExpenseResponse(d appdto.ExpenseDTO) httpdto.Expense`
  - `ToExpenseResponses(dtos []appdto.ExpenseDTO) []httpdto.Expense`

### 4. HTTP Handler (`internal/adapter/in/http/expenses_handler.go`)
- Remove non-existent import `"backend/internal/adapter/in/http/errs"`.
- Update `Create` to return `mapper.ToCreateExpenseResponse(out.Expense)`.
- Refactor `Get`, `Update`, `Delete`, `List`, `GetDistinctUsers`:
  - Bind request with `c.Bind(&req)`.
  - Validate and map with mapper methods.
  - Return errors via `h.returnErr(...)`.
  - Add structured `slog.InfoContext` logging matching `Create` pattern.
  - Return responses via dedicated `{Name}Response` wrapper.
- Update `returnErr` to fallback to default `slog.LogAttrs` if `h.log == nil`.

### 5. Composition Root (`cmd/app/main.go`)
- Pass `logger` as first argument to `adapterhttp.NewExpensesHandler(logger, ...)`.

### 6. Frontend API Service (`frontend/src/services/api.ts`)
- Support unwrapping `{ expense: Expense }` on `createExpense` and `updateExpense` so web panel works seamlessly.

### 7. Unit Tests & Verification
- Update `mapper_test.go` for new signatures.
- Add `error_mapper_test.go` covering `HttpError` scenarios.
- Run `go build ./...`, `go vet ./...`, and `go test ./...`.
