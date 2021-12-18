package domain_test

import (
	"testing"

	"go.ectobit.com/arc/domain"
)

func TestIsValidPassword(t *testing.T) {
	t.Parallel()

	password, err := domain.HashPassword("test")
	if err != nil {
		t.Fatal(err)
	}

	user := domain.User{ //nolint:exhaustivestruct
		Password: password,
	}

	if !user.IsValidPassword("test") {
		t.Errorf(`IsValidPassword("test") = false; want true`)
	}
}
