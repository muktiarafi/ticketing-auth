package service

import (
	"github.com/muktiarafi/ticketing-auth/internal/entity"
	"github.com/muktiarafi/ticketing-auth/internal/model"
)

type AuthService interface {
	Authenticate(userDTO *model.UserDTO) (*entity.User, error)
	SignUp(userDTO *model.UserDTO) (*entity.User, error)
}
