package mapper

import (
	"errors"
	"net/http"

	"backend/internal/adapter/in/http/dto"
	ucerrs "backend/internal/app/errs"
	"backend/internal/domain/models"
	pkgerrs "backend/pkg/errs"
)

func HttpError(err error) *pkgerrs.OutErr {
	if err == nil {
		return nil
	}

	var wErr *ucerrs.WrappedError
	if errors.As(err, &wErr) {
		switch {
		case errors.Is(wErr.Public, ucerrs.ErrInvalidInput):
			msg := wErr.Public.Error()
			if wErr.Reason != nil {
				msg = msg + ": " + wErr.Reason.Error()
			}
			return pkgerrs.NewOutError(http.StatusBadRequest, msg, wErr.Reason)

		case errors.Is(wErr.Public, ucerrs.ErrForbiddenAccess):
			return pkgerrs.NewOutError(http.StatusForbidden, wErr.Public.Error(), wErr.Reason)

		case errors.Is(wErr.Public, ucerrs.ErrExpenseNotFound):
			return pkgerrs.NewOutError(http.StatusNotFound, wErr.Public.Error(), wErr.Reason)

		case errors.Is(wErr.Public, ucerrs.ErrCreateExpenseDB),
			errors.Is(wErr.Public, ucerrs.ErrGetExpenseDB),
			errors.Is(wErr.Public, ucerrs.ErrUpdateExpenseDB),
			errors.Is(wErr.Public, ucerrs.ErrDeleteExpenseDB),
			errors.Is(wErr.Public, ucerrs.ErrListExpensesDB),
			errors.Is(wErr.Public, ucerrs.ErrListLatestExpensesDB),
			errors.Is(wErr.Public, ucerrs.ErrListUserIDSDB),
			errors.Is(wErr.Public, ucerrs.ErrAddMessageStore),
			errors.Is(wErr.Public, ucerrs.ErrGetRecentMessagesStore),
			errors.Is(wErr.Public, ucerrs.ErrClearMessagesStore),
			errors.Is(wErr.Public, ucerrs.ErrGenerateReplyLLM),
			errors.Is(wErr.Public, ucerrs.ErrAgentMaxIterations),
			errors.Is(wErr.Public, ucerrs.ErrAgentFailed):
			return pkgerrs.NewOutError(http.StatusInternalServerError, "internal error", wErr.Reason)

		default:
			return pkgerrs.NewOutError(http.StatusInternalServerError, "internal error", wErr.Reason)
		}
	}

	switch {
	case errors.Is(err, pkgerrs.ErrInvalidJSON),
		errors.Is(err, pkgerrs.ErrInvalidIdentifier),
		errors.Is(err, pkgerrs.ErrValueIsInvalid),
		errors.Is(err, pkgerrs.ErrValueIsRequired),
		errors.Is(err, models.ErrSpentAtAfterCreatedAt),
		errors.Is(err, ucerrs.ErrInvalidInput):
		return pkgerrs.NewOutError(http.StatusBadRequest, err.Error(), err)

	case errors.Is(err, ucerrs.ErrExpenseNotFound),
		errors.Is(err, pkgerrs.ErrObjectNotFound):
		return pkgerrs.NewOutError(http.StatusNotFound, err.Error(), err)

	case errors.Is(err, ucerrs.ErrForbiddenAccess):
		return pkgerrs.NewOutError(http.StatusForbidden, err.Error(), err)

	case errors.Is(err, pkgerrs.ErrObjectAlreadyExists):
		return pkgerrs.NewOutError(http.StatusConflict, err.Error(), err)

	case errors.Is(err, ucerrs.ErrCreateExpenseDB),
		errors.Is(err, ucerrs.ErrGetExpenseDB),
		errors.Is(err, ucerrs.ErrUpdateExpenseDB),
		errors.Is(err, ucerrs.ErrDeleteExpenseDB),
		errors.Is(err, ucerrs.ErrListExpensesDB),
		errors.Is(err, ucerrs.ErrListLatestExpensesDB),
		errors.Is(err, ucerrs.ErrListUserIDSDB),
		errors.Is(err, ucerrs.ErrAddMessageStore),
		errors.Is(err, ucerrs.ErrGetRecentMessagesStore),
		errors.Is(err, ucerrs.ErrClearMessagesStore),
		errors.Is(err, ucerrs.ErrGenerateReplyLLM),
		errors.Is(err, ucerrs.ErrAgentMaxIterations),
		errors.Is(err, ucerrs.ErrAgentFailed):
		return pkgerrs.NewOutError(http.StatusInternalServerError, "internal error", err)
	}

	return pkgerrs.NewOutError(http.StatusInternalServerError, "internal error", err)
}

func MapErrorToResponse(msg string) dto.ErrorResponse {
	return dto.ErrorResponse{Error: msg}
}
