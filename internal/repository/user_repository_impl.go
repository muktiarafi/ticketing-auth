package repository

import (
	"database/sql"

	"github.com/muktiarafi/ticketing-auth/internal/driver"
	"github.com/muktiarafi/ticketing-auth/internal/entity"
	"github.com/muktiarafi/ticketing-auth/internal/model"
	common "github.com/muktiarafi/ticketing-common"
)

type UserRepositoryImpl struct {
	*driver.DB
}

func NewUserRepository(db *driver.DB) UserRepository {
	return &UserRepositoryImpl{
		DB: db,
	}
}

func (r *UserRepositoryImpl) FindOne(email string) (*entity.User, error) {
	ctx, cancel := newDBContext()
	defer cancel()

	stmt := `SELECT id, email, password
	FROM users
	WHERE email = $1`

	user := new(entity.User)
	err := r.SQL.QueryRowContext(ctx, stmt, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
	)

	const op = "UserRepositoryImpl.FindOne"
	if err != nil {
		if err == sql.ErrNoRows {
			return &entity.User{}, &common.Error{Op: op, Code: common.ENOTFOUND, Err: err}
		}
		return &entity.User{}, &common.Error{Op: op, Err: err}
	}

	return user, nil
}

func (r *UserRepositoryImpl) Create(userDTO *model.UserDTO) (*entity.User, error) {
	ctx, cancel := newDBContext()
	defer cancel()

	stmt := `INSERT INTO users (email, password)
	VALUES ($1, $2)
	RETURNING *`

	user := new(entity.User)
	err := r.SQL.QueryRowContext(ctx, stmt, userDTO.Email, userDTO.Password).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
	)
	if err != nil {
		return nil, &common.Error{Op: "UserRepositoryImpl.Create", Err: err}
	}

	return user, nil
}
