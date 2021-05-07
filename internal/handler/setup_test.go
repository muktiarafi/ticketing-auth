package handler

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-playground/validator/v10"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/muktiarafi/ticketing-auth/internal/driver"
	"github.com/muktiarafi/ticketing-auth/internal/repository"
	"github.com/muktiarafi/ticketing-auth/internal/service"
	common "github.com/muktiarafi/ticketing-common"
	"github.com/ory/dockertest/v3"
)

var (
	pool     *dockertest.Pool
	resource *dockertest.Resource
)

var router *echo.Echo

func TestMain(m *testing.M) {
	db := driver.DB{
		SQL: newTestDatabase(),
	}

	router = echo.New()
	router.Use(middleware.Logger())

	val := validator.New()
	trans := common.NewDefaultTranslator(val)
	customValidator := &common.CustomValidator{val, trans}
	router.Validator = customValidator
	router.HTTPErrorHandler = common.CustomErrorHandler

	userRepository := repository.NewUserRepository(&db)
	authService := service.NewAuthService(userRepository)
	authHandler := NewAuthHandler(authService)
	authHandler.Route(router)

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func newTestDatabase() *sql.DB {
	var err error
	pool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err = pool.Run("postgres", "alpine", []string{"POSTGRES_PASSWORD=secret", "POSTGRES_DB=postgres"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	var db *sql.DB
	if err = pool.Retry(func() error {
		db, err = sql.Open(
			"pgx",
			fmt.Sprintf("host=localhost port=%s dbname=postgres user=postgres password=secret", resource.GetPort("5432/tcp")))
		if err != nil {
			return err
		}

		migrationFilePath := filepath.Join("..", "..", "db", "migrations")
		return driver.Migration(migrationFilePath, db)
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	return db
}
