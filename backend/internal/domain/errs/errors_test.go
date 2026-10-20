package errs

import (
	"errors"
	"testing"

	pkgerrs "backend/pkg/errs"
)

func TestDomainErrorsWrapping(t *testing.T) {
	if !errors.Is(ErrInvalidAmount, pkgerrs.ErrValueIsInvalid) {
		t.Errorf("expected ErrInvalidAmount to unwrap to ErrValueIsInvalid")
	}

	if !errors.Is(ErrInvalidCategory, pkgerrs.ErrValueIsInvalid) {
		t.Errorf("expected ErrInvalidCategory to unwrap to ErrValueIsInvalid")
	}

	if !errors.Is(ErrInvalidUserID, pkgerrs.ErrValueIsInvalid) {
		t.Errorf("expected ErrInvalidUserID to unwrap to ErrValueIsInvalid")
	}

	if !errors.Is(ErrInvalidSpentAt, pkgerrs.ErrValueIsInvalid) {
		t.Errorf("expected ErrInvalidSpentAt to unwrap to ErrValueIsInvalid")
	}

	if !errors.Is(ErrExpenseNotFound, pkgerrs.ErrObjectNotFound) {
		t.Errorf("expected ErrExpenseNotFound to match ErrObjectNotFound")
	}

	notFoundErr := NewExpenseNotFoundError("some-id")
	if !errors.Is(notFoundErr, pkgerrs.ErrObjectNotFound) {
		t.Errorf("expected NewExpenseNotFoundError to unwrap to ErrObjectNotFound")
	}

	if !errors.Is(notFoundErr, ErrExpenseNotFound) {
		t.Errorf("expected NewExpenseNotFoundError to match ErrExpenseNotFound")
	}
}
