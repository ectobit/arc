package repository_test

import (
	"testing"

	"go.ectobit.com/arc/repository"
)

func TestUserLoginFromJSON(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		in   string
		want string
	}{
		"empty string":                    {"", ""},
		"multiple spaces":                 {"   ", ""},
		"multiple spaces with characters": {" a  b  c  ", "a b c"},
		"new lines and tabs": {`  a
b  c `, "a b c"},
	}

	for n, test := range tests { //nolint:paralleltest
		test := test

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			got := repository.StripWhitespaces(test.in)

			if got != test.want {
				t.Errorf("StripWhitespaces(%q) = %q; want %q", test.in, got, test.want)
			}
		})
	}
}
