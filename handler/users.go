package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.ectobit.com/arc/handler/public"
	"go.ectobit.com/arc/handler/render"
	"go.ectobit.com/arc/handler/token"
	"go.ectobit.com/arc/repository"
	"go.ectobit.com/arc/send"
	"go.ectobit.com/lax"
)

// UsersHandler contains user related http handlers.
type UsersHandler struct {
	r           render.Renderer
	usersRepo   repository.Users
	jwt         *token.JWT
	sender      send.Sender
	externalURL string
	log         lax.Logger
}

// NewUsersHandler creates users handler.
func NewUsersHandler(r render.Renderer, ur repository.Users, jwt *token.JWT, sender send.Sender, externalURL string,
	log lax.Logger) *UsersHandler {
	return &UsersHandler{
		r:           r,
		usersRepo:   ur,
		jwt:         jwt,
		sender:      sender,
		externalURL: externalURL,
		log:         log,
	}
}

// Register registers new users.
//
// @Tags users
// @Accept json
// @Produce json
// @Router /users [post]
// @Param user body public.UserRegistration true "User"
// @Success 201 {object} public.User
// @Failure 400 {object} render.Error
// @Failure 409 {object} render.Error
// @Failure 500
// @Summary Register user account.
func (h *UsersHandler) Register(res http.ResponseWriter, req *http.Request) {
	user, publicErr := public.UserRegistrationFromJSON(req.Body, h.log)
	if publicErr != nil {
		h.r.Error(res, publicErr.StatusCode, publicErr.Message)

		return
	}

	domainUser, err := h.usersRepo.Create(req.Context(), user.Email, user.HashedPassword)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateKey) {
			h.r.Error(res, http.StatusConflict, "already registered")

			return
		}

		h.log.Warn("create user", lax.Error(err))
		h.r.Render(res, http.StatusInternalServerError, nil)

		return
	}

	message := fmt.Sprintf("%s/users/activate/%s", h.externalURL, domainUser.ActivationToken)

	if err = h.sender.Send(domainUser.Email, "Account activation", message); err != nil {
		h.log.Warn("send activation link", lax.Error(err))

		h.r.Render(res, http.StatusInternalServerError, nil)

		return
	}

	publicUser := public.FromDomainUser(domainUser)
	publicUser.ID = ""

	h.r.Render(res, http.StatusCreated, publicUser)
}

// Activate activates user account.
//
// @Tags users
// @Accept json
// @Produce json
// @Router /users/activate/{token} [get]
// @Param token path string true "Activation token"
// @Success 200 {object} public.User
// @Failure 400 {object} render.Error
// @Failure 500
// @Summary Activate user account.
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

		h.log.Warn("activate user account", lax.Error(err))
		h.r.Render(res, http.StatusInternalServerError, nil)

		return
	}

	publicUser := public.FromDomainUser(user)
	publicUser.ID = ""

	h.r.Render(res, http.StatusOK, publicUser)
}

// Login logins user.
//
// @Tags users
// @Accept json
// @Produce json
// @Router /users/login [post]
// @Param user body public.UserLogin true "User"
// @Success 201 {object} public.User
// @Failure 400 {object} render.Error
// @Failure 401 {object} render.Error
// @Failure 500
// @Summary Login.
func (h *UsersHandler) Login(res http.ResponseWriter, req *http.Request) {
	user, publicErr := public.UserLoginFromJSON(req.Body, h.log)
	if publicErr != nil {
		h.r.Error(res, publicErr.StatusCode, publicErr.Message)

		return
	}

	domainUser, err := h.usersRepo.FindByEmail(req.Context(), user.Email)
	if err != nil {
		h.log.Warn("find user by email", lax.Error(err))
		h.r.Render(res, http.StatusInternalServerError, nil)

		return
	}

	if !domainUser.IsValidPassword(user.Password) {
		h.r.Render(res, http.StatusUnauthorized, nil)

		return
	}

	if !domainUser.IsActive() {
		h.r.Error(res, http.StatusUnauthorized, "account not activated")

		return
	}

	publicUser := public.FromDomainUser(domainUser)
	requestID := middleware.GetReqID(req.Context())

	if publicUser.AuthToken, publicUser.RefreshToken, err = h.jwt.Tokens(publicUser.ID, requestID); err != nil {
		h.log.Warn("tokens", lax.Error(err))
		h.r.Render(res, http.StatusInternalServerError, nil)

		return
	}

	h.r.Render(res, http.StatusOK, publicUser)
}
