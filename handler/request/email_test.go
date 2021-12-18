package request_test

import (
	"bytes"
	"testing"

	"go.ectobit.com/arc/handler/request"
	"go.ectobit.com/lax"
	"go.uber.org/zap/zaptest"
)

func TestEmailFromJSON(t *testing.T) {
	t.Parallel()

	log := lax.NewZapAdapter(zaptest.NewLogger(t))

	tests := map[string]struct {
		in      string
		want    *request.Email
		wantErr string
	}{
		"invalid json body": {``, nil, "invalid json body"},
		"empty body":        {`{}`, nil, "empty email"},
		"all empty":         {`{"email":""}`, nil, "empty email"},
		"invalid email":     {`{"email":"a"}`, nil, "invalid email"},
		"ok": {
			`{"email":"john.doe@sixpack.com"}`,
			&request.Email{Email: "john.doe@sixpack.com"}, "",
		},
	}

	for n, test := range tests { //nolint:paralleltest
		test := test

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			buf := bytes.NewBufferString(test.in)

			got, gotErr := request.EmailFromJSON(buf, log)
			if test.wantErr != "" {
				if gotErr == nil {
					t.Fatalf("EmailFromJSON(%q) = error nil; want error %q", test.in, test.wantErr)
				}

				if gotErr.Error() != test.wantErr {
					t.Fatalf("EmailFromJSON(%q) = error %q; want error %q", test.in, gotErr, test.wantErr)
				}

				return
			}

			if got.Email != test.want.Email {
				t.Errorf("EmailFromJSON(%q) = %v; want %v", test.in, got, test.want)
			}
		})
	}
}
