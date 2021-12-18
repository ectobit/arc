package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nbutton23/zxcvbn-go"
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
// @Param user body request.UserRegistration true "User"
// @Success 201 {object} response.User
// @Failure 400 {object} response.Error
// @Failure 409 {object} response.Error
// @Failure 500
// @Summary Register user account.
func (h *UsersHandler) Register(res http.ResponseWriter, req *http.Request) {
	userRegistration, err := request.UserRegistrationFromJSON(req.Body, h.log)
	if err != nil {
		response.RenderError(res, err, h.log)

		return
	}

	domainUser, err := h.usersRepo.Create(req.Context(), userRegistration.Email, userRegistration.HashedPassword)
	if err != nil {
		if errors.Is(err, repository.ErrUniqueViolation) {
			response.RenderErrorStatus(res, http.StatusConflict, "already registered", h.log)

			return
		}

		h.log.Warn("create user", lax.Error(err))
		response.Render(res, http.StatusInternalServerError, nil, h.log)

		return
	}

	message := fmt.Sprintf("%s/users/activate/%s", h.externalURL, domainUser.ActivationToken)

	if err = h.sender.Send(domainUser.Email, "Account activation", message); err != nil {
		h.log.Warn("send activation link", lax.Error(err))

		response.Render(res, http.StatusInternalServerError, nil, h.log)

		return
	}

	user := response.FromDomainUser(domainUser)
	user.ID = ""

	response.Render(res, http.StatusCreated, user, h.log)
}

// Activate activates user account.
//
// @Tags users
// @Accept json
// @Produce json
// @Router /users/activate/{token} [get]
// @Param token path string true "Activation token"
// @Success 200 {object} response.User
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500
// @Summary Activate user account.
func (h *UsersHandler) Activate(res http.ResponseWriter, req *http.Request) {
	token := chi.URLParam(req, "token")
	if token == "" {
		response.Render(res, http.StatusBadRequest, nil, h.log)

		return
	}

	domainUser, err := h.usersRepo.Activate(req.Context(), token)
	if err != nil {
		if errors.Is(err, repository.ErrResourceNotFound) {
			response.RenderErrorStatus(res, http.StatusNotFound, err.Error(), h.log)

			return
		}

		h.log.Warn("activate user account", lax.Error(err))
		response.Render(res, http.StatusInternalServerError, nil, h.log)

		return
	}

	user := response.FromDomainUser(domainUser)
	user.ID = ""

	response.Render(res, http.StatusOK, user, h.log)
}

// Login logins user.
//
// @Tags users
// @Accept json
// @Produce json
// @Router /users/login [post]
// @Param user body request.UserLogin true "User"
// @Success 201 {object} response.User
// @Failure 400 {object} response.Error
// @Failure 401 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500
// @Summary Login.
func (h *UsersHandler) Login(res http.ResponseWriter, req *http.Request) {
	userLogin, err := request.UserLoginFromJSON(req.Body, h.log)
	if err != nil {
		response.RenderError(res, err, h.log)

		return
	}

	domainUser, err := h.usersRepo.FindOneByEmail(req.Context(), userLogin.Email)
	if err != nil {
		if errors.Is(err, repository.ErrResourceNotFound) {
			response.RenderErrorStatus(res, http.StatusNotFound, err.Error(), h.log)

			return
		}

		h.log.Warn("find user by email", lax.Error(err))
		response.Render(res, http.StatusInternalServerError, nil, h.log)

		return
	}

	if !domainUser.IsValidPassword(userLogin.Password) {
		response.Render(res, http.StatusUnauthorized, nil, h.log)

		return
	}

	if !domainUser.IsActive() {
		response.RenderErrorStatus(res, http.StatusUnauthorized, "account not activated", h.log)

		return
	}

	user := response.FromDomainUser(domainUser)
	requestID := middleware.GetReqID(req.Context())

	if user.AuthToken, user.RefreshToken, err = h.jwt.Tokens(user.ID, requestID); err != nil {
		h.log.Warn("tokens", lax.Error(err))
		response.Render(res, http.StatusInternalServerError, nil, h.log)

		return
	}

	response.Render(res, http.StatusOK, user, h.log)
}

// RequestPasswordReset requests password reset.
//
// @Tags users
// @Accept json
// @Produce json
// @Router /users/reset-password [post]
// @Param email body request.Email true "E-mail address"
// @Success 202
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500
// @Summary Request password reset.
func (h *UsersHandler) RequestPasswordReset(res http.ResponseWriter, req *http.Request) {
	email, err := request.EmailFromJSON(req.Body, h.log)
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
// @Param password body request.Password true "Password"
// @Success 200
// @Failure 400 {object} response.Error
// @Summary Calculate password strength.
func (h *UsersHandler) CheckPasswordStrength(res http.ResponseWriter, req *http.Request) {
	password, err := request.PasswordFromJSON(req.Body, h.log)
	if err != nil {
		response.RenderError(res, err, h.log)

		return
	}

	response.Render(res, http.StatusOK, &response.PasswordStrength{
		Strength: uint8(zxcvbn.PasswordStrength(password.Password, nil).Score),
	}, h.log)
}

// ResetPassword sets new user's password.
//
// @Tags users
// @Accept json
// @Produce json
// @Router /users/reset-password [patch]
// @Param resetPassword body request.ResetPassword true "Password reset token and new password"
// @Success 200 {object} response.User
// @Failure 400 {object} response.Error
// @Failure 404 {object} response.Error
// @Failure 500
// @Summary Set new user's password.
func (h *UsersHandler) ResetPassword(res http.ResponseWriter, req *http.Request) {
	resetPassword, err := request.ResetPasswordFromJSON(req.Body, h.log)
	if err != nil {
		response.RenderError(res, err, h.log)

		return
	}

	domainUser, err := h.usersRepo.ResetPassword(req.Context(), resetPassword.PasswordResetToken, resetPassword.HashedPassword)
	if err != nil {
		if errors.Is(err, repository.ErrResourceNotFound) {
			response.RenderErrorStatus(res, http.StatusNotFound, err.Error(), h.log)

			return
		}

		h.log.Warn("reset password", lax.Error(err))
		response.Render(res, http.StatusInternalServerError, nil, h.log)

		return
	}

	user := response.FromDomainUser(domainUser)
	user.ID = ""

	response.Render(res, http.StatusAccepted, user, h.log)
}

func (h *UsersHandler) RefreshToken(res http.ResponseWriter, req *http.Request) {
	refreshToken, err := request.RefreshTokenFromBody(req.Body, h.log)
	if err != nil {
		response.RenderError(res, err, h.log)

		return
	}

	_ = refreshToken
}
