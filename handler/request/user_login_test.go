package request_test

import (
	"bytes"
	"testing"

	"go.ectobit.com/arc/handler/request"
	"go.ectobit.com/lax"
	"go.uber.org/zap/zaptest"
)

func TestUserLoginFromJSON(t *testing.T) {
	t.Parallel()

	log := lax.NewZapAdapter(zaptest.NewLogger(t))

	tests := map[string]struct {
		in      string
		want    *request.UserLogin
		wantErr string
	}{
		"invalid json body": {``, nil, "invalid json body"},
		"empty body":        {`{}`, nil, "empty email"},
		"all empty":         {`{"email":"","password":""}`, nil, "empty email"},
		"invalid email":     {`{"email":"a","password":""}`, nil, "invalid email"},
		"empty password":    {`{"email":"john.doe@sixpack.com","password":""}`, nil, "empty password"},
		"ok": {
			`{"email":"john.doe@sixpack.com","password":"h+z67{GxLSL~]Cl(I88AqV7w"}`,
			&request.UserLogin{Email: "john.doe@sixpack.com", Password: "h+z67{GxLSL~]Cl(I88AqV7w"}, "",
		},
	}

	for n, test := range tests { //nolint:paralleltest
		test := test

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			buf := bytes.NewBufferString(test.in)

			got, gotErr := request.UserLoginFromJSON(buf, log)
			if test.wantErr != "" {
				if gotErr == nil {
					t.Fatalf("UserLoginFromJSON(%q) = error nil; want error %q", test.in, test.wantErr)
				}

				if gotErr.Error() != test.wantErr {
					t.Fatalf("UserLoginFromJSON(%q) = error %q; want error %q", test.in, gotErr, test.wantErr)
				}

				return
			}

			if got.Email != test.want.Email || got.Password != test.want.Password {
				t.Errorf("UserLoginFromJSON(%q) = %v; want %v", test.in, got, test.want)
			}
		})
	}
}
