package server

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/muktiarafi/ticketing-auth/internal/config"
	"github.com/muktiarafi/ticketing-auth/internal/driver"
	"github.com/muktiarafi/ticketing-auth/internal/handler"
	custommiddleware "github.com/muktiarafi/ticketing-auth/internal/middleware"
	"github.com/muktiarafi/ticketing-auth/internal/repository"
	"github.com/muktiarafi/ticketing-auth/internal/service"
	common "github.com/muktiarafi/ticketing-common"
)

func SetupServer() *echo.Echo {
	e := echo.New()
	p := custommiddleware.NewPrometheus("echo", nil)
	p.Use(e)

	val := validator.New()
	trans := common.NewDefaultTranslator(val)
	customValidator := &common.CustomValidator{val, trans}
	e.Validator = customValidator
	e.HTTPErrorHandler = common.CustomErrorHandler
	e.Use(middleware.Logger())

	db, err := driver.ConnectSQL(config.PostgresDSN())
	if err != nil {
		panic(err)
	}
	userRepository := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepository)
	authHandler := handler.NewAuthHandler(authService)
	authHandler.Route(e)

	return e
}
