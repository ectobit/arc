package handler_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.ectobit.com/arc/domain"
	"go.ectobit.com/arc/handler"
	"go.ectobit.com/arc/handler/token"
	"go.ectobit.com/arc/repository"
	"go.ectobit.com/arc/send"
	"go.ectobit.com/lax"
	"go.uber.org/zap/zaptest"
)

func TestRegister(t *testing.T) { //nolint:funlen
	t.Parallel()

	jwt, err := token.NewJWT("test", "test", time.Hour, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	log := lax.NewZapAdapter(zaptest.NewLogger(t))
	usersHandler := handler.NewUsersHandler(&usersRepositoryFake{}, jwt, &send.Fake{}, "", "", log)
	server := httptest.NewServer(http.HandlerFunc(usersHandler.Register))

	tests := map[string]struct {
		in         string
		wantStatus int
		wantBody   string
	}{
		"invalid json body": {"", http.StatusBadRequest, `{"error":"invalid json body"}`},
		"empty body":        {`{}`, http.StatusBadRequest, `{"error":"empty email"}`},
		"all empty":         {`{"email":"","password":""}`, http.StatusBadRequest, `{"error":"empty email"}`},
		"invalid email":     {`{"email":"a","password":""}`, http.StatusBadRequest, `{"error":"invalid email"}`},
		"empty password":    {`{"email":"john.doe@sixpack.com","password":""}`, http.StatusBadRequest, `{"error":"empty password"}`},    //nolint:lll
		"weak password":     {`{"email":"john.doe@sixpack.com","password":"pass"}`, http.StatusBadRequest, `{"error":"weak password"}`}, //nolint:lll
		"ok": {
			`{"email":"john.doe@sixpack.com","password":"h+z67{GxLSL~]Cl(I88AqV7w"}`,
			http.StatusCreated, "",
		},
	}

	for n, test := range tests { //nolint:paralleltest
		test := test

		t.Run(n, func(t *testing.T) {
			t.Parallel()

			buf := bytes.NewBufferString(test.in)

			gotRes, gotErr := http.DefaultClient.Post(server.URL, "application/json", buf) //nolint:noctx
			if gotErr != nil {
				t.Fatal(gotErr)
			}

			defer func() {
				err := gotRes.Body.Close()
				if err != nil {
					t.Fatal(err)
				}
			}()

			if gotRes.StatusCode != test.wantStatus {
				t.Fatalf("Post() = status %d; want status %d", gotRes.StatusCode, test.wantStatus)
			}

			gotBody, gotErr := io.ReadAll(gotRes.Body)
			if gotErr != nil {
				t.Fatal(gotErr)
			}

			if test.wantBody != "" && string(gotBody) != test.wantBody {
				t.Errorf("ReadAll() = %q; want %q", string(gotBody), test.wantBody)
			}
		})
	}
}

var _ repository.Users = (*usersRepositoryFake)(nil)

type usersRepositoryFake struct{}

func (repo *usersRepositoryFake) Create(ctx context.Context, email string, password []byte) (*domain.User, error) {
	return &domain.User{}, nil //nolint:exhaustruct
}

func (repo *usersRepositoryFake) FindOne(ctx context.Context, id string) (*domain.User, error) {
	panic("unimplemented")
}

func (repo *usersRepositoryFake) FindOneByEmail(ctx context.Context, email string) (*domain.User, error) {
	panic("unimplemented")
}

func (repo *usersRepositoryFake) Activate(ctx context.Context, token string) (*domain.User, error) {
	panic("unimplemented")
}

func (repo *usersRepositoryFake) FetchRecoveryToken(ctx context.Context, email string) (*domain.User, error) {
	panic("unimplemented")
}

func (repo *usersRepositoryFake) ResetPassword(ctx context.Context, recoveryToken string,
	password []byte,
) (*domain.User, error) {
	panic("unimplemented")
}
