package request_test

import (
	"bytes"
	"testing"

	"go.ectobit.com/arc/domain"
	"go.ectobit.com/arc/handler/request"
	"go.ectobit.com/lax"
	"go.uber.org/zap/zaptest"
)

func TestUserRegistrationFromJSON(t *testing.T) {
	t.Parallel()

	log := lax.NewZapAdapter(zaptest.NewLogger(t))

	tests := map[string]struct {
		in      string
		want    *request.UserRegistration
		wantErr string
	}{
		"invalid json body": {``, nil, "invalid json body"},
		"empty body":        {`{}`, nil, "empty email"},
		"all empty":         {`{"email":"","password":""}`, nil, "empty email"},
		"invalid email":     {`{"email":"a","password":""}`, nil, "invalid email"},
		"empty password":    {`{"email":"john.doe@sixpack.com","password":""}`, nil, "empty password"},
		"weak password":     {`{"email":"john.doe@sixpack.com","password":"pass"}`, nil, "weak password"},
		"ok": {
			`{"email":"john.doe@sixpack.com","password":"h+z67{GxLSL~]Cl(I88AqV7w"}`,
			&request.UserRegistration{Email: "john.doe@sixpack.com", Password: "h+z67{GxLSL~]Cl(I88AqV7w"}, "", //nolint:exhaustruct,lll
		},
	}

	for n, test := range tests { //nolint:paralleltest
		test := test

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			buf := bytes.NewBufferString(test.in)

			got, gotErr := request.UserRegistrationFromJSON(buf, log)
			if test.wantErr != "" {
				if gotErr == nil {
					t.Fatalf("UserRegistrationFromJSON(%q) = error nil; want error %q", test.in, test.wantErr)
				}

				if gotErr.Error() != test.wantErr {
					t.Fatalf("UserRegistrationFromJSON(%q) = error %q; want error %q", test.in, test.wantErr, gotErr)
				}

				return
			}

			if got.Email != test.want.Email || got.Password != test.want.Password {
				t.Errorf("UserRegistrationFromJSON(%q) = %v; want %v", test.in, got, test.want)
			}

			domainUser := &domain.User{ //nolint:exhaustruct
				Email:    got.Email,
				Password: got.HashedPassword,
			}

			if !domainUser.IsValidPassword(test.want.Password) {
				t.Errorf("IsValidPassword(%q) = true; got false", test.want.Password)
			}
		})
	}
}
