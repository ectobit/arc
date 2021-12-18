package public_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.ectobit.com/arc/domain"
	"go.ectobit.com/arc/handler/public"
	"go.ectobit.com/lax"
	"go.uber.org/zap/zaptest"
)

func TestUserRegistrationFromJSON(t *testing.T) {
	t.Parallel()

	log := lax.NewZapAdapter(zaptest.NewLogger(t))

	tests := map[string]struct {
		in      string
		want    *public.UserRegistration
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
			&public.UserRegistration{Email: "john.doe@sixpack.com", Password: "h+z67{GxLSL~]Cl(I88AqV7w"}, "", //nolint:exhaustivestruct,lll
		},
	}

	for n, test := range tests { //nolint:paralleltest
		test := test

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			buf := bytes.NewBufferString(test.in)

			got, gotErr := public.UserRegistrationFromJSON(buf, log)
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

			domainUser := &domain.User{ //nolint:exhaustivestruct
				Email:    got.Email,
				Password: got.HashedPassword,
			}

			if !domainUser.IsValidPassword(test.want.Password) {
				t.Errorf("IsValidPassword(%q) = true; got false", test.want.Password)
			}
		})
	}
}

func TestUserLoginFromJSON(t *testing.T) {
	t.Parallel()

	log := lax.NewZapAdapter(zaptest.NewLogger(t))

	tests := map[string]struct {
		in      string
		want    *public.UserLogin
		wantErr string
	}{
		"invalid json body": {``, nil, "invalid json body"},
		"empty body":        {`{}`, nil, "empty email"},
		"all empty":         {`{"email":"","password":""}`, nil, "empty email"},
		"invalid email":     {`{"email":"a","password":""}`, nil, "invalid email"},
		"empty password":    {`{"email":"john.doe@sixpack.com","password":""}`, nil, "empty password"},
		"ok": {
			`{"email":"john.doe@sixpack.com","password":"h+z67{GxLSL~]Cl(I88AqV7w"}`,
			&public.UserLogin{Email: "john.doe@sixpack.com", Password: "h+z67{GxLSL~]Cl(I88AqV7w"}, "",
		},
	}

	for n, test := range tests { //nolint:paralleltest
		test := test

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			buf := bytes.NewBufferString(test.in)

			got, gotErr := public.UserLoginFromJSON(buf, log)
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

func TestEmailFromJSON(t *testing.T) {
	t.Parallel()

	log := lax.NewZapAdapter(zaptest.NewLogger(t))

	tests := map[string]struct {
		in      string
		want    *public.Email
		wantErr string
	}{
		"invalid json body": {``, nil, "invalid json body"},
		"empty body":        {`{}`, nil, "empty email"},
		"all empty":         {`{"email":""}`, nil, "empty email"},
		"invalid email":     {`{"email":"a"}`, nil, "invalid email"},
		"ok": {
			`{"email":"john.doe@sixpack.com"}`,
			&public.Email{Email: "john.doe@sixpack.com"}, "",
		},
	}

	for n, test := range tests { //nolint:paralleltest
		test := test

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			buf := bytes.NewBufferString(test.in)

			got, gotErr := public.EmailFromJSON(buf, log)
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

func TestFromDomainUser(t *testing.T) {
	t.Parallel()

	active := true
	now := time.Now()

	domainUser := &domain.User{
		ID:                 "926c7bed-18a7-4c0f-97fd-f5901b2c52ba",
		Email:              "john.doe@sixpack.com",
		Password:           []byte{},
		Created:            &now,
		Updated:            &now,
		ActivationToken:    "",
		PasswordResetToken: "",
		Active:             &active,
	}

	wantPublicUser := &public.User{
		ID:           domainUser.ID,
		Email:        domainUser.Email,
		Created:      &now,
		Updated:      &now,
		AuthToken:    "",
		RefreshToken: "",
	}

	gotPublicUser := public.FromDomainUser(domainUser)

	if diff := cmp.Diff(wantPublicUser, gotPublicUser); diff != "" {
		t.Errorf("FromDomainUser() mismatch (-want +got):\n%s", diff)
	}
}
