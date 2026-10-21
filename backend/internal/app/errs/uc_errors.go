package errs

import (
	"errors"
)

/*
================ Invalidation & Conflicts ================
*/
var (
	ErrInvalidInput    = errors.New("invalid input")
	ErrForbiddenAccess = errors.New("access forbidden")
)

/*
================ Adapters failures ================
*/
var (
	ErrCreateExpenseDB      = errors.New("failed to create expense using db")
	ErrGetExpenseDB         = errors.New("failed to get expense using db")
	ErrUpdateExpenseDB      = errors.New("failed to update expense using db")
	ErrDeleteExpenseDB      = errors.New("failed to delete expense using db")
	ErrListExpensesDB       = errors.New("failed to get a list of expenses using db")
	ErrListLatestExpensesDB = errors.New("failed to get a list of latest expenses using db")
	ErrListUserIDSDB        = errors.New("failed to get a list of user ids using db")
	ErrExpenseNotFound      = errors.New("expense not found")

	ErrAdMessageStore         = errors.New("failed to append message using store")
	ErrAddMessageStore        = ErrAdMessageStore
	ErrGetRecentMessagesStore = errors.New("failed to get recent message using store")
	ErrClearMessagesStore     = errors.New("failed to clear messages using store")

	ErrGenerateReplyLLM   = errors.New("failed to process message using llm")
	ErrAgentMaxIterations = errors.New("agent reached maximum iteration limit")
	ErrAgentFailed        = errors.New("agent failed to process message")
)
