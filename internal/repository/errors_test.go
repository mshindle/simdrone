package repository

import (
	"errors"
	"testing"
)

func TestDBError_Is(t *testing.T) {
	innerErr := errors.New("mongo: no documents in result")
	wrappedErr := ErrNotFound.Wrap(innerErr)

	t.Run("errors.Is with ErrNotFound", func(t *testing.T) {
		if !errors.Is(wrappedErr, ErrNotFound) {
			t.Errorf("expected errors.Is(wrappedErr, ErrNotFound) to be true")
		}
	})

	t.Run("errors.Is with inner error", func(t *testing.T) {
		if !errors.Is(wrappedErr, innerErr) {
			t.Errorf("expected errors.Is(wrappedErr, innerErr) to be true")
		}
	})

	t.Run("errors.Is with unrelated DBError", func(t *testing.T) {
		unrelated := NewDBError("unrelated error")
		if errors.Is(wrappedErr, unrelated) {
			t.Errorf("expected errors.Is(wrappedErr, unrelated) to be false")
		}
	})
}
