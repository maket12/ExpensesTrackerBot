package errs

type WrappedError struct {
	Public error
	Reason error
}

func (e *WrappedError) Error() string {
	return e.Public.Error()
}

func (e *WrappedError) Unwrap() []error {
	if e.Reason != nil {
		return []error{e.Public, e.Reason}
	}
	return []error{e.Public}
}

func Wrap(public, reason error) error {
	return &WrappedError{
		Public: public,
		Reason: reason,
	}
}
