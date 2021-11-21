package handler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ectobit/arc/handler"
	"github.com/ectobit/arc/handler/render"
	"github.com/ectobit/arc/handler/token"
	"github.com/ectobit/arc/repository/postgres"
	"go.uber.org/zap/zaptest"
)

func TestRegister(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	databaseName := os.Getenv("ARC_DB_HOST")
	if databaseName == "" {
		t.Error("environment variable ARC_DB_HOST not set")
	}

	ctx := context.TODO()

	log := zaptest.NewLogger(t)

	render := render.NewJSON(log)

	conn, err := postgres.Connect(ctx, fmt.Sprintf("postgres://postgres:arc@%s/test?sslmode=disable", databaseName), log)
	if err != nil {
		t.Error(err)
	}

	defer conn.Close()

	usersRepository := postgres.NewUserRepository(conn, log)

	jwt, err := token.NewJWT("test", "test", time.Hour, time.Hour)
	if err != nil {
		t.Error(err)
	}

	usersHandler := handler.NewUsersHandler(render, usersRepository, jwt, nil, "", log)

	server := httptest.NewServer(http.HandlerFunc(usersHandler.Register))

	res, err := http.DefaultClient.Get(server.URL) //nolint:noctx
	if err != nil {
		t.Error(err)
	}

	defer func() {
		err := res.Body.Close()
		if err != nil {
			t.Error(err)
		}
	}()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("want %d, got %d", http.StatusBadRequest, res.StatusCode)
	}
}
