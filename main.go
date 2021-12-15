package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/unrolled/secure"
	"go.ectobit.com/act"
	"go.ectobit.com/arc/docs"
	"go.ectobit.com/arc/handler"
	"go.ectobit.com/arc/handler/render"
	"go.ectobit.com/arc/handler/token"
	"go.ectobit.com/arc/mw"
	"go.ectobit.com/arc/repository/postgres"
	"go.ectobit.com/arc/send/smtp"
	"go.ectobit.com/lax"
)

type config struct {
	Development     bool
	Port            uint          `def:"3000"`
	ShutdownTimeout time.Duration `def:"10s"`
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
	ExternalURL               act.URL `help:"external server base url" def:"http://localhost:3000"`
	FrontendPasswordResetPath string  `def:"frontend-password-reset-path"`
	Log                       struct {
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

	docs.SwaggerInfo.Version = version
	docs.SwaggerInfo.Host = cfg.ExternalURL.Host
	docs.SwaggerInfo.BasePath = cfg.ExternalURL.Path
	docs.SwaggerInfo.Schemes = []string{cfg.ExternalURL.Scheme}

	log := mustCreateLogger(cfg.Log.Format, cfg.Log.Level)

	mux := chi.NewRouter()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.Heartbeat("/health"))
	mux.Use(mw.ZapLogger(log))
	mux.Use(middleware.Recoverer)
	mux.Use(hsts(cfg.Development, cfg.ExternalURL.URL).Handler)

	pool, err := postgres.Connect(context.Background(), cfg.DSN, log, cfg.Log.Level)
	if err != nil {
		exit("postgres", err)
	}

	jwt, err := token.NewJWT(cfg.JWT.Issuer, cfg.JWT.Secret, cfg.JWT.AuthTokenExp, cfg.JWT.RefreshTokenExp)
	if err != nil {
		exit("jwt token", err)
	}

	render := render.NewJSON(log)
	usersRepository := postgres.NewUserRepository(pool)
	mailer := smtp.NewMailer(cfg.SMTP.Host, uint16(cfg.SMTP.Port), cfg.SMTP.Username, cfg.SMTP.Password,
		cfg.SMTP.Sender, log)
	usersHandler := handler.NewUsersHandler(render, usersRepository, jwt, mailer, cfg.ExternalURL.String(),
		cfg.FrontendPasswordResetPath, log)

	mux.Get("/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("%s/doc.json", cfg.ExternalURL)),
	))
	mux.Post("/users", usersHandler.Register)
	mux.Post("/users/login", usersHandler.Login)
	mux.Get("/users/activate/{token}", usersHandler.Activate)
	mux.Get("/users/password-reset/{email}", usersHandler.PasswordResetToken)
	mux.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(jwt.JWTAuth()))
		r.Use(jwtauth.Authenticator)
	})

	server := &http.Server{Addr: fmt.Sprintf(":%d", cfg.Port), Handler: mux} //nolint:exhaustivestruct

	go func() {
		log.Info("listening", lax.Uint("port", cfg.Port), lax.String("version", version),
			lax.String("revision", revision), lax.String("build date", buildDate))

		if err := server.ListenAndServe(); err != nil {
			log.Warn("serve", lax.Error(err))
		}
	}()

	<-signals
	log.Info("graceful shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err = server.Shutdown(ctx); err != nil {
		log.Warn("server shutdown", lax.Error(err))
	}

	pool.Close()

	log.Flush()
}

func mustCreateLogger(logFormat, logLevel string) *lax.ZapAdapter {
	log, err := lax.NewDefaultZapAdapter(logFormat, logLevel)
	if err != nil {
		exit("crate logger", err)
	}

	return log
}

func hsts(development bool, externalURL *url.URL) *secure.Secure {
	return secure.New(secure.Options{ //nolint:exhaustivestruct
		IsDevelopment:         development,
		AllowedHosts:          []string{externalURL.Host},
		HostsProxyHeaders:     []string{"X-Forwarded-Host"},
		SSLRedirect:           true,
		SSLHost:               externalURL.Host,
		SSLProxyHeaders:       map[string]string{"X-Forwarded-Proto": "https"},
		STSSeconds:            31536000, //nolint:gomnd
		STSPreload:            true,
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "script-src $NONCE",
	})
}

func exit(message string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", message, err)
	os.Exit(1)
}
