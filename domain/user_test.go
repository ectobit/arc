package domain_test

import (
	"testing"

	"go.ectobit.com/arc/domain"
)

func TestIsValidPassword(t *testing.T) {
	t.Parallel()

	password, err := domain.HashPassword("test")
	if err != nil {
		t.Error(err)
	}

	user := domain.User{ //nolint:exhaustivestruct
		Password: password,
	}

	if !user.IsValidPassword("test") {
		t.Error("password should be valid")
	}
}
