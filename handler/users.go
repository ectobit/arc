package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.ectobit.com/arc/handler/public"
	"go.ectobit.com/arc/handler/render"
	"go.ectobit.com/arc/handler/request"
	"go.ectobit.com/arc/handler/response"
	"go.ectobit.com/arc/handler/token"
	"go.ectobit.com/arc/repository"
	"go.ectobit.com/arc/send"
	"go.ectobit.com/lax"
)

// UsersHandler contains user related http handlers.
type UsersHandler struct {
	r                         render.Renderer
	usersRepo                 repository.Users
	jwt                       *token.JWT
	sender                    send.Sender
	externalURL               string
	frontendPasswordResetPath string
	log                       lax.Logger
}

// NewUsersHandler creates users handler.
func NewUsersHandler(r render.Renderer, ur repository.Users, jwt *token.JWT, sender send.Sender, externalURL string,
	frontendPasswordResetPath string, log lax.Logger) *UsersHandler {
	return &UsersHandler{
		r:                         r,
		usersRepo:                 ur,
		jwt:                       jwt,
		sender:                    sender,
		externalURL:               externalURL,
		frontendPasswordResetPath: frontendPasswordResetPath,
		log:                       log,
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
	userRegistration, publicErr := public.UserRegistrationFromJSON(req.Body, h.log)
	if publicErr != nil {
		h.r.Error(res, publicErr.StatusCode, publicErr.Message)

		return
	}

	user, err := h.usersRepo.Create(req.Context(), userRegistration.Email, userRegistration.HashedPassword)
	if err != nil {
		if errors.Is(err, repository.ErrUniqueViolation) {
			h.r.Error(res, http.StatusConflict, "already registered")

			return
		}

		h.log.Warn("create user", lax.Error(err))
		h.r.Render(res, http.StatusInternalServerError, nil)

		return
	}

	message := fmt.Sprintf("%s/users/activate/%s", h.externalURL, user.ActivationToken)

	if err = h.sender.Send(user.Email, "Account activation", message); err != nil {
		h.log.Warn("send activation link", lax.Error(err))

		h.r.Render(res, http.StatusInternalServerError, nil)

		return
	}

	publicUser := public.FromDomainUser(user)
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
// @Failure 404 {object} render.Error
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
		if errors.Is(err, repository.ErrResourceNotFound) {
			h.r.Error(res, http.StatusNotFound, err.Error())

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
// @Failure 404 {object} render.Error
// @Failure 500
// @Summary Login.
func (h *UsersHandler) Login(res http.ResponseWriter, req *http.Request) {
	userLogin, publicErr := public.UserLoginFromJSON(req.Body, h.log)
	if publicErr != nil {
		h.r.Error(res, publicErr.StatusCode, publicErr.Message)

		return
	}

	user, err := h.usersRepo.FindOneByEmail(req.Context(), userLogin.Email)
	if err != nil {
		if errors.Is(err, repository.ErrResourceNotFound) {
			h.r.Error(res, http.StatusNotFound, err.Error())

			return
		}

		h.log.Warn("find user by email", lax.Error(err))
		h.r.Render(res, http.StatusInternalServerError, nil)

		return
	}

	if !user.IsValidPassword(userLogin.Password) {
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
		h.log.Warn("tokens", lax.Error(err))
		h.r.Render(res, http.StatusInternalServerError, nil)

		return
	}

	h.r.Render(res, http.StatusOK, publicUser)
}

// RequestPasswordReset requests password reset.
//
// @Tags users
// @Accept json
// @Produce json
// @Router /users/reset-password [post]
// @Param email body public.Email true "E-mail address"
// @Success 202
// @Failure 400 {object} render.Error
// @Failure 404 {object} render.Error
// @Failure 500
// @Summary Request password reset.
func (h *UsersHandler) RequestPasswordReset(res http.ResponseWriter, req *http.Request) {
	email, publicErr := public.EmailFromJSON(req.Body, h.log)
	if publicErr != nil {
		h.r.Error(res, publicErr.StatusCode, publicErr.Message)

		return
	}

	user, err := h.usersRepo.FetchPasswordResetToken(req.Context(), email.Email)
	if err != nil {
		if errors.Is(err, repository.ErrResourceNotFound) {
			h.r.Error(res, http.StatusNotFound, err.Error())

			return
		}

		h.log.Warn("password reset token", lax.Error(err))
		h.r.Render(res, http.StatusInternalServerError, nil)

		return
	}

	message := fmt.Sprintf("%s/%s/%s", h.externalURL, h.frontendPasswordResetPath, user.PasswordResetToken)

	if err = h.sender.Send(user.Email, "Password reset request", message); err != nil {
		h.log.Warn("send password reset token", lax.Error(err))

		h.r.Render(res, http.StatusInternalServerError, nil)

		return
	}

	h.r.Render(res, http.StatusAccepted, nil)
}

// CheckPasswordStrength calculates password strength.
//
// @Tags users
// @Accept json
// @Produce json
// @Router /users/check-password [post]
// @Param password body public.Password true "Password"
// @Success 200
// @Failure 400 {object} render.Error
// @Summary Calculate password strength.
func (h *UsersHandler) CheckPasswordStrength(res http.ResponseWriter, req *http.Request) {
	password, publicErr := public.PasswordFromJSON(req.Body, h.log)
	if publicErr != nil {
		h.r.Error(res, publicErr.StatusCode, publicErr.Message)

		return
	}

	h.r.Render(res, http.StatusOK, password.Strength())
}

// ResetPassword sets new user's password.
//
// @Tags users
// @Accept json
// @Produce json
// @Router /users/reset-password [patch]
// @Param resetPassword body public.ResetPassword true "Password reset token and new password"
// @Success 200 {object} public.User
// @Failure 400 {object} render.Error
// @Failure 404 {object} render.Error
// @Failure 500
// @Summary Set new user's password.
func (h *UsersHandler) ResetPassword(res http.ResponseWriter, req *http.Request) {
	resetPassword, publicErr := public.ResetPasswordFromJSON(req.Body, h.log)
	if publicErr != nil {
		h.r.Error(res, publicErr.StatusCode, publicErr.Message)

		return
	}

	user, err := h.usersRepo.ResetPassword(req.Context(), resetPassword.PasswordResetToken, resetPassword.HashedPassword)
	if err != nil {
		if errors.Is(err, repository.ErrResourceNotFound) {
			h.r.Error(res, http.StatusNotFound, err.Error())

			return
		}

		h.log.Warn("reset password", lax.Error(err))
		h.r.Render(res, http.StatusInternalServerError, nil)

		return
	}

	publicUser := public.FromDomainUser(user)
	publicUser.ID = ""

	h.r.Render(res, http.StatusAccepted, publicUser)
}

func (h *UsersHandler) RefreshToken(res http.ResponseWriter, req *http.Request) {
	refreshToken, err := request.RefreshTokenFromBody(req.Body, h.log)
	if err != nil {
		response.RenderError(res, err, h.log)

		return
	}

	_ = refreshToken
}
