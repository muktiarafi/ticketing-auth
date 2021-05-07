package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/muktiarafi/ticketing-auth/internal/entity"
	"github.com/muktiarafi/ticketing-auth/internal/model"
	"github.com/muktiarafi/ticketing-auth/internal/service"
	common "github.com/muktiarafi/ticketing-common"
)

type AuthHandler struct {
	service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		AuthService: authService,
	}
}

func (h *AuthHandler) Route(r *echo.Echo) {
	auth := r.Group("/auth")
	auth.POST("", h.SignIn)
	auth.POST("/signup", h.New)
	auth.GET("", h.CurrentUser, common.RequireAuth)
	auth.POST("/signout", h.SignOut)
}

func (h *AuthHandler) New(c echo.Context) error {
	userDTO := new(model.UserDTO)
	if err := c.Bind(userDTO); err != nil {
		return &common.Error{Op: "AuthHandler.New", Err: err}
	}

	if err := c.Validate(userDTO); err != nil {
		return err
	}

	newUser, err := h.SignUp(userDTO)
	if err != nil {
		return err
	}

	token, err := common.CreateToken(&common.UserPayload{newUser.ID, newUser.Email})
	if err != nil {
		return err
	}

	cookie := new(http.Cookie)
	cookie.Name = "session"
	cookie.Value = token
	cookie.Expires = time.Now().Add(336 * time.Hour)
	c.SetCookie(cookie)

	return common.NewResponse(http.StatusCreated, "Created", newUser).SendJSON(c)
}

func (h *AuthHandler) SignIn(c echo.Context) error {
	userDTO := new(model.UserDTO)
	if err := c.Bind(userDTO); err != nil {
		return &common.Error{Op: "AuthHandler.SignIn", Err: err}
	}

	if err := c.Validate(userDTO); err != nil {
		return err
	}

	user, err := h.Authenticate(userDTO)
	if err != nil {
		return err
	}

	token, err := common.CreateToken(&common.UserPayload{user.ID, user.Email})
	if err != nil {
		return err
	}

	cookie := new(http.Cookie)
	cookie.Name = "session"
	cookie.Value = token
	cookie.Expires = time.Now().Add(336 * time.Hour)
	c.SetCookie(cookie)

	return common.NewResponse(http.StatusOK, "OK", user).SendJSON(c)
}

func (h *AuthHandler) CurrentUser(c echo.Context) error {
	userPayload, ok := c.Get("userPayload").(*common.UserPayload)
	if !ok {
		return &common.Error{
			Op:  "AuthHandler.CurrentUser",
			Err: errors.New("missing payload in context"),
		}
	}

	return common.NewResponse(http.StatusOK, "OK", &entity.User{ID: userPayload.ID, Email: userPayload.Email}).SendJSON(c)
}

func (h *AuthHandler) SignOut(c echo.Context) error {
	cookie, err := c.Cookie("session")
	if err != nil {
		return &common.Error{
			Op:  "Authandler.SignOut",
			Err: errors.New("missing cookie"),
		}
	}

	cookie.Expires = time.Now().Add(-(336 * time.Hour))
	c.SetCookie(cookie)
	return common.NewResponse(http.StatusOK, "OK", "Logged Out").SendJSON(c)
}
