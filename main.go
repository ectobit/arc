package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.ectobit.com/act"
	"go.ectobit.com/arc/docs"

	// _ "go.ectobit.com/arc/docs"
	"go.ectobit.com/arc/handler"
	"go.ectobit.com/arc/handler/render"
	"go.ectobit.com/arc/handler/token"
	"go.ectobit.com/arc/mw"
	"go.ectobit.com/arc/repository/postgres"
	"go.ectobit.com/arc/send/smtp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type config struct {
	Port            uint          `def:"3000"`
	ShutdownTimeout time.Duration `def:"30s"`
	DSN             string        `def:"postgres://postgres:arc@postgres:5432/arc?sslmode=disable&pool_max_conns=10"` //nolint:lll
	JWT             struct {
		Issuer          string `def:"arc"`
		Secret          string
		AuthTokenExp    time.Duration `def:"15m"`
		RefreshTokenExp time.Duration `def:"168h"`
	}
	SMTP struct {
		Host     string
		Port     uint `def:"25"`
		Username string
		Password string
		Sender   string
	}
	ExtBaseURL string `help:"external server base url" def:"http://localhost:3000"`
	Log        struct {
		Format string `help:"log format [console|json]" def:"console"`
		Level  string `def:"debug"`
	}
}

// @title Arc
// @description REST API providing user accounting and authentication

// @license.name BSD-2-Clause-Patent
// @license.url https://github.com/ectobit/arc/blob/main/LICENSE
func main() { //nolint:funlen
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	cfg := &config{} //nolint:exhaustivestruct

	cli := act.New("arc")

	if err := cli.Parse(cfg, os.Args[1:]); err != nil {
		exit("parsing flags", err)
	}

	docs.SwaggerInfo.Version = "0.1"
	docs.SwaggerInfo.Host = cfg.ExtBaseURL
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http"}

	log := mustCreateLogger(cfg.Log.Format, cfg.Log.Level)

	mux := chi.NewRouter()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.Heartbeat("/health"))
	mux.Use(mw.ZapLogger(log))
	mux.Use(middleware.Recoverer)

	pool, err := postgres.Connect(context.TODO(), cfg.DSN, log, cfg.Log.Level)
	if err != nil {
		exit("postgres", err)
	}

	jwt, err := token.NewJWT(cfg.JWT.Issuer, cfg.JWT.Secret, cfg.JWT.AuthTokenExp, cfg.JWT.RefreshTokenExp)
	if err != nil {
		exit("jwt token", err)
	}

	render := render.NewJSON(log)
	usersRepository := postgres.NewUserRepository(pool, log)
	mailer := smtp.NewMailer(cfg.SMTP.Host, uint16(cfg.SMTP.Port), cfg.SMTP.Username, cfg.SMTP.Password,
		cfg.SMTP.Sender, log)
	usersHandler := handler.NewUsersHandler(render, usersRepository, jwt, mailer, cfg.ExtBaseURL, log)

	mux.Get("/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("%s/doc.json", cfg.ExtBaseURL)),
	))
	mux.Post("/users", usersHandler.Register)
	mux.Post("/users/login", usersHandler.Login)
	mux.Get("/users/activate/{token}", usersHandler.Activate)
	mux.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(jwt.JWTAuth()))
		r.Use(jwtauth.Authenticator)
	})

	server := &http.Server{Addr: fmt.Sprintf(":%d", cfg.Port), Handler: mux} //nolint:exhaustivestruct

	go func() {
		log.Info("listening", zap.Uint("port", cfg.Port))

		if err := server.ListenAndServe(); err != nil {
			log.Warn("serve", zap.Error(err))
		}
	}()

	<-signals
	log.Info("graceful shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err = server.Shutdown(ctx); err != nil {
		exit("server shutdown", err)
	}

	pool.Close()

	_ = log.Sync()
}

func mustCreateLogger(logFormat, logLevel string) *zap.Logger {
	level := zap.NewAtomicLevel()

	encodeLevel := zapcore.LowercaseLevelEncoder
	if logLevel == "debug" {
		encodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config := zap.Config{ //nolint:exhaustivestruct
		Level:       level,
		Development: logLevel == "debug",
		Encoding:    logFormat,
		EncoderConfig: zapcore.EncoderConfig{ //nolint:exhaustivestruct
			CallerKey:      "caller",
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeLevel:    encodeLevel,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			LevelKey:       "level",
			MessageKey:     "msg",
			NameKey:        "logger",
			StacktraceKey:  "stack",
		},

		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build(zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		exit("failed building log config", err)
	}

	return logger
}

func exit(message string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v", message, err)
	os.Exit(1)
}
