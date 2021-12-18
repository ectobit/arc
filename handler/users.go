package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.ectobit.com/arc/handler/public"
	"go.ectobit.com/arc/handler/request"
	"go.ectobit.com/arc/handler/response"
	"go.ectobit.com/arc/handler/token"
	"go.ectobit.com/arc/repository"
	"go.ectobit.com/arc/send"
	"go.ectobit.com/lax"
)

// UsersHandler contains user related http handlers.
type UsersHandler struct {
	usersRepo                 repository.Users
	jwt                       *token.JWT
	sender                    send.Sender
	externalURL               string
	frontendPasswordResetPath string
	log                       lax.Logger
}

// NewUsersHandler creates users handler.
func NewUsersHandler(ur repository.Users, jwt *token.JWT, sender send.Sender, externalURL string,
	frontendPasswordResetPath string, log lax.Logger) *UsersHandler {
	return &UsersHandler{
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
	userRegistration, err := public.UserRegistrationFromJSON(req.Body, h.log)
	if err != nil {
		response.RenderError(res, err, h.log)

		return
	}

	user, err := h.usersRepo.Create(req.Context(), userRegistration.Email, userRegistration.HashedPassword)
	if err != nil {
		if errors.Is(err, repository.ErrUniqueViolation) {
			response.RenderErrorStatus(res, http.StatusConflict, "already registered", h.log)

			return
		}

		h.log.Warn("create user", lax.Error(err))
		response.Render(res, http.StatusInternalServerError, nil, h.log)

		return
	}

	message := fmt.Sprintf("%s/users/activate/%s", h.externalURL, user.ActivationToken)

	if err = h.sender.Send(user.Email, "Account activation", message); err != nil {
		h.log.Warn("send activation link", lax.Error(err))

		response.Render(res, http.StatusInternalServerError, nil, h.log)

		return
	}

	publicUser := public.FromDomainUser(user)
	publicUser.ID = ""

	response.Render(res, http.StatusCreated, publicUser, h.log)
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
		response.Render(res, http.StatusBadRequest, nil, h.log)

		return
	}

	user, err := h.usersRepo.Activate(req.Context(), token)
	if err != nil {
		if errors.Is(err, repository.ErrResourceNotFound) {
			response.RenderErrorStatus(res, http.StatusNotFound, err.Error(), h.log)

			return
		}

		h.log.Warn("activate user account", lax.Error(err))
		response.Render(res, http.StatusInternalServerError, nil, h.log)

		return
	}

	publicUser := public.FromDomainUser(user)
	publicUser.ID = ""

	response.Render(res, http.StatusOK, publicUser, h.log)
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
	userLogin, err := public.UserLoginFromJSON(req.Body, h.log)
	if err != nil {
		response.RenderError(res, err, h.log)

		return
	}

	user, err := h.usersRepo.FindOneByEmail(req.Context(), userLogin.Email)
	if err != nil {
		if errors.Is(err, repository.ErrResourceNotFound) {
			response.RenderErrorStatus(res, http.StatusNotFound, err.Error(), h.log)

			return
		}

		h.log.Warn("find user by email", lax.Error(err))
		response.Render(res, http.StatusInternalServerError, nil, h.log)

		return
	}

	if !user.IsValidPassword(userLogin.Password) {
		response.Render(res, http.StatusUnauthorized, nil, h.log)

		return
	}

	if !user.IsActive() {
		response.RenderErrorStatus(res, http.StatusUnauthorized, "account not activated", h.log)

		return
	}

	publicUser := public.FromDomainUser(user)
	requestID := middleware.GetReqID(req.Context())

	if publicUser.AuthToken, publicUser.RefreshToken, err = h.jwt.Tokens(publicUser.ID, requestID); err != nil {
		h.log.Warn("tokens", lax.Error(err))
		response.Render(res, http.StatusInternalServerError, nil, h.log)

		return
	}

	response.Render(res, http.StatusOK, publicUser, h.log)
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
	email, err := public.EmailFromJSON(req.Body, h.log)
	if err != nil {
		response.RenderError(res, err, h.log)

		return
	}

	user, err := h.usersRepo.FetchPasswordResetToken(req.Context(), email.Email)
	if err != nil {
		if errors.Is(err, repository.ErrResourceNotFound) {
			response.RenderErrorStatus(res, http.StatusNotFound, err.Error(), h.log)

			return
		}

		h.log.Warn("password reset token", lax.Error(err))
		response.Render(res, http.StatusInternalServerError, nil, h.log)

		return
	}

	message := fmt.Sprintf("%s/%s/%s", h.externalURL, h.frontendPasswordResetPath, user.PasswordResetToken)

	if err = h.sender.Send(user.Email, "Password reset request", message); err != nil {
		h.log.Warn("send password reset token", lax.Error(err))

		response.Render(res, http.StatusInternalServerError, nil, h.log)

		return
	}

	response.Render(res, http.StatusAccepted, nil, h.log)
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
	password, err := public.PasswordFromJSON(req.Body, h.log)
	if err != nil {
		response.RenderError(res, err, h.log)

		return
	}

	response.Render(res, http.StatusOK, password.Strength(), h.log)
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
	resetPassword, err := public.ResetPasswordFromJSON(req.Body, h.log)
	if err != nil {
		response.RenderError(res, err, h.log)

		return
	}

	user, err := h.usersRepo.ResetPassword(req.Context(), resetPassword.PasswordResetToken, resetPassword.HashedPassword)
	if err != nil {
		if errors.Is(err, repository.ErrResourceNotFound) {
			response.RenderErrorStatus(res, http.StatusNotFound, err.Error(), h.log)

			return
		}

		h.log.Warn("reset password", lax.Error(err))
		response.Render(res, http.StatusInternalServerError, nil, h.log)

		return
	}

	publicUser := public.FromDomainUser(user)
	publicUser.ID = ""

	response.Render(res, http.StatusAccepted, publicUser, h.log)
}

func (h *UsersHandler) RefreshToken(res http.ResponseWriter, req *http.Request) {
	refreshToken, err := request.RefreshTokenFromBody(req.Body, h.log)
	if err != nil {
		response.RenderError(res, err, h.log)

		return
	}

	_ = refreshToken
}
