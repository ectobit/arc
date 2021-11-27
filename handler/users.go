package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ectobit/arc/handler/public"
	"github.com/ectobit/arc/handler/render"
	"github.com/ectobit/arc/handler/token"
	"github.com/ectobit/arc/repository"
	"github.com/ectobit/arc/send"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// UsersHandler contains user related http handlers.
type UsersHandler struct {
	r         render.Renderer
	usersRepo repository.Users
	jwt       *token.JWT
	sender    send.Sender
	baseURL   string
	log       *zap.Logger
}

// NewUsersHandler creates users handler.
func NewUsersHandler(r render.Renderer, ur repository.Users, jwt *token.JWT, sender send.Sender, baseURL string,
	log *zap.Logger) *UsersHandler {
	return &UsersHandler{
		r:         r,
		usersRepo: ur,
		jwt:       jwt,
		sender:    sender,
		baseURL:   baseURL,
		log:       log,
	}
}

// Register registers new users.
func (h *UsersHandler) Register(res http.ResponseWriter, req *http.Request) {
	user, publicErr := public.UserRegistrationFromJSON(req.Body, h.log)
	if publicErr != nil {
		h.r.Error(res, publicErr.StatusCode, publicErr.Message)

		return
	}

	domainUser, err := h.usersRepo.Create(req.Context(), user.Email, user.HashedPassword)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateKey) {
			h.r.Error(res, http.StatusBadRequest, "already registered")

			return
		}

		h.log.Warn("create user", zap.Error(err))
		h.r.Render(res, http.StatusInternalServerError, nil)

		return
	}

	message := fmt.Sprintf("%s/users/activate/%s", h.baseURL, domainUser.ActivationToken)

	if err = h.sender.Send(domainUser.Email, "Account activation", message); err != nil {
		h.log.Warn("send activation link", zap.Error(err))

		h.r.Render(res, http.StatusInternalServerError, nil)

		return
	}

	publicUser := public.FromDomainUser(domainUser)
	publicUser.ID = ""

	h.r.Render(res, http.StatusCreated, publicUser)
}

// Activate activates user account.
func (h *UsersHandler) Activate(res http.ResponseWriter, req *http.Request) {
	token := chi.URLParam(req, "token")
	if token == "" {
		h.r.Render(res, http.StatusBadRequest, nil)

		return
	}

	user, err := h.usersRepo.Activate(req.Context(), token)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidActivation) {
			h.r.Error(res, http.StatusBadRequest, err.Error())

			return
		}

		h.log.Warn("activate user account", zap.Error(err))
		h.r.Render(res, http.StatusInternalServerError, nil)

		return
	}

	publicUser := public.FromDomainUser(user)
	publicUser.ID = ""

	h.r.Render(res, http.StatusOK, publicUser)
}

// Login logins user.
func (h *UsersHandler) Login(res http.ResponseWriter, req *http.Request) {
	login, err := public.UserLoginFromJSON(req.Body)
	if err != nil {
		h.log.Warn("parse user login", zap.Error(err))
		h.r.Render(res, http.StatusBadRequest, nil)

		return
	}

	user, err := h.usersRepo.FindByEmail(req.Context(), login.Email)
	if err != nil {
		h.log.Warn("fetch user by email", zap.Error(err))

		h.r.Render(res, http.StatusInternalServerError, nil)

		return
	}

	if !user.IsValidPassword(login.Password) {
		h.r.Render(res, http.StatusUnauthorized, nil)

		return
	}

	if !user.IsActive() {
		h.r.Error(res, http.StatusUnauthorized, "account not activated")

		return
	}

	publicUser := public.FromDomainUser(user)
	requestID := middleware.GetReqID(req.Context())

	if publicUser.AuthToken, publicUser.RefreshToken, err = h.jwt.Tokens(publicUser.ID, requestID); err != nil {
		h.log.Warn("tokens", zap.Error(err))

		h.r.Render(res, http.StatusInternalServerError, nil)

		return
	}

	h.r.Render(res, http.StatusCreated, publicUser)
}
