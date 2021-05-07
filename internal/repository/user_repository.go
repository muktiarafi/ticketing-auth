package repository

import (
	"github.com/muktiarafi/ticketing-auth/internal/entity"
	"github.com/muktiarafi/ticketing-auth/internal/model"
)

type UserRepository interface {
	FindOne(email string) (*entity.User, error)
	Create(*model.UserDTO) (*entity.User, error)
}
