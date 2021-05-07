package service

import (
	"database/sql"
	"errors"

	"github.com/muktiarafi/ticketing-auth/internal/entity"
	"github.com/muktiarafi/ticketing-auth/internal/model"
	"github.com/muktiarafi/ticketing-auth/internal/repository"
	common "github.com/muktiarafi/ticketing-common"
	"golang.org/x/crypto/bcrypt"
)

type AuthServiceImpl struct {
	repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &AuthServiceImpl{
		UserRepository: userRepo,
	}
}

func (s *AuthServiceImpl) Authenticate(userDTO *model.UserDTO) (*entity.User, error) {
	user, err := s.FindOne(userDTO.Email)
	er, ok := err.(*common.Error)
	const op = "AuthService.Authenticate"
	if ok {
		if er.Err == sql.ErrNoRows {
			return nil, &common.Error{
				Op:      op,
				Code:    common.EINVALID,
				Message: "Invalid Email or Password",
				Err:     err,
			}
		}

		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userDTO.Password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, &common.Error{
				Op:      op,
				Code:    common.EINVALID,
				Message: "Invalid Email or Password",
				Err:     err,
			}
		}
		return nil, &common.Error{
			Op:  op,
			Err: err,
		}
	}

	return user, nil
}

func (s *AuthServiceImpl) SignUp(userDTO *model.UserDTO) (*entity.User, error) {
	user, err := s.FindOne(userDTO.Email)
	er, ok := err.(*common.Error)
	if ok {
		if er.Err != sql.ErrNoRows {
			return nil, err
		}
	}

	const op = "AuthService.SignUp"
	if len(user.Email) != 0 {
		return nil, &common.Error{
			Op:      op,
			Code:    common.ECONCLICT,
			Message: "Email already taken",
			Err:     errors.New("email already taken"),
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(userDTO.Password), 12)
	if err != nil {
		return nil, &common.Error{
			Op:  op,
			Err: err,
		}
	}

	userDTO.Password = string(hash)

	return s.Create(userDTO)
}
