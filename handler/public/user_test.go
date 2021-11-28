package public_test

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"go.ectobit.com/arc/domain"
	"go.ectobit.com/arc/handler/public"
	"go.uber.org/zap/zaptest"
)

func TestUserRegistrationFromJSON(t *testing.T) {
	t.Parallel()

	log := zaptest.NewLogger(t)

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
					t.Errorf("want error %q, got no error", test.wantErr)

					return
				}

				if gotErr.Error() != test.wantErr {
					t.Errorf("want error %q, got error %q", test.wantErr, gotErr)

					return
				}

				return
			}

			if got.Email != test.want.Email || got.Password != test.want.Password {
				t.Errorf("\nwant %v,\n got %v", test.want, got)
			}

			domainUser := &domain.User{ //nolint:exhaustivestruct
				Email:    got.Email,
				Password: got.HashedPassword,
			}

			if !domainUser.IsValidPassword(test.want.Password) {
				t.Errorf("invalid password")
			}
		})
	}
}

func TestUserLoginFromJSON(t *testing.T) {
	t.Parallel()

	log := zaptest.NewLogger(t)

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
					t.Errorf("want error %q, got no error", test.wantErr)

					return
				}

				if gotErr.Error() != test.wantErr {
					t.Errorf("want error %q, got error %q", test.wantErr, gotErr)

					return
				}

				return
			}

			if got.Email != test.want.Email || got.Password != test.want.Password {
				t.Errorf("\nwant %v,\ngot  %v", test.want, got)
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
	if !reflect.DeepEqual(gotPublicUser, wantPublicUser) {
		t.Errorf("\nwant %v,\ngot  %v", wantPublicUser, gotPublicUser)
	}
}
