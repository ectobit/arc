package handler_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"go.ectobit.com/arc/handler"
	"go.ectobit.com/arc/handler/render"
	"go.ectobit.com/arc/handler/token"
	"go.ectobit.com/arc/repository/postgres"
	"go.ectobit.com/arc/send"
	"go.ectobit.com/lax"
	"go.uber.org/zap/zaptest"
)

func TestRegister(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip()
	}

	usersHandler := setup(t)
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
				t.Error(gotErr)
			}

			defer func() {
				err := gotRes.Body.Close()
				if err != nil {
					t.Error(err)
				}
			}()

			if gotRes.StatusCode != test.wantStatus {
				t.Errorf("want status %d, got status %d", test.wantStatus, gotRes.StatusCode)
			}

			gotBody, gotErr := io.ReadAll(gotRes.Body)
			if gotErr != nil {
				t.Error(gotErr)
			}

			if test.wantBody != "" && string(gotBody) != test.wantBody {
				t.Errorf("want %s, got %s", test.wantBody, string(gotBody))
			}
		})
	}
}

func setup(t *testing.T) *handler.UsersHandler {
	t.Helper()

	databaseName := os.Getenv("ARC_DB_HOST")
	if databaseName == "" {
		t.Fatal("environment variable ARC_DB_HOST not set")
	}

	ctx := context.TODO()

	log := lax.NewZapAdapter(zaptest.NewLogger(t))

	render := render.NewJSON(log)

	conn, err := postgres.Connect(ctx, fmt.Sprintf("postgres://postgres:arc@%s/test?sslmode=disable", databaseName),
		log, "debug")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(conn.Close)

	if _, err := conn.Exec(context.TODO(), "TRUNCATE users"); err != nil {
		t.Error(err)
	}

	usersRepository := postgres.NewUserRepository(conn)

	jwt, err := token.NewJWT("test", "test", time.Hour, time.Hour)
	if err != nil {
		t.Error(err)
	}

	return handler.NewUsersHandler(render, usersRepository, jwt, &send.Fake{}, "", "", log)
}
