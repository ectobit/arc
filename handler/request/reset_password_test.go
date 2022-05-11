package request_test

import (
	"bytes"
	"testing"

	"go.ectobit.com/arc/domain"
	"go.ectobit.com/arc/handler/request"
	"go.ectobit.com/lax"
	"go.uber.org/zap/zaptest"
)

func TestResetPasswordFromJSON(t *testing.T) {
	t.Parallel()

	log := lax.NewZapAdapter(zaptest.NewLogger(t))

	tests := map[string]struct {
		in      string
		want    *request.ResetPassword
		wantErr string
	}{
		"invalid json body": {``, nil, "invalid json body"},
		"empty body":        {`{}`, nil, "empty password reset token"},
		"all empty":         {`{"recoveryToken":"","password":""}`, nil, "empty password reset token"},
		"empty password":    {`{"recoveryToken":"test","password":""}`, nil, "empty password"},
		"weak password":     {`{"recoveryToken":"test","password":"pass"}`, nil, "weak password"},
		"ok": {
			`{"recoveryToken":"test","password":"h+z67{GxLSL~]Cl(I88AqV7w"}`,
			&request.ResetPassword{RecoveryToken: "test", Password: "h+z67{GxLSL~]Cl(I88AqV7w"}, "", //nolint:exhaustruct
		},
	}

	for n, test := range tests { //nolint:paralleltest
		test := test

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			buf := bytes.NewBufferString(test.in)

			got, gotErr := request.ResetPasswordFromJSON(buf, log)
			if test.wantErr != "" {
				if gotErr == nil {
					t.Fatalf("ResetPasswordFromJSON(%q) = error nil; want error %q", test.in, test.wantErr)
				}

				if gotErr.Error() != test.wantErr {
					t.Fatalf("ResetPasswordFromJSON(%q) = error %q; want error %q", test.in, test.wantErr, gotErr)
				}

				return
			}

			if got.RecoveryToken != test.want.RecoveryToken || got.Password != test.want.Password {
				t.Errorf("ResetPasswordFromJSON(%q) = %v; want %v", test.in, got, test.want)
			}

			domainUser := &domain.User{ //nolint:exhaustruct
				Password: got.HashedPassword,
			}

			if !domainUser.IsValidPassword(test.want.Password) {
				t.Errorf("IsValidPassword(%q) = true; got false", test.want.Password)
			}
		})
	}
}
