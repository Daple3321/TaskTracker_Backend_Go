package services

import (
	"context"
	"errors"

	"github.com/Daple3321/TaskTracker/internal/entity"
	"github.com/Daple3321/TaskTracker/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type UserService struct {
	storage *repositories.UserRepository
}

func NewUserService(storage *repositories.UserRepository) *UserService {

	us := UserService{
		storage: storage,
	}

	return &us
}

func (u *UserService) Register(ctx context.Context, user entity.UserDTO) (int, error) {
	if user.Username == "" || user.Password == "" {
		return 0, ErrInvalidCredentials
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	userId, err := u.storage.Create(ctx, user.Username, string(passwordHash))
	if err != nil {
		if errors.Is(err, repositories.ErrDuplicateUsername) {
			return 0, repositories.ErrDuplicateUsername
		}
		return 0, err
	}

	return userId, nil
}

func (us *UserService) Login(ctx context.Context, userDTO entity.UserDTO) (userId int, username string, err error) {
	if userDTO.Username == "" || userDTO.Password == "" {
		return 0, "", ErrInvalidCredentials
	}

	user, err := us.storage.GetByUsername(ctx, userDTO.Username)
	if err != nil {
		return 0, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(userDTO.Password))
	if err != nil {
		return 0, "", ErrInvalidCredentials
	}

	return user.Id, user.Username, nil
}
