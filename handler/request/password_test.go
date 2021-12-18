package request_test

import (
	"bytes"
	"testing"

	"go.ectobit.com/arc/handler/request"
	"go.ectobit.com/lax"
	"go.uber.org/zap/zaptest"
)

func TestPasswordFromJSON(t *testing.T) {
	t.Parallel()

	log := lax.NewZapAdapter(zaptest.NewLogger(t))

	tests := map[string]struct {
		in      string
		want    *request.Password
		wantErr string
	}{
		"invalid json body": {``, nil, "invalid json body"},
		"empty body":        {`{}`, nil, "empty password"},
		"empty password":    {`{"password":""}`, nil, "empty password"},
		"ok": {
			`{"password":"h+z67{GxLSL~]Cl(I88AqV7w"}`,
			&request.Password{Password: "h+z67{GxLSL~]Cl(I88AqV7w"}, "",
		},
	}

	for n, test := range tests { //nolint:paralleltest
		test := test

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			buf := bytes.NewBufferString(test.in)

			got, gotErr := request.PasswordFromJSON(buf, log)
			if test.wantErr != "" {
				if gotErr == nil {
					t.Fatalf("PasswordFromJSON(%q) = error nil; want error %q", test.in, test.wantErr)
				}

				if gotErr.Error() != test.wantErr {
					t.Fatalf("PasswordFromJSON(%q) = error %q; want error %q", test.in, gotErr, test.wantErr)
				}

				return
			}

			if got.Password != test.want.Password {
				t.Errorf("PasswordFromJSON(%q) = %v; want %v", test.in, got, test.want)
			}
		})
	}
}
