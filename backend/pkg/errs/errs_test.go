package errs

import (
	"errors"
	"net/http"
	"testing"
)

func TestPkgErrors(t *testing.T) {
	notFound := NewObjectNotFoundError("user", 123)
	if !errors.Is(notFound, ErrObjectNotFound) {
		t.Errorf("expected notFound to unwrap to ErrObjectNotFound")
	}

	exists := NewObjectAlreadyExistsError("email")
	if !errors.Is(exists, ErrObjectAlreadyExists) {
		t.Errorf("expected exists to unwrap to ErrObjectAlreadyExists")
	}

	unauth := NewNotAuthenticatedError()
	if !errors.Is(unauth, ErrNotAuthenticated) {
		t.Errorf("expected unauth to unwrap to ErrNotAuthenticated")
	}

	invalid := NewValueInvalidError("status")
	if !errors.Is(invalid, ErrValueIsInvalid) {
		t.Errorf("expected invalid to unwrap to ErrValueIsInvalid")
	}

	required := NewValueRequiredError("name")
	if !errors.Is(required, ErrValueIsRequired) {
		t.Errorf("expected required to unwrap to ErrValueIsRequired")
	}

	outErr := NewOutError(http.StatusBadRequest, "bad request", nil)
	if outErr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", outErr.Code)
	}
}
